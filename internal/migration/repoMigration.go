package migration

type RepoMigration struct {
	name                 string
	sourceOrg, targetOrg string
}

func NewRepoMigration(name, sourceOrg, targetOrg string) RepoMigration {
	return RepoMigration{name, sourceOrg, targetOrg}
}

func (rm RepoMigration) Migrate() error {
	return nil
}
