package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gateixeira/gei-migration-helper/internal/github"
	"github.com/gateixeira/gei-migration-helper/internal/migration"
	"github.com/gateixeira/gei-migration-helper/pkg/logging"
	"github.com/spf13/cobra"
)

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
			ctx := context.Background()
			sourceGC, err := github.NewGitHubClient(ctx, logging.NewLoggerFromContext(ctx, false), sourceToken)
			if err != nil {
				slog.Info("error initializing source GitHub Client", err)
				os.Exit(1)
			}

			repositories, err := sourceGC.GetRepositories(ctx, sourceOrg)

			if err != nil {
				slog.Error("error fetching repositories from source organization")
				return
			}

			slog.Info("done")

			for _, repository := range repositories {
				if *repository.Name == ".github" {
					continue
				}

				err := migration.ReactivateTargetWorkflows(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)

				if err != nil {
					slog.Error("error migrating secret scanning for repository: " + *repository.Name)
					continue
				}
			}
		} else {
			migration.ReactivateTargetWorkflows(repository, sourceOrg, targetOrg, sourceToken, targetToken)
		}
	},
}

func init() {
	rootCmd.AddCommand(reactivateTargetWorkflowsCmd)

	reactivateTargetWorkflowsCmd.Flags().String(repositoryFlagName, "", "The repository to reactivate. If not provided, reactivation will be done for all repositories in the organization.")
}
