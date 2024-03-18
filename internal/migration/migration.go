package migration

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/gateixeira/gei-migration-helper/internal/github"
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

type MigrationData struct {
	orgs orgs
	gei  github.GEI
}

type Migration interface {
	Migrate(context.Context) (migrationResult, error)
}

type errWritter struct {
	err error
}

var maxRetries = 5

func NewMigration(ctx context.Context, sourceOrg, targetOrg, sourceToken, targetToken string) (MigrationData, error) {
	sourceGC, err := github.NewGitHubClient(ctx, slog.Default(), sourceToken)
	if err != nil {
		slog.Info("error initializing source GitHub Client", err)
		return MigrationData{}, err
	}

	targetGC, err := github.NewGitHubClient(ctx, slog.Default(), targetToken)
	if err != nil {
		slog.Info("error initializing source GitHub Client", err)
		return MigrationData{}, err
	}

	return MigrationData{orgs{sourceOrg, targetOrg, sourceGC, targetGC}, github.NewGEI(sourceOrg, targetOrg, sourceToken, targetToken)}, nil
}

func (md MigrationData) processRepoMigration(ctx context.Context, logger *slog.Logger, repository github.Repository) error {
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
				return md.orgs.sourceGC.UnarchiveRepository(ctx, md.orgs.source, *repository.Name)
			})
		}
		ew.logAndCallStep(logger, "activating code scanning at source to check for previous analyses", func() error {
			return md.orgs.sourceGC.ChangeGhasRepoSettings(ctx, md.orgs.source, repository, "enabled", "disabled", "disabled")
		})
	}

	codeScanningAnalysis, _ := md.orgs.sourceGC.GetCodeScanningAnalysis(ctx, md.orgs.source, *repository.Name, *repository.DefaultBranch)

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil {
		ew.logAndCallStep(logger, "disabling GHAS settings at source", func() error {
			return md.orgs.sourceGC.ChangeGhasRepoSettings(ctx, md.orgs.source, repository, "disabled", "disabled", "disabled")
		})
	}

	var sourceWorkflows []github.Workflow
	sourceWorkflows, ew.err = md.orgs.sourceGC.GetAllActiveWorkflowsForRepository(ctx, md.orgs.source, *repository.Name)

	if ew.err != nil {
		logger.Error("failed to get workflows")
		return ew.err
	}

	if len(sourceWorkflows) > 0 {
		ew.logAndCallStep(logger, "disabling workflows at source", func() error {
			return md.orgs.sourceGC.DisableWorkflowsForRepository(ctx, md.orgs.source, *repository.Name, sourceWorkflows)
		})
	}

	ew.logAndCallStep(logger, "migrating", func() error {
		return md.gei.MigrateRepo(*repository.Name)
	})

	newRepository, err := md.orgs.targetGC.GetRepository(ctx, *repository.Name, md.orgs.target)

	if err != nil {
		logger.Error("failed to migrate")
		return err
	}

	var targetWorkflows []github.Workflow
	targetWorkflows, ew.err = md.orgs.targetGC.GetAllActiveWorkflowsForRepository(ctx, md.orgs.target, *repository.Name)

	if ew.err != nil {
		logger.Error("failed to get workflows")
		return ew.err
	}

	if len(targetWorkflows) > 0 {
		//this is unfortunately necessary as the workflows get re-enabled after org migration
		ew.logAndCallStep(logger, "disabling workflows at target", func() error {
			return md.orgs.targetGC.DisableWorkflowsForRepository(ctx, md.orgs.target, *repository.Name, targetWorkflows)
		})
	}

	if *newRepository.Archived {
		ew.logAndCallStep(logger, "unarchive target", func() error {
			return md.orgs.targetGC.UnarchiveRepository(ctx, md.orgs.target, *repository.Name)
		})
	}

	ew.logAndCallStep(logger, "deleting branch protections at target", func() error {
		return md.orgs.targetGC.DeleteBranchProtections(ctx, md.orgs.target, *repository.Name)
	})

	if *newRepository.Visibility == "private" {
		ew.logAndCallStep(logger, "changing visibility to internal at target", func() error {
			return md.orgs.targetGC.ChangeRepositoryVisibility(ctx, md.orgs.target, *repository.Name, "internal")
		})

		logger.Debug("waiting 10 seconds for changes to apply...")
		time.Sleep(10 * time.Second)
	} else {
		logger.Info("skipping visibility change because source is already internal or public")
	}

	ew.logAndCallStep(logger, "activating GHAS at target", func() error {
		return md.orgs.targetGC.ChangeGhasRepoSettings(ctx, md.orgs.target, newRepository, "enabled", "enabled", "enabled")
	})

	if len(codeScanningAnalysis) <= 0 {
		logger.Info("no code scan to migrate, skipping.")
	} else {
		logger.Info(fmt.Sprintf("found %d code scanning analysis at source in default branch (%s) before migration", len(codeScanningAnalysis), *repository.DefaultBranch))

		ew.logAndCallStep(logger, "activating code scanning at source to migrate alerts", func() error {
			return md.orgs.sourceGC.ChangeGhasRepoSettings(ctx, md.orgs.source, repository, "enabled", "disabled", "disabled")
		})

		ew.logAndCallStep(logger, "migrating code scanning alerts", func() error {
			return md.gei.MigrateCodeScanning(*repository.Name)
		})

		codeScanningAnalysis, ew.err = md.orgs.targetGC.GetCodeScanningAnalysis(ctx, md.orgs.target, *repository.Name, *repository.DefaultBranch)

		if ew.err != nil {
			logger.Error("failed to get code scanning analysis")
			return ew.err
		}

		logger.Info(fmt.Sprintf("found %d code scanning analysis at target in default branch (%s) after migration", len(codeScanningAnalysis), *repository.DefaultBranch))

		ew.logAndCallStep(logger, "deactivating code scanning at source", func() error {
			return md.orgs.sourceGC.ChangeGhasRepoSettings(ctx, md.orgs.source, repository, "disabled", "disabled", "disabled")
		})
	}

	if *newRepository.Archived {
		ew.logAndCallStep(logger, "archive target", func() error {
			return md.orgs.targetGC.ArchiveRepository(ctx, md.orgs.target, *repository.Name)
		})
	}

	reEnableOrigin(ctx, logger, repository, md.orgs.sourceGC, md.orgs.source, sourceWorkflows)

	//check if repository is not archived
	if !*repository.Archived {
		ew.logAndCallStep(logger, "archiving source", func() error {
			return md.orgs.sourceGC.ArchiveRepository(ctx, md.orgs.source, *repository.Name)
		})
	}

	if ew.err != nil {
		logger.Error("error: %v", ew.err)
		return ew.err
	}

	return nil
}

