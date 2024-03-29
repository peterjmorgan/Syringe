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

		configData, err := utils.ReadConfigFile(&structs.TestConfigData{})
		if err != nil {
			log.Fatalf("Failed to read config file")
			return
		}

		//s, err := Syringe2.NewSyringe(envMap, &opts)
		s, err := Syringe2.NewSyringe(configData, &opts)
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

		if err = s.GetAllLockfilesSerial(); err != nil {
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
					log.Debugf("Analyzing %v from %v\n", inLockfile.Path, project.Name)
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
		}
		wgAnalyze.Wait()
	},
}
