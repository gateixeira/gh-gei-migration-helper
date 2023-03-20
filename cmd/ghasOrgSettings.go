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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		organization, _ := cmd.Flags().GetString("organization")
		token, _ := cmd.Flags().GetString("token")

		changeGHASOrgSettings(organization, token)
	},
}

func init() {
	rootCmd.AddCommand(ghasSettingsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ghasSettingsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	ghasSettingsCmd.Flags().String("token", "t", "The authentication token to use")
	ghasSettingsCmd.MarkFlagRequired("token")
	ghasSettingsCmd.Flags().String("organization", "o", "The organization to run the command against")
	ghasSettingsCmd.MarkFlagRequired("organization")
}

func changeGHASOrgSettings(organization string, token string) {
	fmt.Println("changeGHASOrgSettings called")

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
		
	client := github.NewClient(tc)

	t := false
	//create new organization object
	newOrgSettings := github.Organization{
		AdvancedSecurityEnabledForNewRepos: &t,
		SecretScanningPushProtectionEnabledForNewRepos: &t,
		SecretScanningEnabledForNewRepos: &t,
	}


	// Update the organization
	_, _, err := client.Organizations.Edit(ctx, organization, &newOrgSettings)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}