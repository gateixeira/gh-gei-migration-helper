package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gateixeira/gei-migration-helper/internal/github"
	"github.com/gateixeira/gei-migration-helper/internal/migration"
	"github.com/spf13/cobra"
)

const (
	repositoryFlagName = "repo"
)

var migrateRepoCmd = &cobra.Command{
	Use:   "migrate-repository",
	Short: "Migrate a repository",
	Long: `This script migrates a repositories from one organization to another.

	The target organization has to exist at destination.`,
	Run: func(cmd *cobra.Command, args []string) {
		sourceOrg, _ := cmd.Flags().GetString(sourceOrgFlagName)
		targetOrg, _ := cmd.Flags().GetString(targetOrgFlagName)
		sourceToken, _ := cmd.Flags().GetString(sourceTokenFlagName)
		targetToken, _ := cmd.Flags().GetString(targetTokenFlagName)
		repository, _ := cmd.Flags().GetString(repositoryFlagName)
		maxRetries, _ := cmd.Flags().GetInt(maxRetriesFlagName)

		slog.Info(fmt.Sprintf("migrating repository %s from %s to %s", repository, sourceOrg, targetOrg))

		repo, err := github.GetRepository(repository, sourceOrg, sourceToken)

		if err != nil {
			slog.Info("error getting source repository: " + repository)
			os.Exit(1)
		}

		migration := migration.NewRepoMigration(*repo.Name, sourceOrg, targetOrg, sourceToken, targetToken, maxRetries)
		err = migration.Migrate()

		if err != nil {
			slog.Error("error migrating repository: " + repository)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(migrateRepoCmd)

	migrateRepoCmd.Flags().String(repositoryFlagName, "", "The repository to migrate.")
	migrateRepoCmd.MarkFlagRequired(repositoryFlagName)

}
