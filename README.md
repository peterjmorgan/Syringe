# Syringe

A tool to automate the submission of Gitlab and Github projects to Phylum

# Configuration

Syringe expects several environment variables to be properly configured:
* `SYRINGE_VCS`: "github" | "gitlab" | "azure"
* `PHYLUM_API_KEY`: A token to access the Phylum API
* `PHYLUM_GROUP_NAME`: The name of the Phylum Group to which Syringe project submissions will be correlated. This group must exist before it can be used.

To configure for Gitlab, ensure the following environment variables are properly configured:
* `SYRINGE_VCS_TOKEN_GITLAB`: A token to access the Gitlab API
* `SYRINGE_GITLAB_URL`: The fully-qualified domain name of the GitLab server. Defaults to `https://gitlab.com`
 
To configure for Github, ensure the following environment variables are properly configured:
* `SYRINGE_VCS_TOKEN_GITHUB`: A token to access the Github API
* `SYRINGE_GITHUB_URL`: The fully-qualified domain name of the Github server. Defaults to `https://github.com`

To configure for Azure Devops, ensure the following environment variables are properly configured:
* `SYRINGE_VCS_TOKEN_AZURE`: A token to access the Azure Dev Ops API
* `SYRINGE_AZURE_ORG`: The fully-qualified domain name of the Azure Dev Ops organization. Example: https://dev.azure.com/pete0372

# Quickstart

1. Ensure Phylum is installed and configured
2. Checkout this repository
3. Build `Syringe` with: `go build -o Syringe`
4. Configure the environment variables listed above
5. Examine the subcommands for `Syringe` by running it
6. Execute `Syringe list-projects` to list the projects Syringe can see with the token and configuration provided.
7. Execute `Syringe run-phylum` to submit the identified projects to Phylum for viewing the [Phylum Web UI](https://app.phylum.io)
