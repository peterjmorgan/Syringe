package syringePackage

import (
	"strings"

	Client2 "github.com/peterjmorgan/Syringe/internal/client"
	"github.com/peterjmorgan/Syringe/internal/structs"
)

type Client interface {
	ListProjects() (*[]*structs.SyringeProject, error)
	GetLockfilesByProject(int64, string) ([]*structs.VcsFile, error)
}

// func NewClient(clientType string, envMap map[string]string, mineOnly bool, ratelimit int, proxyUrl string) (Client, error) {
//func NewClient(clientType string, envMap map[string]string, opts *structs.SyringeOptions) (Client, error) {
func NewClient(clientType string, configData *structs.ConfigThing, opts *structs.SyringeOptions) (Client, error) {
	var c Client
	var err error

	clientType = strings.ToLower(clientType)
	switch clientType {
	case "github": // github
		c = Client2.NewGithubClient(configData, opts)
	case "gitlab": // gitlab
		c = Client2.NewGitlabClient(configData, opts)
	case "azure": // azuredevops
		c = Client2.NewAzureClient(configData, opts)
	case "bitbucket_cloud":
		c = Client2.NewBitbucketCloudClient(configData, opts)
	}
	return c, err
}
