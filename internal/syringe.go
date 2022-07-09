package PhylumSyringGitlab

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/exp/slices"
	"io"
	"os"
)

func getMainBranchSlice() []string {
	return []string{
		"master",
		"main",
	}
}

func getSupportedLockfiles() []string {
	return []string{
		"package-lock.json",
		"yarn.lock",
		"requirements.txt",
		"poetry.lock",
		"pom.xml",
		"Gemfile.lock",
	}
}

func getCiFiles() []string {
	return []string{
		".gitlab-ci.yml",
		".gitlab-ci.yaml",
	}
}

func readEnvVar(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	} else {
		return "", fmt.Errorf("Failed to read environment variable: %v\n", key)
	}
}

func init() {
	// setup logging
	log.SetReportCaller(false)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:            true,
		FullTimestamp:          true,
		DisableLevelTruncation: false,
		DisableTimestamp:       false,
	})
	logFile, err := os.OpenFile("LOG_PhylumSyringeGitlab.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error: failed to open logfile")
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	log.SetLevel(log.DebugLevel)
}

func NewSyringe() (*Syringe, error) {
	token_gitlab, err := readEnvVar("GITLAB_TOKEN")
	if err != nil {
		log.Fatalf("Failed to read gitlab token from ENV\n")
	}
	token_phylum, err := readEnvVar("PHYLUM_TOKEN")
	if err != nil {
		log.Fatalf("Failed to read phylum token from ENV\n")
	}

	gitlabClient, err := gitlab.NewClient(token_gitlab)
	if err != nil {
		log.Fatalf("Failed to create gitlab client: %v\n", err)
		return nil, err
	}

	return &Syringe{
		Gitlab:      gitlabClient,
		PhylumToken: token_phylum,
	}, nil
}

func (s *Syringe) ListProjects() ([]*gitlab.Project, error) {
	projects, _, err := s.Gitlab.Projects.ListProjects(&gitlab.ListProjectsOptions{Owned: gitlab.Bool(true)})
	if err != nil {
		log.Errorf("Failed to list gitlab projects: %v\n", err)
		return nil, err
	}
	return projects, nil
}

func (s *Syringe) ListBranches(projectId int) ([]*gitlab.Branch, error) {
	branches, _, err := s.Gitlab.Branches.ListBranches(projectId, &gitlab.ListBranchesOptions{})
	if err != nil {
		log.Errorf("Failed to ListTree from %v: %v\n", projectId, err)
		return nil, err
	}
	return branches, nil
}

func (s *Syringe) ListFiles(projectId int, branch string) ([]*gitlab.TreeNode, error) {
	files, _, err := s.Gitlab.Repositories.ListTree(projectId, &gitlab.ListTreeOptions{
		Path:      gitlab.String("/"),
		Ref:       gitlab.String(branch),
		Recursive: gitlab.Bool(true),
	})
	if err != nil {
		log.Errorf("Failed to ListTree from %v: %v\n", projectId, err)
		return nil, err
	}
	return files, nil
}

func (s *Syringe) IdentifyMainBranch(projectId int) (*gitlab.Branch, error) {
	branches, err := s.ListBranches(projectId)
	if err != nil {
		return nil, err
	}

	mainBranchSlice := getMainBranchSlice()

	foundBranches := make([]*gitlab.Branch, 0)
	for _, branch := range branches {
		if slices.Contains(mainBranchSlice, branch.Name) {
			foundBranches = append(foundBranches, branch)
		}
	}

	var ret *gitlab.Branch
	var retErr error

	switch len(foundBranches) {
	case 0:
		ret = nil
		retErr = fmt.Errorf("No main branch found: %v\n", projectId)
	case 1:
		ret = foundBranches[0]
		retErr = nil
	case 2:
		for _, branch := range foundBranches {
			if branch.Name == "master" {
				ret = branch
				retErr = nil
			}
		}
	default:
		ret = nil
		retErr = fmt.Errorf("IdentifyMainBranch error: shouldn't happen %v\n", projectId)
	}
	return ret, retErr
}

func (s *Syringe) GetFileTreeFromProject(projectId int) ([]*gitlab.TreeNode, error) {
	mainBranch, err := s.IdentifyMainBranch(projectId)
	if err != nil {
		log.Errorf("Failed to IdentifyMainBranch: %v\n", err)
		return nil, err
	}

	projectFiles, err := s.ListFiles(projectId, mainBranch.Name)
	if err != nil {
		log.Errorf("Failed to ListFiles for %v on branch %v\n", projectId, mainBranch.Name)
		return nil, err
	}
	return projectFiles, nil
}

func (s *Syringe) EnumerateTargetFiles(projectId int) ([]*GitlabFile, error) {
	var targetFiles []*GitlabFile

	mainBranch, err := s.IdentifyMainBranch(projectId)
	if err != nil {
		log.Errorf("Failed to IdentifyMainBranch: %v\n", err)
		return nil, err
	}

	projectFiles, err := s.ListFiles(projectId, mainBranch.Name)
	if err != nil {
		log.Errorf("Failed to ListFiles for %v on branch %v\n", projectId, mainBranch.Name)
		return nil, err
	}

	supportedLockfiles := getSupportedLockfiles()
	ciFiles := getCiFiles()
	//TODO: make gofunc
	for _, file := range projectFiles {
		if slices.Contains(supportedLockfiles, file.Name) || slices.Contains(ciFiles, file.Name) {
			data, _, err := s.Gitlab.RepositoryFiles.GetRawFile(projectId, file.Name, &gitlab.GetRawFileOptions{})
			if err != nil {
				log.Errorf("Failed to GetRawFile for %v in projectId %v\n", file.Name, projectId)
			}
			rec := GitlabFile{file.Name, file.Path, file.ID, data}
			targetFiles = append(targetFiles, &rec)
		}
	}
	return targetFiles, nil
}

func (s *Syringe) PhylumGetProjectList() {

}
