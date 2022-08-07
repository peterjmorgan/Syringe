package syringe

import (
	"github.com/xanzy/go-gitlab"
)

type VcsFile struct {
	Name    string
	Path    string
	Id      string
	Content []byte
}

type SyringeProject struct {
	Id               int
	Name             string
	Branch           *gitlab.Branch
	NumPhylumEnabled int
	IsPhylumReady    bool
	Lockfiles        []*VcsFile
	CiFiles          []*VcsFile
}

type PhylumProject struct {
	Name      string `json:"name" yaml:"name"`
	ID        string `json:"id" yaml:"id"`
	UpdatedAt string `json:"updated_at" yaml:"created_at"`
	Ecosystem string `json:"ecosystem"`
}
