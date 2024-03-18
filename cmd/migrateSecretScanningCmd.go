package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gateixeira/gei-migration-helper/internal/migration"
	"github.com/spf13/cobra"
)

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

		ctx := context.Background()
		migration, err := migration.NewSecretScanningMigration(ctx, sourceOrg, targetOrg, sourceToken, targetToken)
		if err != nil {
			slog.Error("error migrating repository: " + repository)
			os.Exit(1)
		}

		err = migration.Migrate(ctx, repository)

		if err != nil {
			slog.Error("error migrating repository: " + repository)
			os.Exit(1)
		}

	},
}

func init() {
	rootCmd.AddCommand(migrateSecretScanningCmd)

	migrateSecretScanningCmd.Flags().String(repositoryFlagName, "", "The repository to migrate. If not provided, Secret Scanning will be migrated for all repositories in the organization.")
}
