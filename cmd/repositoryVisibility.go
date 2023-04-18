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

// repositoryVisibilityCmd represents the repositoryVisibility command
var repositoryVisibilityCmd = &cobra.Command{
	Use:   "change-repository-visibility",
	Short: "Change the visibility of a repository",
	//Long: `Change the visibility of a repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		organization, _ := cmd.Flags().GetString(orgFlagName)
		repository, _ := cmd.Flags().GetString(repositoryFlagName)
		visibility, _ := cmd.Flags().GetString(visibilityFlagName)
		token, _ := cmd.Flags().GetString("token")

		fmt.Println("Changing visibility for repository " + repository + " to " + visibility)
		changeRepositoryVisibility(organization, repository, visibility, token)
	},
}

func init() {
	rootCmd.AddCommand(repositoryVisibilityCmd)

	repositoryVisibilityCmd.Flags().String(tokenFlagName, "t", "The authentication token to use")
	repositoryVisibilityCmd.MarkFlagRequired(tokenFlagName)

	repositoryVisibilityCmd.Flags().String(orgFlagName, "", "The organization to run the command against")
	repositoryVisibilityCmd.MarkFlagRequired(orgFlagName)


	repositoryVisibilityCmd.Flags().String(repositoryFlagName, "", "The repository to change visibility for.")
	repositoryVisibilityCmd.MarkFlagRequired(repositoryFlagName)

	repositoryVisibilityCmd.Flags().String(visibilityFlagName, "", "The new visibility. <private|internal|public>.")
	repositoryVisibilityCmd.MarkFlagRequired(visibilityFlagName)
}

func changeRepositoryVisibility(organization string, repository string, visibility string, token string) {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(tc.Transport)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	client := github.NewClient(rateLimiter)

	//create new repository object
	newRepoSettings := github.Repository{
		Visibility: &visibility,
	}

	// Update the repository
	_, _, err = client.Repositories.Edit(ctx, organization, repository, &newRepoSettings)

	if err != nil {
		//test if error code is 422
		if err, ok := err.(*github.ErrorResponse); ok {
			if err.Response.StatusCode == 422 {
				fmt.Println("Repository is already set to " + visibility)
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}
}