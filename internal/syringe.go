package PhylumSyringGitlab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"
)

// Branches to examine for "main" branch
func getMainBranchSlice() []string {
	return []string{
		"master",
		"main",
	}
}

// Lockfiles to target
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

// CI Files to target
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
	log.SetLevel(log.InfoLevel)
}

func NewSyringe() (*Syringe, error) {
	gitlabToken, err := readEnvVar("GITLAB_TOKEN")
	if err != nil {
		log.Fatalf("Failed to read gitlab token from ENV\n")
	}
	phylumToken, err := readEnvVar("PHYLUM_API_KEY")
	if err != nil {
		log.Fatalf("Failed to read phylum api key from ENV\n")
	}

	phylumGroupName, err := readEnvVar("PHYLUM_GROUP_NAME")
	if err != nil {
		log.Debugf("PHYLUM_GROUP is not set\n")
		phylumGroupName = ""
	}

	gitlabBaseUrl, err := readEnvVar("GITLAB_BASEURL")
	if err != nil {
		log.Debugf("GITLAB_BASEURL is not set\n")
		gitlabBaseUrl = "https://gitlab.com"
	}

	gitlabClient, err := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabBaseUrl))
	if err != nil {
		log.Fatalf("Failed to create gitlab client: %v\n", err)
		return nil, err
	}

	return &Syringe{
		Gitlab:          gitlabClient,
		PhylumToken:     phylumToken,
		PhylumGroupName: phylumGroupName,
		ProjectIDs:      []string{},
	}, nil
}

func (s *Syringe) ListProjects(projects **[]*gitlab.Project) error {
	temp, _, err := s.Gitlab.Projects.ListProjects(&gitlab.ListProjectsOptions{Owned: gitlab.Bool(false)})
	if err != nil {
		log.Errorf("Failed to list gitlab projects: %v\n", err)
		return err
	}
	log.Debugf("Len of gitlab projects: %v\n", len(temp))
	*projects = &temp
	return nil
}

func (s *Syringe) ListBranches(projectId int) ([]*gitlab.Branch, error) {
	branches, _, err := s.Gitlab.Branches.ListBranches(projectId, &gitlab.ListBranchesOptions{})
	if err != nil {
		log.Errorf("Failed to ListTree from %v: %v\n", projectId, err)
		return nil, err
	}
	return branches, nil
}

