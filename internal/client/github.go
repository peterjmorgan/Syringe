package client

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"
	"github.com/peterjmorgan/Syringe/internal/structs"
	"github.com/peterjmorgan/Syringe/internal/utils"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"golang.org/x/oauth2"
)

type GithubClient struct {
	Client  *github.Client
	Ctx     context.Context
	OrgName string
}

func NewGithubClient(envMap map[string]string, opts *structs.SyringeOptions) *GithubClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: envMap["vcsToken"]},
	)

	gh := github.NewClient(oauth2.NewClient(ctx, ts))
	return &GithubClient{
		Client:  gh,
		Ctx:     ctx,
		OrgName: envMap["vcsOrg"],
	}
}

func (g *GithubClient) ListProjects() (*[]*structs.SyringeProject, error) {
	var localProjects []*structs.SyringeProject
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 20},
	}

	_, resp, err := g.Client.Repositories.ListByOrg(g.Ctx, g.OrgName, opt)
	if err != nil {
		log.Errorf("Failed to get github repositories: %v\n", err)
		return nil, err
	}
	count := resp.LastPage
	listProjectsPB := progressbar.New64(int64(count))

	for {
		listProjectsPB.Add(1)

		githubRepos, resp, err := g.Client.Repositories.ListByOrg(g.Ctx, g.OrgName, opt)
		if err != nil {
			log.Errorf("Failed to get github repositories: %v\n", err)
			return nil, err
		}

		for _, repo := range githubRepos {
			localProjects = append(localProjects, &structs.SyringeProject{
				Id:        *repo.ID,
				Name:      *repo.Name,
				Branch:    *repo.DefaultBranch,
				Lockfiles: []*structs.VcsFile{},
				CiFiles:   []*structs.VcsFile{},
				Hydrated:  false,
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return &localProjects, nil
}

func (g *GithubClient) ListFiles(repoName string, branch string) (*github.Tree, error) {

	// fileContent, directoryContent, resp, err := g.Client.Repositories.GetContents(g.Ctx, g.OrgName, repoName, "/", &github.RepositoryContentGetOptions{})
	// if err != nil {
	// 	log.Errorf("Failed to GetContents(%v): %v\n", repoName, err)
	// 	log.Errorf("Resp: %v\n", resp.StatusCode)
	// 	return nil, err
	// }
	//
	// for _, c := range directoryContent {
	// 	switch *c.Type {
	// 	case "file":
	// 		contentHandle, err := g.Client.Repositories.DownloadContents(g.Ctx, g.OrgName, repoName, *c.Path, &github.RepositoryContentGetOptions{})
	// 		if err != nil {
	// 			log.Errorf("Failed to DownloadContents(%v): %v\n", repoName, err)
	// 			return nil, err
	// 		}
	// 		defer contentHandle.Close()
	// 		fileData, err := ioutil.ReadAll(contentHandle)
	// 		if err != nil {
	// 			log.Errorf("Failed to ReadAll(%v): %v\n", contentHandle, err)
	// 			return nil, err
	// 		}
	// 		temp := structs.VcsFile{
	// 			Name:          *c.Name,
	// 			Path:          *c.Path,
	// 			Id:            *c.SHA,
	// 			Content:       fileData,
	// 			PhylumProject: nil,
	// 		}
	//
	// 	case "dir":
	// 	}

	var resultsTree github.Tree

	commits, resp, err := g.Client.Repositories.ListCommits(g.Ctx, g.OrgName, repoName, &github.CommitsListOptions{})
	if err != nil {
		log.Errorf("failed to ListCommits from %v: %v\n", repoName, err)
		log.Errorf("%v\n", resp.StatusCode)
		return nil, err
	}

	// Get the latest commmit SHA for the repo and branch
	lastCommitSHA := *commits[0].SHA

	// Get the tree of objects based on the commit SHA
	ghTree, resp, err := g.Client.Git.GetTree(g.Ctx, g.OrgName, repoName, lastCommitSHA, true)
	if err != nil {
		log.Errorf("Failed to GetTree from %v: %v\n", repoName, err)
		return nil, err
	}
	resultsTree.Truncated = ghTree.Truncated
	if *ghTree.Truncated {
		log.Errorf("GH_ListFiles: GetTree() response is truncated!")
		return nil, err
	}

	for _, treeEntry := range ghTree.Entries {
		switch *treeEntry.Type {
		case "blob": // file
			resultsTree.Entries = append(resultsTree.Entries, treeEntry)
		case "tree": // directory
			// TODO: handle case where GetTree response is truncated here

		default:
			log.Warnf("GH_ListFiles: found treeEntry type: %v\n", *treeEntry.Type)
		}

	}
	return &resultsTree, nil
}

// func (g *GithubClient) GetLockfilesByProject(repoName string, mainBranchName string) ([]*structs.VcsFile, error) {
func (g *GithubClient) GetLockfilesByProject(projectId int64, mainBranchName string) ([]*structs.VcsFile, error) {
	var retLockfiles []*structs.VcsFile

	// Get Repo name via ID
	repo, _, err := g.Client.Repositories.GetByID(g.Ctx, projectId)

	projectTree, err := g.ListFiles(*repo.Name, mainBranchName)
	if err != nil {
		log.Errorf("Failed to ListFiles for %v: %v\n", *repo.Name, err)
		return nil, err
	}

	supportedLockfiles := utils.GetSupportedLockfiles()

	for _, file := range projectTree.Entries {
		fileName := filepath.Base(*file.Path)
		if slices.Contains(supportedLockfiles, fileName) || strings.HasSuffix(fileName, ".csproj") {
			log.Debugf("Lockfile: %v in %v from project: %v\n", fileName, *file.Path, *repo.Name)
			contentHandle, err := g.Client.Repositories.DownloadContents(g.Ctx, g.OrgName, *repo.Name, *file.Path, &github.RepositoryContentGetOptions{})
			if err != nil {
				log.Errorf("Failed to DownloadContents for %v in repo:%v: %v", *file.Path, *repo.Name, err)
				return nil, err
			}
			b, err := ioutil.ReadAll(contentHandle)
			if err != nil {
				log.Errorf("Failed to read bytes from %v: %v\n", *file.Path, err)
				return nil, err
			}
			retLockfiles = append(retLockfiles, &structs.VcsFile{
				Name:          fileName,
				Path:          *file.Path,
				Id:            *file.SHA,
				Content:       b,
				PhylumProject: nil,
			})

		}
	}
	return retLockfiles, nil
}
