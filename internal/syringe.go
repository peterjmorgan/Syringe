package syringePackage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/peterjmorgan/Syringe/internal/structs"
	"github.com/peterjmorgan/Syringe/internal/utils"
	"github.com/schollz/progressbar/v3"
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
}

type Syringe struct {
	Client           Client
	PhylumToken      string
	PhylumGroupName  string
	Projects         *[]*structs.SyringeProject
	ProjectsMap      map[int64]*structs.SyringeProject
	ProjectsMapMutex sync.RWMutex
	MineOnly         bool
	LockfileCount    int
	RateLimit        int
}

func NewSyringe(envMap map[string]string, mineOnly bool, ratelimit int, proxyUrl string) (*Syringe, error) {

	client, err := NewClient(envMap["vcs"], envMap, mineOnly, ratelimit, proxyUrl)
	if err != nil {
		log.Fatalf("Failed to create client: %v\n", err)
	}

	defaultProjects := make([]*structs.SyringeProject, 0)
	defaultProjectMap := make(map[int64]*structs.SyringeProject, 0)

	return &Syringe{
		Client:          client,
		PhylumToken:     envMap["phylumToken"],
		PhylumGroupName: envMap["phylumGroup"],
		Projects:        &defaultProjects,
		ProjectsMap:     defaultProjectMap,
		MineOnly:        mineOnly, // store in map
		LockfileCount:   0,
		RateLimit:       ratelimit,
	}, nil
}

func (s *Syringe) ListProjects() error {

	syringeProjects, err := s.Client.ListProjects()
	if err != nil {
		log.Errorf("Failed to list projects: %v\n", err)
	}
	s.Projects = syringeProjects
	s.ProjectsMap = make(map[int64]*structs.SyringeProject, len(*syringeProjects))
	for _, project := range *syringeProjects {
		s.ProjectsMap[project.Id] = project
	}

	return nil
}

// Returns a pointer to project with the lockfiles in it
func (s *Syringe) GetLockfilesByProject(projectId int64) (*structs.SyringeProject, error) {

	s.ProjectsMapMutex.RLock()
	theProject, ok := s.ProjectsMap[projectId]
	s.ProjectsMapMutex.RUnlock()
	if ok {
		if theProject.Hydrated == true {
			return theProject, nil
		}
	} else {
		// projectID not in map
		log.Debugf("GetLockfilesByProject: Creating empty project")
		theProject = &structs.SyringeProject{}
	}

	lockfiles, err := s.Client.GetLockfilesByProject(theProject.Id, theProject.Branch) // TODO: i think i want s.projects to be a map indexed by projectid
	if err != nil {
		// log.Warnf("Failed to get lockfiles: %v\n", err)
		return nil, err
	}
	theProject.Lockfiles = lockfiles
	theProject.Hydrated = true

	s.ProjectsMapMutex.Lock()
	s.ProjectsMap[projectId] = theProject
	s.ProjectsMapMutex.Unlock()

	return theProject, nil
}

func (s *Syringe) GetAllLockfiles() error {
	var wg sync.WaitGroup
	lockfilesBar := progressbar.NewOptions(len(*s.Projects), progressbar.OptionSetDescription("Getting Lockfiles"))

	for kID, _ := range s.ProjectsMap {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			_, err := s.GetLockfilesByProject(id)
			lockfilesBar.Add(1)
			if err != nil {
				log.Warnf("failed to GetLockFilesByProject() ID=%v: %v\n", kID, err)
			}
		}(kID)
	}
	wg.Wait()

	return nil
}

// This returns a map because usually when I run this, it's concurrent with listProjects. Then, I can integrate them into the syringe struct.
func (s *Syringe) PhylumGetProjectMap(retVal **map[string]structs.PhylumProject) error {
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

	var PhylumProjectList []structs.PhylumProject
	if err := json.Unmarshal(output, &PhylumProjectList); err != nil {
		log.Errorf("Failed to unmarshal JSON: %v\n", err)
		return err
	}

	returnMap := make(map[string]structs.PhylumProject, len(PhylumProjectList))
	for _, elem := range PhylumProjectList {
		returnMap[elem.Name] = elem
	}
	log.Debugf("Found %v phylum projects\n", len(returnMap))
	*retVal = &returnMap
	return nil
}

