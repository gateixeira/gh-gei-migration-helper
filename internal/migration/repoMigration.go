package migration

type RepoMigration struct {
	name string
	orgs orgs
}

func NewRepoMigration(name, sourceOrg, targetOrg, sourceToken, targetToken string, retries int) RepoMigration {
	maxRetries = retries
	return RepoMigration{name, orgs{sourceOrg, targetOrg, sourceToken, targetToken}}
}

func (rm RepoMigration) Migrate() error {
	return nil
}
