package PhylumSyringGitlab

import (
	"github.com/xanzy/go-gitlab"
	"reflect"
	"testing"
)

func TestNewSyringe(t *testing.T) {
	type args struct {
		gitlabToken string
	}
	tests := []struct {
		name    string
		args    args
		want    *Syringe
		wantErr bool
	}{
		{"one", args{"***REMOVED***"}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSyringe(tt.args.gitlabToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSyringe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(&Syringe{}) {
				t.Errorf("NewSyringe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyringe_ListProjects(t *testing.T) {
	type fields struct {
		Gitlab *gitlab.Client
	}
	s, _ := NewSyringe("***REMOVED***")
	tests := []struct {
		name    string
		fields  fields
		want    []*gitlab.Project
		len     int
		wantErr bool
	}{
		{"one", fields{s.Gitlab}, nil, 3, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Syringe{
				Gitlab: tt.fields.Gitlab,
			}
			got, err := s.ListProjects()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("ListProjects() got = %v, want %v", got, tt.want)
			}

			if len(got) != tt.len {
				t.Errorf("ListProjects() len(got) = %v, want %v", len(got), tt.want)
			}
		})
	}
}
