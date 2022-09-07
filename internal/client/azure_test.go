package client

import (
	"fmt"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/peterjmorgan/Syringe/internal/structs"
	"github.com/peterjmorgan/Syringe/internal/utils"
	"reflect"
	"testing"
)

func TestNewAzureClient(t *testing.T) {
	configData, err := utils.ReadConfigFile(&structs.TestConfigData{})
	if err != nil {
		fmt.Printf("failed to read config file: %v\n", err)
	}
	a := NewAzureClient(configData, &testingSyringeOpts)
	_ = a
}

func TestAzureClient_ListProjects(t *testing.T) {
	configData, err := utils.ReadConfigFile(&structs.TestConfigData{})
	if err != nil {
		fmt.Printf("failed to read config file: %v\n", err)
	}
	a := NewAzureClient(configData, &testingSyringeOpts)

	tests := []struct {
		name    string
		want    *[]*structs.SyringeProject
		wantLen int
		wantErr bool
	}{
		{"one", nil, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := a.ListProjects()
			if (err != nil) != tt.wantErr {
				t.Errorf("Azure_ListProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("Azure_ListProjects() TypeOf got = %v, want %v", got, tt.want)
			}

			if len(*got) != tt.wantLen {
				t.Errorf("Azure_ListProjects() len(projects) got = %v, want %v", len(*got), tt.wantLen)
			}
		})
	}
}

func TestAzureClient_ListFiles(t *testing.T) {
	configData, err := utils.ReadConfigFile(&structs.TestConfigData{})
	if err != nil {
		fmt.Printf("failed to read config file: %v\n", err)
	}
	a := NewAzureClient(configData, &testingSyringeOpts)

	type args struct {
		repoID string
		branch string
	}

	tests := []struct {
		name    string
		args    args
		want    []*git.GitItem
		wantLen int
		wantErr bool
	}{
		{"one", args{"4a7912a2-88a1-4697-9cf3-3919e9e07ddf", "master"}, []*git.GitItem{}, 5, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := a.ListFiles(tt.args.repoID, tt.args.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("Azure_ListFiles() TypeOf got = %v, want %v", got, tt.want)
			}

			if len(got) != tt.wantLen {
				t.Errorf("Azure_ListFiles() len(projects) got = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestAzureClient_GetLockfilesByProject(t *testing.T) {
	configData, err := utils.ReadConfigFile(&structs.TestConfigData{})
	if err != nil {
		fmt.Printf("failed to read config file: %v\n", err)
	}
	a := NewAzureClient(configData, &testingSyringeOpts)

	// populate with projects
	_, _ = a.ListProjects()

	type args struct {
		projectId      int64
		mainBranchName string
	}

	tests := []struct {
		name    string
		args    args
		want    []*structs.VcsFile
		wantLen int
		wantErr bool
	}{
		{"one", args{1249448610, "master"}, nil, 4, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := a.GetLockfilesByProject(tt.args.projectId, tt.args.mainBranchName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLockfilesByProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("Azure_GetLockfilesByProject() TypeOf got = %v, want %v", got, tt.want)
			}

			if len(got) != tt.wantLen {
				t.Errorf("Azure_GetLockfilesByProject() len(projects) got = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}
