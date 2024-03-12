/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/gateixeira/gei-migration-helper/cmd/github"
	"github.com/spf13/cobra"
)

type MigrationResult struct {
	Timestamp time.Time `json:"timestamp"`
	SourceOrg string    `json:"sourceOrg"`
	TargetOrg string    `json:"targetOrg"`
	Migrated  []Repo    `json:"migrated"`
	Failed    []Repo    `json:"failed"`
}

type Repo struct {
	Name           string `json:"name"`
	ID             int64  `json:"id"`
	Archived       bool   `json:"archived"`
	CodeScanning   string `json:"codeScanning" default:"disabled"`
	SecretScanning string `json:"secretScanning" default:"disabled"`
	PushProtection string `json:"pushProtection" default:"disabled"`
}

type WorkerError struct {
	Err  error
	Repo github.Repository
}

// migrateOrgCmd represents the migrateOrg command
var migrateOrgCmd = &cobra.Command{
	Use:   "migrate-organization",
	Short: "Migrate all repositories from one organization to another",
	Long: `This script migrates all repositories from one organization to another.

	The target organization has to exist at destination.

	This script will not migrate the .github repository.

	Migration steps:

	- 1. Deactivate GHAS settings at target organization
	- 2. Fetch all repositories from source organization
	- 3. For repositories that are not public, deactivate GHAS settings at source (public repos have this enabled by default)
	- 4. Migrate repository
	- 5. Delete branch protections at target
	- 6. If repository is not private at source, change visibility to internal at target
	- 7. Activate GHAS settings at target`,

	Run: func(cmd *cobra.Command, args []string) {
		initial := time.Now()

		sourceOrg, _ := cmd.Flags().GetString(sourceOrgFlagName)
		targetOrg, _ := cmd.Flags().GetString(targetOrgFlagName)
		sourceToken, _ := cmd.Flags().GetString(sourceTokenFlagName)
		targetToken, _ := cmd.Flags().GetString(targetTokenFlagName)
		maxRetries, _ := cmd.Flags().GetInt(maxRetriesFlagName)

		slog.Info("migrating repositories", "source", sourceOrg, "destination", targetOrg)

		statusRepoName := "migration-status"

		slog.Info("looking for ongoing/past migration")
		repo, err := github.GetRepository(statusRepoName, targetOrg, targetToken)

		if err != nil && err.Error() != github.ErrRepositoryNotFound.Error() {
			slog.Error("error fetching migration status repository", err)
			os.Exit(1)
		}

		if err == nil && *repo.Name == statusRepoName {
			issue, _ := github.GetIssue(targetOrg, statusRepoName, 1, targetToken)

			if issue != nil {
				slog.Error("a migration to this organization was already executed. Please check http://github.com/%s/%s/issues/1 for status", targetOrg, statusRepoName)
				os.Exit(1)
			}

			slog.Error("a migration to this organization is either ongoing or finished in error (remove http://github.com/%s/%s if you want to retry)", targetOrg, statusRepoName)
			os.Exit(1)
		}

		slog.Info("creating migration status repository")
		github.CreateRepository(targetOrg, statusRepoName, targetToken)

		slog.Info("deactivating GHAS settings at target organization")
		github.ChangeGHASOrgSettings(targetOrg, false, targetToken)

		slog.Info("fetching repositories from source organization")
		sourceRepositories, err := github.GetRepositories(sourceOrg, sourceToken)

		if err != nil {
			slog.Error("error fetching repositories from source organization")
			os.Exit(1)
		}

		destinationRepositories, err := github.GetRepositories(targetOrg, targetToken)

		if err != nil {
			slog.Error("error fetching repositories from target organization")
			os.Exit(1)
		}

		m := make(map[string]bool)
		for _, item := range destinationRepositories {
			m[*item.Name] = true
		}
		var sourceRepositoriesToMigrate []github.Repository
		for _, item := range sourceRepositories {
			if _, ok := m[*item.Name]; !ok {
				sourceRepositoriesToMigrate = append(sourceRepositoriesToMigrate, item)
			} else {
				log.Println("[🤝] Repository " + *item.Name + " already exists at target organization")
			}
		}

		slog.Info(strconv.Itoa(len(sourceRepositoriesToMigrate)) + " repositories to migrate")

		jobs := make(chan github.Repository, len(sourceRepositoriesToMigrate))
		results := make(chan WorkerError, len(sourceRepositoriesToMigrate))

		for w := 1; w <= 5; w++ {
			go func(id int, jobs <-chan github.Repository, results chan<- WorkerError) {
				for repository := range jobs {
					slog.Debug("worker started", "job", strconv.Itoa(id), "repository", *repository.Name)

					var repoSummary = Repo{
						Name:     *repository.Name,
						ID:       *repository.ID,
						Archived: *repository.Archived,
					}

					if repository.SecurityAndAnalysis.AdvancedSecurity != nil {
						repoSummary.CodeScanning = *repository.SecurityAndAnalysis.AdvancedSecurity.Status
						repoSummary.SecretScanning = *repository.SecurityAndAnalysis.SecretScanning.Status
						repoSummary.PushProtection = *repository.SecurityAndAnalysis.SecretScanningPushProtection.Status
					}

					err := ProcessRepoMigration(repository, sourceOrg, targetOrg, sourceToken, targetToken, maxRetries)
					results <- WorkerError{Err: err, Repo: repository}

					if err != nil {
						slog.Info("error migrating repository: ", err)
					}
				}
			}(w, jobs, results)
		}

		for _, repository := range sourceRepositoriesToMigrate {
			jobs <- repository
		}
		close(jobs)

		var failed []Repo
		var migrated []Repo
		for a := 1; a <= len(sourceRepositoriesToMigrate); a++ {
			workerResult := <-results
			if workerResult.Err != nil {
				failed = append(failed, Repo{
					Name: *workerResult.Repo.Name,
					ID:   *workerResult.Repo.ID,
				})
			} else {
				migrated = append(migrated, Repo{
					Name: *workerResult.Repo.Name,
					ID:   *workerResult.Repo.ID,
				})
			}
		}

		migrationResult := MigrationResult{
			Timestamp: time.Now().UTC(),
			SourceOrg: sourceOrg,
			TargetOrg: targetOrg,
			Migrated:  migrated,
			Failed:    failed,
		}

		jsonData, err := json.MarshalIndent(migrationResult, "", "  ")
		if err != nil {
			slog.Error("failed to parse result", err)
			os.Exit(1)
		}

		f, err := os.Create("migration-result.json")
		if err != nil {
			slog.Error("failed to create results file", err)
			os.Exit(1)
		}
		defer f.Close()

		_, err = f.Write(jsonData)
		if err != nil {
			slog.Error("failed to write to results file", err)
			os.Exit(1)
		}

		err = github.CreateIssue(targetOrg, statusRepoName, "Migration result", string(jsonData), targetToken)

		if err != nil {
			slog.Error("error creating issue with migration result. Check migration-result.json for details")
			os.Exit(1)
		}

		slog.Info("migration result saved to migration-result.json and written to: " + fmt.Sprintf("https://github.com/%s/%s/issues/1", targetOrg, statusRepoName))

		slog.Info(fmt.Sprintf("migration took %s", time.Since(initial)))
	},
}

func init() {
	rootCmd.AddCommand(migrateOrgCmd)
}
