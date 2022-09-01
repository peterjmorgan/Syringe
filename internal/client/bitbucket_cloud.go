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
	Owner  string
}

//func NewBitbucketCloudClient(envMap map[string]string, opts *structs.SyringeOptions) *BitbucketCloudClient {
func NewBitbucketCloudClient(configData *structs.ConfigThing, opts *structs.SyringeOptions) *BitbucketCloudClient {
	//token := envMap["vcsToken"]
	//owner := "peter_morgan_"

	// Have to use Oauth2 client credentials config. Could not auth to the API any other way.
	//client := bitbucket.NewOAuthClientCredentials("APbFeKnRHr2zBk6v6w", "qP2aBzrzQzmDUbnHnYLScStwxDuHQTFV")
	client := bitbucket.NewOAuthClientCredentials(configData.Associated["bbClientId"], configData.Associated["bbClientSecret"])

	return &BitbucketCloudClient{
		Client: client,
		Owner:  configData.Associated["bbOwner"],
	}
}

func (b *BitbucketCloudClient) ListProjects() (*[]*structs.SyringeProject, error) {
	var retProjects []*structs.SyringeProject

	repos, err := b.Client.Repositories.ListForAccount(&bitbucket.RepositoriesOptions{
		Owner: b.Owner,
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
			Name:      item.Slug,
			Branch:    item.Mainbranch.Name,
			Lockfiles: nil,
			CiFiles:   nil,
			Hydrated:  false,
			GUID:      uuid,
		})
	}

	return &retProjects, nil
}

func (b *BitbucketCloudClient) ListFiles(repoSlug string, branch string) (*[]*bitbucket.RepositoryFile, error) {
	var retFiles []*bitbucket.RepositoryFile

	files, err := b.Client.Repositories.Repository.ListFiles(&bitbucket.RepositoryFilesOptions{
		Owner:    b.Owner,
		RepoSlug: repoSlug,
		Ref:      branch,
		Path:     "/",
		MaxDepth: 500,
	})
	if err != nil {
		errStr := fmt.Sprintf("failed to ListFiles for %v: %v\n", repoSlug, err)
		log.Error(errStr)
		return nil, fmt.Errorf(errStr)
	}
	_ = files

	for _, file := range files {
		if file.Type == "commit_file" {
			newFile := new(bitbucket.RepositoryFile)
			*newFile = file

			retFiles = append(retFiles, newFile)
		}
	}

	return &retFiles, nil
}

func (b *BitbucketCloudClient) GetLockfilesByProject(projectId int64, branch string) ([]*structs.VcsFile, error) {
	var retFiles []*structs.VcsFile

	return retFiles, nil
}