func (s *Syringe) PrintProjectVariables(projectId int) error {
	variables, _, err := s.Gitlab.ProjectVariables.ListVariables(projectId, &gitlab.ListProjectVariablesOptions{})
	if err != nil {
		log.Errorf("Failed to list project variables from %v: %v\n", projectId, err)
		return err
	}
	for _, variable := range variables {
		log.Infof("Variable: %v:%v\n", variable.Key, variable.Value)
	}
	return nil
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

func (s *Syringe) EnumerateTargetFiles(projectId int) ([]*GitlabFile, []*GitlabFile, error) {
	var retLockFiles []*GitlabFile
	var retCiFiles []*GitlabFile

	mainBranch, err := s.IdentifyMainBranch(projectId)
	if err != nil {
		// TODO: this needs to pass when repos don't have code
		log.Infof("Failed to IdentifyMainBranch: %v\n", err)
		return nil, nil, err
	}

	projectFiles, err := s.ListFiles(projectId, mainBranch.Name)
	if err != nil {
		log.Errorf("Failed to ListFiles for %v on branch %v\n", projectId, mainBranch.Name)
		return nil, nil, err
	}

	supportedLockfiles := getSupportedLockfiles()
	supportedciFiles := getCiFiles()
	for _, file := range projectFiles {
		if slices.Contains(supportedLockfiles, file.Name) {
			data, _, err := s.Gitlab.RepositoryFiles.GetRawFile(projectId, file.Path, &gitlab.GetRawFileOptions{&mainBranch.Name})
			if err != nil {
				log.Errorf("Failed to GetRawFile for %v in projectId %v\n", file.Name, projectId)
			}
			rec := GitlabFile{file.Name, file.Path, file.ID, data}
			retLockFiles = append(retLockFiles, &rec)
		}
		if slices.Contains(supportedciFiles, file.Name) {
			data, _, err := s.Gitlab.RepositoryFiles.GetRawFile(projectId, file.Path, &gitlab.GetRawFileOptions{})
			if err != nil {
				log.Errorf("Failed to GetRawFile for %v in projectId %v\n", file.Name, projectId)
			}
			rec := GitlabFile{file.Name, file.Path, file.ID, data}
			retCiFiles = append(retCiFiles, &rec)
		}
	}
	return retLockFiles, retCiFiles, nil
}

func (s *Syringe) PhylumGetProjectMap(retVal **map[string]PhylumProject) error {
	var stdErrBytes bytes.Buffer
	var projectListArgs = []string{"project", "list", "--json"}
	if s.PhylumGroupName != "" {
		projectListArgs = append(projectListArgs, "-g", s.PhylumGroupName)
	}
	projectListCmd := exec.Command("phylum", projectListArgs...)
	projectListCmd.Stderr = &stdErrBytes
	output, err := projectListCmd.Output()
	if err != nil {
		log.Errorf("Failed to exec 'phylum project list': %v\n", err)
		log.Errorf(stdErrBytes.String())
		return err
	}
	stdErrString := stdErrBytes.String()
	_ = stdErrString // prob will need this later

	var PhylumProjectList []PhylumProject
	if err := json.Unmarshal(output, &PhylumProjectList); err != nil {
		log.Errorf("Failed to unmarshal JSON: %v\n", err)
		return err
	}

	returnMap := make(map[string]PhylumProject, 0)
	for _, elem := range PhylumProjectList {
		returnMap[elem.Name] = elem
	}
	log.Debugf("Found %v phylum projects\n", len(returnMap))
	*retVal = &returnMap
	return nil
}

func (s *Syringe) GeneratePhylumProjectName(projectName string, lockfilePath string) string {
	return fmt.Sprintf("SYR-%v__%v", projectName, lockfilePath)
}

func RemoveTempDir(tempDir string) {
	err := os.RemoveAll(tempDir)
	if err != nil {
		log.Errorf("Failed to remove temp directory %v:%v\n", tempDir, err)
	}
}

func (s *Syringe) PhylumCreateProject(projectNames <-chan string, projects chan<- PhylumProject) error {
	for projectName := range projectNames {
		tempDir, err := ioutil.TempDir("", "syringe-create")
		if err != nil {
			log.Errorf("Failed to create temp directory: %v\n", err)
			return err
		}
		defer RemoveTempDir(tempDir)

		var stdErrBytes bytes.Buffer
		var CreateCmdArgs = []string{"project", "create", projectName}
		if s.PhylumGroupName != "" {
			CreateCmdArgs = append(CreateCmdArgs, "-g", s.PhylumGroupName)
		}
		projectCreateCmd := exec.Command("phylum", CreateCmdArgs...)
		projectCreateCmd.Stderr = &stdErrBytes
		projectCreateCmd.Dir = tempDir
		err = projectCreateCmd.Run()
		stdErrString := stdErrBytes.String()
		if err != nil {
			log.Errorf("Failed to exec 'phylum project create %v': %v\n", projectName, err)
			log.Errorf("%v\n", stdErrString)
			return err
		} else {
			log.Debugf("Created phylum project: %v\n", projectName)
		}

		phylumProjectFile := filepath.Join(tempDir, ".phylum_project")
		phylumProjectData, err := os.ReadFile(phylumProjectFile)
		if err != nil {
			log.Errorf("Failed to read created .phylum_project file at %v: %v\n", phylumProjectFile, err)
			return err
		}

		phylumProject := PhylumProject{}
		err = yaml.Unmarshal(phylumProjectData, &phylumProject)
		if err != nil {
			log.Errorf("Failed to unmarshall YAML data from created phylum project %v: %v\n", projectName, err)
			return err
		}
		projects <- phylumProject
	}
	return nil
}

func (s *Syringe) PhylumCreateProjectsFromList(projectsToCreate []string) ([]PhylumProject, error) {
	createdProjects := make([]PhylumProject, 0)

	for _, elem := range projectsToCreate {
		tempDir, err := ioutil.TempDir("", "syringe-create")
		if err != nil {
			log.Errorf("Failed to create temp directory: %v\n", err)
			return nil, err
		}
		defer RemoveTempDir(tempDir)

		var stdErrBytes bytes.Buffer
		var CreateCmdArgs = []string{"project", "create", elem}
		if s.PhylumGroupName != "" {
			CreateCmdArgs = append(CreateCmdArgs, "-g", s.PhylumGroupName)
		}
		projectCreateCmd := exec.Command("phylum", CreateCmdArgs...)
		projectCreateCmd.Stderr = &stdErrBytes
		projectCreateCmd.Dir = tempDir
		err = projectCreateCmd.Run()
		stdErrString := stdErrBytes.String()
		if err != nil {
			log.Errorf("Failed to exec 'phylum project create %v': %v\n", elem, err)
			log.Errorf("%v\n", stdErrString)
			return nil, err
		} else {
			log.Infof("Created phylum project: %v\n", elem)
		}

		phylumProjectFile := filepath.Join(tempDir, ".phylum_project")
		phylumProjectData, err := os.ReadFile(phylumProjectFile)
		if err != nil {
			log.Errorf("Failed to read created .phylum_project file at %v: %v\n", phylumProjectFile, err)
			return nil, err
		}

		phylumProject := PhylumProject{}
		err = yaml.Unmarshal(phylumProjectData, &phylumProject)
		if err != nil {
			log.Errorf("Failed to unmarshall YAML data from created phylum project %v: %v\n", elem, err)
			return nil, err
		}
		createdProjects = append(createdProjects, phylumProject)
	}
	return createdProjects, nil
}

func (s *Syringe) PhylumRunAnalyze(phylumProjectFile PhylumProject, lockfile *GitlabFile, phylumProjectName string) error {
	// create temp directory to write the lockfile content for analyze
	log.Debugf("Analyzing %v\n", phylumProjectFile.Name)
	tempDir, err := ioutil.TempDir("", "syringe-analyze")
	if err != nil {
		log.Errorf("Failed to create temp directory: %v\n", err)
		return err
	}
	defer RemoveTempDir(tempDir)

	// create the lockfile
	tempLockfileName := filepath.Join(tempDir, lockfile.Name)
	err = os.WriteFile(tempLockfileName, lockfile.Content, 0644)
	if err != nil {
		log.Errorf("Failed to write lockfile content to temp file: %v", err)
		return err
	}
	// create the .phylum_project file
	dotPhylumProjectFile := filepath.Join(tempDir, ".phylum_project")
	dotPhylumProjectData, err := yaml.Marshal(phylumProjectFile)
	if err != nil {
		log.Errorf("Failed to marshal phylum project %v to YAML: %v\n", phylumProjectFile.Name, err)
		return err
	}
	err = os.WriteFile(dotPhylumProjectFile, dotPhylumProjectData, 0644)

	var stdErrBytes bytes.Buffer
	var AnalyzeCmdArgs = []string{"analyze", lockfile.Name}
	if s.PhylumGroupName != "" {
		AnalyzeCmdArgs = append(AnalyzeCmdArgs, "-g", s.PhylumGroupName, "--project", phylumProjectName)
	}
	projectAnalyzeCmd := exec.Command("phylum", AnalyzeCmdArgs...)
	projectAnalyzeCmd.Stderr = &stdErrBytes
	projectAnalyzeCmd.Dir = tempDir
	err = projectAnalyzeCmd.Run()
	stdErrString := stdErrBytes.String()
	if err != nil {
		log.Errorf("Failed to exec 'phylum %v': %v\n", strings.Join(AnalyzeCmdArgs, " "), err)
		log.Errorf("%v\n", stdErrString)
		return err
	} else {
		log.Debugf("Phylum Analyzed: %v\n", phylumProjectFile.Name)
	}
	return nil
}

// LoadPidFile Read a text file of project IDs to operate on
// The text file should have one project ID per line
func (s *Syringe) LoadPidFile(filename string) error {
	var pids []string

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read project ID file: %v error: %v\n", filename, err)
		return err
	}
	lineRegex := regexp.MustCompile(`^[\d|\s]+$`)
	lines := bytes.Split(data, []byte("\n"))
	for idx, line := range lines {
		if !lineRegex.Match(line) && len(line) > 0 {
			log.Errorf("Failed to parse project ID file: %v - line #%v\n", filename, idx+1)
			log.Errorf("lines must match regex: %v", lineRegex.String())
			return fmt.Errorf("Failed to parse project ID file: %v - line #%v\n", filename, idx+1)
		}
		pids = append(pids, string(line))
	}

	s.ProjectIDs = pids
	return nil
}
