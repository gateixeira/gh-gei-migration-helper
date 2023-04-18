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
var migrateRepoCmd = &cobra.Command{
	Use:   "migrate-repository",
	Short: "Migrate a repository",
	Long: `This script migrates a repositories from one organization to another.

	The target organization has to exist at destination.

	Migration steps:

	- 1. Deactivate GHAS settings at target organization
	- 2. Deactivate GHAS settings at source repository
	- 3. Migrate repository
	- 4. Delete branch protections at target
	- 5. change repository visibility to internal at target
	- 6. Activate GHAS settings at target`,
	Run: func(cmd *cobra.Command, args []string) {
		sourceOrg, _ := cmd.Flags().GetString(sourceOrgFlagName)
		targetOrg, _ := cmd.Flags().GetString(targetOrgFlagName)
		repository, _ := cmd.Flags().GetString(repositoryFlagName)
		sourceToken, _ := cmd.Flags().GetString(sourceTokenFlagName)
		targetToken, _ := cmd.Flags().GetString(targetTokenFlagName)

		fmt.Println("Migrating repository " + repository + " from " + sourceOrg + " to " + targetOrg)

		changeGHASOrgSettings(targetOrg, false, targetToken)

		fmt.Print(
			"\n\n========================================\n\n" +
				"Migrating repository " + repository +
			"\n\n========================================\n\n")

		changeGhasRepoSettings(sourceOrg, repository, false, sourceToken)
		migrateRepo(repository, sourceOrg, targetOrg, sourceToken, targetToken)
		deleteBranchProtections(targetOrg, repository, targetToken)
		changeRepositoryVisibility(targetOrg, repository, "internal", targetToken)
		changeGhasRepoSettings(targetOrg, repository, true, targetToken)
	},
}

func init() {
	rootCmd.AddCommand(migrateRepoCmd)

	migrateRepoCmd.Flags().String(sourceOrgFlagName, "", "The source organization.")
	migrateRepoCmd.MarkFlagRequired(sourceOrgFlagName)

	migrateRepoCmd.Flags().String(targetOrgFlagName, "", "The target organization.")
	migrateRepoCmd.MarkFlagRequired(targetOrgFlagName)

	migrateRepoCmd.Flags().String(repositoryFlagName, "", "The repository to migrate.")
	migrateRepoCmd.MarkFlagRequired(repositoryFlagName)

	migrateRepoCmd.Flags().String(sourceTokenFlagName, "", "The token of the source organization.")
	migrateRepoCmd.MarkFlagRequired(sourceTokenFlagName)

	migrateRepoCmd.Flags().String(targetTokenFlagName, "", "The token of the target organization.")
	migrateRepoCmd.MarkFlagRequired(targetTokenFlagName)

}

func migrateRepo(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) {
	cmd := exec.Command("gh", "gei", "migrate-repo", "--source-repo", repository, "--github-source-org", sourceOrg, "--github-target-org", targetOrg,  "--github-source-pat", sourceToken, "--github-target-pat", targetToken)

	err := cmd.Run()

	if err != nil {
		log.Fatalf("failed to migrate repository %s: %v", repository, err)
	}
}