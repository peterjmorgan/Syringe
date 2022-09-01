package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/peterjmorgan/Syringe/internal/structs"
	"github.com/peterjmorgan/Syringe/internal/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"strings"
)

func init() {
	rootCmd.AddCommand(configureCmd)
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "configure",
	Run: func(cmd *cobra.Command, args []string) {

		// TODO: Check for existing configuration
		configData, err := utils.ReadConfigFile()
		if err == nil {
			configJson, err3 := json.MarshalIndent(configData, "", "  ")
			if err3 != nil {
				fmt.Printf("Failed to unmarshall config data: %v\n", err3)
				return
			}

			fmt.Printf("Found existing config:\n %s\n", string(configJson))
			reconfigurePrompt := promptui.Prompt{
				Label:     "Would you like to enter new configuration data",
				Default:   "n",
				IsConfirm: true,
			}

			shouldReconfigure, err := reconfigurePrompt.Run()
			if err != nil && err.Error() != "" {
				fmt.Printf(err.Error())
				return
			}
			shouldReconfigure = strings.ToLower(shouldReconfigure)
			if shouldReconfigure == "" || shouldReconfigure == "n" {
				fmt.Printf("No configuration change.\n")
				os.Exit(0)
			}
			// shouldReconfigure == "y" will fall through
		} else {
			// doesn't exist
			fmt.Printf("Could not find 'syringe_config.yaml', let's create one")
		}

		// Configure VCS type
		vcsPrompt := promptui.Select{
			Label: "Choose VCS system",
			Items: []string{"Github.com", "Gitlab.com", "Azure Devops (cloud)", "Bitbucket.com"},
		}

		_, vcsResult, err := vcsPrompt.Run()
		if err != nil {
			fmt.Printf("vcsPrompt failed: %v\n", err)
		}

		config := make(map[string]string, 0)

		ct := structs.ConfigThing{}
		ct.Associated = make(map[string]string, 0)

		switch vcsResult {
		case "Github.com":
			ghToken, err := utils.PromptForString("Enter Github token", 40)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
			ghOrg, err := utils.PromptForString("Enter Github Organization name", -1)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
			config["SYRINGE_VCS"] = "github"
			config["SYRINGE_VCS_TOKEN_GITHUB"] = ghToken
			config["SYRINGE_GITHUB_ORG"] = ghOrg
			ct.VcsType = "github"
			ct.VcsToken = ghToken
			ct.Associated["githubOrg"] = ghOrg

		case "Gitlab.com":
			gitlabToken, err := utils.PromptForString("Enter Gitlab token", 20)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}

			gitlabUrl, err := utils.PromptForString("Enter Gitlab URL", -1)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
			ct.VcsType = "gitlab"
			ct.VcsToken = gitlabToken
			ct.Associated["gitlabUrl"] = gitlabUrl

		case "Azure Devops (cloud)":
			azureToken, err := utils.PromptForString("Enter Azure DevOps token", 52)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
			azureOrg, err := utils.PromptForString("Enter Azure DevOps organization url", -1)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
			ct.VcsType = "azure"
			ct.VcsToken = azureToken
			ct.Associated["azureOrg"] = azureOrg

		case "Bitbucket.com":
			bbOwner, err := utils.PromptForString("Enter Bitbucket Cloud Username", -1)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
			bbClientId, err := utils.PromptForString("Enter Bitbucket Cloud Oauth Client Credential ClientID", -1)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
			bbClientSecret, err := utils.PromptForString("Enter Bitbucket Cloud Oauth Client Credential ClientSecret", -1)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
			ct.VcsType = "bitbucket_cloud"
			ct.Associated["bbOwner"] = bbOwner
			ct.Associated["bbClientId"] = bbClientId
			ct.Associated["bbClientSecret"] = bbClientSecret

		default:
			fmt.Printf("vcsType switch default case. Shouldn't happen\n")
			return
		}

		phylumConfigurePrompt := promptui.Prompt{
			Label:     "Syringe can try to detect your Phylum token. Would you like to OVERRIDE and set it manually",
			Default:   "n",
			IsConfirm: true,
		}

		manuallyConfigPhylum, err := phylumConfigurePrompt.Run()
		if err != nil && err.Error() != "" {
			fmt.Printf(err.Error())
			return
		}
		var phylumToken string
		if strings.ToLower(manuallyConfigPhylum) == "y" {
			phylumToken, err = utils.PromptForString("Enter Phylum Token (result of `phylum auth token`)", -1)
			if err != nil {
				fmt.Printf(err.Error())
				return
			}
		} else {
			//phylumToken = ""
			// TODO: pull from phylum auth token
			phylumToken, err = utils.PhylumGetAuthToken()
			if err != nil {
				fmt.Printf("Failed to read phylum auth token from locally-installed 'phylum'")
				return
			}
			fmt.Printf("Found phylum token from locally-installed 'phylum':\n%v\n", phylumToken)
		}
		config["PHYLUM_API_KEY"] = phylumToken
		ct.PhylumToken = phylumToken

		phylumGroup, err := utils.PromptForString("Enter Phylum group for submitted projects", -1)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		//TODO: check if group exists, or create it
		config["PHYLUM_GROUP_NAME"] = phylumGroup
		ct.PhylumGroup = phylumGroup

		yamlData, err := yaml.Marshal(ct)
		if err != nil {
			fmt.Printf("Failed to marshall config data: %v\n", err)
			return
		}

		err = ioutil.WriteFile("syringe_config.yaml", yamlData, 0644)
		if err != nil {
			fmt.Printf("Failed to write config file: %v\n", err)
			return
		}

		fmt.Printf("Finished configuring Syringe. Wrote configuration to 'syringe_config.yaml`\n")
	},
}
