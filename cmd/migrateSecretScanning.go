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
		sourceOrg, _ := cmd.PersistentFlags().GetString(sourceOrgFlagName)
		targetOrg, _ := cmd.PersistentFlags().GetString(targetOrgFlagName)
		sourceToken, _ := cmd.PersistentFlags().GetString(sourceTokenFlagName)
		targetToken, _ := cmd.PersistentFlags().GetString(targetTokenFlagName)
		repository, _ := cmd.Flags().GetString(repositoryFlagName)
	
		migrateSecretScanning(repository, sourceOrg, targetOrg, sourceToken, targetToken)
	},
}

func init() {
	rootCmd.AddCommand(migrateSecretScanningCmd)

	migrateSecretScanningCmd.Flags().String(repositoryFlagName, "", "The repository to migrate.")
	migrateSecretScanningCmd.MarkFlagRequired(repositoryFlagName)
}

func migrateSecretScanning(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) {
	cmd := exec.Command("gh", "gei", "migrate-secret-alerts", "--source-repo", repository, "--source-org", sourceOrg, "--target-org", targetOrg, "--github-source-pat", sourceToken, "--github-target-pat", targetToken)

	err := cmd.Run()

	if err != nil {
		log.Fatalf("failed to migrate secret scanning remediations %s: %v", repository, err)
	}
}