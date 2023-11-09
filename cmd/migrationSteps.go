package cmd

import (
	"log"
	"time"

	"github.com/gateixeira/gei-migration-helper/cmd/github"
)

type errWritter struct {
	err error
}

func ProcessRepoMigration(repository github.Repository, sourceOrg string, targetOrg string, sourceToken string, targetToken string) error {
	log.Println("========================================")
	log.Println("Repository " + *repository.Name)
	log.Println("========================================")
	log.Println("Archived: ", *repository.Archived)
	log.Println("Visibility: ", *repository.Visibility)

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil {
		log.Println("GHAS Settings:")
		log.Println("Code Scanning: " + *repository.SecurityAndAnalysis.AdvancedSecurity.Status)
		log.Println("Secret Scanning: " + *repository.SecurityAndAnalysis.SecretScanning.Status)
		log.Println("Push Protection: " + *repository.SecurityAndAnalysis.SecretScanningPushProtection.Status)
		log.Println("Dependabot Updates: " + *repository.SecurityAndAnalysis.DependabotSecurityUpdates.Status)
	}

	ew := errWritter{}

	if repository.SecurityAndAnalysis.AdvancedSecurity == nil || *repository.SecurityAndAnalysis.AdvancedSecurity.Status == "disabled" {
		ew.LogAndCallStep("Activating code scanning at source repository to check for previous analysis", func() error {
			return github.ChangeGhasRepoSettings(sourceOrg, repository, "enabled", "disabled", "disabled", sourceToken)
		})
	}

	hasCodeScanningAnalysis, _ := github.HasCodeScanningAnalysis(sourceOrg, *repository.Name, sourceToken)

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil {
		ew.LogAndCallStep("Disabling GHAS settings at source repository", func() error {
			return github.ChangeGhasRepoSettings(sourceOrg, repository, "disabled", "disabled", "disabled", sourceToken)
		})
	}

	sourceWorkflows, _ := github.GetAllActiveWorkflowsForRepository(sourceOrg, *repository.Name, sourceToken)

	if len(sourceWorkflows) > 0 {
		ew.LogAndCallStep("Disabling workflows at source repository", func() error {
			return github.DisableWorkflowsForRepository(sourceOrg, *repository.Name, sourceWorkflows, sourceToken)
		})
	}

	ew.LogAndCallStep("Migrating repository", func() error {
		return github.MigrateRepo(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)
	})

	targetWorkflows, _ := github.GetAllActiveWorkflowsForRepository(targetOrg, *repository.Name, targetToken)

	if len(targetWorkflows) > 0 {
		//this is unfortunately necessary as the workflows get re-enabled after org migration
		ew.LogAndCallStep("Disabling workflows at target repository", func() error {
			return github.DisableWorkflowsForRepository(targetOrg, *repository.Name, targetWorkflows, targetToken)
		})
	}

	ew.LogAndCallStep("Deleting branch protections at target", func() error {
		return github.DeleteBranchProtections(targetOrg, *repository.Name, targetToken)
	})

	newRepository, err := github.GetRepository(*repository.Name, targetOrg, targetToken)

	if err != nil {
		log.Println("failed to get repository: ", err)
		return err
	}

	if *newRepository.Visibility == "private" {

		if *newRepository.Archived {
			ew.LogAndCallStep("Unarchive target repository", func() error {
				return github.UnarchiveRepository(targetOrg, *repository.Name, targetToken)
			})
		}

		ew.LogAndCallStep("Changing visibility to internal at target", func() error {
			return github.ChangeRepositoryVisibility(targetOrg, *repository.Name, "internal", targetToken)
		})

		if *repository.Archived {
			ew.LogAndCallStep("Archive target repository", func() error {
				return github.ArchiveRepository(targetOrg, *repository.Name, targetToken)
			})
		}

		log.Println("[⌛] Waiting 10 seconds for changes to apply...")
		time.Sleep(10 * time.Second)
	} else {
		log.Println("[🚫] Skipping visibility change because repository is already internal or public")
	}

	if hasCodeScanningAnalysis {
		log.Println("[💡] Repository has code scanning analysis")

		if *newRepository.Archived {
			ew.LogAndCallStep("Unarchive target repository", func() error {
				return github.UnarchiveRepository(targetOrg, *repository.Name, targetToken)
			})
		}

		ew.LogAndCallStep("Activating code scanning at target repository to migrate alerts", func() error {
			return github.ChangeGhasRepoSettings(targetOrg, newRepository, "enabled", "disabled", "disabled", targetToken)
		})

		ew.LogAndCallStep("Activating code scanning at source repository to migrate alerts", func() error {
			return github.ChangeGhasRepoSettings(sourceOrg, repository, "enabled", "disabled", "disabled", sourceToken)
		})

		ew.LogAndCallStep("Migrating code scanning alerts", func() error {
			return github.MigrateCodeScanning(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)
		})

		ew.LogAndCallStep("Deactivating code scanning at target", func() error {
			return github.ChangeGhasRepoSettings(targetOrg, repository, "disabled", "disabled", "disabled", targetToken)
		})

		if *newRepository.Archived {
			ew.LogAndCallStep("Archive target repository", func() error {
				return github.ArchiveRepository(targetOrg, *repository.Name, targetToken)
			})
		}
	} else {
		log.Println("[🚫] No code scan to migrate, skipping.")
	}

	reEnableOrigin(repository, sourceOrg, sourceToken, sourceWorkflows)

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil {
		ew.LogAndCallStep("Resetting GHAS settings at source repository", func() error {
			return github.ChangeGhasRepoSettings(sourceOrg, repository,
				*repository.SecurityAndAnalysis.AdvancedSecurity.Status,
				*repository.SecurityAndAnalysis.SecretScanning.Status,
				*repository.SecurityAndAnalysis.SecretScanningPushProtection.Status, sourceToken)
		})
	}

	//check if repository is not archived
	if !*repository.Archived {
		ew.LogAndCallStep("Archiving source repository", func() error {
			return github.ArchiveRepository(sourceOrg, *repository.Name, sourceToken)
		})
	}

	if ew.err != nil {
		return ew.err
	}

	log.Println("[🎉] Finished migrating repository ", *repository.Name)

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

func ReactivateTargetWorkflows(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) error {
	ew := errWritter{}

	var sourceWorkflows []github.Workflow
	sourceWorkflows, ew.err = github.GetAllActiveWorkflowsForRepository(sourceOrg, repository, sourceToken)

	if ew.err != nil {
		return ew.err
	}

	var targetWorkflows []github.Workflow
	targetWorkflows, ew.err = github.GetAllWorkflowsForRepository(targetOrg, repository, targetToken)

	if ew.err != nil {
		return ew.err
	}

	if len(sourceWorkflows) > 0 {

		// add name of sourceWorkflows to a hash map
		sourceWorkflowsMap := make(map[string]bool)
		for _, workflow := range sourceWorkflows {
			sourceWorkflowsMap[*workflow.Name] = true
		}

		// initialize list of workflows to enable with size of sourceWorkflows
		workflows := make([]github.Workflow, 0, len(sourceWorkflows))
		for _, workflow := range targetWorkflows {
			if _, ok := sourceWorkflowsMap[*workflow.Name]; ok {
				workflows = append(workflows, workflow)
			}
		}

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
