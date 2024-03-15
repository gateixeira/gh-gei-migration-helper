/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gateixeira/gei-migration-helper/internal/migration"
	"github.com/spf13/cobra"
)

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
		initial := time.Now()

		sourceOrg, _ := cmd.Flags().GetString(sourceOrgFlagName)
		targetOrg, _ := cmd.Flags().GetString(targetOrgFlagName)
		sourceToken, _ := cmd.Flags().GetString(sourceTokenFlagName)
		targetToken, _ := cmd.Flags().GetString(targetTokenFlagName)
		maxRetries, _ := cmd.Flags().GetInt(maxRetriesFlagName)
		workers, _ := cmd.Flags().GetInt(workersFlagName)

		slog.Info("migrating", "source", sourceOrg, "destination", targetOrg)

		migration := migration.NewOrgMigration(sourceOrg, targetOrg, sourceToken, targetToken, maxRetries, workers)
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
