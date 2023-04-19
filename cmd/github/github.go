package github

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v50/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Repository github.Repository

type BranchProtectionRule struct {
    Nodes []struct {
		Id			  string
    }
	PageInfo struct {
		EndCursor   githubv4.String
		HasNextPage bool
	}
}

var (
	ctx context.Context
	clientV3 *github.Client
	clientV4 *githubv4.Client
	accessToken string
)

func checkClients(token string) {
	if clientV3 == nil || clientV4 == nil || token != accessToken {
		accessToken = token
		ctx = context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(tc.Transport)

		if err != nil {
			panic(err)
		}

		clientV3 = github.NewClient(rateLimiter)
		clientV4 = githubv4.NewClient(rateLimiter)
	}
}

func DeleteBranchProtections(organization string, repository string, token string) {
	checkClients(token)

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
		err := clientV4.Query(ctx, &query, variables)
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
		err := clientV4.Mutate(ctx, &mutate, input, nil)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}	
	}
}

func ChangeGHASOrgSettings(organization string, activate bool, token string) {
	checkClients(token)

	//create new organization object
	newOrgSettings := github.Organization{
		AdvancedSecurityEnabledForNewRepos: &activate,
		SecretScanningPushProtectionEnabledForNewRepos: &activate,
		SecretScanningEnabledForNewRepos: &activate,
	}

	// Update the organization
	_, _, err := clientV3.Organizations.Edit(ctx, organization, &newOrgSettings)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func ChangeGhasRepoSettings(organization string, repository Repository, activate bool, token string) {
	checkClients(token)

	var status string

	if activate {
		status = "enabled"
	} else {
		status = "disabled"
	}

	var payload *github.SecurityAndAnalysis
	//GHAS is always enabled for public repositories and PATCH fails when trying to set to disabled
	if *repository.Visibility == "public" {
		payload = &github.SecurityAndAnalysis{
			SecretScanning: &github.SecretScanning{
				Status: &status,
			},
			SecretScanningPushProtection: &github.SecretScanningPushProtection{
				Status: &status,
			},
		}
	} else {
		payload = &github.SecurityAndAnalysis{
			AdvancedSecurity: &github.AdvancedSecurity{
				Status: &status,
			},
			SecretScanning: &github.SecretScanning{
				Status: &status,
			},
			SecretScanningPushProtection: &github.SecretScanningPushProtection{
				Status: &status,
			},
		}
	}

	//create new repository object
	newRepoSettings := github.Repository{
		SecurityAndAnalysis: payload,
	}

	// Update the repository
	_, _, err := clientV3.Repositories.Edit(ctx, organization, *repository.Name, &newRepoSettings)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func GetRepository(repoName string, org string, token string) Repository {
	checkClients(token)

	repo, _, err := clientV3.Repositories.Get(ctx, org, repoName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return Repository(*repo)
}

func GetRepositories(org string, token string) []Repository {
	checkClients(token)

	// list all repositories for the organization
	opt := &github.RepositoryListByOrgOptions{Type: "all", ListOptions: github.ListOptions{PerPage: 10}}
	var allRepos []*github.Repository
	for {
		repos, resp, err := clientV3.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			log.Fatalf("failed to list repositories: %v", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	var allReposStruct []Repository
	for _, repo := range allRepos {
		allReposStruct = append(allReposStruct, Repository(*repo))
	}

	return allReposStruct

}

func ChangeRepositoryVisibility(organization string, repository string, visibility string, token string) {
	checkClients(token)

	//create new repository object
	newRepoSettings := github.Repository{
		Visibility: &visibility,
	}

	// Update the repository
	_, _, err := clientV3.Repositories.Edit(ctx, organization, repository, &newRepoSettings)

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
