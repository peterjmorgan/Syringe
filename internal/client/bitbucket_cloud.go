package client

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/ktrysmt/go-bitbucket"
	"github.com/peterjmorgan/Syringe/internal/structs"
	log "github.com/sirupsen/logrus"
)

type BitbucketCloudClient struct {
	Client *bitbucket.Client
}

func NewBitbucketCloudClient(envMap map[string]string, opts *structs.SyringeOptions) *BitbucketCloudClient {
	//token := envMap["vcsToken"]

	// Have to use Oauth2 client credentials config. Could not auth to the API any other way.
	client := bitbucket.NewOAuthClientCredentials("APbFeKnRHr2zBk6v6w", "qP2aBzrzQzmDUbnHnYLScStwxDuHQTFV")

	return &BitbucketCloudClient{Client: client}
}

func (b *BitbucketCloudClient) ListProjects() (*[]*structs.SyringeProject, error) {
	var retProjects []*structs.SyringeProject

	repos, err := b.Client.Repositories.ListForAccount(&bitbucket.RepositoriesOptions{
		Owner: "peter_morgan_",
		Role:  "member",
	})
	if err != nil {
		errStr := fmt.Sprintf("failed to ListProjects: %v\n", err)
		log.Error(errStr)
		return nil, fmt.Errorf(errStr)
	}
	for _, item := range repos.Items {
		uuidStr := item.Uuid
		uuid := uuid.MustParse(uuidStr)
		if err != nil {
			log.Errorf("uuid creation failed: %v\n", err)
		}

		retProjects = append(retProjects, &structs.SyringeProject{
			Id:        int64(uuid.ID()),
			Name:      item.Name,
			Branch:    item.Mainbranch.Name,
			Lockfiles: nil,
			CiFiles:   nil,
			Hydrated:  false,
			GUID:      uuid,
		})
	}

	return &retProjects, nil
}
