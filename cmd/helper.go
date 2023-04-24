package cmd

import (
	"log"

	"github.com/gateixeira/gei-migration-helper/cmd/github"
)

func ProcessRepoMigration(repository github.Repository, sourceOrg string, targetOrg string, sourceToken string, targetToken string) error {
	log.Print(
		"\n\n========================================\nRepository " + *repository.Name + "\n========================================\n")

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil && *repository.SecurityAndAnalysis.AdvancedSecurity.Status == "enabled" {
		log.Println("[🔄] Deactivating GHAS settings at source repository")
		github.ChangeGhasRepoSettings(sourceOrg, repository, "disabled", "disabled", "disabled", sourceToken)
		log.Println("[✅] Done")
	}

	workflows, error := github.GetAllActiveWorkflowsForRepository(sourceOrg, *repository.Name, sourceToken)

	if error != nil {
		return error
	}

	if len(workflows) > 0 {
		log.Println("[🔄] Disabling workflows at source repository")
		error := github.DisableWorkflowsForRepository(sourceOrg, *repository.Name, workflows, sourceToken)
		if error != nil {
			return error
		}
		log.Println("[✅] Done")
	}

	log.Println("[🔄] Migrating repository")
	error = github.MigrateRepo(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)
	if error != nil {
		return error
	}

	log.Println("[✅] Done")

	log.Println("[🔄] Deleting branch protections at target")
	error = github.DeleteBranchProtections(targetOrg, *repository.Name, targetToken)
	if error != nil {
		return error
	}
	log.Println("[✅] Done")

	//check if repository is not private
	if !*repository.Private {
		log.Println("[🔄] Repository not private at source. Changing visibility to internal at target")
		error = github.ChangeRepositoryVisibility(targetOrg, *repository.Name, "internal", targetToken)
		if error != nil {
			return error
		}
		log.Println("[✅] Done")
	}

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil && *repository.SecurityAndAnalysis.AdvancedSecurity.Status == "enabled" {
		log.Println("[🔄] Activating GHAS settings at target")
		error = github.ChangeGhasRepoSettings(targetOrg, repository,
			*repository.SecurityAndAnalysis.AdvancedSecurity.Status,
			*repository.SecurityAndAnalysis.SecretScanning.Status,
			*repository.SecurityAndAnalysis.SecretScanningPushProtection.Status, targetToken)
		if error != nil {
			return error
		}
		log.Println("[✅] Finished.")

		log.Println("[🔄] Reactivating GHAS settings at source repository")
		error = github.ChangeGhasRepoSettings(sourceOrg, repository,
			*repository.SecurityAndAnalysis.AdvancedSecurity.Status,
			*repository.SecurityAndAnalysis.SecretScanning.Status,
			*repository.SecurityAndAnalysis.SecretScanningPushProtection.Status, sourceToken)
		if error != nil {
			return error
		}
		log.Println("[✅] Done")
	}

	log.Println("[🔄] Migrating code scanning alerts")
	error = CheckAndMigrateCodeScanning(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)
	if error != nil {
		return error
	}
	log.Println("[✅] Done")	

	if len(workflows) > 0 {
		log.Println("[🔄] Enabling workflows at source repository")
		error = github.EnableWorkflowsForRepository(sourceOrg, *repository.Name, workflows, sourceToken)
		if error != nil {
			return error
		}
		log.Println("[✅] Done")
	}

	log.Println("[🔄] Archiving source repository")
	error = github.ArchiveRepository(sourceOrg, *repository.Name, sourceToken)
	if error != nil {
		return error
	}
	log.Println("[✅] Done")

	return nil
}

func CheckAndMigrateSecretScanning(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) error {
	repo, err := github.GetRepository(repository, sourceOrg, sourceToken)

	if err != nil {
		return err
	}

	if *repo.SecurityAndAnalysis.SecretScanning.Status == "enabled" {
		log.Println("[🔄] Migrating secret scanning alerts for repository", repository)
		err = github.MigrateSecretScanning(repository, sourceOrg, targetOrg, sourceToken, targetToken)
		
		if err != nil {
			return err
		}
		
		log.Println("[✅] Done")
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
		log.Println("[🔄] Migrating code scanning alerts for repository", repository)
		err = github.MigrateCodeScanning(repository, sourceOrg, targetOrg, sourceToken, targetToken)
		
		if err != nil {
			return err
		}
		
		log.Println("[✅] Done")
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
		log.Println("[🔄] Enabling workflows at target repository")
		err := github.EnableWorkflowsForRepository(targetOrg, repository, workflows, targetToken)
		
		if err != nil {
			return err
		}
		
		log.Println("[✅] Done")
	}

	return nil
}