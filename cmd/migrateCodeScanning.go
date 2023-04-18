/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
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
		repository, _ := cmd.Flags().GetString(repositoryFlagName)
		sourceToken, _ := cmd.Flags().GetString(sourceTokenFlagName)
		targetToken, _ := cmd.Flags().GetString(targetTokenFlagName)

		github.MigrateCodeScanning(repository, sourceOrg, targetOrg, sourceToken, targetToken)
	},
}

func init() {
	rootCmd.AddCommand(migrateCodeScanningCmd)

	migrateCodeScanningCmd.Flags().String(sourceOrgFlagName, "", "The source organization.")
	migrateCodeScanningCmd.MarkFlagRequired(sourceOrgFlagName)

	migrateCodeScanningCmd.Flags().String(targetOrgFlagName, "", "The target organization.")
	migrateCodeScanningCmd.MarkFlagRequired(targetOrgFlagName)

	migrateCodeScanningCmd.Flags().String(repositoryFlagName, "", "The repository to migrate.")
	migrateCodeScanningCmd.MarkFlagRequired(repositoryFlagName)

	migrateCodeScanningCmd.Flags().String(sourceTokenFlagName, "", "The token of the source organization.")
	migrateCodeScanningCmd.MarkFlagRequired(sourceTokenFlagName)

	migrateCodeScanningCmd.Flags().String(targetTokenFlagName, "", "The token of the target organization.")
	migrateCodeScanningCmd.MarkFlagRequired(targetTokenFlagName)
}
