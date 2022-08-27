package client

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/peterjmorgan/Syringe/internal/utils"

	"github.com/joho/godotenv"
	"github.com/peterjmorgan/Syringe/internal/structs"
	log "github.com/sirupsen/logrus"
)

func setupEnv(t *testing.T, envFilename string) {
	filename := fmt.Sprintf("../../.env_%v", envFilename)

	err := godotenv.Load(filename)
	if err != nil {
		log.Fatalf("Failed to load .env for testing: %v\n", err)
	}
}

var testingSyringeOpts structs.SyringeOptions = structs.SyringeOptions{
	MineOnly:  true,
	RateLimit: 0,
	ProxyUrl:  "",
}

func TestGitlabClient_ListProjects(t *testing.T) {
	setupEnv(t, "gitlab")
	envMap, _ := utils.ReadEnvironment()
	g := NewGitlabClient(envMap, &testingSyringeOpts)

	tests := []struct {
		name    string
		want    *[]*structs.SyringeProject
		wantLen int
		wantErr bool
	}{
		{"one", nil, 213, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := g.ListProjects()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("GitlabClient_ListProjects() TypeOf got = %v, want %v", got, tt.want)
			}

			if len(*got) != tt.wantLen {
				t.Errorf("GitlabClient_ListProjects() len(projects) got = %v, want %v", len(*got), tt.wantLen)
			}
		})
	}
}

func TestGitlabClient_GetLockfiles(t *testing.T) {
	setupEnv(t, "gitlab")
	envMap, _ := utils.ReadEnvironment()

	g := NewGitlabClient(envMap, &testingSyringeOpts)
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
		{"one", args{38265422, "master"}, nil, 4, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := g.GetLockfilesByProject(tt.args.projectId, tt.args.mainBranchName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLockfilesByProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("GitlabClient_GetLockfiles() TypeOf got = %v, want %v", got, tt.want)
			}

			if len(got) != tt.wantLen {
				t.Errorf("GitlabClient_GetLockfiles() len(projects) got = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}
