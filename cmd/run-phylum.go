package cmd

import (
	PhylumSyringGitlab "github.com/peterjmorgan/PhylumSyringeGitlab/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

		phylumProjectMap, err := s.PhylumGetProjectMap()
		if err != nil {
			log.Fatalf("Failed to PhylumGetProjectMap(): %v\n", err)
			return
		}

		var localProjects []PhylumSyringGitlab.GitlabProject

		for _, project := range gitlabProjects {
			var PhylumProjectsToCreate []string
			mainBranch, err := s.IdentifyMainBranch(project.ID)
			if err != nil {
				log.Fatalf("Failed to IdentifyMainBranch(): %v\n", err)
				return
			}

			lockfiles, ciFiles, err := s.EnumerateTargetFiles(project.ID)

			for _, lf := range lockfiles {
				tempName := s.GeneratePhylumProjectName(project.Name, lf.Path)
				// if the project name is NOT in the slice of keys from the phylum project list map, we have to create it
				if !slices.Contains(maps.Keys(phylumProjectMap), tempName) {
					PhylumProjectsToCreate = append(PhylumProjectsToCreate, tempName)
				} else {
					log.Infof("Found Phylum project for %v : %v\n", project.Name, tempName)
				}
			}

			// TODO: I think i'm breaking this into slices so i can later goroutine it, but I don't really know
			createdPhylumProjects, err := s.PhylumCreateProjectsFromList(PhylumProjectsToCreate)
			if err != nil {
				log.Errorf("Failed to create phylum project: %v\n", err)
				return
			}
			for _, cp := range createdPhylumProjects {
				phylumProjectMap[cp.Name] = cp
			}

			for _, lf := range lockfiles {
				ppName := s.GeneratePhylumProjectName(project.Name, lf.Path)
				phylumProjectFile := phylumProjectMap[ppName]
				err = s.PhylumRunAnalyze(phylumProjectFile, lf)
			}

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
	},
}
