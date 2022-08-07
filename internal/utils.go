package syringe

import (
	"fmt"
	"os"

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

func readEnvVar(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	} else {
		return "", fmt.Errorf("Failed to read environment variable: %v\n", key)
	}
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
