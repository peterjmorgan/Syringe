package utils

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/peterjmorgan/Syringe/internal/structs"
	"os"
	"os/exec"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
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
		"gradle.lockfile",
		"Pipfile.lock",
		"Pipfile",
		"effective-pom.xml",
	}
}

// CI Files to target
//func GetGitlabCIFiles() []string {
//	return []string{
//		".gitlab-ci.yml",
//		".gitlab-ci.yaml",
//	}
//}

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

	var tokenAzure string
	var azureOrg string

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
			log.Debugf("SYRINGE_GITLAB_URL not configured")
		} else {
			log.Debugf("SYRINGE_GITLAB_URL: %v\n", gitlabUrl)
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
	case "azure":
		tokenAzure, err = ReadEnvVar("SYRINGE_VCS_TOKEN_AZURE")
		if err != nil {
			return nil, fmt.Errorf("failed to read 'SYRINGE_VCS_TOKEN_AZURE' from environment\n")
		}
		envMap["vcsToken"] = tokenAzure

		azureOrg, err = ReadEnvVar("SYRINGE_AZURE_ORG")
		if err != nil {
			log.Debugf("SYRINGE_AZURE_ORG not configured")
		} else {
			envMap["vcsOrg"] = azureOrg
		}
	case "bitbucket_cloud":
		tokenBitbucketCloud, err := ReadEnvVar("SYRINGE_VCS_TOKEN_BITBUCKETCLOUD")
		if err != nil {
			return nil, fmt.Errorf("failed to read 'SYRINGE_VCS_TOKEN_BITBUCKETCLOUD' from environment\n")
		}
		envMap["vcsToken"] = tokenBitbucketCloud
	default:
		log.Fatalf("ReadEnvironment(): default case. This shouldn't happen\n")
	}

	phylumToken, err := ReadEnvVar("PHYLUM_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("failed to read 'PHYLUM_API_KEY' from environment\n")
	}
	envMap["phylumToken"] = phylumToken

	phylumGroupName, err := ReadEnvVar("PHYLUM_GROUP_NAME")
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

func GeneratePhylumProjectName(projectName string, lockfilePath string, projectId int64) string {
	return fmt.Sprintf("SYR-%v__%v", projectName, lockfilePath)
}

func PromptForString(message string, lenRequirement int) (string, error) {
	prompt := promptui.Prompt{
		Label: message,
		Validate: func(input string) error {
			strLen := len(input)
			if strLen != lenRequirement && lenRequirement != -1 {
				return errors.New("invalid length")
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("PromptForString: %v failed:%v\n", message, err)
		return "", err
	}
	return result, nil
}

func ReadConfigFile(testConfigData *structs.TestConfigData) (*structs.ConfigThing, error) {
	var filename string = "syringe_config.yaml"

	v := reflect.ValueOf(testConfigData)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		if testConfigData.Filename != "" {
			filename = testConfigData.Filename
		}
	}

	if _, err := os.Stat(filename); err == nil {
		// exists
		fileData, err1 := os.ReadFile(filename)
		if err1 != nil {
			fmt.Printf("Failed to read syringe_config.yaml: %v\n", err1)
			return nil, fmt.Errorf("failed to read file")
		}

		configData := new(structs.ConfigThing)
		err2 := yaml.Unmarshal(fileData, configData)
		if err2 != nil {
			fmt.Printf("Failed to unmarshall config data: %v\n", err2)
			return nil, fmt.Errorf("failed to unmarshall config data")
		}

		return configData, nil
	}
	return nil, fmt.Errorf("config file not found")
}

func PhylumGetAuthToken() (string, error) {
	var retStr string
	var stdErrBytes bytes.Buffer

	var authTokenArgs = []string{"auth", "token"}
	authTokenCmd := exec.Command("phylum", authTokenArgs...)
	authTokenCmd.Stderr = &stdErrBytes

	retBytes, err := authTokenCmd.Output()
	if err != nil {
		log.Errorf("Failed to exec 'phylum auth token': %v\n", err)
		log.Errorf(stdErrBytes.String())
		return "", err
	}
	stdErrString := stdErrBytes.String()
	_ = stdErrString // prob will need this later

	retStr = string(retBytes)
	retStr = strings.Trim(retStr, "\n")

	return retStr, nil
}
