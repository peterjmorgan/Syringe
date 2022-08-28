package client

import (
	"github.com/ktrysmt/go-bitbucket"
	"github.com/peterjmorgan/Syringe/internal/structs"
)

type BitbucketCloudClient struct {
	Client *bitbucket.Client
}

func NewBitbucketCloudClient(envMap map[string]string, opts *structs.SyringeOptions) *BitbucketCloudClient {
	username := envMap["SYRINGE_VCS_BITBUCKETCLOUD_USERNAME"]
	password := envMap["SYRINGE_VCS_BITBUCKETCLOUD_PASSWORD"]

	client := bitbucket.NewBasicAuth(username, password)

	return &BitbucketCloudClient{Client: client}
}

func (b *BitbucketCloudClient) ListProjects() (*[]*structs.SyringeProject, error) {
	var retProjects []*structs.SyringeProject

	repos, err := b.Client.Repositories.ListForAccount(&bitbucket.RepositoriesOptions{
		Owner: "",
		Role:  "",
	})
}
