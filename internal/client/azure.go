package client

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/microsoft/azure-devops-go-api/azuredevops/build"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/peterjmorgan/Syringe/internal/utils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"github.com/peterjmorgan/Syringe/internal/structs"
)

type AzureClient struct {
	Clients         *AzureSubClient
	Ctx             context.Context
	OrgName         string
	ProjectMap      map[int64]*git.GitRepository
	ProjectMapMutex sync.RWMutex
}

type AzureSubClient struct {
	CoreClient  core.Client
	BuildClient build.Client
	GitClient   git.Client
}

// func NewAzureClient(envMap map[string]string, opts *structs.SyringeOptions) *AzureClient {
func NewAzureClient(configData *structs.ConfigThing, opts *structs.SyringeOptions) *AzureClient {

	// TODO: handle alternate urls to ADO
	//org := envMap["vcsOrg"]
	//conn := azuredevops.NewPatConnection(org, envMap["vcsToken"])
	conn := azuredevops.NewPatConnection(configData.Associated["azureOrg"], configData.VcsToken)
	ctx := context.Background()
	coreClient, err := core.NewClient(ctx, conn)
	if err != nil {
		log.Fatalf("NewAzureClient: Failed to create coreClient: %v\n", err)
	}
	buildClient, err := build.NewClient(ctx, conn)
	if err != nil {
		log.Fatalf("NewAzureClient: Failed to create buildClient: %v\n", err)
	}
	gitClient, err := git.NewClient(ctx, conn)
	if err != nil {
		log.Fatalf("NewAzureClient: Failed to create gitClient: %v\n", err)
	}

	return &AzureClient{
		Clients: &AzureSubClient{
			CoreClient:  coreClient,
			BuildClient: buildClient,
			GitClient:   gitClient,
		},
		Ctx:        ctx,
		ProjectMap: make(map[int64]*git.GitRepository, 0),
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
			temp := new(git.GitRepository)
			*temp = repo
			a.ProjectMap[int64(repo.Id.ID())] = temp
		}
	}
	return &retProjects, nil
}

func (a *AzureClient) ListFiles(repoID string, branch string) ([]*git.GitItem, error) {
	var retItems []*git.GitItem

	var recurse git.VersionControlRecursionType = "full"
	var versionType git.GitVersionType = git.GitVersionType("branch")

	items, err := a.Clients.GitClient.GetItems(a.Ctx, git.GetItemsArgs{
		RepositoryId:           &repoID,
		RecursionLevel:         &recurse,
		IncludeContentMetadata: &[]bool{true}[0],
		VersionDescriptor: &git.GitVersionDescriptor{
			Version:        &branch,
			VersionOptions: nil,
			VersionType:    &versionType,
		},
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

	return retItems, nil
}

// This is a little messed up because i'm calling repos "projects" and those don't match up in ADOland
// This should be okay. ListProjects() creates a SyringeProject for each repo in an ADO project.
// TODO: consider renaming SyringeProject to SyringeRepository as that's a better term for the struct
func (a *AzureClient) GetLockfilesByProject(projectId int64, mainBranchName string) ([]*structs.VcsFile, error) {
	var retLockfiles []*structs.VcsFile

	a.ProjectMapMutex.RLock()
	repo := a.ProjectMap[projectId]
	a.ProjectMapMutex.RUnlock()

	guid := repo.Id.String()

	if strings.Contains(mainBranchName, "/") {
		mainBranchName = filepath.Base(mainBranchName)
	}

	projectFiles, err := a.ListFiles(guid, mainBranchName)
	if err != nil {
		errStr := fmt.Sprintf("failed to GetLockfilesByProject for %v: %v\n", guid, err)
		log.Error(errStr)
		return nil, fmt.Errorf(errStr)
	}

	supportedLockfiles := utils.GetSupportedLockfiles()

	for _, file := range projectFiles {
		fileName := filepath.Base(*file.Path)
		if slices.Contains(supportedLockfiles, fileName) || strings.HasSuffix(fileName, ".csproj") {
			log.Debugf("Lockfile: %v in %v from project: %v\n", fileName, *file.Path, *repo.Name)
			// download te file
			item, err := a.Clients.GitClient.GetItem(a.Ctx, git.GetItemArgs{
				RepositoryId:   &guid,
				Path:           file.Path,
				IncludeContent: &[]bool{true}[0],
			})
			if err != nil {
				errStr := fmt.Sprintf("failed to GetItem for %v: %v\n", fileName, err)
				log.Error(errStr)
				return nil, fmt.Errorf(errStr)
			}

			retLockfiles = append(retLockfiles, &structs.VcsFile{
				Name:          fileName,
				Path:          *file.Path,
				Id:            *file.CommitId, // TODO: look at this
				Content:       []byte(*item.Content),
				PhylumProject: nil,
			})
		}
	}

	return retLockfiles, nil
}
