package client

import (
	"github.com/peterjmorgan/Syringe/internal/structs"
	utils "github.com/peterjmorgan/Syringe/internal/utils"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/exp/slices"
)

type GitlabClient struct {
	Client   *gitlab.Client
	MineOnly bool
}

func NewGitlabClient(gitlabToken string, gitlabBaseUrl string, mineOnly bool) *GitlabClient {
	var gitlabClient *gitlab.Client
	var err error

	if gitlabBaseUrl != "" {
		gitlabClient, err = gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabBaseUrl))
	} else {
		gitlabClient, err = gitlab.NewClient(gitlabToken)
	}
	if err != nil {
		log.Fatalf("Failed to create gitlab client: %v\n", err)
	}

	return &GitlabClient{
		Client:   gitlabClient,
		MineOnly: mineOnly,
	}
}

func (g *GitlabClient) ListProjects() (*[]*structs.SyringeProject, error) {
	var localProjects []*structs.SyringeProject
	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 10,
			Page:    0,
		},
		Owned: gitlab.Bool(g.MineOnly),
	}

	_, resp, err := g.Client.Projects.ListProjects(opt)
	if err != nil {
		log.Errorf("Failed to list gitlab projects: %v\n", err)
		return nil, err
	}
	count := resp.TotalPages
	listProjectsPB := progressbar.New64(int64(count))

	for {
		listProjectsPB.Add(1)

		gitlabProjects, resp, err := g.Client.Projects.ListProjects(opt)
		if err != nil {
			log.Errorf("Failed to list gitlab projects: %v\n", err)
		}

		// Iterate through gitlabProjects and create SyringeProjects for each
		for _, gitlabProject := range gitlabProjects {
			localProjects = append(localProjects, &structs.SyringeProject{
				Id:        int64(gitlabProject.ID),
				Name:      gitlabProject.Name,
				Branch:    gitlabProject.DefaultBranch,
				Lockfiles: []*structs.VcsFile{},
				CiFiles:   []*structs.VcsFile{},
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		log.Debugf("ListProjects() paging to page #%v\n", opt.Page)
	}

	log.Debugf("Len of gitlab projects: %v\n", len(localProjects))
	return &localProjects, nil
}

// No longer needed
// return []string of branch names
// func (g *GitlabClient) ListBranches(projectId int) ([]string, error) {
// 	var branchNames []string
//
// 	gitlabBranches, _, err := g.Client.Branches.ListBranches(projectId, &gitlab.ListBranchesOptions{})
// 	if err != nil {
// 		log.Errorf("Failed to ListTree from %v: %v\n", projectId, err)
// 		return nil, err
// 	}
// 	for _, elem := range gitlabBranches {
// 		branchName := elem.Name
// 		branchNames = append(branchNames, branchName)
// 	}
//
// 	return branchNames, nil
// }

func (g *GitlabClient) PrintProjectVariables(projectId int) error {
	variables, _, err := g.Client.ProjectVariables.ListVariables(projectId, &gitlab.ListProjectVariablesOptions{})
	if err != nil {
		log.Errorf("Failed to list project variables from %v: %v\n", projectId, err)
		return err
	}
	for _, variable := range variables {
		log.Infof("Variable: %v:%v\n", variable.Key, variable.Value)
	}
	return nil
}

func (g *GitlabClient) ListFiles(projectId int, branch string) ([]*gitlab.TreeNode, error) {
	files, _, err := g.Client.Repositories.ListTree(projectId, &gitlab.ListTreeOptions{
		Path:      gitlab.String("/"),
		Ref:       gitlab.String(branch),
		Recursive: gitlab.Bool(true),
	})
	if err != nil {
		log.Errorf("Failed to ListTree from %v: %v\n", projectId, err)
		return nil, err
	}
	return files, nil
}

// No longer needed! gitlab projects have a DefaultBranch property that has this info!
// func (g *GitlabClient) IdentifyMainBranch(projectId int) (string, error) {
// 	branches, err := g.ListBranches(projectId)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	mainBranchSlice := Syringe.GetMainBranchSlice()
//
// 	foundBranches := make([]string, 0)
// 	for _, branch := range branches {
// 		if slices.Contains(mainBranchSlice, branch) {
// 			foundBranches = append(foundBranches, branch)
// 		}
// 	}
//
// 	var ret string
// 	var retErr error
//
// 	switch len(foundBranches) {
// 	case 0:
// 		ret = ""
// 		retErr = fmt.Errorf("No main branch found: %v\n", projectId)
// 	case 1:
// 		ret = foundBranches[0]
// 		retErr = nil
// 	case 2:
// 		// If there is more than one, opt for "master"
// 		for _, branch := range foundBranches {
// 			if branch == "master" {
// 				ret = branch
// 				retErr = nil
// 			}
// 		}
// 	default:
// 		ret = ""
// 		retErr = fmt.Errorf("IdentifyMainBranch error: shouldn't happen %v\n", projectId)
// 	}
// 	return ret, retErr
// }

func (g *GitlabClient) EnumerateTargetFiles(projectId int, mainBranchName string) ([]*structs.VcsFile, []*structs.VcsFile, error) {
	var retLockFiles []*structs.VcsFile
	var retCiFiles []*structs.VcsFile

	// mainBranchName, err := g.IdentifyMainBranch(projectId)
	// if err != nil {
	// 	// TODO: this needs to pass when repos don't have code
	// 	log.Infof("Failed to IdentifyMainBranch: %v\n", err)
	// 	return nil, nil, err
	// }

	projectFiles, err := g.ListFiles(projectId, mainBranchName)
	if err != nil {
		log.Errorf("Failed to ListFiles for %v on branch %v\n", projectId, mainBranchName)
		return nil, nil, err
	}

	supportedLockfiles := utils.GetSupportedLockfiles()
	supportedciFiles := utils.GetGitlabCIFiles()

	for _, file := range projectFiles {
		if slices.Contains(supportedLockfiles, file.Name) {
			data, _, err := g.Client.RepositoryFiles.GetRawFile(projectId, file.Path, &gitlab.GetRawFileOptions{&mainBranchName})
			if err != nil {
				log.Errorf("Failed to GetRawFile for %v in projectId %v\n", file.Name, projectId)
			}

			rec := structs.VcsFile{file.Name, file.Path, file.ID, data}
			retLockFiles = append(retLockFiles, &rec)
		}
		if slices.Contains(supportedciFiles, file.Name) {
			data, _, err := g.Client.RepositoryFiles.GetRawFile(projectId, file.Path, &gitlab.GetRawFileOptions{})
			if err != nil {
				log.Errorf("Failed to GetRawFile for %v in projectId %v\n", file.Name, projectId)
			}
			rec := structs.VcsFile{file.Name, file.Path, file.ID, data}
			retCiFiles = append(retCiFiles, &rec)
		}
	}
	return retLockFiles, retCiFiles, nil
}

// func (s *Syringe) GetFileTreeFromProject(projectId int) ([]*gitlab.TreeNode, error) {
// 	mainBranch, err := s.IdentifyMainBranch(projectId)
// 	if err != nil {
// 		log.Errorf("Failed to IdentifyMainBranch: %v\n", err)
// 		return nil, err
// 	}
//
// 	projectFiles, err := s.ListFiles(projectId, mainBranch.Name)
// 	if err != nil {
// 		log.Errorf("Failed to ListFiles for %v on branch %v\n", projectId, mainBranch.Name)
// 		return nil, err
// 	}
// 	return projectFiles, nil
// }
