package client

import (
	"github.com/peterjmorgan/Syringe/internal/structs"
	"github.com/peterjmorgan/Syringe/internal/utils"
	"reflect"
	"testing"
)

func TestBitbucketCloudClient_ListProjects(t *testing.T) {
	setupEnv(t, "bitbucket")
	envMap, _ := utils.ReadEnvironment()
	b := NewBitbucketCloudClient(envMap, &structs.SyringeOptions{})

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
