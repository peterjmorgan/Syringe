package PhylumSyringGitlab

import (
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"os"
	"reflect"
	"testing"
)

func setupEnv(t *testing.T) func(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Failed to load .env for testing: %v\n", err)
	}

	return func(t *testing.T) {
		err := os.Unsetenv("GITLAB_TOKEN")
		if err != nil {
			log.Fatalf("Failed to UNSET gitlab token: %v\n", err)
		}
		os.Unsetenv("PHYLUM_TOKEN")
		if err != nil {
			log.Fatalf("Failed to UNSET phylum token: %v\n", err)
		}
	}
}

func TestNewSyringe(t *testing.T) {
	tearDown := setupEnv(t)
	defer tearDown(t)

	type args struct {
		gitlabToken string
	}
	tests := []struct {
		name    string
		want    *Syringe
		wantErr bool
	}{
		{"one", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSyringe()
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
	tearDown := setupEnv(t)
	defer tearDown(t)

	s, _ := NewSyringe()

	tests := []struct {
		name    string
		want    []*gitlab.Project
		len     int
		wantErr bool
	}{
		{"one", nil, 3, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	tearDown := setupEnv(t)
	defer tearDown(t)

	s, _ := NewSyringe()

	type args struct {
		projectId int
	}
	tests := []struct {
		name    string
		args    args
		want    []*gitlab.TreeNode
		len     int
		wantErr bool
	}{
		{"one", args{31479523}, nil, 5, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	tearDown := setupEnv(t)
	defer tearDown(t)

	s, _ := NewSyringe()

	type args struct {
		projectId int
	}
	tests := []struct {
		name    string
		args    args
		want    []*gitlab.Branch
		len     int
		wantErr bool
	}{
		{"one", args{31479523}, nil, 20, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	tearDown := setupEnv(t)
	defer tearDown(t)

	s, _ := NewSyringe()

	type args struct {
		projectId int
	}
	tests := []struct {
		name    string
		args    args
		want    *gitlab.Branch
		wantErr bool
	}{
		{"one", args{31479523}, nil, false},
		{"two", args{0}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	tearDown := setupEnv(t)
	defer tearDown(t)

	s, _ := NewSyringe()

	type args struct {
		projectId int
	}
	tests := []struct {
		name    string
		args    args
		want    []*gitlab.TreeNode
		wantLen int
		wantErr bool
	}{
		{"one", args{31479523}, nil, 5, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	tearDown := setupEnv(t)
	defer tearDown(t)

	s, _ := NewSyringe()

	type args struct {
		projectId int
	}
	tests := []struct {
		name    string
		args    args
		want    []*GitlabFile
		wantLen int
		wantErr bool
	}{
		{"one", args{31479523}, nil, 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestSyringe_PhylumGetProjectList(t *testing.T) {
	tearDown := setupEnv(t)
	defer tearDown(t)

	s, _ := NewSyringe()

	tests := []struct {
		name    string
		want    []PhylumProject
		wantErr bool
	}{
		{"one", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.PhylumGetProjectList()
			if (err != nil) != tt.wantErr {
				t.Errorf("PhylumGetProjectList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("EnumerateTargetFiles() got = %v, want %v", got, tt.want)
			}
		})
	}
}
