package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/spf13/cobra"
)

var (
	repository     string = "https://github.com/peterjmorgan/Syringe"
	currentVersion string = "v0.8.1"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update to latest release",
	Run: func(cmd *cobra.Command, args []string) {
		// func ConfirmAndSelfUpdate(repository string, currentVersion string) {
		latest, found, err := selfupdate.DetectLatest(repository)
		if err != nil {
			log.Println("Error occurred while detecting version:", err)
			return
		}

		fmt.Printf("Version: %s\n", currentVersion)
		v := semver.MustParse(currentVersion)
		if !found || latest.Version.Equals(v) {
			log.Println("Current version is the latest")
			return
		}

		update := false
		prompt := &survey.Confirm{
			Message: "Do you want to update to " + latest.Version.String() + "?",
		}
		survey.AskOne(prompt, &update, nil)

		if !update {
			return
		}

		cmdPath, err := os.Executable()
		if err != nil {
			log.Errorf("os executable: %v\n", err)
			return
		}

		if err := selfupdate.UpdateTo(latest.AssetURL, cmdPath); err != nil {
			log.Println("Error occurred while updating binary:", err)
			return
		}
		log.Println("Successfully updated to version", latest.Version)
	},
}
