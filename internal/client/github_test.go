package client

import (
	"reflect"
	"testing"

	"github.com/peterjmorgan/Syringe/internal/utils"

	"github.com/peterjmorgan/Syringe/internal/structs"

	"github.com/google/go-github/github"
)

func TestGithubClient_ListProjects(t *testing.T) {
	setupEnv(t, "github")
	envMap, _ := utils.ReadEnvironment()
	g := NewGithubClient(envMap, &structs.SyringeOptions{})

	tests := []struct {
		name    string
		want    *[]*structs.SyringeProject
		wantLen int
		wantErr bool
	}{
		{"phylum-dev", nil, 58, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := g.ListProjects()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("GithubClient_ListProjects() TypeOf got = %v, want %v", got, tt.want)
			}

			if len(*got) != tt.wantLen {
				t.Errorf("GithubClient_ListProjects() len(projects) got = %v, want %v", len(*got), tt.wantLen)
			}
		})
	}
}

func TestGithubClient_ListFiles(t *testing.T) {
	setupEnv(t, "github")
	envMap, _ := utils.ReadEnvironment()
	g := NewGithubClient(envMap, &structs.SyringeOptions{})

	g.OrgName = "phylum-dev"

	type args struct {
		repoName string
		branch   string
	}

	tests := []struct {
		name    string
		args    args
		want    *github.Tree
		wantLen int
		wantErr bool
	}{
		{"phylum-dev/cli", args{"cli", "main"}, nil, 191, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := g.ListFiles(tt.args.repoName, tt.args.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("GithubClient_ListFiles() TypeOf got = %v, want %v", got, tt.want)
			}

			if len(got.Entries) != tt.wantLen {
				t.Errorf("GithubClient_ListFiles() len(files) got = %v, want %v", len(got.Entries), tt.wantLen)
			}
		})
	}
}

func TestGithubClient_GetLockfilesByProject(t *testing.T) {
	setupEnv(t, "github")
	envMap, _ := utils.ReadEnvironment()
	g := NewGithubClient(envMap, &structs.SyringeOptions{})
	g.OrgName = "phylum-dev"

	type args struct {
		projectId int64
		branch    string
	}

	tests := []struct {
		name    string
		args    args
		want    []*structs.VcsFile
		wantLen int
		wantErr bool
	}{
		// Only has 1 yarn.lock
		{"phylum-dev/phylum-ui", args{325083799, "main"}, nil, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := g.GetLockfilesByProject(tt.args.projectId, tt.args.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLockfilesByProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("GithubClient_GetLockfilesByProject() TypeOf got = %v, want %v", got, tt.want)
			}

			if len(got) != tt.wantLen {
				t.Errorf("GithubClient_GetLockfilesByProject() len(files) got = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}
