package migration

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/gateixeira/gei-migration-helper/internal/github"
	"github.com/gateixeira/gei-migration-helper/pkg/logging"
)

type orgs struct {
	source, target     string
	sourceGC, targetGC *github.GitHubClient
}

type migrationResult struct {
	Timestamp time.Time    `json:"timestamp"`
	SourceOrg string       `json:"sourceOrg"`
	TargetOrg string       `json:"targetOrg"`
	Migrated  []repoStatus `json:"migrated"`
	Failed    []repoStatus `json:"failed"`
}

type repoStatus struct {
	Name           string `json:"name"`
	ID             int64  `json:"id"`
	Archived       bool   `json:"archived"`
	CodeScanning   string `json:"codeScanning" default:"disabled"`
	SecretScanning string `json:"secretScanning" default:"disabled"`
	PushProtection string `json:"pushProtection" default:"disabled"`
}

type Migration interface {
	Migrate() (migrationResult, error)
}

var maxRetries = 5

func processRepoMigration(ctx context.Context, logger *slog.Logger, repository github.Repository, orgs orgs, gei github.GEI) error {
	logger.Info("migration", "repository", *repository.Name, slog.String("archived", strconv.FormatBool(*repository.Archived)), slog.String("visibility", *repository.Visibility))

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil {
		logger.Info("GHAS Settings",
			"repository", *repository.Name,
			slog.String("code Scanning", *repository.SecurityAndAnalysis.AdvancedSecurity.Status),
			slog.String("secret Scanning", *repository.SecurityAndAnalysis.SecretScanning.Status),
			slog.String("push Protection", *repository.SecurityAndAnalysis.SecretScanningPushProtection.Status),
			slog.String("dependabot Updates", *repository.SecurityAndAnalysis.DependabotSecurityUpdates.Status))
	}

	ew := errWritter{}

	if repository.SecurityAndAnalysis.AdvancedSecurity == nil || *repository.SecurityAndAnalysis.AdvancedSecurity.Status == "disabled" {
		if *repository.Archived {
			ew.logAndCallStep(logger, "unarchive source", func() error {
				return orgs.sourceGC.UnarchiveRepository(ctx, orgs.source, *repository.Name)
			})
		}
		ew.logAndCallStep(logger, "activating code scanning at source to check for previous analyses", func() error {
			return orgs.sourceGC.ChangeGhasRepoSettings(ctx, orgs.source, repository, "enabled", "disabled", "disabled")
		})
	}

	codeScanningAnalysis, _ := orgs.sourceGC.GetCodeScanningAnalysis(ctx, orgs.source, *repository.Name, *repository.DefaultBranch)

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil {
		ew.logAndCallStep(logger, "disabling GHAS settings at source", func() error {
			return orgs.sourceGC.ChangeGhasRepoSettings(ctx, orgs.source, repository, "disabled", "disabled", "disabled")
		})
	}

	var sourceWorkflows []github.Workflow
	sourceWorkflows, ew.err = orgs.sourceGC.GetAllActiveWorkflowsForRepository(ctx, orgs.source, *repository.Name)

	if ew.err != nil {
		logger.Error("failed to get workflows")
		return ew.err
	}

	if len(sourceWorkflows) > 0 {
		ew.logAndCallStep(logger, "disabling workflows at source", func() error {
			return orgs.sourceGC.DisableWorkflowsForRepository(ctx, orgs.source, *repository.Name, sourceWorkflows)
		})
	}

	ew.logAndCallStep(logger, "migrating", func() error {
		return gei.MigrateRepo(*repository.Name)
	})

	newRepository, err := orgs.targetGC.GetRepository(ctx, *repository.Name, orgs.target)

	if err != nil {
		logger.Error("failed to migrate")
		return err
	}

	var targetWorkflows []github.Workflow
	targetWorkflows, ew.err = orgs.targetGC.GetAllActiveWorkflowsForRepository(ctx, orgs.target, *repository.Name)

	if ew.err != nil {
		logger.Error("failed to get workflows")
		return ew.err
	}

	if len(targetWorkflows) > 0 {
		//this is unfortunately necessary as the workflows get re-enabled after org migration
		ew.logAndCallStep(logger, "disabling workflows at target", func() error {
			return orgs.targetGC.DisableWorkflowsForRepository(ctx, orgs.target, *repository.Name, targetWorkflows)
		})
	}

	if *newRepository.Archived {
		ew.logAndCallStep(logger, "unarchive target", func() error {
			return orgs.targetGC.UnarchiveRepository(ctx, orgs.target, *repository.Name)
		})
	}

	ew.logAndCallStep(logger, "deleting branch protections at target", func() error {
		return orgs.targetGC.DeleteBranchProtections(ctx, orgs.target, *repository.Name)
	})

	if *newRepository.Visibility == "private" {
		ew.logAndCallStep(logger, "changing visibility to internal at target", func() error {
			return orgs.targetGC.ChangeRepositoryVisibility(ctx, orgs.target, *repository.Name, "internal")
		})

		logger.Debug("waiting 10 seconds for changes to apply...")
		time.Sleep(10 * time.Second)
	} else {
		logger.Info("skipping visibility change because source is already internal or public")
	}

	ew.logAndCallStep(logger, "activating GHAS at target", func() error {
		return orgs.targetGC.ChangeGhasRepoSettings(ctx, orgs.target, newRepository, "enabled", "enabled", "enabled")
	})

	if len(codeScanningAnalysis) <= 0 {
		logger.Info("no code scan to migrate, skipping.")
	} else {
		logger.Info(fmt.Sprintf("found %d code scanning analysis at source in default branch (%s) before migration", len(codeScanningAnalysis), *repository.DefaultBranch))

		ew.logAndCallStep(logger, "activating code scanning at source to migrate alerts", func() error {
			return orgs.sourceGC.ChangeGhasRepoSettings(ctx, orgs.source, repository, "enabled", "disabled", "disabled")
		})

		ew.logAndCallStep(logger, "migrating code scanning alerts", func() error {
			return gei.MigrateCodeScanning(*repository.Name)
		})

		codeScanningAnalysis, ew.err = orgs.targetGC.GetCodeScanningAnalysis(ctx, orgs.target, *repository.Name, *repository.DefaultBranch)

		if ew.err != nil {
			logger.Error("failed to get code scanning analysis")
			return ew.err
		}

		logger.Info(fmt.Sprintf("found %d code scanning analysis at target in default branch (%s) after migration", len(codeScanningAnalysis), *repository.DefaultBranch))

		ew.logAndCallStep(logger, "deactivating code scanning at source", func() error {
			return orgs.sourceGC.ChangeGhasRepoSettings(ctx, orgs.source, repository, "disabled", "disabled", "disabled")
		})
	}

	if *newRepository.Archived {
		ew.logAndCallStep(logger, "archive target", func() error {
			return orgs.targetGC.ArchiveRepository(ctx, orgs.target, *repository.Name)
		})
	}

	reEnableOrigin(ctx, logger, repository, orgs.sourceGC, orgs.source, sourceWorkflows)

	//check if repository is not archived
	if !*repository.Archived {
		ew.logAndCallStep(logger, "archiving source", func() error {
			return orgs.sourceGC.ArchiveRepository(ctx, orgs.source, *repository.Name)
		})
	}

	if ew.err != nil {
		logger.Error("error: %v", ew.err)
		return ew.err
	}

	return nil
}

