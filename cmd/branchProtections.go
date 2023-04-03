/*
Package cmd provides a command-line interface for deleting branch protections for a given repository.
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

type BranchProtectionRule struct {
    Nodes []struct {
        Pattern       string
        ID            githubv4.ID
    }
}

// branchProtectionsCmd represents the branchProtections command
var branchProtectionsCmd = &cobra.Command{
	Use:   "branchProtections",
	Short: "Delete branch protections for a given repository",
	Long: `Delete branch protections for a given repository.

Provide the name of the repository to delete branch protections from.`,
	Run: func(cmd *cobra.Command, args []string) {
		organization, err := cmd.Flags().GetString(orgFlagName)
		if err != nil {
			log.Fatalf("failed to get organization flag value: %v", err)
		}
		repository, err := cmd.Flags().GetString(repositoryFlagName)
		if err != nil {
			log.Fatalf("failed to get repository flag value: %v", err)
		}
		token, err := cmd.Flags().GetString("token")
		if err != nil {
			log.Fatalf("failed to get token flag value: %v", err)
		}

		deleteBranchProtections(organization, repository, token)
	},
}

func init() {
	rootCmd.AddCommand(branchProtectionsCmd)

	branchProtectionsCmd.Flags().String(repositoryFlagName, "", "The repository to delete branch protections from.")
	branchProtectionsCmd.MarkFlagRequired(repositoryFlagName)
}

func deleteBranchProtections(organization string, repository string, token string) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	//client := github.NewClient(tc)

	clientv4 := githubv4.NewClient(tc)

	var query struct {
		Repository struct {
			BranchProtectionRules BranchProtectionRule `graphql:"branchProtectionRules(first: 100)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner": githubv4.String(organization),
		"name": githubv4.String(repository),
	}

	err := clientv4.Query(ctx, &query, variables)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	
	branchProtections := query.Repository.BranchProtectionRules.Nodes

	// // delete all branch protections
	for _, branchProtection := range branchProtections {
		var mutate struct {
			DeleteBranchProtectionRule struct { // Empty struct does not work
				ClientMutationId githubv4.ID
			} `graphql:"deleteBranchProtectionRule(input: $input)"`
		}
		input := githubv4.DeleteBranchProtectionRuleInput{
			BranchProtectionRuleID: branchProtection.ID,
		}
	
		ctx := context.WithValue(context.Background(), ctx, branchProtection.ID)
		err := clientv4.Mutate(ctx, &mutate, input, nil)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}	
	}

	os.Exit(0)
}