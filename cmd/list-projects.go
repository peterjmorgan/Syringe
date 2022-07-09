package cmd

import (
	PhylumSyringGitlab "github.com/peterjmorgan/PhylumSyringeGitlab/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listProjectsCmd)
}

var listProjectsCmd = &cobra.Command{
	Use:   "list-projects",
	Short: "List Gitlab Projects",
	Run: func(cmd *cobra.Command, args []string) {
		s, err := PhylumSyringGitlab.NewSyringe()
		if err != nil {
			log.Fatal("Failed to create NewSyringe(): %v\n", err)
			return
		}
		gitlabProjects, err := s.ListProjects()
		if err != nil {
			log.Fatalf("Failed to ListProjects(): %v\n", err)
			return
		}

		var localProjects []PhylumSyringGitlab.GitlabProject
		for _, project := range gitlabProjects {
			mainBranch, err := s.IdentifyMainBranch(project.ID)
			if err != nil {
				log.Fatalf("Failed to IdentifyMainBranch(): %v\n", err)
				return
			}

			lockfiles, ciFiles, err := s.EnumerateTargetFiles(project.ID)

		}
	},
}
