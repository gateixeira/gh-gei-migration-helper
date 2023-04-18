package github

import (
	"log"
	"os/exec"
)

func MigrateCodeScanning(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) {
	cmd := exec.Command("gh", "gei", "migrate-code-scanning-alerts", "--source-repo", repository,
		"--source-org", sourceOrg, "--target-org", targetOrg,"--github-source-pat", sourceToken,
			"--github-target-pat", targetToken)

	err := cmd.Run()

	if err != nil {
		log.Fatalf("failed to migrate code scanning alerts %s: %v", repository, err)
	}
}

func MigrateRepo(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) {
	cmd := exec.Command("gh", "gei", "migrate-repo", "--source-repo",
		repository,"--github-source-org", sourceOrg, "--github-target-org", targetOrg,
			"--github-source-pat", sourceToken, "--github-target-pat", targetToken)

	err := cmd.Run()

	if err != nil {
		log.Fatalf("failed to migrate repository %s: %v", repository, err)
	}
}