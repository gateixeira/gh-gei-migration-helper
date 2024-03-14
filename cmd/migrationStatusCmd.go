/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/gateixeira/gei-migration-helper/internal/github"
	"github.com/spf13/cobra"
)

var migrationStatusCmd = &cobra.Command{
	Use:   "migration-status",
	Short: "Check status for organization migration",
	Run: func(cmd *cobra.Command, args []string) {
		sourceOrg, _ := cmd.Flags().GetString(sourceOrgFlagName)
		targetOrg, _ := cmd.Flags().GetString(targetOrgFlagName)
		sourceToken, _ := cmd.Flags().GetString(sourceTokenFlagName)
		targetToken, _ := cmd.Flags().GetString(targetTokenFlagName)

		slog.Info(fmt.Sprintf("checking migration status from %s to %s", sourceOrg, targetOrg))

		statusRepoName := "migration-status"

		migrationRepository, _ := github.GetRepository(statusRepoName, targetOrg, targetToken)

		if migrationRepository == nil {
			slog.Info("there is no migration-status repository in the target organization, likely the migration has not been started")
			os.Exit(0)
		}

		migrationIssue, _ := github.GetIssue(targetOrg, statusRepoName, 1, targetToken)

		if migrationIssue != nil {
			slog.Info(fmt.Sprintf("migration finished. Check https://github.com/%s/%s/issues/1 for details", targetOrg, statusRepoName))
			os.Exit(0)
		}

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

		sourceRepos := make(map[string]bool)
		for _, repo := range sourceRepositories {
			sourceRepos[*repo.Name] = true
		}

		intersection := make(map[string]bool)
		for _, repo := range destinationRepositories {
			if sourceRepos[*repo.Name] {
				intersection[*repo.Name] = true
			}
		}

		slog.Info("=========================================================")
		slog.Info("a migration is ongoing from %s to %s (or finished in error)", sourceOrg, targetOrg)
		slog.Info("=========================================================")
		slog.Info(fmt.Sprintf("%d/%d repositories are migrated", len(intersection), len(sourceRepositories)))
		slog.Info("=========================================================")
		slog.Info("migrated repositories:")
		var migrated []string
		for repo := range intersection {
			migrated = append(migrated, repo)
		}
		slog.Info(strings.Join(migrated, ", "))
		slog.Info("=========================================================")
		slog.Info("repositories to be migrated:")

		var toMigrate []string
		for _, repo := range sourceRepositories {
			if !intersection[*repo.Name] {
				toMigrate = append(toMigrate, *repo.Name)
			}
		}

		if len(toMigrate) == 0 {
			slog.Info("no repositories to migrate. Finishing the last repository migration")
			os.Exit(0)
		}

		slog.Info(strings.Join(toMigrate, ", "))
	},
}

func init() {
	rootCmd.AddCommand(migrationStatusCmd)
}
