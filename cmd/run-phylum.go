package cmd

import (
	"context"
	"sync"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/semaphore"

	Syringe2 "github.com/peterjmorgan/Syringe/internal"
	"github.com/peterjmorgan/Syringe/internal/structs"
	"github.com/peterjmorgan/Syringe/internal/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var projectIDFileName string

func init() {
	runPhylumCmd.Flags().StringVar(&projectIDFileName, "pidFilename", "", "project id filename")
	rootCmd.AddCommand(runPhylumCmd)
}

var runPhylumCmd = &cobra.Command{
	Use:   "run-phylum",
	Short: "Run Phylum on GitLab Projects",
	Run: func(cmd *cobra.Command, args []string) {
		var mineOnly bool = false
		var ratelimit int = 0
		var proxyUrl string = ""
		var err error

		if cmd.Flags().Lookup("debug").Changed {
			log.SetLevel(log.DebugLevel)
		}

		if cmd.Flags().Lookup("mine-only").Changed {
			mineOnly = true
		}
		if cmd.Flags().Lookup("ratelimit").Changed {
			ratelimit, err = cmd.Flags().GetInt("ratelimit")
			if err != nil {
				log.Errorf("Failed to read int value from ratelimit")
			}
		}
		if cmd.Flags().Lookup("proxyUrl").Changed {
			proxyUrl, err = cmd.Flags().GetString("proxyUrl")
			if err != nil {
				log.Errorf("Failed to read string value from proxyUrl")
			}
		}

		opts := structs.SyringeOptions{
			MineOnly:  mineOnly,
			RateLimit: ratelimit,
			ProxyUrl:  proxyUrl,
		}

		envMap, err := utils.ReadEnvironment()
		if err != nil {
			log.Fatalf("Failed to read environment variables: %v\n", err)
			return
		}
		s, err := Syringe2.NewSyringe(envMap, &opts)
		if err != nil {
			log.Fatal("Failed to create NewSyringe(): %v\n", err)
			return
		}

		// if cmd.Flags().Lookup("pidFilename").Changed {
		// 	pidFile := cmd.Flags().Lookup("pidFilename").Value.String()
		// 	err = s.LoadPidFile(pidFile)
		// 	if err != nil {
		// 		return
		// 	}
		// }

		var phylumProjectMap *map[string]structs.PhylumProject
		var wg sync.WaitGroup

		wg.Add(2)
		go func() {
			defer wg.Done()
			if err := s.ListProjects(); err != nil {
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

		// // Enumerate list of Gitlab projects that do not have an associated Phylum project.
		// chCreateProjects := make(chan string, 3000)
		// chProjectResults := make(chan structs.PhylumProject, 3000)
		// var wgAnalyze sync.WaitGroup
		//
		// go func() {
		//	wgAnalyze.Wait()
		//	close(chCreateProjects)
		// }()
		// for _, project := range *s.Projects {
		//	wgAnalyze.Add(1)
		//	go func(inProject structs.SyringeProject) {
		//		defer wgAnalyze.Done()
		//
		//		// lockfiles, _, err := s.EnumerateTargetFiles(inProject.ID)
		//		// if err != nil {
		//		// 	log.Debugf("Failed to GetLockFiles(): %v\n", err)
		//		// 	return
		//		// }
		//
		//		resultProject, err := s.GetLockfilesByProject(inProject.Id)
		//		if err != nil {
		//			log.Debugf("Failed to GetLockFiles(): %v\n", err)
		//			return
		//		}
		//
		//		for _, lf := range resultProject.Lockfiles {
		//			phylumProjectName := utils.GeneratePhylumProjectName(inProject.Name, lf.Path)
		//			// if the project name is NOT in the slice of keys from the phylum project list map, we have to create it
		//			if !slices.Contains(maps.Keys(*phylumProjectMap), phylumProjectName) {
		//				log.Debugf("sending %v to chCreateProjects\n", phylumProjectName)
		//				chCreateProjects <- phylumProjectName
		//				go func() {
		//					err = s.PhylumCreateProject(chCreateProjects, chProjectResults)
		//					if err != nil {
		//						log.Errorf("PhylumCreateProject failed: %v\n", err)
		//						return
		//					}
		//				}()
		//			} else {
		//				log.Debugf("Found Phylum project for %v : %v\n", inProject.Name, phylumProjectName)
		//			}
		//		}
		//	}(*project)
		// }
		//
		// // recv from channel to block until create loop is complete
		// go func() {
		//	for item := range chProjectResults {
		//		log.Debugf("recv'd projectResult: %v\n", item.Name)
		//		(*phylumProjectMap)[item.Name] = item
		//	}
		//	close(chProjectResults)
		// }()

		if err = s.GetAllLockfiles(); err != nil {
			log.Errorf("Failed to GetAllLockfiles(): %v\n", err)
		}

		newProjects := s.IntegratePhylumProjectList(phylumProjectMap)

		var wgNewProjects sync.WaitGroup
		chCreatedProjects := make(chan *structs.PhylumProject, 3000) // responses

		for _, project := range newProjects {
			wgNewProjects.Add(1)
			go func(projectName string) {
				defer wgNewProjects.Done()
				err = s.PhylumCreateProject(projectName, chCreatedProjects)
				if err != nil {
					log.Errorf("Failed to create phylum project %v: %v\n", projectName, err)
				}
			}(project)
		}

		go func() {
			for item := range chCreatedProjects {
				log.Debugf("recv'd projectResult: %v\n", item.Name)
				(*phylumProjectMap)[item.Name] = *item
			}
			close(chCreatedProjects)
		}()
		wgNewProjects.Wait()

		// update the struct with the new project data
		_ = s.IntegratePhylumProjectList(phylumProjectMap)

		analyzeBar := progressbar.NewOptions(s.LockfileCount, progressbar.OptionSetDescription("Analyzing lockfiles"))
		var wgAnalyze sync.WaitGroup
		sem := semaphore.NewWeighted(50)
		ctx := context.TODO()

		// Phylum analyze loop
		for _, project := range *s.Projects {
			for _, lockfile := range project.Lockfiles {
				wgAnalyze.Add(1)
				go func(inLockfile *structs.VcsFile) {
					defer wgAnalyze.Done()
					if err = sem.Acquire(ctx, 1); err != nil {
						log.Errorf("Failed to acquire semaphore: %v\n", err)
						return
					}
					defer sem.Release(1)
					err = s.PhylumRunAnalyze(*inLockfile.PhylumProject, inLockfile, inLockfile.PhylumProject.Name)
					if err != nil {
						log.Errorf("Failed to analyze %v: %v\n", inLockfile.PhylumProject.Name, err)
					}
					analyzeBar.Add(1)
				}(lockfile)

			}
			// go func(inProject structs.SyringeProject) {
			// 	defer wgAnalyze.Done()
			//
			// 	analyzeBar.Add(1)
			// 	for _, lf := range inProject.Lockfiles {
			// 		ppName := utils.GeneratePhylumProjectName(inProject.Name, lf.Path)
			// 		phylumProjectFile := (*phylumProjectMap)[ppName]
			// 		err = s.PhylumRunAnalyze(phylumProjectFile, lf, ppName)
			// 	}
			// 	analyzeBar.Add(1)
			// }(*project)
		}
		wgAnalyze.Wait()
	},
}
