/*
Package cmd provides a command-line interface for changing GHAS settings for a given organization.
*/
package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v50/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
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

		fmt.Println("Migrating all repositories from " + sourceOrg + " to " + targetOrg)

		changeGHASOrgSettings(targetOrg, false, targetToken)

		repositories := getRepositories(sourceOrg, sourceToken)

		for _, repository := range repositories {
			if *repository.Name == ".github" {
				continue
			}

			fmt.Print(
				"\n\n========================================\n\n" +
					"Migrating repository " + *repository.Name +
				"\n\n========================================\n\n")

			if *repository.Visibility != "public" {
				changeGhasRepoSettings(sourceOrg, *repository.Name, false, sourceToken)
			}

			migrateRepo(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)

			deleteBranchProtections(targetOrg, *repository.Name, targetToken)

			//check if repository is not private
			if !*repository.Private {
				fmt.Println("Source repo is internal. Changing from private at destination.")
				changeRepositoryVisibility(targetOrg, *repository.Name, "internal", targetToken)
			}
			changeGhasRepoSettings(targetOrg, *repository.Name, true, targetToken)
		}
	},
}

func init() {
	rootCmd.AddCommand(migrateOrgCmd)

	migrateOrgCmd.Flags().String(sourceOrgFlagName, "", "The source organization.")
	migrateOrgCmd.MarkFlagRequired(sourceOrgFlagName)

	migrateOrgCmd.Flags().String(targetOrgFlagName, "", "The target organization.")
	migrateOrgCmd.MarkFlagRequired(targetOrgFlagName)

	migrateOrgCmd.Flags().String(sourceTokenFlagName, "", "The token of the source organization.")
	migrateOrgCmd.MarkFlagRequired(sourceTokenFlagName)

	migrateOrgCmd.Flags().String(targetTokenFlagName, "", "The token of the target organization.")
	migrateOrgCmd.MarkFlagRequired(targetTokenFlagName)

}

func getRepositories(sourceOrg string, sourceToken string) []*github.Repository {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: sourceToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(tc.Transport)

	if err != nil {
		panic(err)
	}

	client := github.NewClient(rateLimiter)

	// list all repositories for the organization
	opt := &github.RepositoryListByOrgOptions{Type: "all", ListOptions: github.ListOptions{PerPage: 10}}
	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, sourceOrg, opt)
		if err != nil {
			log.Fatalf("failed to list repositories: %v", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos

}