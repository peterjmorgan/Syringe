package PhylumSyringGitlab

import "github.com/xanzy/go-gitlab"

type GitlabFile struct {
	Name    string
	Path    string
	Id      string
	Content []byte
}

type GitlabProject struct {
	Id              int
	Name            string
	Branch          *gitlab.Branch
	IsPhylumEnabled bool
	IsPhylumReady   bool
	Lockfiles       []*GitlabFile
	CiFiles         []*GitlabFile
}

type Syringe struct {
	Gitlab          *gitlab.Client
	PhylumToken     string
	PhylumGroupName string
}

type PhylumProject struct {
	Name      string `json:"name" yaml:"name"`
	ID        string `json:"id" yaml:"id"`
	UpdatedAt string `json:"updated_at" yaml:"created_at"`
	Ecosystem string `json:"ecosystem"`
}
