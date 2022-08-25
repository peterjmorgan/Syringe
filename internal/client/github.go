package client

import (
	"context"
	"github.com/peterjmorgan/Syringe/internal/utils"
	"golang.org/x/exp/slices"
	"io/ioutil"
	"strings"

	"github.com/google/go-github/github"
	"github.com/peterjmorgan/Syringe/internal/structs"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type GithubClient struct {
	Client  *github.Client
	Ctx     context.Context
	OrgName string
}

func NewGithubClient(envMap map[string]string) *GithubClient {
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

//TODO: write tests
func (g *GithubClient) ListFiles(repoName string, branch string) (*github.Tree, error) {
	fileContent, directoryContent, resp, err := g.Client.Repositories.GetContents(g.Ctx, g.OrgName, repoName, "/", &github.RepositoryContentGetOptions{})
	if err != nil {
		log.Errorf("Failed to GetContents(%v): %v\n", repoName, err)
		log.Errorf("Resp: %v\n", resp.StatusCode)
		return nil, err
	}

	for _, c := range directoryContent {
		switch *c.Type {
		case "file":
			contentHandle, err := g.Client.Repositories.DownloadContents(g.Ctx, g.OrgName, repoName, *c.Path, &github.RepositoryContentGetOptions{})
			if err != nil {
				log.Errorf("Failed to DownloadContents(%v): %v\n", repoName, err)
				return nil, err
			}
			defer contentHandle.Close()
			fileData, err := ioutil.ReadAll(contentHandle)
			if err != nil {
				log.Errorf("Failed to ReadAll(%v): %v\n", contentHandle, err)
				return nil, err
			}
			temp := structs.VcsFile{
				Name:          *c.Name,
				Path:          *c.Path,
				Id:            *c.SHA,
				Content:       fileData,
				PhylumProject: nil,
			}

		case "dir":
		}

	}
	//commits, resp, err := g.Client.Repositories.ListCommits(g.Ctx, g.OrgName, repoName, &github.CommitsListOptions{})
	//if err != nil {
	//	log.Errorf("failed to ListCommits from %v: %v\n", repoName, err)
	//	log.Errorf("%v\n", resp.StatusCode)
	//	return nil, err
	//}
	//
	//lastCommitSHA := *commits[0].SHA
	//
	////files, resp, err := g.Client.Git.GetTree(g.Ctx, g.OrgName, repoName, lastCommitSHA, true)
	//files, resp, err := g.Client(g.Ctx, g.OrgName, repoName, lastCommitSHA, true)
	//if err != nil {
	//	log.Errorf("Failed to GetTree from %v: %v\n", repoName, err)
	//	return nil, err
	//}
	//return files, nil
}

func (g *GithubClient) GetLockfilesByProject(repoName string, mainBranchName string) ([]*structs.VcsFile, error) {
	var retLockfiles []*structs.VcsFile

	projectFiles, err := g.ListFiles(repoName, mainBranchName)
	if err != nil {
		log.Errorf("Failed to ListFiles for %v: %v\n", repoName, err)
	}

	supportedLockfiles := utils.GetSupportedLockfiles()

	for _, file := range projectFiles {
		if slices.Contains(supportedLockfiles, file.Name) || strings.HasSuffix(file.Name, ".csproj") {

		}
	}
	return nil, nil
}
