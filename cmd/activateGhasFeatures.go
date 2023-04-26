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

// migrateRepoCmd represents the migrateRepo command
var activateGhasFeaturesCmd = &cobra.Command{
	Use:   "activate-ghas-features",
	Short: "Activate GHAS features for all orgs in an enterprise",
	Run: func(cmd *cobra.Command, args []string) {
		enterprise, _ := cmd.Flags().GetString("enterprise")
		token, _ := cmd.Flags().GetString("token")
		organization, _ := cmd.Flags().GetString("organization")
		url, _ := cmd.Flags().GetString("url")

		if organization != "" {
			log.Println("[🔄] Activating GHAS settings for organization: " + organization)
			error := github.ChangeGHASOrgSettings(organization, true, token, url)

			if error != nil {
				log.Println("[❌] Error activating GHAS settings for organization: " + organization)
				os.Exit(1)
			}

			log.Println("[✅] Done")
			os.Exit(0)
		}

		log.Println("[🔄] Fetching organizations from enterprise...")
		organizations, err := github.GetOrganizationsInEnterprise(enterprise, token, url)
		log.Println("[✅] Done")

		if err != nil {
			log.Println("[❌] Error fetching organizations from enterprise")
			os.Exit(1)
		}

		for _, organization := range organizations {
			log.Println("[🔄] Activating GHAS settings for organization: " + organization)
			error := github.ChangeGHASOrgSettings(organization, true, token, url)

			if error != nil {
				log.Println("[❌] Error activating GHAS settings for organization: " + organization)
				continue
			}
			log.Println("[✅] Done")
		}
	},
}

func init() {
	rootCmd.AddCommand(activateGhasFeaturesCmd)

	activateGhasFeaturesCmd.Flags().String("enterprise", "", "The slug of the enterprise")
	activateGhasFeaturesCmd.Flags().String("token", "", "The access token")
	activateGhasFeaturesCmd.Flags().String("organization", "", "To filter for a single organization")
	activateGhasFeaturesCmd.Flags().String("url", "", "URL of the GitHub instance")
}