// IntegratePhylumProjectList
// Iterate through all VCS projects matching each project with a Phylum project if it exists.
// If it doesn't, return the slice of phylum project names to be created.
func (s *Syringe) IntegratePhylumProjectList(phylumProjectMap *map[string]structs.PhylumProject) []string {
	var phylumProjectsToCreate = make([]string, 0)
	var lockfileCount int = 0

	for _, syringeProject := range s.ProjectsMap {
		for _, lockfile := range syringeProject.Lockfiles {
			lockfileCount++
			phylumProjectName := utils.GeneratePhylumProjectName(syringeProject.Name, lockfile.Path, syringeProject.Id)
			phylumProject, ok := (*phylumProjectMap)[phylumProjectName]
			if ok {
				lockfile.PhylumProject = &phylumProject
			} else {
				phylumProjectsToCreate = append(phylumProjectsToCreate, phylumProjectName)
			}
		}
	}
	s.LockfileCount = lockfileCount
	return phylumProjectsToCreate
}

// Identify vcsProjects that do not have an associated phylum project. Create those projects. Update the syringe struct with new phylum project information
// func (s *Syringe) CreatePhylumProjects(phylumProjectMap *map[string]structs.PhylumProject, syringeProjects *[]*structs.SyringeProject) error {
// 	// Enumerate list of SyringeProjects that do not have an associated Phylum project.
// 	chCreateProjects := make(chan string, 3000)
// 	chProjectResults := make(chan structs.PhylumProject, 3000)
// 	var wgLoop sync.WaitGroup
//
// 	go func() {
// 		wgLoop.Wait()
// 		close(chCreateProjects)
// 	}()
// 	for _, project := range *syringeProjects {
// 		wgLoop.Add(1)
// 		go func(inProject structs.SyringeProject) {
// 			defer wgLoop.Done()
//
// 			// lockfiles, _, err := s.EnumerateTargetFiles(inProject.ID)
// 			// if err != nil {
// 			// 	log.Debugf("Failed to GetLockFiles(): %v\n", err)
// 			// 	return
// 			// }
//
// 			log.Debugf("calling GetLockfilesByProject: %v\n", inProject.Name)
// 			resultProject, err := s.GetLockfilesByProject(inProject.Id)
// 			if err != nil {
// 				log.Debugf("Failed to GetLockFiles(): %v\n", err)
// 				return
// 			}
//
// 			for _, lf := range resultProject.Lockfiles {
// 				phylumProjectName := utils.GeneratePhylumProjectName(inProject.Name, lf.Path)
// 				// if the project name is NOT in the slice of keys from the phylum project list map, we have to create it
// 				// if !slices.Contains(maps.Keys(*phylumProjectMap), phylumProjectName) {
// 				if !slices.Contains(maps.Keys(*phylumProjectMap), phylumProjectName) {
// 					log.Debugf("sending %v to chCreateProjects\n", phylumProjectName)
// 					chCreateProjects <- phylumProjectName
// 					go func() {
// 						err = s.PhylumCreateProject(chCreateProjects, chProjectResults)
// 						if err != nil {
// 							log.Errorf("PhylumCreateProject failed: %v\n", err)
// 							return
// 						}
// 					}()
// 				} else {
// 					log.Debugf("Found Phylum project for %v : %v\n", inProject.Name, phylumProjectName)
// 				}
// 			}
// 		}(*project)
// 	}
//
// 	// recv from channel to block until create loop is complete
// 	go func() {
// 		for item := range chProjectResults {
// 			log.Debugf("recv'd projectResult: %v\n", item.Name)
// 			(*phylumProjectMap)[item.Name] = item
// 		}
// 		close(chProjectResults)
// 	}()
// 	return nil
// }

func (s *Syringe) PhylumCreateProject(projectName string, projects chan<- *structs.PhylumProject) error {
	tempDir, err := ioutil.TempDir("", "syringe-create")
	if err != nil {
		log.Errorf("Failed to create temp directory: %v\n", err)
		return err
	}
	defer utils.RemoveTempDir(tempDir)

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

	phylumProject := structs.PhylumProject{}
	err = yaml.Unmarshal(phylumProjectData, &phylumProject)
	if err != nil {
		log.Errorf("Failed to unmarshall YAML data from created phylum project %v: %v\n", projectName, err)
		return err
	}
	projects <- &phylumProject
	return nil
}

