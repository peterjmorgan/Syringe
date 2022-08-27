package client

import (
	"context"
	"fmt"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"github.com/peterjmorgan/Syringe/internal/structs"
)

type AzureClient struct {
	Client  core.Client
	Ctx     context.Context
	OrgName string
}

func NewAzureClient(envMap map[string]string, opts *structs.SyringeOptions) *AzureClient {
	var connUrl string

	if vcsUrl, ok := envMap["vcsUrl"]; ok {
		connUrl = vcsUrl
	}
	conn := azuredevops.NewPatConnection(connUrl, envMap["vcsToken"])
	ctx := context.Background()
	client, err := core.NewClient(ctx, conn)
	if err != nil {
		// handle
	}
	return &AzureClient{
		Client: client,
		Ctx:    ctx,
	}
}

func (a *AzureClient) ListProjects() (*[]*structs.SyringeProject, error) {
	var localProjects []core.TeamProjectReference
	var retProjects []*structs.SyringeProject

	projectResp, err := a.Client.GetProjects(a.Ctx, core.GetProjectsArgs{})
	if err != nil {
		return nil, fmt.Errorf("Failed to GetProjects: %v\n")
	}

	for projectResp != nil {
		for _, proj := range (*projectResp).Value {
			localProjects = append(localProjects, proj)
		}
		if projectResp.ContinuationToken != "" {
			projectArgs := core.GetProjectsArgs{
				ContinuationToken: &projectResp.ContinuationToken,
			}
			projectResp, err = a.Client.GetProjects(a.Ctx, projectArgs)
			if err != nil {
				return nil, fmt.Errorf("Failed to GetProjects (cont) %v\n", err)
			}
		} else {
			projectResp = nil
		}
	}
	for _, proj := range localProjects {
		a.Client.A

		retProjects = append(retProjects, structs.SyringeProject{
			Id:        *proj.Id,
			Name:      *proj.Name,
			Branch:    "",
			Lockfiles: nil,
			CiFiles:   nil,
			Hydrated:  false,
		})
	}
}
