package syringePackage

import (
	Client2 "github.com/peterjmorgan/Syringe/internal/client"
	"github.com/peterjmorgan/Syringe/internal/structs"
	"strings"
)

type Client interface {
	ListProjects() (*[]*structs.SyringeProject, error)
	GetLockfilesByProject(int64, string) ([]*structs.VcsFile, error)
}

func NewClient(clientType string, envMap map[string]string, mineOnly bool) (Client, error) {
	var c Client
	var err error

	clientType = strings.ToLower(clientType)
	switch clientType {
	case "github": // github
		c = Client2.NewGithubClient(envMap)
	case "gitlab": // gitlab
		c = Client2.NewGitlabClient(envMap, mineOnly)
	}

	return c, err
}
