package cmd

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/gateixeira/gei-migration-helper/cmd/github"
)

var MaxRetries = 5

type errWritter struct {
	err error
}

var logger *slog.Logger

func ProcessRepoMigration(repository github.Repository, sourceOrg string, targetOrg string, sourceToken string, targetToken string, maxRetries int) error {
	logger = slog.With(
		slog.String("repository", *repository.Name),
	)

	logger.Info("starting migration", slog.String("archived", strconv.FormatBool(*repository.Archived)), slog.String("visibility", *repository.Visibility))

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil {
		logger.Info("GHAS Settings",
			slog.String("code Scanning", *repository.SecurityAndAnalysis.AdvancedSecurity.Status),
			slog.String("secret Scanning", *repository.SecurityAndAnalysis.SecretScanning.Status),
			slog.String("push Protection", *repository.SecurityAndAnalysis.SecretScanningPushProtection.Status),
			slog.String("dependabot Updates", *repository.SecurityAndAnalysis.DependabotSecurityUpdates.Status))
	}

	ew := errWritter{}

	if repository.SecurityAndAnalysis.AdvancedSecurity == nil || *repository.SecurityAndAnalysis.AdvancedSecurity.Status == "disabled" {
		if *repository.Archived {
			ew.LogAndCallStep("unarchive source", func() error {
				return github.UnarchiveRepository(sourceOrg, *repository.Name, sourceToken)
			})
		}
		ew.LogAndCallStep("activating code scanning at source to check for previous analyses", func() error {
			return github.ChangeGhasRepoSettings(sourceOrg, repository, "enabled", "disabled", "disabled", sourceToken)
		})
	}

	codeScanningAnalysis, _ := github.GetCodeScanningAnalysis(sourceOrg, *repository.Name, *repository.DefaultBranch, sourceToken)

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil {
		ew.LogAndCallStep("disabling GHAS settings at source", func() error {
			return github.ChangeGhasRepoSettings(sourceOrg, repository, "disabled", "disabled", "disabled", sourceToken)
		})
	}

	var sourceWorkflows []github.Workflow
	sourceWorkflows, ew.err = github.GetAllActiveWorkflowsForRepository(sourceOrg, *repository.Name, sourceToken)

	if ew.err != nil {
		logger.Error("failed to get workflows")
		return ew.err
	}

	if len(sourceWorkflows) > 0 {
		ew.LogAndCallStep("disabling workflows at source", func() error {
			return github.DisableWorkflowsForRepository(sourceOrg, *repository.Name, sourceWorkflows, sourceToken)
		})
	}

	ew.LogAndCallStep("migrating", func() error {
		return github.MigrateRepo(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)
	})

	newRepository, err := github.GetRepository(*repository.Name, targetOrg, targetToken)

	if err != nil {
		logger.Error("failed to migrate")
		return err
	}

	var targetWorkflows []github.Workflow
	targetWorkflows, ew.err = github.GetAllActiveWorkflowsForRepository(targetOrg, *repository.Name, targetToken)

	if ew.err != nil {
		logger.Error("failed to get workflows")
		return ew.err
	}

	if len(targetWorkflows) > 0 {
		//this is unfortunately necessary as the workflows get re-enabled after org migration
		ew.LogAndCallStep("disabling workflows at target", func() error {
			return github.DisableWorkflowsForRepository(targetOrg, *repository.Name, targetWorkflows, targetToken)
		})
	}

	if *newRepository.Archived {
		ew.LogAndCallStep("unarchive target", func() error {
			return github.UnarchiveRepository(targetOrg, *repository.Name, targetToken)
		})
	}

	ew.LogAndCallStep("deleting branch protections at target", func() error {
		return github.DeleteBranchProtections(targetOrg, *repository.Name, targetToken)
	})

	if *newRepository.Visibility == "private" {
		ew.LogAndCallStep("changing visibility to internal at target", func() error {
			return github.ChangeRepositoryVisibility(targetOrg, *repository.Name, "internal", targetToken)
		})

		logger.Debug("waiting 10 seconds for changes to apply...")
		time.Sleep(10 * time.Second)
	} else {
		logger.Info("skipping visibility change because source is already internal or public")
	}

	ew.LogAndCallStep("activating GHAS at target", func() error {
		return github.ChangeGhasRepoSettings(targetOrg, newRepository, "enabled", "enabled", "enabled", targetToken)
	})

	if len(codeScanningAnalysis) <= 0 {
		logger.Info("no code scan to migrate, skipping.")
	} else {
		logger.Info(fmt.Sprintf("found %d code scanning analysis at source in default branch (%s) before migration", len(codeScanningAnalysis), *repository.DefaultBranch))

		ew.LogAndCallStep("activating code scanning at source to migrate alerts", func() error {
			return github.ChangeGhasRepoSettings(sourceOrg, repository, "enabled", "disabled", "disabled", sourceToken)
		})

		ew.LogAndCallStep("migrating code scanning alerts", func() error {
			return github.MigrateCodeScanning(*repository.Name, sourceOrg, targetOrg, sourceToken, targetToken)
		})

		codeScanningAnalysis, ew.err = github.GetCodeScanningAnalysis(targetOrg, *repository.Name, *repository.DefaultBranch, targetToken)

		if ew.err != nil {
			logger.Error("failed to get code scanning analysis")
			return ew.err
		}

		logger.Info(fmt.Sprintf("found %d code scanning analysis at target in default branch (%s) after migration", len(codeScanningAnalysis), *repository.DefaultBranch))

		ew.LogAndCallStep("deactivating code scanning at source", func() error {
			return github.ChangeGhasRepoSettings(sourceOrg, repository, "disabled", "disabled", "disabled", sourceToken)
		})
	}

	if *newRepository.Archived {
		ew.LogAndCallStep("archive target", func() error {
			return github.ArchiveRepository(targetOrg, *repository.Name, targetToken)
		})
	}

	reEnableOrigin(repository, sourceOrg, sourceToken, sourceWorkflows)

	//check if repository is not archived
	if !*repository.Archived {
		ew.LogAndCallStep("archiving source", func() error {
			return github.ArchiveRepository(sourceOrg, *repository.Name, sourceToken)
		})
	}

	if ew.err != nil {
		logger.Error("error: %v", ew.err)
		return ew.err
	}

	logger.Info("finished migrating")

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
		ew.LogAndCallStep("migrating secret scanning alerts", func() error {
			return github.MigrateSecretScanning(repository, sourceOrg, targetOrg, sourceToken, targetToken)
		})
	} else {
		logger.Info("skipping because secret scanning is not enabled")
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

		ew.LogAndCallStep("Enabling workflows at target", func() error {
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
	logger.Debug(fmt.Sprintf("%s", stepName))

	for i := 0; i < MaxRetries; i++ {
		ew.err = f()
		if ew.err == nil {
			break
		}

		// Exponential backoff: 2^i * 100ms
		time.Sleep(time.Duration((1<<uint(i))*1000) * time.Millisecond)
		logger.Debug(fmt.Sprintf("retrying %d/%d...", i+1, MaxRetries))
	}

	if ew.err != nil {
		logger.Error("%s error: %v", stepName, ew.err)
		return
	}
	logger.Debug(fmt.Sprintf("done: %s", stepName))
}

func reEnableOrigin(repository github.Repository, sourceOrg string, sourceToken string, workflows []github.Workflow) {
	ew := errWritter{}

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil {
		ew.LogAndCallStep("resetting GHAS settings at source", func() error {
			return github.ChangeGhasRepoSettings(sourceOrg, repository,
				*repository.SecurityAndAnalysis.AdvancedSecurity.Status,
				*repository.SecurityAndAnalysis.SecretScanning.Status,
				*repository.SecurityAndAnalysis.SecretScanningPushProtection.Status, sourceToken)
		})
	}

	if len(workflows) > 0 {
		ew.LogAndCallStep("re-enabling workflows at source", func() error {
			return github.EnableWorkflowsForRepository(sourceOrg, *repository.Name, workflows, sourceToken)
		})
	}
}
