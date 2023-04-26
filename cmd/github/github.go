package github

import (
	"context"
	"log"
	"time"

	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v50/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Repository github.Repository

const githubDelay = 720 * time.Millisecond

type BranchProtectionRule struct {
	Nodes []struct {
		Id string
	}
	PageInfo struct {
		EndCursor   githubv4.String
		HasNextPage bool
	}
}

var (
	ctx         context.Context
	clientV3    *github.Client
	clientV4    *githubv4.Client
	accessToken string
)

func checkClients(token string) error {

	// Sleep to avoid hitting the rate limit. Not ideal but a valid trade-off as the migration step in GEI consumes the majority of the time
	time.Sleep(githubDelay)

	if clientV3 == nil || clientV4 == nil || token != accessToken {
		accessToken = token
		ctx = context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(tc.Transport)

		if err != nil {
			return err
		}

		clientV3 = github.NewClient(rateLimiter)
		clientV4 = githubv4.NewClient(rateLimiter)
	}

	return nil
}

func logTokenRateLimit(response *github.Response) {
	log.Printf("[ℹ️] Quota remaining: %d, Limit: %d, Reset: %s", response.Rate.Remaining, response.Rate.Limit, response.Rate.Reset)
}

func DeleteBranchProtections(organization string, repository string, token string) error {
	checkClients(token)

	var query struct {
		Repository struct {
			BranchProtectionRules BranchProtectionRule `graphql:"branchProtectionRules(first: 100, after: $cursor)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":  githubv4.String(organization),
		"name":   githubv4.String(repository),
		"cursor": (*githubv4.String)(nil),
	}

	results := make([]string, 0)
	for {
		err := clientV4.Query(ctx, &query, variables)
		if err != nil {
			log.Println("Error querying branch protection rules: ", err)
			return err
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
			log.Println("Error deleting branch protection rule: ", err)
			return err
		}
	}

	return nil
}

func ChangeGHASOrgSettings(organization string, activate bool, token string) error {
	checkClients(token)

	//create new organization object
	newOrgSettings := github.Organization{
		AdvancedSecurityEnabledForNewRepos:             &activate,
		SecretScanningPushProtectionEnabledForNewRepos: &activate,
		SecretScanningEnabledForNewRepos:               &activate,
	}

	// Update the organization
	_, response, err := clientV3.Organizations.Edit(ctx, organization, &newOrgSettings)

	logTokenRateLimit(response)

	if err != nil {
		log.Println("Error updating organization settings: ", err)
	}

	return err
}

func ChangeGhasRepoSettings(organization string, repository Repository, ghas string, secretScanning string, pushProtection string, token string) error {
	checkClients(token)

	var payload *github.SecurityAndAnalysis
	//GHAS is always enabled for public repositories and PATCH fails when trying to set to disabled
	if *repository.Visibility == "public" {
		payload = &github.SecurityAndAnalysis{
			SecretScanning: &github.SecretScanning{
				Status: &secretScanning,
			},
			SecretScanningPushProtection: &github.SecretScanningPushProtection{
				Status: &pushProtection,
			},
		}
	} else {
		payload = &github.SecurityAndAnalysis{
			AdvancedSecurity: &github.AdvancedSecurity{
				Status: &ghas,
			},
			SecretScanning: &github.SecretScanning{
				Status: &secretScanning,
			},
			SecretScanningPushProtection: &github.SecretScanningPushProtection{
				Status: &pushProtection,
			},
		}
	}

	//create new repository object
	newRepoSettings := github.Repository{
		SecurityAndAnalysis: payload,
	}

	// Update the repository
	_, response, err := clientV3.Repositories.Edit(ctx, organization, *repository.Name, &newRepoSettings)

	logTokenRateLimit(response)

	if err != nil {
		log.Println("Error updating repository settings: ", err)
	}

	return err
}

func GetRepository(repoName string, org string, token string) (Repository, error) {
	checkClients(token)

	repo, _, err := clientV3.Repositories.Get(ctx, org, repoName)
	if err != nil {
		log.Println("Error getting repository: ", err)
		return Repository{}, err
	}

	return Repository(*repo), nil
}

func GetRepositories(org string, token string) ([]Repository, error) {
	checkClients(token)

	// list all repositories for the organization
	opt := &github.RepositoryListByOrgOptions{Type: "all", ListOptions: github.ListOptions{PerPage: 10}}
	var allRepos []*github.Repository
	for {
		repos, response, err := clientV3.Repositories.ListByOrg(ctx, org, opt)

		logTokenRateLimit(response)

		if err != nil {
			log.Println("Error getting repositories: ", err)
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if response.NextPage == 0 {
			break
		}
		opt.Page = response.NextPage
	}

	var allReposStruct []Repository
	for _, repo := range allRepos {
		allReposStruct = append(allReposStruct, Repository(*repo))
	}

	return allReposStruct, nil

}

func ChangeRepositoryVisibility(organization string, repository string, visibility string, token string) error {
	checkClients(token)

	//create new repository object
	newRepoSettings := github.Repository{
		Visibility: &visibility,
	}

	// Update the repository
	_, response, err := clientV3.Repositories.Edit(ctx, organization, repository, &newRepoSettings)

	logTokenRateLimit(response)

	if err != nil {
		//test if error code is 422
		if err, ok := err.(*github.ErrorResponse); ok {
			if err.Response.StatusCode == 422 {
				log.Println("Repository is already set to " + visibility)
				return nil
			} else {
				log.Println("Error changing repository visibility: ", err)
			}
		}
	}

	return err
}

func GetAllActiveWorkflowsForRepository(organization string, repository string, token string) ([]*github.Workflow, error) {
	checkClients(token)

	// list all workflows for the repository
	opt := &github.ListOptions{PerPage: 10}
	var allWorkflows []*github.Workflow
	for {
		workflows, response, err := clientV3.Actions.ListWorkflows(ctx, organization, repository, opt)

		logTokenRateLimit(response)

		if err != nil {
			log.Println("failed to list workflows: ", err)
			return nil, err
		}
		allWorkflows = append(allWorkflows, workflows.Workflows...)
		if response.NextPage == 0 {
			break
		}
		opt.Page = response.NextPage
	}

	var activeWorkflowsStruct []*github.Workflow
	for _, workflow := range allWorkflows {
		if *workflow.State == "active" {
			activeWorkflowsStruct = append(activeWorkflowsStruct, workflow)
		}
	}

	return activeWorkflowsStruct, nil
}

func DisableWorkflowsForRepository(organization string, repository string, workflows []*github.Workflow, token string) error {
	checkClients(token)

	// disable all workflows
	for _, workflow := range workflows {
		response, err := clientV3.Actions.DisableWorkflowByID(ctx, organization, repository, *workflow.ID)

		logTokenRateLimit(response)

		if err, ok := err.(*github.ErrorResponse); ok {
			log.Println("failed to disable workflow: ", workflow.Name)

			if err.Response.StatusCode == 422 {
				log.Println("error is 422. Skipping...")
				return nil
			}
			return err
		}
	}

	return nil
}

func EnableWorkflowsForRepository(organization string, repository string, workflows []*github.Workflow, token string) error {
	checkClients(token)

	// enable all workflows
	for _, workflow := range workflows {
		response, err := clientV3.Actions.EnableWorkflowByID(ctx, organization, repository, *workflow.ID)

		logTokenRateLimit(response)

		if err, ok := err.(*github.ErrorResponse); ok {
			log.Println("failed to enable workflow: ", workflow.Name)

			if err.Response.StatusCode == 422 {
				log.Println("error is 422. Skipping...")
				return nil
			}
			return err
		}
	}

	return nil
}

func HasCodeScanningAnalysis(organization string, repository string, token string) (bool, error) {
	checkClients(token)

	//list code scanning alerts
	opt := &github.AlertListOptions{}

	_, response, err := clientV3.CodeScanning.ListAlertsForRepo(ctx, organization, repository, opt)

	logTokenRateLimit(response)

	if err != nil {
		//test if error code is 404
		if err, ok := err.(*github.ErrorResponse); ok {
			if err.Response.StatusCode == 404 {
				return false, nil
			} else {
				log.Println("Error getting code scanning alerts: ", err)
				return false, err
			}
		}
	}

	return true, nil
}

func ArchiveRepository(organization string, repository string, token string) error {
	checkClients(token)

	archive := true
	newRepoSettings := github.Repository{
		Archived: &archive,
	}

	_, response, err := clientV3.Repositories.Edit(ctx, organization, repository, &newRepoSettings)

	logTokenRateLimit(response)

	if err != nil {
		if err, ok := err.(*github.ErrorResponse); ok {
			if err.Response.StatusCode == 403 {
				//repository is already archived
				return nil
			} else {
				log.Println("Error archiving repository: ", err)
				return err
			}
		}
	}

	return err
}
