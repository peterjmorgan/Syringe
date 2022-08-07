package syringe

import (
	"fmt"

	Syringe "github.com/peterjmorgan/Syringe/internal"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/exp/slices"
)

type GitlabClient struct {
	Client   *gitlab.Client
	MineOnly bool
}

func NewGitlabClient(gitlabToken string, gitlabBaseUrl string) *GitlabClient {
	gitlabClient, err := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabBaseUrl))
	if err != nil {
		log.Fatalf("Failed to create gitlab client: %v\n", err)
	}

	return &GitlabClient{
		Client:   gitlabClient,
		MineOnly: false,
	}
}

func (g *GitlabClient) ListProjects(projects **[]*Syringe.SyringeProject) error {
	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 10,
			Page:    0,
		},
		// Owned: gitlab.Bool(s.MineOnly),
	}

	// ch := make(chan []*gitlab.Project)
	var localProjects []*Syringe.SyringeProject

	_, resp, err := g.Client.Projects.ListProjects(opt)
	if err != nil {
		log.Errorf("Failed to list gitlab projects: %v\n", err)
		return err
	}
	count := resp.TotalPages
	// listProjectsBar := uiprogress.AddBar(count).AppendCompleted().PrependElapsed()
	lPbar := progressbar.New64(int64(count))

	for {
		// listProjectsBar.Incr()
		lPbar.Add(1)

		temp, resp, err := g.Client.Projects.ListProjects(opt)
		_ = temp
		if err != nil {
			log.Errorf("Failed to list gitlab projects: %v\n", err)
			return err
		}
		// TODO: convert gitlab projects slice to syringeprojects slice
		// localProjects = append(localProjects, temp...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		log.Debugf("ListProjects() paging to page #%v\n", opt.Page)

	}

	log.Debugf("Len of gitlab projects: %v\n", len(localProjects))
	*projects = &localProjects
	return nil
}

// return []string of branch names
func (g *GitlabClient) ListBranches(projectId int) ([]string, error) {
	var branchNames []string

	gitlabBranches, _, err := g.Client.Branches.ListBranches(projectId, &gitlab.ListBranchesOptions{})
	if err != nil {
		log.Errorf("Failed to ListTree from %v: %v\n", projectId, err)
		return nil, err
	}
	for _, elem := range gitlabBranches {
		branchName := elem.Name
		branchNames = append(branchNames, branchName)
	}

	return branchNames, nil
}

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

func (g *GitlabClient) IdentifyMainBranch(projectId int) (string, error) {
	branches, err := g.ListBranches(projectId)
	if err != nil {
		return "", err
	}

	mainBranchSlice := Syringe.GetMainBranchSlice()

	foundBranches := make([]string, 0)
	for _, branch := range branches {
		if slices.Contains(mainBranchSlice, branch) {
			foundBranches = append(foundBranches, branch)
		}
	}

	var ret string
	var retErr error

	switch len(foundBranches) {
	case 0:
		ret = ""
		retErr = fmt.Errorf("No main branch found: %v\n", projectId)
	case 1:
		ret = foundBranches[0]
		retErr = nil
	case 2:
		// If there is more than one, opt for "master"
		for _, branch := range foundBranches {
			if branch == "master" {
				ret = branch
				retErr = nil
			}
		}
	default:
		ret = ""
		retErr = fmt.Errorf("IdentifyMainBranch error: shouldn't happen %v\n", projectId)
	}
	return ret, retErr
}

func (g *GitlabClient) EnumerateTargetFiles(projectId int) ([]*Syringe.VcsFile, []*Syringe.VcsFile, error) {
	var retLockFiles []*Syringe.VcsFile
	var retCiFiles []*Syringe.VcsFile

	mainBranchName, err := g.IdentifyMainBranch(projectId)
	if err != nil {
		// TODO: this needs to pass when repos don't have code
		log.Infof("Failed to IdentifyMainBranch: %v\n", err)
		return nil, nil, err
	}

	projectFiles, err := g.ListFiles(projectId, mainBranchName)
	if err != nil {
		log.Errorf("Failed to ListFiles for %v on branch %v\n", projectId, mainBranchName)
		return nil, nil, err
	}

	supportedLockfiles := Syringe.GetSupportedLockfiles()
	supportedciFiles := Syringe.GetGitlabCIFiles()

	for _, file := range projectFiles {
		if slices.Contains(supportedLockfiles, file.Name) {
			data, _, err := g.Client.RepositoryFiles.GetRawFile(projectId, file.Path, &gitlab.GetRawFileOptions{&mainBranchName})
			if err != nil {
				log.Errorf("Failed to GetRawFile for %v in projectId %v\n", file.Name, projectId)
			}

			rec := Syringe.VcsFile{file.Name, file.Path, file.ID, data}
			retLockFiles = append(retLockFiles, &rec)
		}
		if slices.Contains(supportedciFiles, file.Name) {
			data, _, err := g.Client.RepositoryFiles.GetRawFile(projectId, file.Path, &gitlab.GetRawFileOptions{})
			if err != nil {
				log.Errorf("Failed to GetRawFile for %v in projectId %v\n", file.Name, projectId)
			}
			rec := Syringe.VcsFile{file.Name, file.Path, file.ID, data}
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
