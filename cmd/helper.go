package cmd

import (
	"fmt"

	"github.com/gateixeira/gei-migration-helper/cmd/github"
)

func ProcessRepoMigration(repository github.Repository, sourceOrg string, targetOrg string, sourceToken string, targetToken string) {
	fmt.Print(
		"\n\n========================================\nRepository " + *repository.Name + "\n========================================\n")

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil && *repository.SecurityAndAnalysis.AdvancedSecurity.Status == "enabled" {
		fmt.Println("[🔄] Deactivating GHAS settings at source repository")
		github.ChangeGhasRepoSettings(sourceOrg, repository, "disabled", "disabled", "disabled", sourceToken)
		fmt.Println("[✅] Done")
	}

	workflows := github.GetAllActiveWorkflowsForRepository(sourceOrg, *repository.Name, sourceToken)

	if len(workflows) > 0 {
		fmt.Println("[🔄] Disabling workflows at source repository")
		github.DisableWorkflowsForRepository(sourceOrg, *repository.Name, workflows, sourceToken)
		fmt.Println("[✅] Done")
	}

	fmt.Println("[🔄] Migrating repository")
	github.MigrateRepo(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)
	fmt.Println("[✅] Done")

	fmt.Println("[🔄] Deleting branch protections at target")
	github.DeleteBranchProtections(targetOrg, *repository.Name, targetToken)
	fmt.Println("[✅] Done")

	//check if repository is not private
	if !*repository.Private {
		fmt.Println("[🔄] Repository not private at source. Changing visibility to internal at target")
		github.ChangeRepositoryVisibility(targetOrg, *repository.Name, "internal", targetToken)
		fmt.Println("[✅] Done")
	}

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil && *repository.SecurityAndAnalysis.AdvancedSecurity.Status == "enabled" {
		fmt.Println("[🔄] Activating GHAS settings at target")
		github.ChangeGhasRepoSettings(targetOrg, repository,
			*repository.SecurityAndAnalysis.AdvancedSecurity.Status,
			*repository.SecurityAndAnalysis.SecretScanning.Status,
			*repository.SecurityAndAnalysis.SecretScanningPushProtection.Status, targetToken)
		fmt.Println("[✅] Finished.")

		fmt.Println("[🔄] Reactivating GHAS settings at source repository")
		github.ChangeGhasRepoSettings(sourceOrg, repository,
			*repository.SecurityAndAnalysis.AdvancedSecurity.Status,
			*repository.SecurityAndAnalysis.SecretScanning.Status,
			*repository.SecurityAndAnalysis.SecretScanningPushProtection.Status, sourceToken)
		fmt.Println("[✅] Done")
	}

	if len(workflows) > 0 {
		fmt.Println("[🔄] Enabling workflows at source repository")
		github.EnableWorkflowsForRepository(sourceOrg, *repository.Name, workflows, sourceToken)
		fmt.Println("[✅] Done")
	}
}

func CheckAndMigrateSecretScanning(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) {
	repo := github.GetRepository(repository, sourceOrg, sourceToken)
	if *repo.SecurityAndAnalysis.SecretScanning.Status == "enabled" {
		fmt.Println("[🔄] Migrating secret scanning alerts for repository", repository)
		github.MigrateSecretScanning(repository, sourceOrg, targetOrg, sourceToken, targetToken)
		fmt.Println("[✅] Done")
	} else {
		fmt.Println("[🚫] Skipping repository", repository, "because it secret scanning is not enabled")
	}
}

func CheckAndMigrateCodeScanning(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) {
	repo := github.GetRepository(repository, sourceOrg, sourceToken)

	if github.HasCodeScanningAnalysis(*repo.Name, sourceOrg, sourceToken) {
		fmt.Println("[🔄] Migrating code scanning alerts for repository", repository)
		github.MigrateCodeScanning(repository, sourceOrg, targetOrg, sourceToken, targetToken)
		fmt.Println("[✅] Done")
	} else {
		fmt.Println("[🚫] Skipping repository", repository, "because it does not have code scanning analysis")
	}
}