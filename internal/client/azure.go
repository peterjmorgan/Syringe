package client

import (
	"context"
	"fmt"
	"github.com/microsoft/azure-devops-go-api/azuredevops/build"
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
}

func NewAzureClient(envMap map[string]string, opts *structs.SyringeOptions) *AzureClient {
	//var connUrl string
	//
	//if vcsUrl, ok := envMap["vcsUrl"]; ok {
	//	connUrl = vcsUrl
	//}

	org := envMap["vcsOrg"]
	conn := azuredevops.NewPatConnection(org, envMap["vcsToken"])
	ctx := context.Background()
	coreClient, err := core.NewClient(ctx, conn)
	buildClient, err := build.NewClient(ctx, conn)
	if err != nil {
		// handle
	}
	return &AzureClient{
		Clients: &AzureSubClient{
			CoreClient:  coreClient,
			BuildClient: buildClient,
		},
		Ctx: ctx,
	}
}

func (a *AzureClient) ListProjects() (*[]*structs.SyringeProject, error) {
	var localProjects []core.TeamProjectReference
	var retProjects []*structs.SyringeProject

	projectResp, err := a.Clients.CoreClient.GetProjects(a.Ctx, core.GetProjectsArgs{})
	if err != nil {
		return nil, fmt.Errorf("Failed to GetProjects: %v\n", err)
	}

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
	for _, proj := range localProjects {

		// TODO: get the branch of the project
		branches, err := a.Clients.BuildClient.ListBranches(a.Ctx, build.ListBranchesArgs{
			Project:           proj.Name,
			ProviderName:
		})
		_ = branches

		if err != nil {
			errStr := fmt.Sprintf("failed to list branches for %v:%v\n", proj.Name, err)
			log.Error(errStr)
			return nil, fmt.Errorf(errStr)
		}

		retProjects = append(retProjects, &structs.SyringeProject{
			Id:        int64(proj.Id.ID()),
			Name:      *proj.Name,
			Branch:    "",
			Lockfiles: nil,
			CiFiles:   nil,
			Hydrated:  false,
		})
	}
	return &retProjects, nil
}