func (md MigrationData) CheckAndMigrateSecretScanning(ctx context.Context, logger *slog.Logger, repository github.Repository) error {
	ew := errWritter{}

	if *repository.SecurityAndAnalysis.SecretScanning.Status == "enabled" {
		ew.logAndCallStep(slog.Default(), "migrating secret scanning alerts", func() error {
			return md.gei.MigrateSecretScanning(*repository.Name)
		})
	} else {
		slog.Info("skipping because secret scanning is not enabled")
	}

	if ew.err != nil {
		return ew.err
	}

	return nil
}

func (md MigrationData) ReactivateTargetWorkflows(ctx context.Context, repository string) error {
	var repositories []github.Repository
	if repository == "" {
		slog.Info("fetching repositories from source organization")

		repositories, _ = md.orgs.sourceGC.GetRepositories(ctx, md.orgs.source)
	} else {
		repo, err := md.orgs.sourceGC.GetRepository(ctx, repository, md.orgs.source)

		if err != nil {
			slog.Error("error getting repository: "+repository, err)
			return err
		}

		repositories = append(repositories, repo)
	}

	for _, repository := range repositories {
		if *repository.Name == ".github" {
			continue
		}

		ew := errWritter{}

		var sourceWorkflows []github.Workflow
		sourceWorkflows, ew.err = md.orgs.sourceGC.GetAllActiveWorkflowsForRepository(ctx, md.orgs.source, *repository.Name)

		if ew.err != nil {
			return ew.err
		}

		var targetWorkflows []github.Workflow
		targetWorkflows, ew.err = md.orgs.targetGC.GetAllWorkflowsForRepository(ctx, md.orgs.target, *repository.Name)

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
				return md.orgs.targetGC.EnableWorkflowsForRepository(ctx, md.orgs.target, *repository.Name, workflows)
			})
		}

		if ew.err != nil {
			return ew.err
		}
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
