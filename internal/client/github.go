package client

import (
	"context"

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

func (g *GithubClient) GetLockfilesByProject(projectId int64, mainBranchName string) ([]*structs.VcsFile, error) {
	return nil, nil
}
