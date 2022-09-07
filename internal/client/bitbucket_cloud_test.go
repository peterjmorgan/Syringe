package client

import (
	"fmt"
	"github.com/ktrysmt/go-bitbucket"
	"github.com/peterjmorgan/Syringe/internal/structs"
	"github.com/peterjmorgan/Syringe/internal/utils"
	"reflect"
	"testing"
)

func TestBitbucketCloudClient_ListProjects(t *testing.T) {
	configData, err := utils.ReadConfigFile(&structs.TestConfigData{})
	if err != nil {
		fmt.Printf("failed to read config file: %v\n", err)
	}
	b := NewBitbucketCloudClient(configData, &testingSyringeOpts)

	tests := []struct {
		name    string
		want    *[]*structs.SyringeProject
		wantErr bool
	}{
		{"one", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := b.ListProjects()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListProjects() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBitbucketCloudClient_ListFiles(t *testing.T) {
	configData, err := utils.ReadConfigFile(&structs.TestConfigData{})
	if err != nil {
		fmt.Printf("failed to read config file: %v\n", err)
	}
	b := NewBitbucketCloudClient(configData, &testingSyringeOpts)

	type args struct {
		repoSlug string
		branch   string
	}
	tests := []struct {
		name    string
		args    args
		want    *[]*bitbucket.RepositoryFile
		wantLen int
		wantErr bool
	}{
		{"one", args{"beta", "master"}, nil, 4, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := b.ListFiles(tt.args.repoSlug, tt.args.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("BitbucketCloud_ListProjects() TypeOf got = %v, want %v", got, tt.want)
			}

			if len(*got) != tt.wantLen {
				t.Errorf("BitbucketCloud_ListProjects() len(projects) got = %v, want %v", len(*got), tt.wantLen)
			}
		})
	}
}

func TestBitbucketCloudClient_GetLockfilesByProject(t *testing.T) {
	configData, err := utils.ReadConfigFile(&structs.TestConfigData{})
	if err != nil {
		fmt.Printf("failed to read config file: %v\n", err)
	}
	b := NewBitbucketCloudClient(configData, &testingSyringeOpts)

	_, _ = b.ListProjects()

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
		{"one", args{3407435279, "master"}, nil, 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := b.GetLockfilesByProject(tt.args.projectId, tt.args.mainBranchName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLockfilesByProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("BitbucketCloud_GetLockfilesByProject() TypeOf got = %v, want %v", got, tt.want)
			}

			if len(got) != tt.wantLen {
				t.Errorf("BitbucketCloud_GetLockfilesByProject() len(projects) got = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}
