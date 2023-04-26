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

		log.Println("[🔄] Deactivating GHAS settings at target organization")
		github.ChangeGHASOrgSettings(targetOrg, false, targetToken)
		log.Println("[✅] Done")

		log.Println("[🔄] Fetching repositories from source organization")
		sourceRepositories, err := github.GetRepositories(sourceOrg, sourceToken)

		if err != nil {
			log.Println("[❌] Error fetching repositories from source organization")
			os.Exit(1)
		}

		destinationRepositories, err := github.GetRepositories(targetOrg, targetToken)

		if err != nil {
			log.Println("[❌] Error fetching repositories from target organization")
			os.Exit(1)
		}

		log.Println("[✅] Done")

		// Remove intersection from source and destination repositories
		m := make(map[string]bool)
		for _, item := range destinationRepositories {
			m[*item.Name] = true
		}
		var sourceRepositoriesToMigrate []github.Repository
		for _, item := range sourceRepositories {
			if _, ok := m[*item.Name]; !ok {
				sourceRepositoriesToMigrate = append(sourceRepositoriesToMigrate, item)
			} else {
				log.Println("[🤝] Repository " + *item.Name + " already exists at target organization")
			}
		}

		// initialize new empty string array
		var failedRepositories []string
		for _, repository := range sourceRepositoriesToMigrate {
			if *repository.Name == ".github" {
				continue
			}

			err := ProcessRepoMigration(repository, sourceOrg, targetOrg, sourceToken, targetToken)
			if err != nil {
				log.Println("[❌] Error migrating repository: ", err)
				failedRepositories = append(failedRepositories, *repository.Name)
				continue
			}
		}

		if len(failedRepositories) > 0 {
			//save failed repositories to file
			f, err := os.Create("failed-repositories.txt")
			if err != nil {
				log.Println("[❌] Error creating file: ", err)
				os.Exit(1)
			}
			defer f.Close()

			for _, repository := range failedRepositories {
				_, err := f.WriteString(repository + "\n")
				if err != nil {
					log.Println("[❌] Error writing to file: ", err)
					os.Exit(1)
				}
			}
			log.Println("[❌] Failed repositories saved to file failed-repositories.txt")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(migrateOrgCmd)
}
