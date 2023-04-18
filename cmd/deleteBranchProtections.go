/*
Package cmd provides a command-line interface for deleting branch protections for a given repository.
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

type BranchProtectionRule struct {
    Nodes []struct {
		Id			  string
    }
	PageInfo struct {
		EndCursor   githubv4.String
		HasNextPage bool
	}
}

// deleteBranchProtectionsCmd represents the branchProtections command
var deleteBranchProtectionsCmd = &cobra.Command{
	Use:   "delete-branch-protections",
	Short: "Delete branch protections for a given repository",
	Long: `Delete branch protections for a given repository.

Provide the name of the repository to delete branch protections from.`,
	Run: func(cmd *cobra.Command, args []string) {
		organization, _ := cmd.Flags().GetString(orgFlagName)
		repository, _ := cmd.Flags().GetString(repositoryFlagName)
		token, _ := cmd.Flags().GetString("token")

		fmt.Println("Deleting branch protections for repository " + repository + " in organization " + organization)
		deleteBranchProtections(organization, repository, token)
	},
}

func init() {
	rootCmd.AddCommand(deleteBranchProtectionsCmd)

	deleteBranchProtectionsCmd.Flags().String(tokenFlagName, "t", "The authentication token to use")
	deleteBranchProtectionsCmd.MarkFlagRequired(tokenFlagName)

	deleteBranchProtectionsCmd.Flags().String(orgFlagName, "", "The organization to run the command against")
	deleteBranchProtectionsCmd.MarkFlagRequired(orgFlagName)

	deleteBranchProtectionsCmd.Flags().String(repositoryFlagName, "", "The repository to delete branch protections from.")
	deleteBranchProtectionsCmd.MarkFlagRequired(repositoryFlagName)
}

func deleteBranchProtections(organization string, repository string, token string) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(tc.Transport)

	if err != nil {
		panic(err)
	}

	clientv4 := githubv4.NewClient(rateLimiter)

	var query struct {
		Repository struct {
			BranchProtectionRules BranchProtectionRule `graphql:"branchProtectionRules(first: 100, after: $cursor)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner": githubv4.String(organization),
		"name": githubv4.String(repository),
		"cursor": (*githubv4.String)(nil),
	}

	results := make([]string, 0)
	for {
		err := clientv4.Query(ctx, &query, variables)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for _, protection := range query.Repository.BranchProtectionRules.Nodes {
			results = append(results, protection.Id)
		}

		variables["cursor"] = query.Repository.BranchProtectionRules.PageInfo.EndCursor

		if !query.Repository.BranchProtectionRules.PageInfo.HasNextPage {
			break
		}
	}

	// // delete all branch protections
	for _, branchProtection := range results {
		var mutate struct {
			DeleteBranchProtectionRule struct { // Empty struct does not work
				ClientMutationId githubv4.ID
			} `graphql:"deleteBranchProtectionRule(input: $input)"`
		}
		input := githubv4.DeleteBranchProtectionRuleInput{
			BranchProtectionRuleID: branchProtection,
		}
	
		ctx := context.WithValue(context.Background(), ctx, branchProtection)
		err := clientv4.Mutate(ctx, &mutate, input, nil)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}	
	}
}