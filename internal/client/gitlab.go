package client

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/peterjmorgan/Syringe/internal/structs"
	utils "github.com/peterjmorgan/Syringe/internal/utils"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/exp/slices"
	"golang.org/x/time/rate"
)

type GitlabClient struct {
	Client   *gitlab.Client
	MineOnly bool
}

// func NewGitlabClient(envMap map[string]string, mineOnly bool, ratelimit int, proxyUrl string) *GitlabClient {
func NewGitlabClient(envMap map[string]string, opts *structs.SyringeOptions) *GitlabClient {
	var gitlabClient *gitlab.Client
	var err error
	var mineOnly bool = false

	clientOptions := []gitlab.ClientOptionFunc{
		gitlab.WithCustomRetry(retryablehttp.DefaultRetryPolicy),
	}

	if vcsUrl, ok := envMap["vcsUrl"]; ok {
		gitlab.WithBaseURL(vcsUrl)
	}

	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		if opts.ProxyUrl != "" {
			theProxyUrl, err := url.Parse(opts.ProxyUrl)
			if err != nil {
				log.Errorf("failed to parse burpurl: %v\n", err)
			}

			proxyHttp := &http.Client{
				Transport: &http.Transport{
					Proxy:           http.ProxyURL(theProxyUrl),
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}
			clientOptions = append(clientOptions, gitlab.WithHTTPClient(proxyHttp))
		}
		if opts.RateLimit != 0 {
			clientOptions = append(clientOptions, gitlab.WithCustomLimiter(rate.NewLimiter(rate.Every(time.Second), opts.RateLimit)))
		}

		mineOnly = opts.MineOnly
	}

	gitlabClient, err = gitlab.NewClient(envMap["vcsToken"], clientOptions...)

	if err != nil {
		log.Fatalf("Failed to create gitlab client: %v\n", err)
	}

	return &GitlabClient{
		Client:   gitlabClient,
		MineOnly: mineOnly,
	}
}

func (g *GitlabClient) CheckResponse(resp *gitlab.Response) error {
	return nil
}

func (g *GitlabClient) ListProjects() (*[]*structs.SyringeProject, error) {
	var localProjects []*structs.SyringeProject
	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 50,
			Page:    0,
		},
		Owned: gitlab.Bool(g.MineOnly),
	}

	_, resp, err := g.Client.Projects.ListProjects(opt)
	if err != nil {
		log.Errorf("Failed to list gitlab projects: %v\n", err)
		return nil, err
	}
	count := resp.TotalPages
	listProjectsPB := progressbar.NewOptions(count, progressbar.OptionSetDescription("Getting Projects"))

	for {
		listProjectsPB.Add(1)

		gitlabProjects, resp, err := g.Client.Projects.ListProjects(opt)
		if err != nil {
			log.Errorf("Failed to list gitlab projects: %v\n", err)
		}

		// Iterate through gitlabProjects and create SyringeProjects for each
		for _, gitlabProject := range gitlabProjects {
			localProjects = append(localProjects, &structs.SyringeProject{
				Id:        int64(gitlabProject.ID),
				Name:      gitlabProject.Name,
				Branch:    gitlabProject.DefaultBranch,
				Lockfiles: []*structs.VcsFile{},
				CiFiles:   []*structs.VcsFile{},
				Hydrated:  false,
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		log.Debugf("ListProjects() paging to page #%v\n", opt.Page)
	}

	log.Debugf("Len of gitlab projects: %v\n", len(localProjects))
	return &localProjects, nil
}

func (g *GitlabClient) ListFiles(projectId int64, branch string) ([]*gitlab.TreeNode, error) {
	files, _, err := g.Client.Repositories.ListTree(int(projectId), &gitlab.ListTreeOptions{
		Path:      gitlab.String("/"),
		Ref:       gitlab.String(branch),
		Recursive: gitlab.Bool(true),
	})
	if err != nil {
		// log.Warnf("Failed to ListTree from %v: %v\n", projectId, err)
		return nil, err
	}
	return files, nil
}

func (g *GitlabClient) GetLockfilesByProject(projectId int64, mainBranchName string) ([]*structs.VcsFile, error) {
	// TODO: check if mainBranchName isn't set or is "". Bail if that's the case, there are repos without code and will not have a branch
	var retLockFiles []*structs.VcsFile

	projectFiles, err := g.ListFiles(projectId, mainBranchName)
	if err != nil {
		// log.Errorf("Failed to ListFiles for %v on branch %v\n", projectId, mainBranchName)
		return nil, err
	}

	supportedLockfiles := utils.GetSupportedLockfiles()

	for _, file := range projectFiles {
		if slices.Contains(supportedLockfiles, file.Name) || strings.HasSuffix(file.Name, ".csproj") {
			log.Debugf("Lockfile: %v in %v from projectID: %v\n", file.Name, file.Path, projectId)
			data, _, err := g.Client.RepositoryFiles.GetRawFile(int(projectId), file.Path, &gitlab.GetRawFileOptions{&mainBranchName})
			if err != nil {
				log.Errorf("Failed to GetRawFile for %v in projectId %v: %v\n", file.Name, projectId, err)
			}

			rec := structs.VcsFile{file.Name, file.Path, file.ID, data, nil}
			retLockFiles = append(retLockFiles, &rec)
		}
	}
	return retLockFiles, nil
}

// func (g *GitlabClient) PrintProjectVariables(projectId int) error {
// 	variables, _, err := g.Client.ProjectVariables.ListVariables(projectId, &gitlab.ListProjectVariablesOptions{})
// 	if err != nil {
// 		log.Errorf("Failed to list project variables from %v: %v\n", projectId, err)
// 		return err
// 	}
// 	for _, variable := range variables {
// 		log.Infof("Variable: %v:%v\n", variable.Key, variable.Value)
// 	}
// 	return nil
// }
