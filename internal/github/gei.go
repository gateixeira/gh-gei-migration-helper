package github

import (
	"log/slog"
	"os/exec"
)

type GEI struct {
	sourceOrg, targetOrg     string
	sourceToken, targetToken string
}

func NewGEI(source, target, sourceToken, targetToken string) GEI {
	return GEI{source, target, sourceToken, targetToken}
}

func (gei *GEI) MigrateCodeScanning(repository string) error {
	cmd := exec.Command("gh", "gei", "migrate-code-scanning-alerts", "--source-repo", repository,
		"--source-org", gei.sourceOrg, "--target-org", gei.targetOrg, "--github-source-pat", gei.sourceToken,
		"--github-target-pat", gei.targetToken)

	err := cmd.Run()

	if err != nil {
		slog.Error("failed to migrate code scanning alerts: ", err)
		return err
	}

	return nil
}

func (gei *GEI) MigrateSecretScanning(repository string) error {
	cmd := exec.Command(
		"gh", "gei", "migrate-secret-alerts", "--source-repo",
		repository, "--source-org", gei.sourceOrg, "--target-org",
		gei.targetOrg, "--github-source-pat", gei.sourceToken,
		"--github-target-pat", gei.targetToken)

	err := cmd.Run()

	if err != nil {
		slog.Error("failed to migrate secret scanning remediations: ", err)
		return err
	}

	return nil
}

func (gei *GEI) MigrateRepo(repository string) error {
	cmd := exec.Command("gh", "gei", "migrate-repo", "--source-repo",
		repository, "--github-source-org", gei.sourceOrg, "--github-target-org", gei.targetOrg,
		"--github-source-pat", gei.sourceToken, "--github-target-pat", gei.targetToken)

	err := cmd.Run()

	if err != nil {
		slog.Error("failed to migrate repository: ", err)
		return err
	}

	return nil
}
