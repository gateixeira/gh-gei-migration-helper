/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

const (
	repositoryFlagName      = "repo"
	visibilityFlagName = "visibility"
)

// repositoryVisibilityCmd represents the repositoryVisibility command
var repositoryVisibilityCmd = &cobra.Command{
	Use:   "repositoryVisibility",
	Short: "Change the visibility of a repository",
	//Long: `Change the visibility of a repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		repository, err := cmd.Flags().GetString(repositoryFlagName)
		if err != nil {
			log.Fatalf("failed to get repository flag value: %v", err)
		}
		visibility, err := cmd.Flags().GetString(visibilityFlagName)
		if err != nil {
			log.Fatalf("failed to get activate flag value: %v", err)
		}
		token, err := cmd.Flags().GetString("token")
		if err != nil {
			log.Fatalf("failed to get token flag value: %v", err)
		}

		changerepositoryVisibility(repository, visibility, token)
	},
}

func init() {
	rootCmd.AddCommand(repositoryVisibilityCmd)

	repositoryVisibilityCmd.Flags().String(repositoryFlagName, "", "The repository to change visibility for. Format: owner/repo")
	repositoryVisibilityCmd.MarkFlagRequired(repositoryFlagName)

	repositoryVisibilityCmd.Flags().String(visibilityFlagName, "", "The new visibility. <private|internal|public>")
	repositoryVisibilityCmd.MarkFlagRequired(visibilityFlagName)
}

func changerepositoryVisibility(repository string, visibility string, token string) {
	fmt.Println("Changing visibility for repository " + repository + " to " + visibility)

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	//create new repository object
	newRepoSettings := github.Repository{
		Visibility: &visibility,
	}

	// Update the repository
	_, _, err := client.Repositories.Edit(ctx, strings.Split(repository, "/")[0], strings.Split(repository, "/")[1], &newRepoSettings)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}