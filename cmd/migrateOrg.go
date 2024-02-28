/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
	"encoding/json"
	"log"
	"os"
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
		sourceOrg, _ := cmd.Flags().GetString(sourceOrgFlagName)
		targetOrg, _ := cmd.Flags().GetString(targetOrgFlagName)
		sourceToken, _ := cmd.Flags().GetString(sourceTokenFlagName)
		targetToken, _ := cmd.Flags().GetString(targetTokenFlagName)
		maxRetries, _ := cmd.Flags().GetInt(maxRetriesFlagName)

		statusRepoName := "migration-status"

		log.Println("[🔄] Looking for ongoing/past migration")
		repo, err := github.GetRepository(statusRepoName, targetOrg, targetToken)

		if err != nil && err.Error() != github.ErrRepositoryNotFound.Error() {
			log.Println("[❌] Error fetching migration status repository", err)
			os.Exit(1)
		}

		if err == nil && *repo.Name == statusRepoName {
			log.Println("[❌] A migration to this organization is either ongoing or has finished. Please check the migration status repository.")
			os.Exit(1)
		}

		log.Println("[🔄] Creating migration status repository")
		github.CreateRepository(targetOrg, statusRepoName, targetToken)

		log.Println("[🔄] Deactivating GHAS settings at target organization")
		github.ChangeGHASOrgSettings(targetOrg, false, targetToken)

		log.Println("[🔄] Fetching repositories from source organization")
		sourceRepositories, err := github.GetRepositories(sourceOrg, sourceToken)

		if err != nil {
			log.Println("[❌] Error fetching repositories from source organization")
			os.Exit(1)
		}

		destinationRepositories, err := github.GetRepositories(targetOrg, targetToken)

		if err != nil {
			log.Println("[❌] Error fetching repositories from target organization")
			os.Exit(1)
		}

		// Remove intersection from source and destination repositories
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

		log.Printf("%d repositories to migrate", len(sourceRepositoriesToMigrate))

		var migratedRepos []Repo = []Repo{}
		var failedRepos []Repo = []Repo{}
		for i, repository := range sourceRepositoriesToMigrate {
			log.Println("========================================")
			log.Printf("[🔄] Migrating repository %d of %d", i+1, len(sourceRepositoriesToMigrate))

			err := ProcessRepoMigration(repository, sourceOrg, targetOrg, sourceToken, targetToken, maxRetries)

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

			if err != nil {
				log.Println("[❌] Error migrating repository: ", err)
				failedRepos = append(failedRepos, repoSummary)
				continue
			}

			migratedRepos = append(migratedRepos, repoSummary)
		}

		migrationResult := MigrationResult{
			Timestamp: time.Now().UTC(),
			SourceOrg: sourceOrg,
			TargetOrg: targetOrg,
			Migrated:  migratedRepos,
			Failed:    failedRepos,
		}

		jsonData, err := json.MarshalIndent(migrationResult, "", "  ")
		if err != nil {
			log.Println(err)
			return
		}

		f, err := os.Create("migration-result.json")
		if err != nil {
			log.Println(err)
			return
		}
		defer f.Close()

		_, err = f.Write(jsonData)
		if err != nil {
			log.Println(err)
			return
		}

		err = github.CreateIssue(targetOrg, statusRepoName, "Migration result", string(jsonData), targetToken)

		if err != nil {
			log.Println("[❌] Error creating issue with migration result. Check migration-result.json for details")
			return
		}

		log.Printf("\nMigration result saved to migration-result.json and written to http://github.com/%s/%s/issues/1", targetOrg, statusRepoName)
	},
}

func init() {
	rootCmd.AddCommand(migrateOrgCmd)
}
