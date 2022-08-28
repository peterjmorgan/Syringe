package syringePackage

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/peterjmorgan/Syringe/internal/structs"
	"github.com/peterjmorgan/Syringe/internal/utils"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func setupEnv(t *testing.T, envFilename string) {
	filename := fmt.Sprintf("../.env_%v", envFilename)

	err := godotenv.Load(filename)
	if err != nil {
		log.Fatalf("Failed to load .env for testing: %v\n", err)
	}
}

var gitlabOpts *structs.SyringeOptions = &structs.SyringeOptions{
	MineOnly:  true,
	RateLimit: 0,
	ProxyUrl:  "",
}

func TestNewSyringe(t *testing.T) {

	tests := []struct {
		name    string
		opts    *structs.SyringeOptions
		want    *Syringe
		wantErr bool
	}{
		// {"gitlab", true, nil, false},
		// {"github", false, nil, false},
		{"gitlab", gitlabOpts, nil, false},
		{"github", &structs.SyringeOptions{}, nil, false},
		{"azure", &structs.SyringeOptions{}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupEnv(t, tt.name)
			envMap, _ := utils.ReadEnvironment()
			got, err := NewSyringe(envMap, tt.opts)
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

	// s, _ := NewSyringe(envMap, true)

	tests := []struct {
		name    string
		opts    *structs.SyringeOptions
		want    *[]*structs.SyringeProject
		wantLen int
		wantErr bool
	}{
		{"gitlab", gitlabOpts, nil, 213, false},
		{"github", &structs.SyringeOptions{}, nil, 58, false},
		{"azure", &structs.SyringeOptions{}, nil, 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupEnv(t, tt.name)
			envMap, _ := utils.ReadEnvironment()
			s, err := NewSyringe(envMap, tt.opts)
			err = s.ListProjects()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if reflect.TypeOf(s.Projects) != reflect.TypeOf(tt.want) {
				t.Errorf("ListProjects() got = %v, want %v", s.Projects, tt.want)
			}

			if len(*s.Projects) != tt.wantLen {
				t.Errorf("ListProjects() wantLen(got) = %v, want %v", len(*s.Projects), tt.wantLen)
			}
		})
	}
}

func TestSyringe_GetLockfilesByProject(t *testing.T) {

	type args struct {
		projectId int64
	}

	tests := []struct {
		name    string
		args    args
		opts    *structs.SyringeOptions
		want    *structs.SyringeProject
		wantLen int
		wantErr bool
	}{
		{"gitlab", args{38265422}, gitlabOpts, &structs.SyringeProject{}, 4, false},
		{"github", args{325083799}, &structs.SyringeOptions{}, &structs.SyringeProject{}, 1, false},
		{"azure", args{1249448610}, &structs.SyringeOptions{}, &structs.SyringeProject{}, 4, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupEnv(t, tt.name)
			envMap, _ := utils.ReadEnvironment()
			s, err := NewSyringe(envMap, tt.opts)
			if err = s.ListProjects(); err != nil {
				fmt.Printf("failed to list projects: %v\n", err)
			}
			got, err := s.GetLockfilesByProject(tt.args.projectId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLockfilesByProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("GetLockfilesByProject() got = %v, want %v", s.Projects, tt.want)
			}

			if len(got.Lockfiles) != tt.wantLen {
				t.Errorf("GetLockfilesByProject() wantLen(got) = %v, want %v", len(got.Lockfiles), tt.want)
			}
		})
	}
}

