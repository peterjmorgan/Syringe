package PhylumSyringGitlab

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"io"
	"os"
)

func init() {
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

func NewSyringe(gitlabToken string) (*Syringe, error) {
	gitlabClient, err := gitlab.NewClient(gitlabToken)
	if err != nil {
		log.Fatalf("Failed to create gitlab client: %v\n", err)
		return nil, err
	}

	return &Syringe{Gitlab: gitlabClient}, nil
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

func getMainBranchSlice() []string {
	return []string{
		"master",
		"main",
	}
}

func (s *Syringe) IdentifyMainBranch(projectId int) (*gitlab.Branch, error) {
	branches, err := s.ListBranches(projectId)
	if err != nil {
		return nil, err
	}

	mainBranchSlice := getMainBranchSlice()

	foundBranches := make([]*gitlab.Branch, 0)
	for _, branch := range branches {
		for _, mainBranchName := range mainBranchSlice {
			if branch.Name == mainBranchName {
				foundBranches = append(foundBranches, branch)
			}
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