type errWritter struct {
	err error
}

func CheckAndMigrateSecretScanning(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) error {
	ew := errWritter{}

	ctx := context.Background()
	logger := logging.NewLoggerFromContext(ctx, false)
	sourceGC, err := github.NewGitHubClient(ctx, logger, sourceToken)
	if err != nil {
		slog.Info("error initializing source GitHub Client", err)
		return err
	}

	gei := github.NewGEI(sourceOrg, targetOrg, sourceToken, targetToken)

	var repo github.Repository
	repo, ew.err = sourceGC.GetRepository(ctx, repository, sourceOrg)

	if ew.err != nil {
		return ew.err
	}

	if *repo.SecurityAndAnalysis.SecretScanning.Status == "enabled" {
		ew.logAndCallStep(slog.Default(), "migrating secret scanning alerts", func() error {
			return gei.MigrateSecretScanning(repository)
		})
	} else {
		slog.Info("skipping because secret scanning is not enabled")
	}

	if ew.err != nil {
		return ew.err
	}

	return nil
}

func ReactivateTargetWorkflows(repository string, sourceOrg string, targetOrg string, sourceToken string, targetToken string) error {
	ew := errWritter{}

	ctx := context.Background()
	logger := logging.NewLoggerFromContext(ctx, false)
	sourceGC, err := github.NewGitHubClient(ctx, logger, sourceToken)
	if err != nil {
		slog.Info("error initializing source GitHub Client", err)
		return err
	}

	targetGC, err := github.NewGitHubClient(ctx, logger, targetToken)
	if err != nil {
		slog.Info("error initializing source GitHub Client", err)
		return err
	}

	var sourceWorkflows []github.Workflow
	sourceWorkflows, ew.err = sourceGC.GetAllActiveWorkflowsForRepository(ctx, sourceOrg, repository)

	if ew.err != nil {
		return ew.err
	}

	var targetWorkflows []github.Workflow
	targetWorkflows, ew.err = targetGC.GetAllWorkflowsForRepository(ctx, targetOrg, repository)

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

		ew.logAndCallStep(slog.Default(), "Enabling workflows at target", func() error {
			return targetGC.EnableWorkflowsForRepository(ctx, targetOrg, repository, workflows)
		})
	}

	if ew.err != nil {
		return ew.err
	}

	return nil
}

func (ew *errWritter) logAndCallStep(logger *slog.Logger, stepName string, f func() error) {
	if ew.err != nil {
		return
	}
	logger.Debug(stepName)

	for i := 0; i < maxRetries; i++ {
		ew.err = f()
		if ew.err == nil {
			break
		}

		// Exponential backoff: 2^i * 100ms
		time.Sleep(time.Duration((1<<uint(i))*1000) * time.Millisecond)
		logger.Debug(fmt.Sprintf("retrying %d/%d...", i+1, maxRetries))
	}

	if ew.err != nil {
		logger.Error("%s error: %v", stepName, ew.err)
		return
	}
	logger.Debug(fmt.Sprintf("done: %s", stepName))
}

func reEnableOrigin(ctx context.Context, logger *slog.Logger,
	repository github.Repository, sourceGC *github.GitHubClient, sourceOrg string, workflows []github.Workflow) {
	ew := errWritter{}

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil {
		ew.logAndCallStep(logger, "resetting GHAS settings at source", func() error {
			return sourceGC.ChangeGhasRepoSettings(ctx, sourceOrg, repository,
				*repository.SecurityAndAnalysis.AdvancedSecurity.Status,
				*repository.SecurityAndAnalysis.SecretScanning.Status,
				*repository.SecurityAndAnalysis.SecretScanningPushProtection.Status)
		})
	}

	if len(workflows) > 0 {
		ew.logAndCallStep(logger, "re-enabling workflows at source", func() error {
			return sourceGC.EnableWorkflowsForRepository(ctx, sourceOrg, *repository.Name, workflows)
		})
	}
}
