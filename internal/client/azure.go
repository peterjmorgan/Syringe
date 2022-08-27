package client

import (
	"context"
	"fmt"

	"github.com/microsoft/azure-devops-go-api/azuredevops/build"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	log "github.com/sirupsen/logrus"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"github.com/peterjmorgan/Syringe/internal/structs"
)

type AzureClient struct {
	Clients *AzureSubClient
	Ctx     context.Context
	OrgName string
}

type AzureSubClient struct {
	CoreClient  core.Client
	BuildClient build.Client
	GitClient   git.Client
}

func NewAzureClient(envMap map[string]string, opts *structs.SyringeOptions) *AzureClient {
	// var connUrl string
	//
	// if vcsUrl, ok := envMap["vcsUrl"]; ok {
	//	connUrl = vcsUrl
	// }

	org := envMap["vcsOrg"]
	conn := azuredevops.NewPatConnection(org, envMap["vcsToken"])
	ctx := context.Background()
	coreClient, err := core.NewClient(ctx, conn)
	buildClient, err := build.NewClient(ctx, conn)
	gitClient, err := git.NewClient(ctx, conn)
	if err != nil {
		// handle
	}
	return &AzureClient{
		Clients: &AzureSubClient{
			CoreClient:  coreClient,
			BuildClient: buildClient,
			GitClient:   gitClient,
		},
		Ctx: ctx,
	}
}

func (a *AzureClient) ListProjects() (*[]*structs.SyringeProject, error) {
	var localProjects []core.TeamProjectReference
	var retProjects []*structs.SyringeProject

	// Projects are not 1-to-1 with repositories in ADO
	projectResp, err := a.Clients.CoreClient.GetProjects(a.Ctx, core.GetProjectsArgs{})
	if err != nil {
		return nil, fmt.Errorf("Failed to GetProjects: %v\n", err)
	}

	// Paginate through ADO Projects
	for projectResp != nil {
		for _, proj := range (*projectResp).Value {
			localProjects = append(localProjects, proj)
		}
		if projectResp.ContinuationToken != "" {
			projectArgs := core.GetProjectsArgs{
				ContinuationToken: &projectResp.ContinuationToken,
			}
			projectResp, err = a.Clients.CoreClient.GetProjects(a.Ctx, projectArgs)
			if err != nil {
				return nil, fmt.Errorf("Failed to GetProjects (cont) %v\n", err)
			}
		} else {
			projectResp = nil
		}
	}

	// Iterate through ADO repositories
	for _, proj := range localProjects {

		projId := proj.Id.String()
		repos, err := a.Clients.GitClient.GetRepositories(a.Ctx, git.GetRepositoriesArgs{
			Project: &projId,
		})
		if err != nil {
			errStr := fmt.Sprintf("failed to GetRepositories for %v: %v\n", proj.Name, err)
			log.Error(errStr)
			return nil, fmt.Errorf(errStr)
		}

		for _, repo := range *repos {
			retProjects = append(retProjects, &structs.SyringeProject{
				Id:        int64(repo.Id.ID()),
				GUID:      *repo.Id,
				Name:      *repo.Name,
				Branch:    *repo.DefaultBranch,
				Lockfiles: nil,
				CiFiles:   nil,
				Hydrated:  false,
			})
		}
	}
	return &retProjects, nil
}

func (a *AzureClient) ListFiles(repoID string, branch string) ([]*git.GitItem, error) {
	// var retTree *git.GitTreeRef
	var retItems []*git.GitItem

	var recurse git.VersionControlRecursionType = "full"
	var versionType git.GitVersionType = git.GitVersionType("branch")

	items, err := a.Clients.GitClient.GetItems(a.Ctx, git.GetItemsArgs{
		RepositoryId: &repoID,
		// Path:                   nil,
		// Project:                nil,
		// ScopePath:              nil,
		RecursionLevel:         &recurse,
		IncludeContentMetadata: &[]bool{true}[0],
		// LatestProcessedChange:  nil,
		// Download:               nil,
		VersionDescriptor: &git.GitVersionDescriptor{
			Version:        &branch,
			VersionOptions: nil,
			VersionType:    &versionType,
			// VersionType:    &[]git.GitVersionType{git.GitVersionType(branch)}[0],
		},
		// IncludeContent:         nil,
		// ResolveLfs:             nil,
	})
	if err != nil {
		errStr := fmt.Sprintf("failed to GetItems for %v: %v\n", repoID, err)
		log.Error(errStr)
		return nil, fmt.Errorf(errStr)
	}
	for _, item := range *items {
		if *item.GitObjectType == "blob" {
			blah := new(git.GitItem)
			*blah = item
			retItems = append(retItems, blah)
		}
	}

	// // TODO: get commmit SHA first
	// branchResp, err := a.Clients.GitClient.GetBranch(a.Ctx, git.GetBranchArgs{
	// 	RepositoryId: &repoID,
	// 	Name:         &branch,
	// 	// Project:               nil,
	// 	// BaseVersionDescriptor: nil,
	// })
	// if err != nil {
	// 	errStr := fmt.Sprintf("failed to GetBranch for %v: %v\n", repoID, err)
	// 	log.Error(errStr)
	// 	return nil, fmt.Errorf(errStr)
	// }
	// commitSHA := branchResp.Commit.CommitId
	//
	// tree, err := a.Clients.GitClient.GetTree(a.Ctx, git.GetTreeArgs{
	// 	RepositoryId: &repoID,
	// 	Sha1:         commitSHA,
	// 	Recursive:    &[]bool{true}[0],
	// 	// Project:      nil,
	// 	// ProjectId:    nil,
	// 	// FileName:  nil,
	// })
	// if err != nil {
	// 	errStr := fmt.Sprintf("failed to GetTree for %v: %v\n", repoID, err)
	// 	log.Error(errStr)
	// 	return nil, fmt.Errorf(errStr)
	// }
	// _ = tree
	return retItems, nil
}

func (a *AzureClient) GetLockfilesByProject(projectId int64, mainBranchName string) ([]*structs.VcsFile, error) {
	var retLockfiles []*structs.VcsFile

	return retLockfiles, nil
}
