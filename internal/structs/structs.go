package structs

import (
	"github.com/google/uuid"
)

type VcsFile struct {
	Name          string
	Path          string
	Id            string
	Content       []byte
	PhylumProject *PhylumProject
}

type SyringeProject struct {
	Id        int64
	Name      string
	Branch    string
	Lockfiles []*VcsFile
	CiFiles   []*VcsFile
	Hydrated  bool
	GUID      uuid.UUID
}

type PhylumProject struct {
	Name      string `json:"name" yaml:"name"`
	ID        string `json:"id" yaml:"id"`
	UpdatedAt string `json:"updated_at" yaml:"created_at"`
	Ecosystem string `json:"ecosystem"`
}

type SyringeOptions struct {
	MineOnly  bool
	RateLimit int
	ProxyUrl  string
}

type ConfigThing struct {
	VcsType     string
	VcsToken    string
	Associated  map[string]string
	PhylumToken string
	PhylumGroup string
}

type TestConfigData struct {
	Filename string
}