func TestSyringe_PhylumGetProjectMap(t *testing.T) {

	setupEnv(t, "gitlab")
	envMap, _ := utils.ReadEnvironment()
	s, err := NewSyringe(envMap, gitlabOpts)
	if err != nil {
		fmt.Printf("failed to create syringe: %v\n", err)
	}
	err = s.ListProjects()
	if err != nil {
		fmt.Printf("failed to ListProjects: %v\n", err)
	}
	for k, _ := range s.ProjectsMap {
		_, err = s.GetLockfilesByProject(k)
		if err != nil {
			fmt.Printf("failed to GetLockfilesByProject: %v\n", err)
		}
	}

	tests := []struct {
		name    string
		want    *map[string]structs.PhylumProject
		wantErr bool
	}{
		{"one", nil, false},
	}

	var phylumProjects *map[string]structs.PhylumProject

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.PhylumGetProjectMap(&phylumProjects)
			if (err != nil) != tt.wantErr {
				t.Errorf("PhylumGetProjectMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(phylumProjects) != reflect.TypeOf(tt.want) {
				t.Errorf("PhylumGetProjectMap() got = %v, want %v", phylumProjects, tt.want)
			}

			_ = s.IntegratePhylumProjectList(phylumProjects)
			fmt.Println("test")
		})
	}
}

