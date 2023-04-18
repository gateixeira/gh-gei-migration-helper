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

// ghasRepoSettingsCmd represents the ghasRepoSettings command
var ghasRepoSettingsCmd = &cobra.Command{
	Use:   "change-ghas-repo-settings",
	Short: "Change GHAS settings for a repository",
	Long: `Change GHAS settings for a given repository.

By default, Advanced Security and Secret Scanning will be deactivated.
Pass the --activate flag to activate the features.`,
	Run: func(cmd *cobra.Command, args []string) {
		organization, _ := cmd.Flags().GetString(orgFlagName)
		repository, _ := cmd.Flags().GetString(repositoryFlagName)
		activate, _ := cmd.Flags().GetBool(activateFlagName)
		token, _ := cmd.Flags().GetString("token")

		fmt.Println("Changing GHAS settings for repository " + repository + " in organization " + organization)
		changeGhasRepoSettings(organization, repository, activate, token)
	},
}

func init() {
	rootCmd.AddCommand(ghasRepoSettingsCmd)

	ghasRepoSettingsCmd.Flags().String(tokenFlagName, "t", "The authentication token to use")
	ghasRepoSettingsCmd.MarkFlagRequired(tokenFlagName)

	ghasRepoSettingsCmd.Flags().String(orgFlagName, "", "The organization to run the command against")
	ghasRepoSettingsCmd.MarkFlagRequired(orgFlagName)

	ghasRepoSettingsCmd.Flags().String(repositoryFlagName, "", "The repository to change visibility for.")
	ghasRepoSettingsCmd.MarkFlagRequired(repositoryFlagName)

	ghasRepoSettingsCmd.Flags().BoolP(activateFlagName, "a", false, "Activate GHAS for the organization.")
}

func changeGhasRepoSettings(organization string, repository string, activate bool, token string) {
	var status string
	if activate {
		status = "enabled"
	} else {
		status = "disabled"
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
	_, _, err = client.Repositories.Edit(ctx, organization, repository, &newRepoSettings)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}