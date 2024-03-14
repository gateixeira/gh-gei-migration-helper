package migration

import (
	"time"
)

type MigrationResult struct {
	Timestamp time.Time `json:"timestamp"`
	SourceOrg string    `json:"sourceOrg"`
	TargetOrg string    `json:"targetOrg"`
	Migrated  []repo    `json:"migrated"`
	Failed    []repo    `json:"failed"`
}

type repo struct {
	Name           string `json:"name"`
	ID             int64  `json:"id"`
	Archived       bool   `json:"archived"`
	CodeScanning   string `json:"codeScanning" default:"disabled"`
	SecretScanning string `json:"secretScanning" default:"disabled"`
	PushProtection string `json:"pushProtection" default:"disabled"`
}

type migration interface {
	migrate() error
}
