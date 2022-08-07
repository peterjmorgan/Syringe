package syringe

import (
	"os"
	"reflect"
	"testing"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
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
		err = os.Unsetenv("PHYLUM_TOKEN")
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
		wantTwo []*GitlabFile
		wantLen int
		wantErr bool
	}{
		{"one", args{31479523}, nil, nil, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotTwo, err := s.EnumerateTargetFiles(tt.args.projectId)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnumerateTargetFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("EnumerateTargetFiles() got = %v, want %v", got, tt.want)
			}
			if reflect.TypeOf(gotTwo) != reflect.TypeOf(tt.wantTwo) {
				t.Errorf("EnumerateTargetFiles() got = %v, want %v", gotTwo, tt.wantTwo)
			}
			if len(got) != tt.wantLen {
				t.Errorf("EnumerateTargetFiles() got = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func TestSyringe_PhylumGetProjectMap(t *testing.T) {
	tearDown := setupEnv(t)
	defer tearDown(t)

	s, _ := NewSyringe()

	tests := []struct {
		name    string
		want    map[string]PhylumProject
		wantErr bool
	}{
		{"one", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.PhylumGetProjectMap()
			if (err != nil) != tt.wantErr {
				t.Errorf("PhylumGetProjectMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("PhylumGetProjectMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyringe_PhylumCreateProjectsFromList(t *testing.T) {
	tearDown := setupEnv(t)
	defer tearDown(t)
	s, _ := NewSyringe()

	type args struct {
		projectsToCreate []string
	}
	tests := []struct {
		name    string
		args    args
		want    []PhylumProject
		wantLen int
		wantErr bool
	}{
		{"one", args{[]string{
			"ZXY-test4",
			"ZXY-test5",
		}}, nil, 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.PhylumCreateProjectsFromList(tt.args.projectsToCreate)
			if (err != nil) != tt.wantErr {
				t.Errorf("PhylumCreateProjectsFromList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("PhylumCreateProjectsFromList() got = %v, want %v", got, tt.want)
			}
			if len(got) != tt.wantLen {
				t.Errorf("PhylumCreateProjectsFromList() gotLen = %v, wantLen %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestSyringe_PhylumRunAnalyze(t *testing.T) {
	tearDown := setupEnv(t)
	defer tearDown(t)
	s, _ := NewSyringe()

	type args struct {
		phylumProjectFile PhylumProject
		lockfile          *GitlabFile
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"one",
			args{
				PhylumProject{
					ID:        "70376347-6295-4bf8-969c-458f04ab92f0",
					Name:      "ZXY-test1",
					UpdatedAt: "2022-07-10T13:55:25.231350500-07:00",
				},
				&GitlabFile{
					Name:    "package-lock.json",
					Path:    "package-lock.json",
					Id:      "38029830",
					Content: []byte("{\n  \"name\": \"foo-pkg\",\n  \"version\": \"1.0.0\",\n  \"lockfileVersion\": 2,\n  \"requires\": true,\n  \"packages\": {\n    \"\": {\n      \"name\": \"foo-pkg\",\n      \"version\": \"1.0.0\",\n      \"license\": \"ISC\",\n      \"dependencies\": {\n        \"ansi-red\": \"0.1.1\",\n        \"axios\": \"0.19.0\",\n        \"ini\": \"1.3.5\"\n      }\n    },\n    \"node_modules/ansi-red\": {\n      \"version\": \"0.1.1\",\n      \"resolved\": \"https://registry.npmjs.org/ansi-red/-/ansi-red-0.1.1.tgz\",\n      \"integrity\": \"sha1-jGOPnRCAgAo1PJwoyKgcpHBdlGw=\",\n      \"dependencies\": {\n        \"ansi-wrap\": \"0.1.0\"\n      },\n      \"engines\": {\n        \"node\": \">=0.10.0\"\n      }\n    },\n    \"node_modules/ansi-wrap\": {\n      \"version\": \"0.1.0\",\n      \"resolved\": \"https://registry.npmjs.org/ansi-wrap/-/ansi-wrap-0.1.0.tgz\",\n      \"integrity\": \"sha1-qCJQ3bABXponyoLoLqYDu/pF768=\",\n      \"engines\": {\n        \"node\": \">=0.10.0\"\n      }\n    },\n    \"node_modules/axios\": {\n      \"version\": \"0.19.0\",\n      \"resolved\": \"https://registry.npmjs.org/axios/-/axios-0.19.0.tgz\",\n      \"integrity\": \"sha512-1uvKqKQta3KBxIz14F2v06AEHZ/dIoeKfbTRkK1E5oqjDnuEerLmYTgJB5AiQZHJcljpg1TuRzdjDR06qNk0DQ==\",\n      \"deprecated\": \"Critical security vulnerability fixed in v0.21.1. For more information, see https://github.com/axios/axios/pull/3410\",\n      \"dependencies\": {\n        \"follow-redirects\": \"1.5.10\",\n        \"is-buffer\": \"^2.0.2\"\n      }\n    },\n    \"node_modules/debug\": {\n      \"version\": \"3.1.0\",\n      \"resolved\": \"https://registry.npmjs.org/debug/-/debug-3.1.0.tgz\",\n      \"integrity\": \"sha512-OX8XqP7/1a9cqkxYw2yXss15f26NKWBpDXQd0/uK/KPqdQhxbPa994hnzjcE2VqQpDslf55723cKPUOGSmMY3g==\",\n      \"dependencies\": {\n        \"ms\": \"2.0.0\"\n      }\n    },\n    \"node_modules/follow-redirects\": {\n      \"version\": \"1.5.10\",\n      \"resolved\": \"https://registry.npmjs.org/follow-redirects/-/follow-redirects-1.5.10.tgz\",\n      \"integrity\": \"sha512-0V5l4Cizzvqt5D44aTXbFZz+FtyXV1vrDN6qrelxtfYQKW0KO0W2T/hkE8xvGa/540LkZlkaUjO4ailYTFtHVQ==\",\n      \"dependencies\": {\n        \"debug\": \"=3.1.0\"\n      },\n      \"engines\": {\n        \"node\": \">=4.0\"\n      }\n    },\n    \"node_modules/ini\": {\n      \"version\": \"1.3.5\",\n      \"resolved\": \"https://registry.npmjs.org/ini/-/ini-1.3.5.tgz\",\n      \"integrity\": \"sha512-RZY5huIKCMRWDUqZlEi72f/lmXKMvuszcMBduliQ3nnWbx9X/ZBQO7DijMEYS9EhHBb2qacRUMtC7svLwe0lcw==\",\n      \"deprecated\": \"Please update to ini >=1.3.6 to avoid a prototype pollution issue\",\n      \"engines\": {\n        \"node\": \"*\"\n      }\n    },\n    \"node_modules/is-buffer\": {\n      \"version\": \"2.0.5\",\n      \"resolved\": \"https://registry.npmjs.org/is-buffer/-/is-buffer-2.0.5.tgz\",\n      \"integrity\": \"sha512-i2R6zNFDwgEHJyQUtJEk0XFi1i0dPFn/oqjK3/vPCcDeJvW5NQ83V8QbicfF1SupOaB0h8ntgBC2YiE7dfyctQ==\",\n      \"funding\": [\n        {\n          \"type\": \"github\",\n          \"url\": \"https://github.com/sponsors/feross\"\n        },\n        {\n          \"type\": \"patreon\",\n          \"url\": \"https://www.patreon.com/feross\"\n        },\n        {\n          \"type\": \"consulting\",\n          \"url\": \"https://feross.org/support\"\n        }\n      ],\n      \"engines\": {\n        \"node\": \">=4\"\n      }\n    },\n    \"node_modules/ms\": {\n      \"version\": \"2.0.0\",\n      \"resolved\": \"https://registry.npmjs.org/ms/-/ms-2.0.0.tgz\",\n      \"integrity\": \"sha1-VgiurfwAvmwpAd9fmGF4jeDVl8g=\"\n    }\n  },\n  \"dependencies\": {\n    \"ansi-red\": {\n      \"version\": \"0.1.1\",\n      \"resolved\": \"https://registry.npmjs.org/ansi-red/-/ansi-red-0.1.1.tgz\",\n      \"integrity\": \"sha1-jGOPnRCAgAo1PJwoyKgcpHBdlGw=\",\n      \"requires\": {\n        \"ansi-wrap\": \"0.1.0\"\n      }\n    },\n    \"ansi-wrap\": {\n      \"version\": \"0.1.0\",\n      \"resolved\": \"https://registry.npmjs.org/ansi-wrap/-/ansi-wrap-0.1.0.tgz\",\n      \"integrity\": \"sha1-qCJQ3bABXponyoLoLqYDu/pF768=\"\n    },\n    \"axios\": {\n      \"version\": \"0.19.0\",\n      \"resolved\": \"https://registry.npmjs.org/axios/-/axios-0.19.0.tgz\",\n      \"integrity\": \"sha512-1uvKqKQta3KBxIz14F2v06AEHZ/dIoeKfbTRkK1E5oqjDnuEerLmYTgJB5AiQZHJcljpg1TuRzdjDR06qNk0DQ==\",\n      \"requires\": {\n        \"follow-redirects\": \"1.5.10\",\n        \"is-buffer\": \"^2.0.2\"\n      }\n    },\n    \"debug\": {\n      \"version\": \"3.1.0\",\n      \"resolved\": \"https://registry.npmjs.org/debug/-/debug-3.1.0.tgz\",\n      \"integrity\": \"sha512-OX8XqP7/1a9cqkxYw2yXss15f26NKWBpDXQd0/uK/KPqdQhxbPa994hnzjcE2VqQpDslf55723cKPUOGSmMY3g==\",\n      \"requires\": {\n        \"ms\": \"2.0.0\"\n      }\n    },\n    \"follow-redirects\": {\n      \"version\": \"1.5.10\",\n      \"resolved\": \"https://registry.npmjs.org/follow-redirects/-/follow-redirects-1.5.10.tgz\",\n      \"integrity\": \"sha512-0V5l4Cizzvqt5D44aTXbFZz+FtyXV1vrDN6qrelxtfYQKW0KO0W2T/hkE8xvGa/540LkZlkaUjO4ailYTFtHVQ==\",\n      \"requires\": {\n        \"debug\": \"=3.1.0\"\n      }\n    },\n    \"ini\": {\n      \"version\": \"1.3.5\",\n      \"resolved\": \"https://registry.npmjs.org/ini/-/ini-1.3.5.tgz\",\n      \"integrity\": \"sha512-RZY5huIKCMRWDUqZlEi72f/lmXKMvuszcMBduliQ3nnWbx9X/ZBQO7DijMEYS9EhHBb2qacRUMtC7svLwe0lcw==\"\n    },\n    \"is-buffer\": {\n      \"version\": \"2.0.5\",\n      \"resolved\": \"https://registry.npmjs.org/is-buffer/-/is-buffer-2.0.5.tgz\",\n      \"integrity\": \"sha512-i2R6zNFDwgEHJyQUtJEk0XFi1i0dPFn/oqjK3/vPCcDeJvW5NQ83V8QbicfF1SupOaB0h8ntgBC2YiE7dfyctQ==\"\n    },\n    \"ms\": {\n      \"version\": \"2.0.0\",\n      \"resolved\": \"https://registry.npmjs.org/ms/-/ms-2.0.0.tgz\",\n      \"integrity\": \"sha1-VgiurfwAvmwpAd9fmGF4jeDVl8g=\"\n    }\n  }\n}"),
				},
			},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.PhylumRunAnalyze(tt.args.phylumProjectFile, tt.args.lockfile)
			if (err != nil) != tt.wantErr {
				t.Errorf("PhylumRunAnalyze() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSyringe_LoadPidFile(t *testing.T) {
	type fields struct {
		Gitlab          *gitlab.Client
		PhylumToken     string
		PhylumGroupName string
	}
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Syringe{
				Gitlab:          tt.fields.Gitlab,
				PhylumToken:     tt.fields.PhylumToken,
				PhylumGroupName: tt.fields.PhylumGroupName,
			}
			if err := s.LoadPidFile(tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("LoadPidFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
