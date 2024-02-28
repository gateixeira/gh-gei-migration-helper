/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
	"log"
	"os"
	"strings"

	"github.com/gateixeira/gei-migration-helper/cmd/github"
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

		log.Printf("Checking migration status from %s to %s", sourceOrg, targetOrg)

		statusRepoName := "migration-status"

		migrationRepository, _ := github.GetRepository(statusRepoName, targetOrg, targetToken)

		if migrationRepository == nil {
			log.Println("[❌] There is no migration-status repository in the target organization, likely the migration has not been started")
			os.Exit(0)
		}

		migrationIssue, _ := github.GetIssue(targetOrg, statusRepoName, 1, targetToken)

		if migrationIssue != nil {
			log.Printf("[✅] Migration finished. Check https://github.com/%s/%s/issues/1 for details", targetOrg, statusRepoName)
			os.Exit(0)
		}

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

		log.Println("=========================================================")
		log.Printf("[🔄] A migration is ongoing from %s to %s (or finished in error)", sourceOrg, targetOrg)
		log.Println("=========================================================")
		log.Printf("[ℹ️] %d/%d repositories are migrated", len(intersection), len(sourceRepositories))
		log.Println("=========================================================")
		log.Println("[✅] Migrated repositories:")
		var migrated []string
		for repo := range intersection {
			migrated = append(migrated, repo)
		}
		log.Println(strings.Join(migrated, ", "))
		log.Println("=========================================================")
		log.Println("[🔄] Repositories to be migrated:")

		var toMigrate []string
		for _, repo := range sourceRepositories {
			if !intersection[*repo.Name] {
				toMigrate = append(toMigrate, *repo.Name)
			}
		}
		log.Println(strings.Join(toMigrate, ", "))
	},
}

func init() {
	rootCmd.AddCommand(migrationStatusCmd)
}
