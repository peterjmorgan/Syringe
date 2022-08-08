package client

import (
	"os"
	"reflect"
	"testing"

	"github.com/joho/godotenv"
	"github.com/peterjmorgan/Syringe/internal/structs"
	log "github.com/sirupsen/logrus"
)

func setupEnv(t *testing.T, envVarName string) func(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Failed to load .env for testing: %v\n", err)
	}

	return func(t *testing.T) {
		err := os.Unsetenv(envVarName)
		if err != nil {
			log.Fatalf("Failed to unset %v: %v\n", envVarName, err)
		}
	}
}

func TestGitlabClient_ListProjects(t *testing.T) {
	tearDown := setupEnv(t, "SYRINGE_VCS_TOKEN_GITLAB")
	defer tearDown(t)

	g := NewGitlabClient(os.Getenv("SYRINGE_VCS_TOKEN_GITLAB"), "", true)

	tests := []struct {
		name    string
		want    *[]*structs.SyringeProject
		wantLen int
		wantErr bool
	}{
		{"one", nil, 212, false},
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
				t.Errorf("GitlabClient_ListProjects() len(projects) got = %v, want %v", len(*got), tt.want)
			}
		})
	}
}

func TestGitlabClient_GetLockfiles(t *testing.T) {
	tearDown := setupEnv(t, "SYRINGE_VCS_TOKEN_GITLAB")
	defer tearDown(t)

	g := NewGitlabClient(os.Getenv("SYRINGE_VCS_TOKEN_GITLAB"), "", true)
	type args struct {
		projectId      int64
		mainBranchName string
	}
	tests := []struct {
		name    string
		args    args
		want    []*structs.VcsFile
		wantErr bool
	}{
		{"one", args{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := g.GetLockfiles(tt.args.projectId, tt.args.mainBranchName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLockfiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLockfiles() got = %v, want %v", got, tt.want)
			}
		})
	}
}
