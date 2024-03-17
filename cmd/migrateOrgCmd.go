package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gateixeira/gei-migration-helper/internal/migration"
	"github.com/spf13/cobra"
)

var migrateOrgCmd = &cobra.Command{
	Use:   "migrate-organization",
	Short: "Migrate all repositories from one organization to another",
	Long: `This script migrates all repositories from one organization to another.

	The target organization has to exist at destination.

	This script will not migrate the .github repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		initial := time.Now()

		sourceOrg, _ := cmd.Flags().GetString(sourceOrgFlagName)
		targetOrg, _ := cmd.Flags().GetString(targetOrgFlagName)
		sourceToken, _ := cmd.Flags().GetString(sourceTokenFlagName)
		targetToken, _ := cmd.Flags().GetString(targetTokenFlagName)
		maxRetries, _ := cmd.Flags().GetInt(maxRetriesFlagName)
		workers, _ := cmd.Flags().GetInt(workersFlagName)

		slog.Info("migrating", "source", sourceOrg, "destination", targetOrg)

		migration, err := migration.NewOrgMigration(
			context.Background(), sourceOrg, targetOrg, sourceToken, targetToken, maxRetries, workers)

		if err != nil {
			slog.Error("error creating migration", err)
			os.Exit(1)
		}

		migrationResult, err := migration.Migrate()

		if err != nil {
			slog.Error("error migrating", err)
			os.Exit(1)
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

		slog.Info("migration result saved to migration-result.json")

		slog.Info(fmt.Sprintf("migration took %s", time.Since(initial)))
	},
}

func init() {
	rootCmd.AddCommand(migrateOrgCmd)
}
