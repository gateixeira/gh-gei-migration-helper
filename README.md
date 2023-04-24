# GEI Migration Helper

This is a wrapper tool to GitHub Enterprise Importer that orchestrate necessary steps between [GitHub's GEI](https://github.com/github/gh-gei) repository migration and GHAS secret and code scanning migrations.

It is a collection of scripts that can be used to help with the migration process by wrapping GEI commands and performing pre and post-migration changes

## Usage

Run the tool via command line:
```
$ ./gei-migration-helper --help
```

## Migration process

Read all repositories from source organization and for each repository:

1. Deactivate GitHub Advanced Security features at source repository if enabled
2. Disable workflows at source repository, if any
3. Migrate repository to target organization
4. Delete branch protections at target (To be removed in a later version)
5. If repository is internal at source it should be reset to internal at target after migration
6. Activate GitHub Advanced Security at target
7. Reactivate GitHub Advanced Security at source
8. Migrate code scanning alerts
9. Re-enable workflows at source, if any
10. Archive source repository


## Manual steps to execute a migration

1. Download the [GitHub CLI](https://cli.github.com/)
2. Install the [GEI extension](https://github.com/github/gh-gei)
3. Create a [personal access token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token) for the source and target organization according to [these scopes](https://docs.github.com/en/enterprise-cloud@latest/migrations/using-github-enterprise-importer/preparing-to-migrate-with-github-enterprise-importer/managing-access-for-github-enterprise-importer#personal-access-tokens-for-github-products)
4. Run the migration helper scripts
  1. `migrate-organization` to migrate all repositories in an organization
  2. Wait for secret scanning to execute on the target organization
  3. `migrate-secret-scanning` to migrate secret scanning results
  4. `reactivate-target-workflow` to reactivate workflows at target that were deactivated during the migration process

## Scripts

### `migrate-organization`

This script can be used to migrate all repositories in an organization

#### Usage

```
$ ./gei-migration-helper migrate-organization --source-org <source_org> --target-org <target_org> --source-token <source_token> --target-token <target_token>
```

### `migrate-repository`

This script can be used to migrate a single repository

#### Usage

```
$ ./gei-migration-helper migrate-repository --repo <repository_name> --source-org <source_org> --target-org <target_org> --source-token <source_token> --target-token <target_token>
```

### `migrate-secret-scanning`

Wrapper to migrate secret scan results. It migrates for all repositories in an org if no `--repo` is provided.

#### Usage

```
$ ./gei-migration-helper migrate-secret-scanning --repo <repository_name> --source-org <source_org> --target-org <target_org> --source-token <source_token> --target-token <target_token>
```

### `reactivate-target-workflow`

Resets the target repository workflows to their original state. It reactivates all workflows that were deactivated during the migration process. 

Omit the repository flag to run against the whole organization.

#### Usage

```
$ ./gei-migration-helper reactivate-target-workflow --source-org <source_org> --target-org <target_org> --source-token <source_token> --target-token <target_token>
```
