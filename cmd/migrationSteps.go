package cmd

import (
	"log"
	"time"

	"github.com/gateixeira/gei-migration-helper/cmd/github"
)

var MaxRetries = 5

type errWritter struct {
	err error
}

func ProcessRepoMigration(repository github.Repository, sourceOrg string, targetOrg string, sourceToken string, targetToken string, maxRetries int) error {
	log.Println("========================================")
	log.Println("Repository " + *repository.Name)
	log.Println("========================================")
	log.Println("Archived: ", *repository.Archived)
	log.Println("Visibility: ", *repository.Visibility)

	MaxRetries = maxRetries

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil {
		log.Println("GHAS Settings:")
		log.Println("Code Scanning: " + *repository.SecurityAndAnalysis.AdvancedSecurity.Status)
		log.Println("Secret Scanning: " + *repository.SecurityAndAnalysis.SecretScanning.Status)
		log.Println("Push Protection: " + *repository.SecurityAndAnalysis.SecretScanningPushProtection.Status)
		log.Println("Dependabot Updates: " + *repository.SecurityAndAnalysis.DependabotSecurityUpdates.Status)
	}

	ew := errWritter{}

	if repository.SecurityAndAnalysis.AdvancedSecurity == nil || *repository.SecurityAndAnalysis.AdvancedSecurity.Status == "disabled" {
		ew.LogAndCallStep("Activating code scanning at source repository to check for previous analyses", func() error {
			return github.ChangeGhasRepoSettings(sourceOrg, repository, "enabled", "disabled", "disabled", sourceToken)
		})
	}

	codeScanningAnalysis, _ := github.GetCodeScanningAnalysis(sourceOrg, *repository.Name, *repository.DefaultBranch, sourceToken)

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

	newRepository, _ := github.GetRepository(*repository.Name, targetOrg, targetToken)

	targetWorkflows, _ := github.GetAllActiveWorkflowsForRepository(targetOrg, *repository.Name, targetToken)

	if len(targetWorkflows) > 0 {
		//this is unfortunately necessary as the workflows get re-enabled after org migration
		ew.LogAndCallStep("Disabling workflows at target repository", func() error {
			return github.DisableWorkflowsForRepository(targetOrg, *repository.Name, targetWorkflows, targetToken)
		})
	}

	if *newRepository.Archived {
		ew.LogAndCallStep("Unarchive target repository", func() error {
			return github.UnarchiveRepository(targetOrg, *repository.Name, targetToken)
		})
	}

	ew.LogAndCallStep("Deleting branch protections at target", func() error {
		return github.DeleteBranchProtections(targetOrg, *repository.Name, targetToken)
	})

	if *newRepository.Visibility == "private" {
		ew.LogAndCallStep("Changing visibility to internal at target", func() error {
			return github.ChangeRepositoryVisibility(targetOrg, *repository.Name, "internal", targetToken)
		})

		log.Println("[⌛] Waiting 10 seconds for changes to apply...")
		time.Sleep(10 * time.Second)
	} else {
		log.Println("[🚫] Skipping visibility change because repository is already internal or public")
	}

	ew.LogAndCallStep("Activating GHAS at target", func() error {
		return github.ChangeGhasRepoSettings(targetOrg, newRepository, "enabled", "enabled", "enabled", targetToken)
	})

	if len(codeScanningAnalysis) > 0 {
		log.Printf("[💡] Found %d code scanning analysis at source in default branch (%s) before migration", len(codeScanningAnalysis), *repository.DefaultBranch)

		ew.LogAndCallStep("Activating code scanning at source repository to migrate alerts", func() error {
			return github.ChangeGhasRepoSettings(sourceOrg, repository, "enabled", "disabled", "disabled", sourceToken)
		})

		ew.LogAndCallStep("Migrating code scanning alerts", func() error {
			return github.MigrateCodeScanning(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)
		})

		codeScanningAnalysis, _ := github.GetCodeScanningAnalysis(targetOrg, *repository.Name, *repository.DefaultBranch, targetToken)
		log.Printf("[💡] Found %d code scanning analysis at target in default branch (%s) after migration", len(codeScanningAnalysis), *repository.DefaultBranch)

		ew.LogAndCallStep("Deactivating code scanning at source", func() error {
			return github.ChangeGhasRepoSettings(sourceOrg, repository, "disabled", "disabled", "disabled", sourceToken)
		})
	} else {
		log.Println("[🚫] No code scan to migrate, skipping.")
	}

	if *newRepository.Archived {
		ew.LogAndCallStep("Archive target repository", func() error {
			return github.ArchiveRepository(targetOrg, *repository.Name, targetToken)
		})
	}

	reEnableOrigin(repository, sourceOrg, sourceToken, sourceWorkflows)

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

	for i := 0; i < MaxRetries; i++ {
		ew.err = f()
		if ew.err == nil {
			break
		}

		// Exponential backoff: 2^i * 100ms
		time.Sleep(time.Duration((1<<uint(i))*1000) * time.Millisecond)
		log.Printf("[⏳] Retrying %d/%d...", i+1, MaxRetries)
	}

	if ew.err != nil {
		log.Printf("[❌] %s Error: %v\n", stepName, ew.err)
		return
	}
	log.Printf("[✅] Done: %s\n", stepName)
}

func reEnableOrigin(repository github.Repository, sourceOrg string, sourceToken string, workflows []github.Workflow) {
	ew := errWritter{}

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil {
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
