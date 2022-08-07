package syringe

import syringe "github.com/peterjmorgan/Syringe/internal/client"

type ClientType int

const (
	GithubType = iota
	GitlabType
)

type Client interface {
	ListProjects(**[]*SyringeProject) error
}

func NewClient(clientType ClientType, token string, baseUrl string) (Client, error) {
	var c Client
	var err error

	switch clientType {
	case GithubType: // github
		c = syringe.NewGithubClient(token, baseUrl)
	case GitlabType: // gitlab
		c = syringe.NewGitlabClient(token, baseUrl)
	}

	return c, err
}
