package cmd

import (
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
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
		var phylumProjectMap *map[string]Syringe.PhylumProject
		var wg sync.WaitGroup

		wg.Add(2)
		go func() {
			err := s.ListProjects(&gitlabProjects)
			if err != nil {
				log.Fatalf("Failed to ListProjects(): %v\n", err)
				return
			}
			wg.Done()
		}()

		go func() {
			defer wg.Done()
			err := s.PhylumGetProjectMap(&phylumProjectMap)
			if err != nil {
				log.Fatalf("Failed to PhylumGetProjectMap(): %v\n", err)
				return
			}
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

				var NumPhylumEnabled int
				for _, lf := range lockfiles {
					generatedName := s.GeneratePhylumProjectName(inProject.Name, lf.Path)
					if slices.Contains(maps.Keys(*phylumProjectMap), generatedName) {
						NumPhylumEnabled++
					}
				}

				chProject <- Syringe.GitlabProject{
					inProject.ID,
					inProject.Name,
					mainBranch,
					NumPhylumEnabled,
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
		rowConfigAutoMerge := table.RowConfig{AutoMerge: true}
		//t.SetAutoIndex(true)
		t.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, AutoMerge: true},
			{Number: 2, AutoMerge: true},
			{Number: 3, AutoMerge: true},
			{Number: 4, AutoMerge: true},
		})
		t.SetStyle(table.StyleLight)
		t.SetOutputMirror(os.Stdout)
		// t.AppendHeader(table.Row{"Project Name", "ID", "Main Branch", "Protected", "# LockFiles", "# Phylum Enabled"})
		t.AppendHeader(table.Row{"Project Name", "ID", "Main Branch", "Protected", "Lockfile Path", "Phylum Enabled"}, rowConfigAutoMerge)
		for _, lp := range localProjects {
			//phylumEnabled := fmt.Sprintf("%v/%v", lp.NumPhylumEnabled, len(lp.Lockfiles))
			//t.AppendRow(table.Row{
			//	lp.Name, lp.Id, lp.Branch.Name, lp.Branch.Protected, len(lp.Lockfiles), phylumEnabled,
			//})
			for _, lockfile := range lp.Lockfiles {
				t.AppendRow(table.Row{lp.Name, lp.Id, lp.Branch.Name, lp.Branch.Protected, lockfile.Path, "Y"}, rowConfigAutoMerge)
			}
		}
		t.Style().Options.SeparateRows = true
		t.Render()
	},
}
