package cmd

import (
	"log"

	"github.com/gateixeira/gei-migration-helper/cmd/github"
)

type errWritter struct {
	err error
}

func ProcessRepoMigration(repository github.Repository, sourceOrg string, targetOrg string, sourceToken string, targetToken string) error {
	log.Print(
		"\n\n========================================\nRepository " + *repository.Name + "\n========================================\n")

	ew := errWritter{}

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil && *repository.SecurityAndAnalysis.AdvancedSecurity.Status == "enabled" {
		ew.LogAndCallStep("Deactivating GHAS settings at source repository", func() error {
			return github.ChangeGhasRepoSettings(sourceOrg, repository, "disabled", "disabled", "disabled", sourceToken)
		})
	}

	workflows, err := github.GetAllActiveWorkflowsForRepository(sourceOrg, *repository.Name, sourceToken)

	if err != nil {
		return err
	}

	if len(workflows) > 0 {
		ew.LogAndCallStep("Disabling workflows at source repository", func() error {
			return github.DisableWorkflowsForRepository(sourceOrg, *repository.Name, workflows, sourceToken)
		})

		if err != nil {
			return err
		}
	}

	ew.LogAndCallStep("Migrating repository", func() error {
		return github.MigrateRepo(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)
	})

	if err != nil {
		return err
	}

	ew.LogAndCallStep("Deleting branch protections at target", func() error {
		return github.DeleteBranchProtections(targetOrg, *repository.Name, targetToken)
	})

	//check if repository is not private
	if !*repository.Private {
		ew.LogAndCallStep("Changing visibility to internal at target", func() error {
			return github.ChangeRepositoryVisibility(targetOrg, *repository.Name, "internal", targetToken)
		})
	}

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil && *repository.SecurityAndAnalysis.AdvancedSecurity.Status == "enabled" {
		ew.LogAndCallStep("Activating GHAS settings at target", func() error {
			return github.ChangeGhasRepoSettings(targetOrg, repository,
				*repository.SecurityAndAnalysis.AdvancedSecurity.Status,
				*repository.SecurityAndAnalysis.SecretScanning.Status,
				*repository.SecurityAndAnalysis.SecretScanningPushProtection.Status, targetToken)
		})
	}

	ew.LogAndCallStep("Migrating code scanning alerts", func() error {
		return CheckAndMigrateCodeScanning(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)
	})

	ew.LogAndCallStep("Archiving source repository", func() error {
		return github.ArchiveRepository(sourceOrg, *repository.Name, sourceToken)
	})

	reEnableOrigin(repository, sourceOrg, sourceToken, workflows)

	if ew.err != nil {
		return ew.err
	}

	return nil
}

func CheckAndMigrateSecretScanning(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) error {
	ew := errWritter{}

	var repo github.Repository
	repo, ew.err = github.GetRepository(repository, sourceOrg, sourceToken)

	if ew.err != nil {
		return ew.err
	}

	if *repo.SecurityAndAnalysis.SecretScanning.Status == "enabled" {
		ew.LogAndCallStep("Migrating secret scanning alerts for repository "+repository, func() error {
			return github.MigrateSecretScanning(repository, sourceOrg, targetOrg, sourceToken, targetToken)
		})
	} else {
		log.Println("[🚫] Skipping repository", repository, "because secret scanning is not enabled")
	}

	if ew.err != nil {
		return ew.err
	}

	return nil
}

func CheckAndMigrateCodeScanning(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) error {
	ew := errWritter{}

	var repo github.Repository
	repo, ew.err = github.GetRepository(repository, sourceOrg, sourceToken)

	if ew.err != nil {
		return ew.err
	}

	var hasCodeScanningAnalysis bool
	hasCodeScanningAnalysis, ew.err = github.HasCodeScanningAnalysis(*repo.Name, sourceOrg, sourceToken)

	if ew.err != nil {
		return ew.err
	}

	if hasCodeScanningAnalysis {
		ew.LogAndCallStep("Migrating code scanning alerts for repository", func() error {
			return github.MigrateCodeScanning(repository, sourceOrg, targetOrg, sourceToken, targetToken)
		})
	} else {
		log.Println("[🚫] Skipping repository", repository, "because it does not have code scanning analysis")
	}

	if ew.err != nil {
		return ew.err
	}

	return nil
}

func ReactivateTargetWorkflows(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) error {
	ew := errWritter{}

	var workflows []github.Workflow
	workflows, ew.err = github.GetAllActiveWorkflowsForRepository(sourceOrg, repository, sourceToken)

	if ew.err != nil {
		return ew.err
	}

	if len(workflows) > 0 {
		ew.LogAndCallStep("Enabling workflows at target repository", func() error {
			return github.EnableWorkflowsForRepository(targetOrg, repository, workflows, targetToken)
		})
	}

	if ew.err != nil {
		return ew.err
	}

	return nil
}

func (ew *errWritter) LogAndCallStep(stepName string, f func() error) {
	if ew.err != nil {
		return
	}
	log.Printf("[🔄] %s\n", stepName)
	ew.err = f()
	if ew.err != nil {
		log.Printf("[❌] %s Error: %v\n", stepName, ew.err)
		return
	}
	log.Printf("[✅] Done: %s\n", stepName)
}

func reEnableOrigin(repository github.Repository, sourceOrg string, sourceToken string, workflows []github.Workflow) {
	ew := errWritter{}

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil && *repository.SecurityAndAnalysis.AdvancedSecurity.Status == "enabled" {
		ew.LogAndCallStep("Resetting GHAS settings at source repository", func() error {
			return github.ChangeGhasRepoSettings(sourceOrg, repository,
				*repository.SecurityAndAnalysis.AdvancedSecurity.Status,
				*repository.SecurityAndAnalysis.SecretScanning.Status,
				*repository.SecurityAndAnalysis.SecretScanningPushProtection.Status, sourceToken)
		})
	}

	if len(workflows) > 0 {
		ew.LogAndCallStep("Re-enabling workflows at source repository", func() error {
			return github.EnableWorkflowsForRepository(sourceOrg, *repository.Name, workflows, sourceToken)
		})
	}
}
