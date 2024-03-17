package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/gateixeira/gei-migration-helper/internal/github"
	"github.com/gateixeira/gei-migration-helper/pkg/logging"
	"github.com/gateixeira/gei-migration-helper/pkg/worker"
)

type OrgMigration struct {
	parallelMigrations int
	orgs               orgs
	gei                github.GEI
}

const statusRepoName = "migration-status"

func NewOrgMigration(ctx context.Context, source, target, sourceToken, targetToken string, retries int, parallelMigrations int) (OrgMigration, error) {
	maxRetries = retries

	sourceGC, err := github.NewGitHubClient(ctx, slog.Default(), sourceToken)
	if err != nil {
		slog.Info("error initializing source GitHub Client", err)
		return OrgMigration{}, err
	}

	targetGC, err := github.NewGitHubClient(ctx, slog.Default(), targetToken)
	if err != nil {
		slog.Info("error initializing source GitHub Client", err)
		return OrgMigration{}, err
	}

	return OrgMigration{parallelMigrations, orgs{
		source, target, sourceGC, targetGC}, github.NewGEI(source, target, sourceToken, targetToken)}, nil
}

func (om OrgMigration) checkOngoing(ctx context.Context) error {
	slog.Info("looking for ongoing/past migration")
	repo, err := om.orgs.targetGC.GetRepository(ctx, statusRepoName, om.orgs.target)

	if err != nil && err.Error() != github.ErrRepositoryNotFound.Error() {
		slog.Error("error fetching migration status repository", err)
		return err
	}

	if err == nil && *repo.Name == statusRepoName {
		issue, _ := om.orgs.targetGC.GetIssue(ctx, om.orgs.target, statusRepoName, 1)

		if issue != nil {
			err := fmt.Errorf("a migration to this organization was already executed. Please check http://github.com/%s/%s/issues/1 for status", om.orgs.target, statusRepoName)
			slog.Error(err.Error())
			return err
		}

		err := fmt.Errorf("a migration to this organization is either ongoing or finished in error (remove http://github.com/%s/%s if you want to retry)", om.orgs.target, statusRepoName)
		return err
	}

	slog.Info("creating migration status repository")
	om.orgs.targetGC.CreateRepository(ctx, om.orgs.target, statusRepoName)

	return nil
}

func (om OrgMigration) prepareMigration(ctx context.Context) error {
	if err := om.checkOngoing(ctx); err != nil {
		return err
	}

	slog.Info("deactivating GHAS settings at target organization")
	om.orgs.targetGC.ChangeGHASOrgSettings(ctx, om.orgs.target, false)

	return nil
}

func (om OrgMigration) Process(repo interface{}, ctx context.Context) error {
	repository, ok := repo.(github.Repository)
	if !ok {
		return fmt.Errorf("could not cast repository to github.Repository")
	}

	var repoSummary = repoStatus{
		Name:     *repository.Name,
		ID:       *repository.ID,
		Archived: *repository.Archived,
	}

	if repository.SecurityAndAnalysis.AdvancedSecurity != nil {
		repoSummary.CodeScanning = *repository.SecurityAndAnalysis.AdvancedSecurity.Status
		repoSummary.SecretScanning = *repository.SecurityAndAnalysis.SecretScanning.Status
		repoSummary.PushProtection = *repository.SecurityAndAnalysis.SecretScanningPushProtection.Status
	}

	slog.Info("starting migration", "name", *repository.Name)
	logger := logging.NewLoggerFromContext(ctx, false)
	err := processRepoMigration(ctx, logger, repository, om.orgs, om.gei)
	slog.Info("finished migrating", "name", *repository.Name)
	if err != nil {
		slog.Error("error migrating repository: ", err)
		return err
	}

	return nil
}

func (om OrgMigration) Migrate() (migrationResult, error) {
	ctx := context.Background()

	if err := om.prepareMigration(ctx); err != nil {
		return migrationResult{}, err
	}

	slog.Info("fetching repositories from source organization")
	sourceRepositories, err := om.orgs.sourceGC.GetRepositories(ctx, om.orgs.source)

	if err != nil {
		slog.Error("error fetching repositories from source organization")
		return migrationResult{}, err
	}

	destinationRepositories, err := om.orgs.targetGC.GetRepositories(ctx, om.orgs.target)

	if err != nil {
		slog.Error("error fetching repositories from target organization")
		return migrationResult{}, err
	}

	m := make(map[string]bool)
	for _, item := range destinationRepositories {
		m[*item.Name] = true
	}
	var sourceRepositoriesToMigrate []github.Repository
	for _, item := range sourceRepositories {
		if _, ok := m[*item.Name]; !ok {
			sourceRepositoriesToMigrate = append(sourceRepositoriesToMigrate, item)
		} else {
			slog.Info("repository " + *item.Name + " already exists at target organization")
		}
	}

	slog.Info(strconv.Itoa(len(sourceRepositoriesToMigrate)) + " repositories to migrate")

	jobs := make(chan interface{}, len(sourceRepositoriesToMigrate))
	results := make(chan worker.Error, len(sourceRepositoriesToMigrate))

	w, _ := worker.New(om.Process, jobs, results)

	for i := 1; i <= om.parallelMigrations; i++ {
		go w.Start(context.WithValue(ctx, logging.IDKey, i))
	}

	for _, repository := range sourceRepositoriesToMigrate {
		jobs <- repository
	}
	close(jobs)

	var failed []repoStatus
	var migrated []repoStatus
	for a := 1; a <= len(sourceRepositoriesToMigrate); a++ {
		workerResult := <-results
		slog.Debug("result received")
		entity := workerResult.Entity.(github.Repository)
		if workerResult.Err != nil {
			failed = append(failed, repoStatus{
				Name: *entity.Name,
				ID:   *entity.ID,
			})
		} else {
			migrated = append(migrated, repoStatus{
				Name: *entity.Name,
				ID:   *entity.ID,
			})
		}
	}

	mr := migrationResult{
		Timestamp: time.Now().UTC(),
		SourceOrg: om.orgs.source,
		TargetOrg: om.orgs.target,
		Migrated:  migrated,
		Failed:    failed,
	}

	jsonData, err := json.MarshalIndent(mr, "", "  ")
	if err != nil {
		slog.Error("failed to parse result", err)
		return migrationResult{}, err
	}

	err = om.orgs.targetGC.CreateIssue(ctx, om.orgs.target, statusRepoName, "Migration result", string(jsonData))

	if err != nil {
		slog.Error("error creating issue with migration result. Check migration-result.json for details")
		return migrationResult{}, err
	}

	return mr, nil
}
