/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
	"fmt"

	"github.com/gateixeira/gei-migration-helper/cmd/github"
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
		sourceOrg, _ := cmd.Flags().GetString(sourceOrgFlagName)
		targetOrg, _ := cmd.Flags().GetString(targetOrgFlagName)
		sourceToken, _ := cmd.Flags().GetString(sourceTokenFlagName)
		targetToken, _ := cmd.Flags().GetString(targetTokenFlagName)

		fmt.Println("[🔄] Deactivating GHAS settings at target organization")
		github.ChangeGHASOrgSettings(targetOrg, false, targetToken)
		fmt.Println("[✅] Done")

		fmt.Println("[🔄] Fetching repositories from source organization")
		repositories := github.GetRepositories(sourceOrg, sourceToken)
		fmt.Println("[✅] Done")

		for _, repository := range repositories {
			if *repository.Name == ".github" {
				continue
			}

			fmt.Print(
				"\n\n========================================\nRepository " + *repository.Name + "\n========================================")

			if *repository.Visibility != "public" {
				fmt.Println("[🔄] Deactivating GHAS settings at source repository")
				github.ChangeGhasRepoSettings(sourceOrg, *repository.Name, false, sourceToken)
				fmt.Println("[✅] Done")
			}

			fmt.Println("[🔄] Migrating repository")
			github.MigrateRepo(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)
			fmt.Println("[✅] Done")

			fmt.Println("[🔄] Deleting branch protections at target")
			github.DeleteBranchProtections(targetOrg, *repository.Name, targetToken)
			fmt.Println("[✅] Done")

			//check if repository is not private
			if !*repository.Private {
				fmt.Println("[🔄] Repository not private at source. Changing visibility to internal at target")
				github.ChangeRepositoryVisibility(targetOrg, *repository.Name, "internal", targetToken)
				fmt.Println("[✅] Done")
			}

			fmt.Println("[🔄] Activating GHAS settings at target")
			github.ChangeGhasRepoSettings(targetOrg, *repository.Name, true, targetToken)
			fmt.Println("[✅] Finished. Next...")
		}
	},
}

func init() {
	rootCmd.AddCommand(migrateOrgCmd)

	migrateOrgCmd.Flags().String(sourceOrgFlagName, "", "The source organization.")
	migrateOrgCmd.MarkFlagRequired(sourceOrgFlagName)

	migrateOrgCmd.Flags().String(targetOrgFlagName, "", "The target organization.")
	migrateOrgCmd.MarkFlagRequired(targetOrgFlagName)

	migrateOrgCmd.Flags().String(sourceTokenFlagName, "", "The token of the source organization.")
	migrateOrgCmd.MarkFlagRequired(sourceTokenFlagName)

	migrateOrgCmd.Flags().String(targetTokenFlagName, "", "The token of the target organization.")
	migrateOrgCmd.MarkFlagRequired(targetTokenFlagName)

}
