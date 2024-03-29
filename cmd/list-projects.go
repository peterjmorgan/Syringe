package cmd

import (
	"os"
	"sync"

	"github.com/peterjmorgan/Syringe/internal/utils"

	"github.com/jedib0t/go-pretty/v6/table"
	Syringe2 "github.com/peterjmorgan/Syringe/internal"
	"github.com/peterjmorgan/Syringe/internal/structs"
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

		s, err := Syringe2.NewSyringe(configData, &opts)
		if err != nil {
			log.Fatal("Failed to create NewSyringe(): %v\n", err)
			return
		}

		var phylumProjectMap *map[string]structs.PhylumProject
		var wg sync.WaitGroup

		wg.Add(2)
		go func() {
			err := s.ListProjects()
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

		if err = s.GetAllLockfiles(); err != nil {
			log.Errorf("Failed to GetAllLockfiles: %v\n", err)
		}

		_ = s.IntegratePhylumProjectList(phylumProjectMap)

		t := table.NewWriter()
		// rowConfigAutoMerge := table.RowConfig{AutoMerge: true}
		// // t.SetAutoIndex(true)
		// t.SetColumnConfigs([]table.ColumnConfig{
		// 	{Number: 1, AutoMerge: true},
		// 	{Number: 2, AutoMerge: true},
		// 	{Number: 3, AutoMerge: true},
		// 	{Number: 4, AutoMerge: true},
		// })
		t.SetStyle(table.StyleLight)
		t.SetOutputMirror(os.Stdout)
		// t.AppendHeader(table.Row{"Project Name", "ID", "Main Branch", "Protected", "Lockfile Path"}, rowConfigAutoMerge)
		t.AppendHeader(table.Row{"Project Name", "ID", "Main Branch", "Lockfile Path"})
		for _, p := range *s.Projects {
			for _, lockfile := range p.Lockfiles {
				t.AppendRow(table.Row{p.Name, p.Id, p.Branch, lockfile.Path})
			}
			// t.AppendRow(table.Row{p.Name, p.Id, p.Branch})
		}
		t.Style().Options.SeparateRows = true
		t.Render()
	},
}