//
// func TestSyringe_PhylumCreateProjectsFromList(t *testing.T) {
// 	tearDown := setupEnv(t)
// 	defer tearDown(t)
// 	s, _ := NewSyringe()
//
// 	type args struct {
// 		projectsToCreate []string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    []structs.PhylumProject
// 		wantLen int
// 		wantErr bool
// 	}{
// 		{"one", args{[]string{
// 			"ZXY-test4",
// 			"ZXY-test5",
// 		}}, nil, 2, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := s.PhylumCreateProjectsFromList(tt.args.projectsToCreate)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("PhylumCreateProjectsFromList() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
// 				t.Errorf("PhylumCreateProjectsFromList() got = %v, want %v", got, tt.want)
// 			}
// 			if wantLen(got) != tt.wantLen {
// 				t.Errorf("PhylumCreateProjectsFromList() gotLen = %v, wantLen %v", wantLen(got), tt.wantLen)
// 			}
// 		})
// 	}
// }
//
// func TestSyringe_PhylumRunAnalyze(t *testing.T) {
// 	tearDown := setupEnv(t)
// 	defer tearDown(t)
// 	s, _ := NewSyringe()
//
// 	type args struct {
// 		phylumProjectFile structs.PhylumProject
// 		lockfile          *GitlabFile
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 		{"one",
// 			args{
// 				structs.PhylumProject{
// 					ID:        "70376347-6295-4bf8-969c-458f04ab92f0",
// 					Name:      "ZXY-test1",
// 					UpdatedAt: "2022-07-10T13:55:25.231350500-07:00",
// 				},
// 				&GitlabFile{
// 					Name:    "package-lock.json",
// 					Path:    "package-lock.json",
// 					Id:      "38029830",
// 					Content: []byte("{\n  \"name\": \"foo-pkg\",\n  \"version\": \"1.0.0\",\n  \"lockfileVersion\": 2,\n  \"requires\": true,\n  \"packages\": {\n    \"\": {\n      \"name\": \"foo-pkg\",\n      \"version\": \"1.0.0\",\n      \"license\": \"ISC\",\n      \"dependencies\": {\n        \"ansi-red\": \"0.1.1\",\n        \"axios\": \"0.19.0\",\n        \"ini\": \"1.3.5\"\n      }\n    },\n    \"node_modules/ansi-red\": {\n      \"version\": \"0.1.1\",\n      \"resolved\": \"https://registry.npmjs.org/ansi-red/-/ansi-red-0.1.1.tgz\",\n      \"integrity\": \"sha1-jGOPnRCAgAo1PJwoyKgcpHBdlGw=\",\n      \"dependencies\": {\n        \"ansi-wrap\": \"0.1.0\"\n      },\n      \"engines\": {\n        \"node\": \">=0.10.0\"\n      }\n    },\n    \"node_modules/ansi-wrap\": {\n      \"version\": \"0.1.0\",\n      \"resolved\": \"https://registry.npmjs.org/ansi-wrap/-/ansi-wrap-0.1.0.tgz\",\n      \"integrity\": \"sha1-qCJQ3bABXponyoLoLqYDu/pF768=\",\n      \"engines\": {\n        \"node\": \">=0.10.0\"\n      }\n    },\n    \"node_modules/axios\": {\n      \"version\": \"0.19.0\",\n      \"resolved\": \"https://registry.npmjs.org/axios/-/axios-0.19.0.tgz\",\n      \"integrity\": \"sha512-1uvKqKQta3KBxIz14F2v06AEHZ/dIoeKfbTRkK1E5oqjDnuEerLmYTgJB5AiQZHJcljpg1TuRzdjDR06qNk0DQ==\",\n      \"deprecated\": \"Critical security vulnerability fixed in v0.21.1. For more information, see https://github.com/axios/axios/pull/3410\",\n      \"dependencies\": {\n        \"follow-redirects\": \"1.5.10\",\n        \"is-buffer\": \"^2.0.2\"\n      }\n    },\n    \"node_modules/debug\": {\n      \"version\": \"3.1.0\",\n      \"resolved\": \"https://registry.npmjs.org/debug/-/debug-3.1.0.tgz\",\n      \"integrity\": \"sha512-OX8XqP7/1a9cqkxYw2yXss15f26NKWBpDXQd0/uK/KPqdQhxbPa994hnzjcE2VqQpDslf55723cKPUOGSmMY3g==\",\n      \"dependencies\": {\n        \"ms\": \"2.0.0\"\n      }\n    },\n    \"node_modules/follow-redirects\": {\n      \"version\": \"1.5.10\",\n      \"resolved\": \"https://registry.npmjs.org/follow-redirects/-/follow-redirects-1.5.10.tgz\",\n      \"integrity\": \"sha512-0V5l4Cizzvqt5D44aTXbFZz+FtyXV1vrDN6qrelxtfYQKW0KO0W2T/hkE8xvGa/540LkZlkaUjO4ailYTFtHVQ==\",\n      \"dependencies\": {\n        \"debug\": \"=3.1.0\"\n      },\n      \"engines\": {\n        \"node\": \">=4.0\"\n      }\n    },\n    \"node_modules/ini\": {\n      \"version\": \"1.3.5\",\n      \"resolved\": \"https://registry.npmjs.org/ini/-/ini-1.3.5.tgz\",\n      \"integrity\": \"sha512-RZY5huIKCMRWDUqZlEi72f/lmXKMvuszcMBduliQ3nnWbx9X/ZBQO7DijMEYS9EhHBb2qacRUMtC7svLwe0lcw==\",\n      \"deprecated\": \"Please update to ini >=1.3.6 to avoid a prototype pollution issue\",\n      \"engines\": {\n        \"node\": \"*\"\n      }\n    },\n    \"node_modules/is-buffer\": {\n      \"version\": \"2.0.5\",\n      \"resolved\": \"https://registry.npmjs.org/is-buffer/-/is-buffer-2.0.5.tgz\",\n      \"integrity\": \"sha512-i2R6zNFDwgEHJyQUtJEk0XFi1i0dPFn/oqjK3/vPCcDeJvW5NQ83V8QbicfF1SupOaB0h8ntgBC2YiE7dfyctQ==\",\n      \"funding\": [\n        {\n          \"type\": \"github\",\n          \"url\": \"https://github.com/sponsors/feross\"\n        },\n        {\n          \"type\": \"patreon\",\n          \"url\": \"https://www.patreon.com/feross\"\n        },\n        {\n          \"type\": \"consulting\",\n          \"url\": \"https://feross.org/support\"\n        }\n      ],\n      \"engines\": {\n        \"node\": \">=4\"\n      }\n    },\n    \"node_modules/ms\": {\n      \"version\": \"2.0.0\",\n      \"resolved\": \"https://registry.npmjs.org/ms/-/ms-2.0.0.tgz\",\n      \"integrity\": \"sha1-VgiurfwAvmwpAd9fmGF4jeDVl8g=\"\n    }\n  },\n  \"dependencies\": {\n    \"ansi-red\": {\n      \"version\": \"0.1.1\",\n      \"resolved\": \"https://registry.npmjs.org/ansi-red/-/ansi-red-0.1.1.tgz\",\n      \"integrity\": \"sha1-jGOPnRCAgAo1PJwoyKgcpHBdlGw=\",\n      \"requires\": {\n        \"ansi-wrap\": \"0.1.0\"\n      }\n    },\n    \"ansi-wrap\": {\n      \"version\": \"0.1.0\",\n      \"resolved\": \"https://registry.npmjs.org/ansi-wrap/-/ansi-wrap-0.1.0.tgz\",\n      \"integrity\": \"sha1-qCJQ3bABXponyoLoLqYDu/pF768=\"\n    },\n    \"axios\": {\n      \"version\": \"0.19.0\",\n      \"resolved\": \"https://registry.npmjs.org/axios/-/axios-0.19.0.tgz\",\n      \"integrity\": \"sha512-1uvKqKQta3KBxIz14F2v06AEHZ/dIoeKfbTRkK1E5oqjDnuEerLmYTgJB5AiQZHJcljpg1TuRzdjDR06qNk0DQ==\",\n      \"requires\": {\n        \"follow-redirects\": \"1.5.10\",\n        \"is-buffer\": \"^2.0.2\"\n      }\n    },\n    \"debug\": {\n      \"version\": \"3.1.0\",\n      \"resolved\": \"https://registry.npmjs.org/debug/-/debug-3.1.0.tgz\",\n      \"integrity\": \"sha512-OX8XqP7/1a9cqkxYw2yXss15f26NKWBpDXQd0/uK/KPqdQhxbPa994hnzjcE2VqQpDslf55723cKPUOGSmMY3g==\",\n      \"requires\": {\n        \"ms\": \"2.0.0\"\n      }\n    },\n    \"follow-redirects\": {\n      \"version\": \"1.5.10\",\n      \"resolved\": \"https://registry.npmjs.org/follow-redirects/-/follow-redirects-1.5.10.tgz\",\n      \"integrity\": \"sha512-0V5l4Cizzvqt5D44aTXbFZz+FtyXV1vrDN6qrelxtfYQKW0KO0W2T/hkE8xvGa/540LkZlkaUjO4ailYTFtHVQ==\",\n      \"requires\": {\n        \"debug\": \"=3.1.0\"\n      }\n    },\n    \"ini\": {\n      \"version\": \"1.3.5\",\n      \"resolved\": \"https://registry.npmjs.org/ini/-/ini-1.3.5.tgz\",\n      \"integrity\": \"sha512-RZY5huIKCMRWDUqZlEi72f/lmXKMvuszcMBduliQ3nnWbx9X/ZBQO7DijMEYS9EhHBb2qacRUMtC7svLwe0lcw==\"\n    },\n    \"is-buffer\": {\n      \"version\": \"2.0.5\",\n      \"resolved\": \"https://registry.npmjs.org/is-buffer/-/is-buffer-2.0.5.tgz\",\n      \"integrity\": \"sha512-i2R6zNFDwgEHJyQUtJEk0XFi1i0dPFn/oqjK3/vPCcDeJvW5NQ83V8QbicfF1SupOaB0h8ntgBC2YiE7dfyctQ==\"\n    },\n    \"ms\": {\n      \"version\": \"2.0.0\",\n      \"resolved\": \"https://registry.npmjs.org/ms/-/ms-2.0.0.tgz\",\n      \"integrity\": \"sha1-VgiurfwAvmwpAd9fmGF4jeDVl8g=\"\n    }\n  }\n}"),
// 				},
// 			},
// 			false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			err := s.PhylumRunAnalyze(tt.args.phylumProjectFile, tt.args.lockfile)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("PhylumRunAnalyze() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
//
// func TestSyringe_LoadPidFile(t *testing.T) {
// 	type fields struct {
// 		Gitlab          *gitlab.Client
// 		PhylumToken     string
// 		PhylumGroupName string
// 	}
// 	type args struct {
// 		filename string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &Syringe{
// 				Gitlab:          tt.fields.Gitlab,
// 				PhylumToken:     tt.fields.PhylumToken,
// 				PhylumGroupName: tt.fields.PhylumGroupName,
// 			}
// 			if err := s.LoadPidFile(tt.args.filename); (err != nil) != tt.wantErr {
// 				t.Errorf("LoadPidFile() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

func TestSyringe_GetAllLockfiles(t *testing.T) {

	tests := []struct {
		name    string
		opts    *structs.SyringeOptions
		want    *structs.SyringeProject
		wantLen int
		wantErr bool
	}{
		{"gitlab", gitlabOpts, &structs.SyringeProject{}, 4, false},
		{"github", &structs.SyringeOptions{}, &structs.SyringeProject{}, 57, false},
		{"azure", &structs.SyringeOptions{}, &structs.SyringeProject{}, 4, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupEnv(t, tt.name)
			envMap, _ := utils.ReadEnvironment()
			s, err := NewSyringe(envMap, tt.opts)
			if err = s.ListProjects(); err != nil {
				fmt.Printf("failed to list projects: %v\n", err)
			}
			err = s.GetAllLockfiles()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllLockfiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
			//	t.Errorf("GetLockfilesByProject() got = %v, want %v", s.Projects, tt.want)
			// }

			// if len(got.Lockfiles) != tt.wantLen {
			//	t.Errorf("GetLockfilesByProject() wantLen(got) = %v, want %v", len(got.Lockfiles), tt.want)
			// }
		})
	}
}

