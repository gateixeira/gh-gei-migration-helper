package migration

import (
	"context"
	"log/slog"

	"github.com/gateixeira/gei-migration-helper/internal/github"
	"github.com/gateixeira/gei-migration-helper/pkg/logging"
)

type RepoMigration struct {
	name string
	md   MigrationData
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

	return RepoMigration{name, MigrationData{orgs{sourceOrg, targetOrg, sourceGC, targetGC}, github.NewGEI(sourceOrg, targetOrg, sourceToken, targetToken)}}, nil
}

func (rm RepoMigration) Migrate(ctx context.Context) error {
	logger := logging.NewLoggerFromContext(ctx, false)

	repo, err := rm.md.orgs.sourceGC.GetRepository(ctx, rm.name, rm.md.orgs.source)

	if err != nil {
		slog.Info("error getting repository: "+*repo.Name, err)
		return err
	}

	err = rm.md.processRepoMigration(ctx, logger, repo)

	if err != nil {
		slog.Error("error migrating repository: "+*repo.Name, err)
		return err
	}

	return nil
}
