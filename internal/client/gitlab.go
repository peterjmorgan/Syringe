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

func NewGitlabClient(envMap map[string]string, mineOnly bool) *GitlabClient {
	var gitlabClient *gitlab.Client
	var err error

	if vcsUrl, ok := envMap["vcsUrl"]; ok {
		gitlabClient, err = gitlab.NewClient(envMap["vcsToken"], gitlab.WithBaseURL(vcsUrl))
	} else {
		gitlabClient, err = gitlab.NewClient(envMap["vcsToken"])
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
				Hydrated:  false,
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

func (g *GitlabClient) ListFiles(projectId int64, branch string) ([]*gitlab.TreeNode, error) {
	files, _, err := g.Client.Repositories.ListTree(int(projectId), &gitlab.ListTreeOptions{
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

func (g *GitlabClient) GetLockfilesByProject(projectId int64, mainBranchName string) ([]*structs.VcsFile, error) {
	// TODO: check if mainBranchName isn't set or is "". Bail if that's the case, there are repos without code and will not have a branch
	var retLockFiles []*structs.VcsFile

	projectFiles, err := g.ListFiles(projectId, mainBranchName)
	if err != nil {
		log.Errorf("Failed to ListFiles for %v on branch %v\n", projectId, mainBranchName)
		return nil, err
	}

	supportedLockfiles := utils.GetSupportedLockfiles()

	for _, file := range projectFiles {
		if slices.Contains(supportedLockfiles, file.Name) {
			data, _, err := g.Client.RepositoryFiles.GetRawFile(int(projectId), file.Path, &gitlab.GetRawFileOptions{&mainBranchName})
			if err != nil {
				log.Errorf("Failed to GetRawFile for %v in projectId %v\n", file.Name, projectId)
			}

			rec := structs.VcsFile{file.Name, file.Path, file.ID, data, nil}
			retLockFiles = append(retLockFiles, &rec)
		}
	}
	return retLockFiles, nil
}

// func (g *GitlabClient) PrintProjectVariables(projectId int) error {
// 	variables, _, err := g.Client.ProjectVariables.ListVariables(projectId, &gitlab.ListProjectVariablesOptions{})
// 	if err != nil {
// 		log.Errorf("Failed to list project variables from %v: %v\n", projectId, err)
// 		return err
// 	}
// 	for _, variable := range variables {
// 		log.Infof("Variable: %v:%v\n", variable.Key, variable.Value)
// 	}
// 	return nil
// }
