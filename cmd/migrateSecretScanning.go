/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
	"log"
	"os/exec"

	"github.com/spf13/cobra"
)

// migrateRepoCmd represents the migrateRepo command
var migrateSecretScanningCmd = &cobra.Command{
	Use:   "migrate-secret-scanning",
	Short: "Migrate secret scanning remediations for a repository",
	Run: func(cmd *cobra.Command, args []string) {
		sourceOrg, _ := cmd.Flags().GetString(sourceOrgFlagName)
		targetOrg, _ := cmd.Flags().GetString(targetOrgFlagName)
		repository, _ := cmd.Flags().GetString(repositoryFlagName)
		sourceToken, _ := cmd.Flags().GetString(sourceTokenFlagName)
		targetToken, _ := cmd.Flags().GetString(targetTokenFlagName)
	
		migrateSecretScanning(repository, sourceOrg, targetOrg, sourceToken, targetToken)
	},
}

func init() {
	rootCmd.AddCommand(migrateSecretScanningCmd)

	migrateSecretScanningCmd.Flags().String(sourceOrgFlagName, "", "The source organization.")
	migrateSecretScanningCmd.MarkFlagRequired(sourceOrgFlagName)

	migrateSecretScanningCmd.Flags().String(targetOrgFlagName, "", "The target organization.")
	migrateSecretScanningCmd.MarkFlagRequired(targetOrgFlagName)

	migrateSecretScanningCmd.Flags().String(repositoryFlagName, "", "The repository to migrate.")
	migrateSecretScanningCmd.MarkFlagRequired(repositoryFlagName)

	migrateSecretScanningCmd.Flags().String(sourceTokenFlagName, "", "The token of the source organization.")
	migrateSecretScanningCmd.MarkFlagRequired(sourceTokenFlagName)

	migrateSecretScanningCmd.Flags().String(targetTokenFlagName, "", "The token of the target organization.")
	migrateSecretScanningCmd.MarkFlagRequired(targetTokenFlagName)
}

func migrateSecretScanning(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) {
	cmd := exec.Command("gh", "gei", "migrate-secret-alerts", "--source-repo", repository, "--source-org", sourceOrg, "--target-org", targetOrg, "--github-source-pat", sourceToken, "--github-target-pat", targetToken)

	err := cmd.Run()

	if err != nil {
		log.Fatalf("failed to migrate secret scanning remediations %s: %v", repository, err)
	}
}