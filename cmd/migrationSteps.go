package cmd

import (
	"log"

	"github.com/gateixeira/gei-migration-helper/cmd/github"
)

func ProcessRepoMigration(repository github.Repository, sourceOrg string, targetOrg string, sourceToken string, targetToken string) error {
	log.Print(
		"\n\n========================================\nRepository " + *repository.Name + "\n========================================\n")

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil && *repository.SecurityAndAnalysis.AdvancedSecurity.Status == "enabled" {
		LogAndCallStep("Deactivating GHAS settings at source repository", func() error {
			return github.ChangeGhasRepoSettings(sourceOrg, repository, "disabled", "disabled", "disabled", sourceToken)
		})
	}

	workflows, err := github.GetAllActiveWorkflowsForRepository(sourceOrg, *repository.Name, sourceToken)

	if err != nil {
		return err
	}

	if len(workflows) > 0 {
		LogAndCallStep("Disabling workflows at source repository", func() error {
			return github.DisableWorkflowsForRepository(sourceOrg, *repository.Name, workflows, sourceToken)
		})
	}

	LogAndCallStep("Migrating repository", func() error {
		return github.MigrateRepo(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)
	})

	LogAndCallStep("Deleting branch protections at target", func() error {
		return github.DeleteBranchProtections(targetOrg, *repository.Name, targetToken)
	})

	//check if repository is not private
	if !*repository.Private {
		LogAndCallStep("Changing visibility to internal at target", func() error {
			return github.ChangeRepositoryVisibility(targetOrg, *repository.Name, "internal", targetToken)
		})
	}

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil && *repository.SecurityAndAnalysis.AdvancedSecurity.Status == "enabled" {
		LogAndCallStep("Activating GHAS settings at target", func() error {
			return github.ChangeGhasRepoSettings(targetOrg, repository,
				*repository.SecurityAndAnalysis.AdvancedSecurity.Status,
				*repository.SecurityAndAnalysis.SecretScanning.Status,
				*repository.SecurityAndAnalysis.SecretScanningPushProtection.Status, targetToken)
		})

		LogAndCallStep("Reactivating GHAS settings at source repository", func() error {
			return github.ChangeGhasRepoSettings(sourceOrg, repository,
				*repository.SecurityAndAnalysis.AdvancedSecurity.Status,
				*repository.SecurityAndAnalysis.SecretScanning.Status,
				*repository.SecurityAndAnalysis.SecretScanningPushProtection.Status, sourceToken)
		})
	}

	LogAndCallStep("Migrating code scanning alerts", func() error {
		return CheckAndMigrateCodeScanning(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)
	})

	if len(workflows) > 0 {
		LogAndCallStep("Enabling workflows at source repository", func() error {
			return github.EnableWorkflowsForRepository(sourceOrg, *repository.Name, workflows, sourceToken)
		})
	}

	LogAndCallStep("Archiving source repository", func() error {
		return github.ArchiveRepository(sourceOrg, *repository.Name, sourceToken)
	})

	return nil
}

func CheckAndMigrateSecretScanning(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) error {
	repo, err := github.GetRepository(repository, sourceOrg, sourceToken)

	if err != nil {
		return err
	}

	if *repo.SecurityAndAnalysis.SecretScanning.Status == "enabled" {
		LogAndCallStep("Migrating secret scanning alerts for repository", func() error {
			return github.MigrateSecretScanning(repository, sourceOrg, targetOrg, sourceToken, targetToken)
		})
	} else {
		log.Println("[🚫] Skipping repository", repository, "because it secret scanning is not enabled")
	}

	return nil
}

func CheckAndMigrateCodeScanning(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) error {
	repo, err := github.GetRepository(repository, sourceOrg, sourceToken)

	if err != nil {
		return err
	}

	hasCodeScanningAnalysis, err := github.HasCodeScanningAnalysis(*repo.Name, sourceOrg, sourceToken)

	if err != nil {
		return err
	}

	if hasCodeScanningAnalysis {
		LogAndCallStep("Migrating code scanning alerts for repository", func() error {
			return github.MigrateCodeScanning(repository, sourceOrg, targetOrg, sourceToken, targetToken)
		})
	} else {
		log.Println("[🚫] Skipping repository", repository, "because it does not have code scanning analysis")
	}

	return nil
}

func ReactivateTargetWorkflows(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) error {
	workflows, err := github.GetAllActiveWorkflowsForRepository(sourceOrg, repository, sourceToken)

	if err != nil {
		return err
	}

	if len(workflows) > 0 {
		LogAndCallStep("Enabling workflows at target repository", func() error {
			return github.EnableWorkflowsForRepository(targetOrg, repository, workflows, targetToken)
		})
	}

	return nil
}

func LogAndCallStep(stepName string, f func() error) {
	log.Printf("[🔄] %s\n", stepName)
	err := f()
	if err != nil {
		log.Printf("[❌] %s Error: %v\n", stepName, err)
		log.Fatal("[❌] Aborting migration")
	}
	log.Printf("[✅]Done: %s\n", stepName)
}
