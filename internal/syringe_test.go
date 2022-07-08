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

// 31479523
func TestSyringe_ListFiles(t *testing.T) {
	s, _ := NewSyringe("***REMOVED***")
	type fields struct {
		Gitlab *gitlab.Client
	}
	type args struct {
		projectId int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*gitlab.TreeNode
		len     int
		wantErr bool
	}{
		{"one", fields{s.Gitlab}, args{31479523}, nil, 5, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Syringe{
				Gitlab: tt.fields.Gitlab,
			}
			got, err := s.ListFiles(tt.args.projectId, "main")
			if (err != nil) != tt.wantErr {
				t.Errorf("ListFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("ListFiles() got = %v, want %v", got, tt.want)
			}
			if len(got) != tt.len {
				t.Errorf("ListFiles() got = %v, want %v", len(got), tt.len)
			}
		})
	}
}

func TestSyringe_ListBranches(t *testing.T) {
	s, _ := NewSyringe("***REMOVED***")
	type fields struct {
		Gitlab *gitlab.Client
	}
	type args struct {
		projectId int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*gitlab.Branch
		len     int
		wantErr bool
	}{
		{"one", fields{s.Gitlab}, args{31479523}, nil, 20, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Syringe{
				Gitlab: tt.fields.Gitlab,
			}
			got, err := s.ListBranches(tt.args.projectId)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListBranches() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("ListBranches() got = %v, want %v", got, tt.want)
			}
			if len(got) != tt.len {
				t.Errorf("ListBranches() got = %v, want %v", len(got), tt.len)
			}
		})
	}
}

func TestSyringe_IdentifyMainBranch(t *testing.T) {
	s, _ := NewSyringe("***REMOVED***")
	type fields struct {
		Gitlab *gitlab.Client
	}
	type args struct {
		projectId int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *gitlab.Branch
		wantErr bool
	}{
		{"one", fields{s.Gitlab}, args{31479523}, nil, false},
		{"two", fields{s.Gitlab}, args{0}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Syringe{
				Gitlab: tt.fields.Gitlab,
			}
			got, err := s.IdentifyMainBranch(tt.args.projectId)
			if (err != nil) != tt.wantErr {
				t.Errorf("IdentifyMainBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("IdentifyMainBranch() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyringe_GetFileTreeFromProject(t *testing.T) {
	s, _ := NewSyringe("***REMOVED***")
	type fields struct {
		Gitlab *gitlab.Client
	}
	type args struct {
		projectId int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*gitlab.TreeNode
		wantLen int
		wantErr bool
	}{
		{"one", fields{s.Gitlab}, args{31479523}, nil, 5, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Syringe{
				Gitlab: tt.fields.Gitlab,
			}
			got, err := s.GetFileTreeFromProject(tt.args.projectId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFileTreeFromProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("GetFileTreeFromProject() got = %v, want %v", got, tt.want)
			}
			if len(got) != tt.wantLen {
				t.Errorf("GetFileTreeFromProject() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyringe_EnumerateTargetFiles(t *testing.T) {
	s, _ := NewSyringe("***REMOVED***")
	type fields struct {
		Gitlab *gitlab.Client
	}
	type args struct {
		projectId int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*GitlabFile
		wantLen int
		wantErr bool
	}{
		{"one", fields{s.Gitlab}, args{31479523}, nil, 5, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Syringe{
				Gitlab: tt.fields.Gitlab,
			}
			got, err := s.EnumerateTargetFiles(tt.args.projectId)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnumerateTargetFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("EnumerateTargetFiles() got = %v, want %v", got, tt.want)
			}
			if len(got) != tt.wantLen {
				t.Errorf("EnumerateTargetFiles() got = %v, want %v", got, tt.want)
			}
		})
	}
}
