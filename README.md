# GEI Migration Helper

This tool offers support to configurations around [GitHub's GEI](https://github.com/github/gh-gei) migration tool.

It is a collection of scripts that can be used to help with the migration process by wrapping GEI commands and performing pre and post-migration changes

## Usage

Run the tool via command line:
```
$ ./gei-migration-helper --help
```

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

### `migrate-code-scanning`

Wrapper to migrate code scan alerts. It migrates for all repositories in an org if no `--repo` is provided.

#### Usage

```
$ ./gei-migration-helper migrate-code-scanning --repo <repository_name> --source-org <source_org> --target-org <target_org> --source-token <source_token> --target-token <target_token>
```
