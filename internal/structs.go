package PhylumSyringGitlab

import "github.com/xanzy/go-gitlab"

type GitlabFile struct {
	Name    string
	Path    string
	Id      string
	Content string
}

type GitlabProject struct {
	Id              int
	Name            string
	IsPhylumEnabled bool
	IsPhylumReady   bool
	Lockfiles       []*GitlabFile
	CiFiles         []*GitlabFile
}

type Syringe struct {
	Gitlab *gitlab.Client
	//TODO: need a db or storage thing
}