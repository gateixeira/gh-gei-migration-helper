# GEI Migration Helper

This tool offers support to configurations around [GitHub's GEI](https://github.com/github/gh-gei) migration tool.

It is a collection of scripts that can be used to help with the migration process.

## Usage

Run the tool via command line:
```
$ ./gei-migration-helper --help
```

## Scripts

### `repositoryVisibility`

This script can be used to change the visibility of a repository.

#### Usage

```
$ ./gei-migration-helper repositoryVisibility --repo <repository_name> --token <pat_token> --org <org_name> --visibility <visibility>
```

The visibility can be either `public`, `private` or `internal`.

### `ghasOrgSettings`

This script can be used to change the GitHub Advanced Security settings of an organization.

#### Usage

```
$ ./gei-migration-helper ghasOrgSettings --token <pat_token> --org <org_name> --activate
```

Omit the `activate` flag to deactivate settings

### `ghasRepoSettings`

This script can be used to change the GitHub Advanced Security settings of a repository.

#### Usage

```
$ ./gei-migration-helper ghasRepoSettings --repo <repository_name> --token <pat_token> --org <org_name> --activate
```

Omit the `activate` flag to deactivate settings

### `deleteBranchProtections`

This script can be used to delete branch protections for a repository.

#### Usage

```
$ ./gei-migration-helper deleteBranchProtections --repo <repo_name> --token <pat_token> --org <org_name>
```
