/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "PhylumSyringeGitlab",
	Short: "Inject Phylum goodness into your VCS",
	Long:  `Inject Phylum goodness into your VCS`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.PhylumSyringeGitlab.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug logging")
	rootCmd.PersistentFlags().BoolP("mine-only", "m", false, "Mine (owned) projects only")
	rootCmd.PersistentFlags().Int32P("ratelimit", "r", 100, "Rate Limit (X/reqs/sec) ")
	rootCmd.PersistentFlags().StringP("proxyUrl", "p", "", "proxy (https://url:port)")
}
