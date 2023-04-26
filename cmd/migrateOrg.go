/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
	"log"
	"os"

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

		log.Println("\n[🔄] Deactivating GHAS settings at target organization")
		github.ChangeGHASOrgSettings(targetOrg, false, targetToken, "")
		log.Println("[✅] Done")

		log.Println("[🔄] Fetching repositories from source organization")
		repositories, err := github.GetRepositories(sourceOrg, sourceToken)
		if err != nil {
			log.Println("[❌] Error fetching repositories from source organization")
			os.Exit(1)
		}

		log.Println("[✅] Done")

		for _, repository := range repositories {
			if *repository.Name == ".github" {
				continue
			}

			error := ProcessRepoMigration(repository, sourceOrg, targetOrg, sourceToken, targetToken)
			if error != nil {
				log.Println("[❌] Error migrating repository " + *repository.Name)
				continue
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(migrateOrgCmd)
}
