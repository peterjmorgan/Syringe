package syringe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gosuri/uiprogress"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func init() {
	// setup logging
	log.SetReportCaller(false)
	// log.SetFormatter(&log.TextFormatter{
	// 	ForceColors:            false,
	// 	FullTimestamp:          true,
	// 	DisableLevelTruncation: false,
	// 	DisableTimestamp:       false,
	// })
	logFile, err := os.OpenFile("LOG_Syringe.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error: failed to open logfile")
	}
	// mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(logFile)
	log.SetLevel(log.InfoLevel)
	uiprogress.Start()
}

type Syringe struct {
	Client          Client
	PhylumToken     string
	PhylumGroupName string
	ProjectIDs      []string
	MineOnly        bool
}

func NewSyringe(mineOnly bool) (*Syringe, error) {
	var vcsType ClientType
	vcsValue, err := readEnvVar("SYRINGE_VCS")
	if err != nil {
		log.Fatalf("Failed to read syringe vcs type from ENV\n")
	}

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

	if vcsValue == "gitlab" {
		vcsType = 1
	} else if vcsValue == "github" {
		vcsType = 0
	}
	client, err := NewClient(vcsType, gitlabToken, gitlabBaseUrl)
	if err != nil {
		log.Fatalf("Failed to create client: %v\n", err)
	}

	return &Syringe{
		Client:          client,
		PhylumToken:     phylumToken,
		PhylumGroupName: phylumGroupName,
		ProjectIDs:      []string{},
		MineOnly:        mineOnly,
	}, nil
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
