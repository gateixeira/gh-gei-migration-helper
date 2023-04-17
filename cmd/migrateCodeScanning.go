/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/spf13/cobra"
)

// migrateRepoCmd represents the migrateRepo command
var migrateCodeScanningCmd = &cobra.Command{
	Use:   "migrateCodeScanning",
	Short: "Migrate code scanning alerts for a repository",
	Run: func(cmd *cobra.Command, args []string) {
		sourceOrg, _ := cmd.Flags().GetString(sourceOrgFlagName)
		targetOrg, _ := cmd.Flags().GetString(targetOrgFlagName)
		repository, _ := cmd.Flags().GetString(repositoryFlagName)
		sourceToken, _ := cmd.Flags().GetString(sourceTokenFlagName)
		targetToken, _ := cmd.Flags().GetString(targetTokenFlagName)

		migrateCodeScanning(repository, sourceOrg, targetOrg, sourceToken, targetToken)
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

func migrateCodeScanning(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) {
	
	fmt.Println("Invoking GEI to migrate code scanning alerts for repository " + repository + " from " + sourceOrg + " to " + targetOrg)

	cmd := exec.Command("gh", "gei", "migrate-code-scanning-alerts", "--source-repo", repository, "--source-org", sourceOrg, "--target-org", targetOrg, "--github-source-pat", sourceToken, "--github-target-pat", targetToken)

	err := cmd.Run()

	if err != nil {
		log.Fatalf("failed to migrate code scanning alerts %s: %v", repository, err)
	}
}