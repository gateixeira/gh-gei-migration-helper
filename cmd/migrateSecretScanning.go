/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
	"github.com/gateixeira/gei-migration-helper/cmd/github"
	"github.com/spf13/cobra"
)

// migrateRepoCmd represents the migrateRepo command
var migrateSecretScanningCmd = &cobra.Command{
	Use:   "migrate-secret-scanning",
	Short: "Migrate secret scanning remediations for a repository",
	Run: func(cmd *cobra.Command, args []string) {
		sourceOrg, _ := cmd.PersistentFlags().GetString(sourceOrgFlagName)
		targetOrg, _ := cmd.PersistentFlags().GetString(targetOrgFlagName)
		sourceToken, _ := cmd.PersistentFlags().GetString(sourceTokenFlagName)
		targetToken, _ := cmd.PersistentFlags().GetString(targetTokenFlagName)
		repository, _ := cmd.Flags().GetString(repositoryFlagName)
	
		github.MigrateSecretScanning(repository, sourceOrg, targetOrg, sourceToken, targetToken)
	},
}

func init() {
	rootCmd.AddCommand(migrateSecretScanningCmd)

	migrateSecretScanningCmd.Flags().String(repositoryFlagName, "", "The repository to migrate.")
	migrateSecretScanningCmd.MarkFlagRequired(repositoryFlagName)
}