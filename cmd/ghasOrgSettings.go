/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v50/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

const (
	orgFlagName      = "org"
	activateFlagName = "activate"
)

// ghasOrgSettingsCmd represents the ghasOrgSettings command
var ghasOrgSettingsCmd = &cobra.Command{
	Use:   "ghasOrgSettings",
	Short: "Change GHAS settings for an organization",
	Long: `Change GHAS settings for a given organization.

By default, Advanced Security and Secret Scanning are deactivated.
Pass the --activate flag to activate the features.`,
	Run: func(cmd *cobra.Command, args []string) {
		organization, err := cmd.Flags().GetString(orgFlagName)
		if err != nil {
			log.Fatalf("failed to get organization flag value: %v", err)
		}
		activate, err := cmd.Flags().GetBool(activateFlagName)
		if err != nil {
			log.Fatalf("failed to get activate flag value: %v", err)
		}
		token, err := cmd.Flags().GetString("token")
		if err != nil {
			log.Fatalf("failed to get token flag value: %v", err)
		}

		changeGHASOrgSettings(organization, activate, token)
	},
}

func init() {
	rootCmd.AddCommand(ghasOrgSettingsCmd)

	ghasOrgSettingsCmd.Flags().String(orgFlagName, "", "The organization to run the command against")
	ghasOrgSettingsCmd.MarkFlagRequired(orgFlagName)

	ghasOrgSettingsCmd.Flags().BoolP(activateFlagName, "a", false, "Activate GHAS for the organization")
}

func changeGHASOrgSettings(organization string, activate bool, token string) {
	if activate {
		fmt.Println("Activating GHAS for organization " + organization)
	} else {
		fmt.Println("Deactivating GHAS for organization " + organization)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
		
	client := github.NewClient(tc)

	//create new organization object
	newOrgSettings := github.Organization{
		AdvancedSecurityEnabledForNewRepos: &activate,
		SecretScanningPushProtectionEnabledForNewRepos: &activate,
		SecretScanningEnabledForNewRepos: &activate,
	}


	// Update the organization
	_, _, err := client.Organizations.Edit(ctx, organization, &newOrgSettings)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}