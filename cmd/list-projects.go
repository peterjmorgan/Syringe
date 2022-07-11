package cmd

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
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

		// phylumProjectList, err := s.PhylumGetProjectList()
		// if err != nil {
		//	log.Fatalf("Failed to PhylumGetProjectList(): %v\n", err)
		//	return
		// }
		// _ = phylumProjectList

		var localProjects []PhylumSyringGitlab.GitlabProject
		for _, project := range gitlabProjects {
			mainBranch, err := s.IdentifyMainBranch(project.ID)
			if err != nil {
				log.Fatalf("Failed to IdentifyMainBranch(): %v\n", err)
				return
			}

			lockfiles, ciFiles, err := s.EnumerateTargetFiles(project.ID)

			localProjects = append(localProjects, PhylumSyringGitlab.GitlabProject{
				project.ID,
				project.Name,
				mainBranch,
				false,
				false,
				lockfiles,
				ciFiles,
			})

		}
		t := table.NewWriter()
		t.SetStyle(table.StyleLight)
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Project Name", "ID", "Main Branch", "Protected", "# LockFiles"})
		for _, lp := range localProjects {
			t.AppendRow(table.Row{
				lp.Name, lp.Id, lp.Branch.Name, lp.Branch.Protected, len(lp.Lockfiles),
			})
		}
		t.Render()

	},
}
