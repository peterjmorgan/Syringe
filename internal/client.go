package syringePackage

import (
	Client2 "github.com/peterjmorgan/Syringe/internal/client"
	"github.com/peterjmorgan/Syringe/internal/structs"
)

type ClientType int

const (
	GithubType = iota
	GitlabType
)

type Client interface {
	ListProjects() (*[]*structs.SyringeProject, error)
}

func NewClient(clientType ClientType, token string, baseUrl string, mineOnly bool) (Client, error) {
	var c Client
	var err error

	switch clientType {
	case GithubType: // github
		c = Client2.NewGithubClient(token, baseUrl)
	case GitlabType: // gitlab
		c = Client2.NewGitlabClient(token, baseUrl, mineOnly)
	}

	return c, err
}
