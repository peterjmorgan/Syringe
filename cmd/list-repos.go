package cmd

//
// import (
// 	"errors"
// 	"fmt"
// 	"os"
//
// 	"github.com/jedib0t/go-pretty/v6/table"
// 	PhylumSyringGitlab "github.com/peterjmorgan/Syringe/internal"
// 	"github.com/spf13/cobra"
// )
//
// func init() {
// 	rootCmd.AddCommand(listReposCmd)
// }
//
// var listReposCmd = &cobra.Command{
// 	Use:   "list-repos",
// 	Short: "List VCS Repos",
// 	Args: func(cmd *cobra.Command, args []string) error {
// 		if len(args) < 1 {
// 			return errors.New("requires a Gitlab token")
// 		}
// 		if len(args[0]) > 5 && len(args[0]) < 100 {
// 			return nil
// 		}
// 		return fmt.Errorf("missing Gitlab token")
// 	},
// 	Run: func(cmd *cobra.Command, args []string) {
// 		s, err := PhylumSyringGitlab.NewSyringe()
// 		if err != nil {
// 			fmt.Println(err)
// 		}
//
// 		projects, err := s.ListProjects()
// 		if err != nil {
// 			fmt.Println(err)
// 		}
//
// 		t := table.NewWriter()
// 		t.SetStyle(table.StyleLight)
// 		t.SetOutputMirror(os.Stdout)
// 		t.AppendHeader(table.Row{"Project Name", "ID", "Main Branch", "Protected"})
// 		for _, project := range projects {
// 			mainBranch, err := s.IdentifyMainBranch(project.ID)
// 			if err != nil {
// 				fmt.Println(err)
// 			}
//
// 			t.AppendRow(table.Row{
// 				project.Name,
// 				project.ID,
// 				mainBranch.Name,
// 				mainBranch.Protected,
// 			})
// 		}
// 		t.Render()
// 	},
// }
