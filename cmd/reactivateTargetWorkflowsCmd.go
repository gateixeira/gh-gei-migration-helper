package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gateixeira/gei-migration-helper/internal/migration"
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

		ctx := context.Background()
		migrationData, err := migration.NewMigration(ctx, sourceOrg, targetOrg, sourceToken, targetToken)
		if err != nil {
			slog.Error("error migrating repository: " + repository)
			os.Exit(1)
		}

		err = migrationData.ReactivateTargetWorkflows(ctx, repository)

		if err != nil {
			slog.Error("error migrating repository: " + repository)
			os.Exit(1)
		}

	},
}

func init() {
	rootCmd.AddCommand(reactivateTargetWorkflowsCmd)

	reactivateTargetWorkflowsCmd.Flags().String(repositoryFlagName, "", "The repository to reactivate. If not provided, reactivation will be done for all repositories in the organization.")
}
