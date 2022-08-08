package client

import (
	"os"
	"reflect"
	"testing"

	"github.com/peterjmorgan/Syringe/internal/structs"
)

func TestGithubClient_ListProjects(t *testing.T) {
	tearDown := setupEnv(t, "SYRINGE_VCS_TOKEN_GITHUB")
	defer tearDown(t)

	g := NewGithubClient(os.Getenv("SYRINGE_VCS_TOKEN_GITHUB"), os.Getenv("SYRINGE_ORG_GITHUB"))

	tests := []struct {
		name    string
		want    *[]*structs.SyringeProject
		wantLen int
		wantErr bool
	}{
		{"phylum-dev", nil, 57, false},
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
				t.Errorf("GithubClient_ListProjects() len(projects) got = %v, want %v", len(*got), tt.want)
			}
		})
	}
}