// func TestSyringe_CreatePhylumProjects(t *testing.T) {
// 	type args struct {
// 		phylumProjectMap *map[string]structs.PhylumProject
// 		syringeProjects  *[]*structs.SyringeProject
// 	}
//
// 	ppMap := make(map[string]structs.PhylumProject, 1)
// 	ppMap["one"] = structs.PhylumProject{
// 		Name:      "one",
// 		ID:        "232980293840928340",
// 		UpdatedAt: "2022-08-12",
// 		Ecosystem: "npm",
// 	}
// 	lfOne := structs.VcsFile{
// 		Name:          "requirements.txt",
// 		Path:          "backend/requirements.txt",
// 		Id:            "SHAhash380823",
// 		Content:       nil,
// 		PhylumProject: nil,
// 	}
// 	syringeProject := structs.SyringeProject{
// 		Id:        38265422,
// 		Name:      "testing-1234-Syringe",
// 		Branch:    "master",
// 		Lockfiles: nil,
// 		CiFiles:   nil,
// 		Hydrated:  true,
// 	}
// 	syringeProject.Lockfiles = make([]*structs.VcsFile, 0)
// 	syringeProject.Lockfiles = append(syringeProject.Lockfiles, &lfOne)
// 	syringeProjects := make([]*structs.SyringeProject, 0)
// 	syringeProjects = append(syringeProjects, &syringeProject)
// 	oneArgs := args{
// 		phylumProjectMap: &ppMap,
// 		syringeProjects:  &syringeProjects,
// 	}
//
// 	tests := []struct {
// 		name    string
// 		mine    bool
// 		args    args
// 		wantErr bool
// 	}{
// 		{"gitlab", true, oneArgs, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			setupEnv(t, tt.name)
// 			envMap, _ := utils.ReadEnvironment()
// 			s, err := NewSyringe(envMap, tt.mine)
// 			*s.Projects = append(*s.Projects, &syringeProject)
// 			s.ProjectsMap[syringeProject.Id] = &syringeProject
// 			if err != nil {
// 				fmt.Printf("Failed to create new syringe: %v\n", err)
// 			}
// 			if err := s.CreatePhylumProjects(tt.args.phylumProjectMap, tt.args.syringeProjects); (err != nil) != tt.wantErr {
// 				t.Errorf("CreatePhylumProjects() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
