/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
	"os"
)

// ghasSettingsCmd represents the ghasSettings command
var ghasSettingsCmd = &cobra.Command{
	Use:   "ghasOrgSettings",
	Short: "Change GHAS settings for an organization",
	Long: `This command changes GHAS settings for a given organization
	
	It deactivates Advanced Security and Secret Scanning by default.
	
	Pass --activate in order to activate the features.`,
	Run: func(cmd *cobra.Command, args []string) {
 
		organization, _ := cmd.Flags().GetString("organization")
		activate, _ := cmd.Flags().GetBool("activate")
		token, _ := cmd.Flags().GetString("token")

		changeGHASOrgSettings(organization, activate, token)
	},
}

func init() {
	rootCmd.AddCommand(ghasSettingsCmd)

	ghasSettingsCmd.Flags().String("organization", "o", "The organization to run the command against")
	ghasSettingsCmd.MarkFlagRequired("organization")

	ghasSettingsCmd.Flags().BoolP("activate", "a", false, "Activate GHAS for the organization")
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