package migration

import (
	"context"
	"log/slog"

	"github.com/gateixeira/gei-migration-helper/internal/github"
	"github.com/gateixeira/gei-migration-helper/pkg/logging"
)

type SecretScanningMigration struct {
	md MigrationData
}

func NewSecretScanningMigration(ctx context.Context, sourceOrg, targetOrg, sourceToken, targetToken string) (SecretScanningMigration, error) {
	sourceGC, err := github.NewGitHubClient(ctx, slog.Default(), sourceToken)
	if err != nil {
		slog.Info("error initializing source GitHub Client", err)
		return SecretScanningMigration{}, err
	}

	targetGC, err := github.NewGitHubClient(ctx, slog.Default(), sourceToken)
	if err != nil {
		slog.Info("error initializing source GitHub Client", err)
		return SecretScanningMigration{}, err
	}

	return SecretScanningMigration{MigrationData{orgs{sourceOrg, targetOrg, sourceGC, targetGC}, github.NewGEI(sourceOrg, targetOrg, sourceToken, targetToken)}}, nil
}

func (scm SecretScanningMigration) Migrate(ctx context.Context, repository string) error {
	logger := logging.NewLoggerFromContext(ctx, false)

	var repositories []github.Repository
	if repository == "" {
		slog.Info("fetching repositories from source organization")

		repositories, _ = scm.md.orgs.sourceGC.GetRepositories(ctx, scm.md.orgs.source)
	} else {
		repo, err := scm.md.orgs.sourceGC.GetRepository(ctx, repository, scm.md.orgs.source)

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
		err := scm.md.CheckAndMigrateSecretScanning(ctx, logger, repository)

		if err != nil {
			slog.Error("error migrating secret scanning for repository: " + *repository.Name)
			continue
		}
	}

	return nil
}
