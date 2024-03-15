package migration

import (
	"time"
)

type MigrationResult struct {
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
	Migrate() (MigrationResult, error)
}
