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

		// Enumerate list of Gitlab projects that do not have an associated Phylum project.
		chCreateProjects := make(chan string, 3000)
		chProjectResults := make(chan Syringe.PhylumProject, 3000)
		var wgLoop sync.WaitGroup

		go func() {
			wgLoop.Wait()
			close(chCreateProjects)
			log.Debugf("chProjects channel closed")
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
						// log.Debugf("sending %v to chCreateProjects\n", tempName)
						chCreateProjects <- tempName
						go func() {
							err = s.PhylumCreateProject(chCreateProjects, chProjectResults)
							if err != nil {
								log.Errorf("PhylumCreateProject failed: %v\n", err)
								return
							}
						}()
					} else {
						log.Infof("Found Phylum project for %v : %v\n", project.Name, tempName)
					}
				}
			}(*project)
		}

		// recv from channel to block until create loop is complete
		close(chProjectResults)
		for item := range chProjectResults {
			// createdPhylumProjects = append(createdPhylumProjects, item)
			log.Debugf("recv'd %v\n", item.Name)
			(*phylumProjectMap)[item.Name] = item
		}

		// Phylum analyze loop
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
					ppName := s.GeneratePhylumProjectName(inProject.Name, lf.Path)
					phylumProjectFile := (*phylumProjectMap)[ppName]
					err = s.PhylumRunAnalyze(phylumProjectFile, lf)
				}
			}(*project)
		}
		wgLoop.Wait()
	},
}
