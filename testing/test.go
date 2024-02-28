package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v59/github"
	"golang.org/x/oauth2"
)

type RepoConfig struct {
	token string
	org   string
	repo  string
}

func disableGHAS(ctx context.Context, client *github.Client, config RepoConfig) *github.Repository {
	fmt.Println("Disabling GHAS code scanning")
	repo := setCodeScanningState(ctx, client, config, "disabled")
	return repo
}

func enableCodeScanning(ctx context.Context, client *github.Client, config RepoConfig) *github.Repository {
	fmt.Println("Enabling GHAS code scanning")
	repo := setCodeScanningState(ctx, client, config, "enabled")
	return repo
}

func setCodeScanningState(ctx context.Context, client *github.Client, config RepoConfig, state string) *github.Repository {
	disabled := "disabled"

	payload := &github.SecurityAndAnalysis{
		AdvancedSecurity: &github.AdvancedSecurity{
			Status: &state,
		},
		SecretScanning: &github.SecretScanning{
			Status: &disabled,
		},
		SecretScanningPushProtection: &github.SecretScanningPushProtection{
			Status: &disabled,
		},
	}

	//create new repository object
	newRepoSettings := &github.Repository{
		SecurityAndAnalysis: payload,
	}

	repo, resp, err := client.Repositories.Edit(ctx, config.org, config.repo, newRepoSettings)
	if err != nil {
		fmt.Println("Error updating repository GHAS settings")
		fmt.Println(err)
	}
	fmt.Println("HTTP status for request: ", resp.Status)

	return repo
}

func listCodeScanningAnalyses(ctx context.Context, client *github.Client, config RepoConfig) {
	analyses, response, err := client.CodeScanning.ListAnalysesForRepo(ctx, config.org, config.repo, nil)
	if err != nil {
		fmt.Println("Error listing the code scanning analyses")
		fmt.Println(err)
		fmt.Println(" HTTP response: ", response.Status)
	} else {
		fmt.Println(response.Status)
		for _, analysis := range analyses {
			fmt.Printf("Analysis: %v, sarif id:%v\n", *analysis.ID, *analysis.SarifID)
		}
	}
}

func fetchRepository(ctx context.Context, client *github.Client, config RepoConfig) *github.Repository {
	repo, _, err := client.Repositories.Get(ctx, config.org, config.repo)
	if err != nil {
		fmt.Println("Error fetching repository")
		fmt.Println(err)
		return nil
	}

	fmt.Println("Repository: ", *repo.Name)
	fmt.Println("  is archived?   ", *repo.Archived)
	fmt.Println("  GHAS enabled?  ", *repo.SecurityAndAnalysis.AdvancedSecurity.Status)

	return repo
}

func main() {
	config := new(RepoConfig)
	config.org = "octodemo"
	config.repo = "pm-bs-testing-8001"
	config.token = os.Getenv("GH_TEST_TOKEN")

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Check the repository settings that we are starting out with
	repo := fetchRepository(ctx, client, *config)

	if repo == nil {
		return
	}

	if repo.SecurityAndAnalysis.AdvancedSecurity.Status == nil || *repo.SecurityAndAnalysis.AdvancedSecurity.Status == "enabled" {
		// Switch to disabled
		repo := disableGHAS(ctx, client, *config)
		fmt.Println("  GHAS enabled?  ", *repo.SecurityAndAnalysis.AdvancedSecurity.Status)

		fmt.Println("Pausing for settings to flush")
		time.Sleep(10 * time.Second)
	}

	// This should fail to list
	listCodeScanningAnalyses(ctx, client, *config)

	// Re-enable GHAS code scanning
	enableCodeScanning(ctx, client, *config)
	time.Sleep(10 * time.Second) // Arbitary wait to ensure that things flush through
	fmt.Println("-----------------------------------------------------")
	fmt.Println("Checking repository GHAS settings")
	fetchRepository(ctx, client, *config)
	listCodeScanningAnalyses(ctx, client, *config)
}
