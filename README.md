# GEI Migration Helper

GitHub CLI Extension.

This is a wrapper tool to GitHub Enterprise Importer that orchestrate necessary steps between [GitHub's GEI](https://github.com/github/gh-gei) repository migration and GHAS secret and code scanning migrations.

It is a collection of scripts that can be used to help with the migration process by wrapping GEI commands and performing pre and post-migration changes

## Installation

the GEI Migration Helper can be installed via this command:

```
$ gh extension install gateixeira/gh-gei-migration-helper
```

## Usage

Run the tool via command line:

```
$ gh gei-migration-helper --help
```

## Migration process

Read all repositories from source organization and for each repository:

1. Check if code scanning analysis exist at source in default branch
    - Activate code scanning at source if not already activated
    - 1.2 Check if code scanning analysis exist at source
2. Disable GHAS at source
3. Disable workflows at source
4. Migrate repository
5. Disable workflows at target (they get re-enabled after a migration)
6. Check if target repository is archived
    - 6.1 Unarchive target repository
7. Delete branch protections at target
8. Check if target repository is private
    - 8.1 Change visibility of target repository to internal 
9. Activate GHAS at target
10. If source repository has code scanning analysis
    - 10.1 Activate code scanning at source
    - 10.2 Migrate code scanning alerts
    - 10.3 Deactivate code Scanning at source
11. Check if target repository is archived
    - 11.1 Archive target repository
12. Reset origin
    - 12.1 Reset GHAS settings at source
    - 12.2 Reset workflows at source
13. Check if source repository is archived
    - 13.1 Archive source repository

## Manual steps to execute a migration

1. Download the [GitHub CLI](https://cli.github.com/)
2. Install the [GEI extension](https://github.com/github/gh-gei)
3. Create a [personal access token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token) for the source and target organization according to [these scopes](https://docs.github.com/en/enterprise-cloud@latest/migrations/using-github-enterprise-importer/preparing-to-migrate-with-github-enterprise-importer/managing-access-for-github-enterprise-importer#personal-access-tokens-for-github-products)
4. Run the migration helper scripts
5. `migrate-organization` to migrate all repositories in an organization
6. Wait for secret scanning to execute on the target organization
7. `migrate-secret-scanning` to migrate secret scanning results
8. `reactivate-target-workflow` to reactivate workflows at target that were deactivated during the migration process

## Scripts

### `migrate-organization`

This script can be used to migrate all repositories in an organization

#### Usage

```
$ gh gh-gei-migration-helper migrate-organization --source-org <source_org> --target-org <target_org> --source-token <source_token> --target-token <target_token>
```

### `migrate-repository`

This script can be used to migrate a single repository

#### Usage

```
$ gh gh-gei-migration-helper migrate-repository --repo <repository_name> --source-org <source_org> --target-org <target_org> --source-token <source_token> --target-token <target_token>
```

### `migrate-secret-scanning`

Wrapper to migrate secret scan results. It migrates for all repositories in an org if no `--repo` is provided.

#### Usage

```
$ gh gh-gei-migration-helper migrate-secret-scanning --repo <repository_name> --source-org <source_org> --target-org <target_org> --source-token <source_token> --target-token <target_token>
```

### `reactivate-target-workflow`

Resets the target repository workflows to their original state. It reactivates all workflows that were deactivated during the migration process.

Omit the repository flag to run against the whole organization.

#### Usage

```
$ gh gh-gei-migration-helper reactivate-target-workflow --source-org <source_org> --target-org <target_org> --source-token <source_token> --target-token <target_token>
```
