/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v50/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// ghasOrgSettingsCmd represents the ghasOrgSettings command
var ghasOrgSettingsCmd = &cobra.Command{
	Use:   "ghasOrgSettings",
	Short: "Change GHAS settings for an organization",
	Long: `Change GHAS settings for a given organization.

By default, Advanced Security and Secret Scanning will be deactivated.
Pass the --activate flag to activate the features.`,
	Run: func(cmd *cobra.Command, args []string) {
		organization, _ := cmd.Flags().GetString(orgFlagName)
		activate, _ := cmd.Flags().GetBool(activateFlagName)
		token, _ := cmd.Flags().GetString("token")

		fmt.Println("Changing GHAS settings for organization " + organization)
		changeGHASOrgSettings(organization, activate, token)
	},
}

func init() {
	rootCmd.AddCommand(ghasOrgSettingsCmd)

	ghasOrgSettingsCmd.Flags().String(tokenFlagName, "t", "The authentication token to use")
	ghasOrgSettingsCmd.MarkFlagRequired(tokenFlagName)

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
	rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(tc.Transport)

	if err != nil {
		panic(err)
	}
		
	client := github.NewClient(rateLimiter)

	//create new organization object
	newOrgSettings := github.Organization{
		AdvancedSecurityEnabledForNewRepos: &activate,
		SecretScanningPushProtectionEnabledForNewRepos: &activate,
		SecretScanningEnabledForNewRepos: &activate,
	}


	// Update the organization
	_, _, err = client.Organizations.Edit(ctx, organization, &newOrgSettings)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}