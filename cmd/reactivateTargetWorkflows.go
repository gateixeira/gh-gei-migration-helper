/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
	"fmt"
	"log/slog"

	"github.com/gateixeira/gei-migration-helper/cmd/github"
	"github.com/spf13/cobra"
)

// migrateRepoCmd represents the migrateRepo command
var reactivateTargetWorkflowsCmd = &cobra.Command{
	Use:   "reactivate-target-workflows",
	Short: "Reactivate workflows for a migrated repository based on source",
	Run: func(cmd *cobra.Command, args []string) {
		sourceOrg, _ := cmd.Flags().GetString(sourceOrgFlagName)
		targetOrg, _ := cmd.Flags().GetString(targetOrgFlagName)
		sourceToken, _ := cmd.Flags().GetString(sourceTokenFlagName)
		targetToken, _ := cmd.Flags().GetString(targetTokenFlagName)
		repository, _ := cmd.Flags().GetString(repositoryFlagName)

		slog.Info(fmt.Sprintf("reactivating target workflows for repository %s from %s to %s", repository, sourceOrg, targetOrg))

		if repository == "" {
			slog.Info("fetching repositories from source organization")
			repositories, err := github.GetRepositories(sourceOrg, sourceToken)

			if err != nil {
				slog.Error("error fetching repositories from source organization")
				return
			}

			slog.Info("done")

			for _, repository := range repositories {
				if *repository.Name == ".github" {
					continue
				}

				err := ReactivateTargetWorkflows(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)

				if err != nil {
					slog.Error("error migrating secret scanning for repository: " + *repository.Name)
					continue
				}
			}
		} else {
			ReactivateTargetWorkflows(repository, sourceOrg, targetOrg, sourceToken, targetToken)
		}
	},
}

func init() {
	rootCmd.AddCommand(reactivateTargetWorkflowsCmd)

	reactivateTargetWorkflowsCmd.Flags().String(repositoryFlagName, "", "The repository to reactivate. If not provided, reactivation will be done for all repositories in the organization.")
}
