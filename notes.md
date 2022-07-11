# Questions
- Is Gitlab self hosted?
- How do you authenticate to your gitlab instance? 
  - Creds?
  - Oauth / OIDC?
- How do you use protected branches? 
  - Might be able to use protected values to identify the main branch


# Todo
- Create a docker image with Syringe
  - Install Phylum CLI
  - Set environment variable: GITLAB_TOKEN
  - Set environment variable: PHYLUM_API_KEY

- Handle creating groups
- Add group name parameter to run-phylum command

- run-phylum: add flag to read list of project ids (maybe from an environment variable)


# Readme
1. Setup GITLAB_TOKEN environment variable
2. Setup PHYLUM_TOKEN environment variable
3. Create a Phylum Group


# Ideas
One .phylum_project file at the root of the project
- list of projects in the form
```yaml
id: <GUID>
name: <NAME>
created_at: <TIMESTAMP>
lockfile_path: relative path to lockfile
ecosystem: might not be necessary
group_name: <GROUP>
```

We could identify the root of the git project to reference what is the top

Would have to figure out a project name scheme