package syringe

import (
	"context"

	"github.com/google/go-github/github"
	Syringe "github.com/peterjmorgan/Syringe/internal"
	"golang.org/x/oauth2"
)

type GithubClient struct {
	Client *github.Client
	Ctx    context.Context
}

func NewGithubClient(githubToken string, githubBaseUrl string) *GithubClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)

	gh := github.NewClient(oauth2.NewClient(ctx, ts))
	return &GithubClient{Client: gh, Ctx: ctx}
}

func (g *GithubClient) ListProjects(projects **[]*Syringe.SyringeProject) error {

	return nil
}