// func (s *Syringe) OldPhylumCreateProject(projectNames <-chan string, projects chan<- structs.PhylumProject) error {
// 	for projectName := range projectNames {
// 		tempDir, err := ioutil.TempDir("", "syringe-create")
// 		if err != nil {
// 			log.Errorf("Failed to create temp directory: %v\n", err)
// 			return err
// 		}
// 		defer utils.RemoveTempDir(tempDir)
//
// 		var stdErrBytes bytes.Buffer
// 		var CreateCmdArgs = []string{"project", "create", projectName}
// 		if s.PhylumGroupName != "" {
// 			CreateCmdArgs = append(CreateCmdArgs, "-g", s.PhylumGroupName)
// 		}
// 		projectCreateCmd := exec.Command("phylum", CreateCmdArgs...)
// 		projectCreateCmd.Stderr = &stdErrBytes
// 		projectCreateCmd.Dir = tempDir
// 		err = projectCreateCmd.Run()
// 		stdErrString := stdErrBytes.String()
// 		if err != nil {
// 			log.Errorf("Failed to exec 'phylum project create %v': %v\n", projectName, err)
// 			log.Errorf("%v\n", stdErrString)
// 			return err
// 		} else {
// 			log.Debugf("Created phylum project: %v\n", projectName)
// 		}
//
// 		phylumProjectFile := filepath.Join(tempDir, ".phylum_project")
// 		phylumProjectData, err := os.ReadFile(phylumProjectFile)
// 		if err != nil {
// 			log.Errorf("Failed to read created .phylum_project file at %v: %v\n", phylumProjectFile, err)
// 			return err
// 		}
//
// 		phylumProject := structs.PhylumProject{}
// 		err = yaml.Unmarshal(phylumProjectData, &phylumProject)
// 		if err != nil {
// 			log.Errorf("Failed to unmarshall YAML data from created phylum project %v: %v\n", projectName, err)
// 			return err
// 		}
// 		projects <- phylumProject
// 	}
// 	return nil
// }

func (s *Syringe) PhylumRunAnalyze(phylumProjectFile structs.PhylumProject, lockfile *structs.VcsFile, phylumProjectName string) error {
	// create temp directory to write the lockfile content for analyze
	log.Debugf("Analyzing %v\n", phylumProjectFile.Name)
	tempDir, err := ioutil.TempDir("", "syringe-analyze")
	if err != nil {
		log.Errorf("Failed to create temp directory: %v\n", err)
		return err
	}
	defer utils.RemoveTempDir(tempDir)

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

// // LoadPidFile Read a text file of project IDs to operate on
// // The text file should have one project ID per line
// func (s *Syringe) LoadPidFile(filename string) error {
// 	var pids []string
//
// 	data, err := os.ReadFile(filename)
// 	if err != nil {
// 		log.Fatalf("Failed to read project ID file: %v error: %v\n", filename, err)
// 		return err
// 	}
// 	lineRegex := regexp.MustCompile(`^[\d|\s]+$`)
// 	lines := bytes.Split(data, []byte("\n"))
// 	for idx, line := range lines {
// 		if !lineRegex.Match(line) && wantLen(line) > 0 {
// 			log.Errorf("Failed to parse project ID file: %v - line #%v\n", filename, idx+1)
// 			log.Errorf("lines must match regex: %v", lineRegex.String())
// 			return fmt.Errorf("Failed to parse project ID file: %v - line #%v\n", filename, idx+1)
// 		}
// 		pids = append(pids, string(line))
// 	}
//
// 	s.ProjectIDs = pids
// 	return nil
// }

// func (s *Syringe) PhylumCreateProjectsFromList(projectsToCreate []string) ([]structs.PhylumProject, error) {
//	createdProjects := make([]structs.PhylumProject, 0)
//
//	for _, elem := range projectsToCreate {
//		tempDir, err := ioutil.TempDir("", "syringe-create")
//		if err != nil {
//			log.Errorf("Failed to create temp directory: %v\n", err)
//			return nil, err
//		}
//		defer utils.RemoveTempDir(tempDir)
//
//		var stdErrBytes bytes.Buffer
//		var CreateCmdArgs = []string{"project", "create", elem}
//		if s.PhylumGroupName != "" {
//			CreateCmdArgs = append(CreateCmdArgs, "-g", s.PhylumGroupName)
//		}
//		projectCreateCmd := exec.Command("phylum", CreateCmdArgs...)
//		projectCreateCmd.Stderr = &stdErrBytes
//		projectCreateCmd.Dir = tempDir
//		err = projectCreateCmd.Run()
//		stdErrString := stdErrBytes.String()
//		if err != nil {
//			log.Errorf("Failed to exec 'phylum project create %v': %v\n", elem, err)
//			log.Errorf("%v\n", stdErrString)
//			return nil, err
//		} else {
//			log.Infof("Created phylum project: %v\n", elem)
//		}
//
//		phylumProjectFile := filepath.Join(tempDir, ".phylum_project")
//		phylumProjectData, err := os.ReadFile(phylumProjectFile)
//		if err != nil {
//			log.Errorf("Failed to read created .phylum_project file at %v: %v\n", phylumProjectFile, err)
//			return nil, err
//		}
//
//		phylumProject := structs.PhylumProject{}
//		err = yaml.Unmarshal(phylumProjectData, &phylumProject)
//		if err != nil {
//			log.Errorf("Failed to unmarshall YAML data from created phylum project %v: %v\n", elem, err)
//			return nil, err
//		}
//		createdProjects = append(createdProjects, phylumProject)
//	}
//	return createdProjects, nil
// }
