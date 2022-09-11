package client

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/ktrysmt/go-bitbucket"
	"github.com/peterjmorgan/Syringe/internal/structs"
	"github.com/peterjmorgan/Syringe/internal/utils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

type BitbucketCloudClient struct {
	Client          *bitbucket.Client
	Owner           string
	ProjectMap      map[int64]*bitbucket.Repository
	ProjectMapMutex sync.RWMutex
}

// func NewBitbucketCloudClient(envMap map[string]string, opts *structs.SyringeOptions) *BitbucketCloudClient {
func NewBitbucketCloudClient(configData *structs.ConfigThing, opts *structs.SyringeOptions) *BitbucketCloudClient {
	//token := envMap["vcsToken"]
	//owner := "peter_morgan_"

	// Have to use Oauth2 client credentials config. Could not auth to the API any other way.
	//client := bitbucket.NewOAuthClientCredentials("APbFeKnRHr2zBk6v6w", "qP2aBzrzQzmDUbnHnYLScStwxDuHQTFV")
	client := bitbucket.NewOAuthClientCredentials(configData.Associated["bbClientId"], configData.Associated["bbClientSecret"])

	return &BitbucketCloudClient{
		Client:     client,
		Owner:      configData.Associated["bbOwner"],
		ProjectMap: make(map[int64]*bitbucket.Repository, 0),
	}
}

func (b *BitbucketCloudClient) ListProjects() (*[]*structs.SyringeProject, error) {
	var retProjects []*structs.SyringeProject

	repos, err := b.Client.Repositories.ListForAccount(&bitbucket.RepositoriesOptions{
		Owner: b.Owner,
		Role:  "member",
	})
	if err != nil {
		errStr := fmt.Sprintf("BitBucket: failed to ListProjects: %v\n", err)
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
		temp := new(bitbucket.Repository)
		*temp = item
		b.ProjectMap[int64(uuid.ID())] = temp
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
		errStr := fmt.Sprintf("BitBucket: failed to ListFiles for %v: %v\n", repoSlug, err)
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

func (b *BitbucketCloudClient) GetLockfilesByProject(projectId int64, mainBranchName string) ([]*structs.VcsFile, error) {
	var retLockfiles []*structs.VcsFile

	b.ProjectMapMutex.RLock()
	repo := b.ProjectMap[projectId]
	b.ProjectMapMutex.RUnlock()

	projectFiles, err := b.ListFiles(repo.Name, repo.Mainbranch.Name)
	if err != nil {
		errStr := fmt.Sprintf("BitBucket: failed to GetLockfilesByProject for %v: %v\n", repo.Name, err)
		log.Error(errStr)
		return nil, fmt.Errorf(errStr)
	}

	supportedLockfiles := utils.GetSupportedLockfiles()

	for _, file := range *projectFiles {
		// var filePath string
		fileName := filepath.Base(file.Path)
		if slices.Contains(supportedLockfiles, fileName) || strings.HasSuffix(fileName, ".csproj") {
			log.Debugf("Lockfile: %v in %v from project: %v\n", fileName, file.Path, repo.Name)
			// download te file

			// if !strings.Contains(file.Path, "/") {
			// 	filePath = fmt.Sprintf("./%v", file.Path)
			// } else {
			// 	filePath = file.Path
			// }

			content, err := b.Client.Repositories.Repository.GetFileBlob(&bitbucket.RepositoryBlobOptions{
				Owner:    b.Owner,
				RepoSlug: repo.Name,
				Ref:      mainBranchName,
				Path:     file.Path,
			})

			// content, err := b.Client.Repositories.Repository.GetFileContent(&bitbucket.RepositoryFilesOptions{
			// 	Owner:    b.Owner,
			// 	RepoSlug: repo.Name,
			// 	Ref:      mainBranchName,
			// 	Path:     file.Path,
			// 	MaxDepth: 500,
			// })
			if err != nil {
				errStr := fmt.Sprintf("BitBucket: failed to GetFileContent for %v: %v\n", fileName, err)
				log.Error(errStr)
				return nil, fmt.Errorf(errStr)
			}

			retLockfiles = append(retLockfiles, &structs.VcsFile{
				Name:          fileName,
				Path:          file.Path,
				Id:            file.Path,
				Content:       content.Content,
				PhylumProject: nil,
			})

		}
	}

	return retLockfiles, nil
}
