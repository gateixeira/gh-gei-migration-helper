/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gateixeira/gei-migration-helper/cmd/github"
	"github.com/spf13/cobra"
)

// migrateRepoCmd represents the migrateRepo command
var migrateSecretScanningCmd = &cobra.Command{
	Use:   "migrate-secret-scanning",
	Short: "Migrate secret scanning remediations for a repository",
	Run: func(cmd *cobra.Command, args []string) {
		sourceOrg, _ := cmd.Flags().GetString(sourceOrgFlagName)
		targetOrg, _ := cmd.Flags().GetString(targetOrgFlagName)
		sourceToken, _ := cmd.Flags().GetString(sourceTokenFlagName)
		targetToken, _ := cmd.Flags().GetString(targetTokenFlagName)
		repository, _ := cmd.Flags().GetString(repositoryFlagName)

		slog.Info(fmt.Sprintf("migrating secret scanning for repository %s from %s to %s", repository, sourceOrg, targetOrg))

		if repository == "" {
			slog.Info("fetching repositories from source organization")
			repositories, err := github.GetRepositories(sourceOrg, sourceToken)

			if err != nil {
				slog.Error("error fetching repositories from source organization")
				os.Exit(1)
			}

			slog.Info("done")

			for _, repository := range repositories {
				if *repository.Name == ".github" {
					continue
				}
				err := CheckAndMigrateSecretScanning(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)

				if err != nil {
					slog.Error("error migrating secret scanning for repository: " + *repository.Name)
					continue
				}
			}
		} else {
			err := CheckAndMigrateSecretScanning(repository, sourceOrg, targetOrg, sourceToken, targetToken)

			if err != nil {
				slog.Error("error migrating secret scanning for repository: " + repository)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(migrateSecretScanningCmd)

	migrateSecretScanningCmd.Flags().String(repositoryFlagName, "", "The repository to migrate. If not provided, Secret Scanning will be migrated for all repositories in the organization.")
}
