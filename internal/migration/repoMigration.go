package migration

import (
	"context"
	"log/slog"

	"github.com/gateixeira/gei-migration-helper/internal/github"
	"github.com/gateixeira/gei-migration-helper/pkg/logging"
)

type RepoMigration struct {
	name string
	orgs orgs
	gei  github.GEI
}

func NewRepoMigration(ctx context.Context, name, sourceOrg, targetOrg, sourceToken, targetToken string, retries int) (RepoMigration, error) {
	maxRetries = retries
	sourceGC, err := github.NewGitHubClient(ctx, slog.Default(), sourceToken)
	if err != nil {
		slog.Info("error initializing source GitHub Client", err)
		return RepoMigration{}, err
	}

	targetGC, err := github.NewGitHubClient(ctx, slog.Default(), targetToken)
	if err != nil {
		slog.Info("error initializing source GitHub Client", err)
		return RepoMigration{}, err
	}

	return RepoMigration{name, orgs{sourceOrg, targetOrg, sourceGC, targetGC}, github.NewGEI(sourceOrg, targetOrg, sourceToken, targetToken)}, nil
}

func (rm RepoMigration) Migrate() error {
	ctx := context.Background()
	logger := logging.NewLoggerFromContext(ctx, false)

	repo, err := rm.orgs.sourceGC.GetRepository(ctx, rm.name, rm.orgs.source)

	if err != nil {
		slog.Info("error getting repository: "+*repo.Name, err)
		return err
	}

	err = processRepoMigration(ctx, logger, repo, rm.orgs, rm.gei)

	if err != nil {
		slog.Error("error migrating repository: "+*repo.Name, err)
		return err
	}

	return nil
}
