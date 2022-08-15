package utils

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Branches to examine for "main" branch
func GetMainBranchSlice() []string {
	return []string{
		"master",
		"main",
	}
}

// Lockfiles to target
func GetSupportedLockfiles() []string {
	return []string{
		"package-lock.json",
		"yarn.lock",
		"requirements.txt",
		"poetry.lock",
		"pom.xml",
		"Gemfile.lock",
	}
}

// CI Files to target
func GetGitlabCIFiles() []string {
	return []string{
		".gitlab-ci.yml",
		".gitlab-ci.yaml",
	}
}

func ReadEnvVar(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	} else {
		return "", fmt.Errorf("Failed to read environment variable: %v\n", key)
	}
}

func ReadEnvironment() (map[string]string, error) {
	var tokenGitlab string
	var gitlabUrl string

	var tokenGithub string
	var githubUrl string
	var githubOrg string

	envMap := make(map[string]string, 5)

	vcsType, err := ReadEnvVar("SYRINGE_VCS")
	if err != nil {
		return nil, fmt.Errorf("failed to read 'SYRINGE_VCS' from ENV\n")
	}
	vcsType = strings.ToLower(vcsType)
	envMap["vcs"] = vcsType
	log.Debugf("SYRINGE_VCS: %v selected\n", vcsType)

	switch vcsType {
	case "gitlab":
		tokenGitlab, err = ReadEnvVar("SYRINGE_VCS_TOKEN_GITLAB")
		if err != nil {
			return nil, fmt.Errorf("failed to read 'SYRINGE_VCS_TOKEN_GITLAB'\n")
		}
		envMap["vcsToken"] = tokenGitlab

		gitlabUrl, err = ReadEnvVar("SYRINGE_GITLAB_URL")
		if err != nil {
			log.Debugf("GITLAB_URL not configured")
		} else {
			envMap["vcsUrl"] = gitlabUrl
		}

	case "github":
		tokenGithub, err = ReadEnvVar("SYRINGE_VCS_TOKEN_GITHUB")
		if err != nil {
			return nil, fmt.Errorf("failed to read 'SYRINGE_VCS_TOKEN_GITHUB' from environment\n")
		}
		envMap["vcsToken"] = tokenGithub

		githubUrl, err = ReadEnvVar("SYRINGE_GITHUB_URL")
		if err != nil {
			log.Debugf("GITHUB_URL not configured")
		} else {
			envMap["vcsUrl"] = githubUrl
		}

		githubOrg, err = ReadEnvVar("SYRINGE_GITHUB_ORG")
		if err != nil {
			log.Debugf("GITHUB_ORG not configured")
		} else {
			envMap["vcsOrg"] = githubOrg
		}
	default:
		log.Fatalf("ReadEnvironment(): default case. This shouldn't happen\n")
	}

	phylumToken, err := ReadEnvVar("PHYLUM_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("failed to read 'PHYLUM_API_KEY' from environment\n")
	}
	envMap["phylumToken"] = phylumToken

	phylumGroupName, err := ReadEnvVar("PHYLUM_GROUP")
	if err != nil {
		log.Debugf("PHYLUM_GROUP not configured\n")
	} else {
		envMap["phylumGroup"] = phylumGroupName
	}

	return envMap, nil
}

func RemoveTempDir(tempDir string) {
	err := os.RemoveAll(tempDir)
	if err != nil {
		log.Errorf("Failed to remove temp directory %v:%v\n", tempDir, err)
	}
}

func GeneratePhylumProjectName(projectName string, lockfilePath string) string {
	return fmt.Sprintf("SYR-%v__%v", projectName, lockfilePath)
}
