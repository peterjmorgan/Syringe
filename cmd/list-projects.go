package cmd

import (
	"os"
	"sync"

	"github.com/jedib0t/go-pretty/v6/table"
	Syringe "github.com/peterjmorgan/Syringe/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

func init() {
	rootCmd.AddCommand(listProjectsCmd)
}

var listProjectsCmd = &cobra.Command{
	Use:   "list-projects",
	Short: "List Gitlab Projects",
	Run: func(cmd *cobra.Command, args []string) {
		s, err := Syringe.NewSyringe()
		if err != nil {
			log.Fatal("Failed to create NewSyringe(): %v\n", err)
			return
		}
		var gitlabProjects *[]*gitlab.Project
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			err := s.ListProjects(&gitlabProjects)
			if err != nil {
				log.Fatalf("Failed to ListProjects(): %v\n", err)
				return
			}
			wg.Done()
		}()
		wg.Wait()

		var localProjects []Syringe.GitlabProject
		chProject := make(chan Syringe.GitlabProject)
		var wgLoop sync.WaitGroup

		go func() {
			wgLoop.Wait()
			close(chProject)
		}()

		for _, project := range *gitlabProjects {
			wgLoop.Add(1)
			go func(inProject gitlab.Project) {
				defer wgLoop.Done()
				mainBranch, err := s.IdentifyMainBranch(inProject.ID)
				if err != nil {
					log.Fatalf("Failed to IdentifyMainBranch(): %v\n", err)
					return
				}

				lockfiles, ciFiles, err := s.EnumerateTargetFiles(inProject.ID)

				// for _, lf := range lockfiles {
				//
				// }

				chProject <- Syringe.GitlabProject{
					inProject.ID,
					inProject.Name,
					mainBranch,
					false,
					false,
					lockfiles,
					ciFiles,
				}
			}(*project)
		}

		for item := range chProject {
			localProjects = append(localProjects, item)
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
