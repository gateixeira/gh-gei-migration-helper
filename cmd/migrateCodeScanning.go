/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
	"fmt"

	"github.com/gateixeira/gei-migration-helper/cmd/github"
	"github.com/spf13/cobra"
)

// migrateRepoCmd represents the migrateRepo command
var migrateCodeScanningCmd = &cobra.Command{
	Use:   "migrate-code-scanning",
	Short: "Migrate code scanning alerts for a repository",
	Run: func(cmd *cobra.Command, args []string) {
		sourceOrg, _ := cmd.Flags().GetString(sourceOrgFlagName)
		targetOrg, _ := cmd.Flags().GetString(targetOrgFlagName)
		sourceToken, _ := cmd.Flags().GetString(sourceTokenFlagName)
		targetToken, _ := cmd.Flags().GetString(targetTokenFlagName)
		repository, _ := cmd.Flags().GetString(repositoryFlagName)

		if repository == "" {
			fmt.Println("\n[🔄] Fetching repositories from source organization")
			repositories := github.GetRepositories(sourceOrg, sourceToken)
			fmt.Println("[✅] Done")

			for _, repository := range repositories {
				if *repository.Name == ".github" {
					continue
				}

				CheckAndMigrateCodeScanning(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)
			}
		} else {
			CheckAndMigrateCodeScanning(repository, sourceOrg, targetOrg, sourceToken, targetToken)
		}
	},
}

func init() {
	rootCmd.AddCommand(migrateCodeScanningCmd)

	migrateCodeScanningCmd.Flags().String(repositoryFlagName, "", "The repository to migrate. If not provided, Code Scanning will be migrated for all repositories in the organization.")
}
