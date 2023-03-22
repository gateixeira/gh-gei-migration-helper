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

// ghasRepoSettingsCmd represents the ghasRepoSettings command
var ghasRepoSettingsCmd = &cobra.Command{
	Use:   "ghasRepoSettings",
	Short: "Change GHAS settings for a repository",
	Long: `Change GHAS settings for a given repository.

By default, Advanced Security and Secret Scanning will be deactivated.
Pass the --activate flag to activate the features.`,
	Run: func(cmd *cobra.Command, args []string) {
		organization, err := cmd.Flags().GetString(orgFlagName)
		if err != nil {
			log.Fatalf("failed to get organization flag value: %v", err)
		}
		repository, err := cmd.Flags().GetString(repositoryFlagName)
		if err != nil {
			log.Fatalf("failed to get repository flag value: %v", err)
		}
		activate, err := cmd.Flags().GetBool(activateFlagName)
		if err != nil {
			log.Fatalf("failed to get activate flag value: %v", err)
		}
		token, err := cmd.Flags().GetString("token")
		if err != nil {
			log.Fatalf("failed to get token flag value: %v", err)
		}

		changeGhasRepoSettings(organization, repository, activate, token)
	},
}

func init() {
	rootCmd.AddCommand(ghasRepoSettingsCmd)

	ghasRepoSettingsCmd.Flags().String(repositoryFlagName, "", "The repository to change visibility for.")
	ghasRepoSettingsCmd.MarkFlagRequired(repositoryFlagName)

	ghasRepoSettingsCmd.Flags().BoolP(activateFlagName, "a", false, "Activate GHAS for the organization.")
}

func changeGhasRepoSettings(organization string, repository string, activate bool, token string) {
	var status string
	if activate {
		fmt.Println("Activating GHAS for repository " + repository)
		status = "enabled"
	} else {
		fmt.Println("Deactivating GHAS for repository " + repository)
		status = "disabled"
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	//create new repository object
	newRepoSettings := github.Repository{
		SecurityAndAnalysis: &github.SecurityAndAnalysis{
			AdvancedSecurity: &github.AdvancedSecurity{
				Status: &status,
			},
			SecretScanning: &github.SecretScanning{
				Status: &status,
			},
			SecretScanningPushProtection: &github.SecretScanningPushProtection{
				Status: &status,
			},
		},
	}

	// Update the repository
	_, _, err := client.Repositories.Edit(ctx, organization, repository, &newRepoSettings)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}