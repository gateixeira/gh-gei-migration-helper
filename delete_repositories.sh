#!/bin/sh

# Delete all repositories in a given GitHub organization

# Usage: delete_repositories.sh <organization> <token>

ORG=$1
TOKEN=$2
URL="https://api.github.com/orgs/$ORG/repos?per_page=100"

# CURL the API to get the repos
REPOS=$(curl -H "Authorization: Bearer $TOKEN" -s $URL)

REPOS_NAME=$(echo "$REPOS" | jq -r '.[].name')

# Loop over the repository names and delete each one
for repo in $REPOS_NAME
do
    echo "Deleting repository $repo"
    curl -X DELETE -H "Authorization: Bearer $TOKEN" "https://api.github.com/repos/$ORG/$repo"
done