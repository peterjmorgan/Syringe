package cmd

import (
	"sync"

	Syringe "github.com/peterjmorgan/Syringe/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func init() {
	rootCmd.AddCommand(runPhylumCmd)
}

var runPhylumCmd = &cobra.Command{
	Use:   "run-phylum",
	Short: "Run Phylum on GitLab Projects",
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
			defer wg.Done()
			err := s.ListProjects(&gitlabProjects)
			if err != nil {
				log.Fatalf("Failed to ListProjects(): %v\n", err)
				return
			}
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

		// Enumerate list of Gitlab projects that do not have an associated Phylum project.
		chCreateProjects := make(chan string)
		var wgLoop sync.WaitGroup

		go func() {
			wgLoop.Wait()
			close(chCreateProjects)
		}()
		for _, project := range *gitlabProjects {
			wgLoop.Add(1)
			go func(inProject gitlab.Project) {
				defer wgLoop.Done()

				lockfiles, _, err := s.EnumerateTargetFiles(inProject.ID)
				if err != nil {
					log.Errorf("Failed to EnumerateTargetFiles(): %v\n", err)
					return
				}

				for _, lf := range lockfiles {
					tempName := s.GeneratePhylumProjectName(inProject.Name, lf.Path)
					// if the project name is NOT in the slice of keys from the phylum project list map, we have to create it
					if !slices.Contains(maps.Keys(*phylumProjectMap), tempName) {
						// PhylumProjectsToCreate = append(PhylumProjectsToCreate, tempName)
						chCreateProjects <- tempName
					} else {
						log.Infof("Found Phylum project for %v : %v\n", project.Name, tempName)
					}
				}
			}(*project)
		}

		var PhylumProjectsToCreate []string
		for item := range chCreateProjects {
			PhylumProjectsToCreate = append(PhylumProjectsToCreate, item)
		}

		chNewProjects := make(chan Syringe.PhylumProject)

		go func() {

		}()

			// TODO: I think i'm breaking this into slices so i can later goroutine it, but I don't really know
			createdPhylumProjects, err := s.PhylumCreateProjectsFromList(PhylumProjectsToCreate)
			if err != nil {
				log.Errorf("Failed to create phylum project: %v\n", err)
				return
			}

			for _, cp := range createdPhylumProjects {
				(*phylumProjectMap)[cp.Name] = cp
			}

			for _, lf := range lockfiles {
				ppName := s.GeneratePhylumProjectName(project.Name, lf.Path)
				phylumProjectFile := (*phylumProjectMap)[ppName]
				err = s.PhylumRunAnalyze(phylumProjectFile, lf)
			}

			localProjects = append(localProjects, Syringe.GitlabProject{
				project.ID,
				project.Name,
				mainBranch,
				false,
				false,
				lockfiles,
				ciFiles,
			})
		}
	},
}
