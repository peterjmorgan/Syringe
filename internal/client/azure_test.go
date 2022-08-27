package client

import (
	"reflect"
	"testing"

	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/peterjmorgan/Syringe/internal/structs"
	"github.com/peterjmorgan/Syringe/internal/utils"
)

func TestNewAzureClient(t *testing.T) {
	setupEnv(t, "azure")
	envMap, _ := utils.ReadEnvironment()
	a := NewAzureClient(envMap, &testingSyringeOpts)
	_ = a

}

func TestAzureClient_ListProjects(t *testing.T) {
	setupEnv(t, "azure")
	envMap, _ := utils.ReadEnvironment()
	a := NewAzureClient(envMap, &testingSyringeOpts)

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
	setupEnv(t, "azure")
	envMap, _ := utils.ReadEnvironment()
	a := NewAzureClient(envMap, &testingSyringeOpts)

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
