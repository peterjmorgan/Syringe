package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/peterjmorgan/Syringe/internal/structs"
	"github.com/peterjmorgan/Syringe/internal/utils"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"golang.org/x/oauth2"
)

type GithubClient struct {
	Client  *github.Client
	Ctx     context.Context
	OrgName string
}

// func NewGithubClient(envMap map[string]string, opts *structs.SyringeOptions) *GithubClient {
func NewGithubClient(configData *structs.ConfigThing, opts *structs.SyringeOptions) *GithubClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: configData.VcsToken},
	)

	oac := oauth2.NewClient(ctx, ts)

	oac.Transport = utils.NewEtagTransport(oac.Transport)
	oac.Transport = utils.NewRateLimitTransport(oac.Transport, utils.WithWriteDelay(5), utils.WithReadDelay(1))

	gh := github.NewClient(oac)
	return &GithubClient{
		Client:  gh,
		Ctx:     ctx,
		OrgName: configData.Associated["githubOrg"],
	}
}

func (g *GithubClient) ListProjects() (*[]*structs.SyringeProject, error) {
	var localProjects []*structs.SyringeProject
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 200},
	}

	_, resp, err := g.Client.Repositories.ListByOrg(g.Ctx, g.OrgName, opt)
	if err != nil {
		log.Errorf("Failed to get github repositories: %v\n", err)
		return nil, err
	}
	count := resp.LastPage
	listProjectsPB := progressbar.New64(int64(count))

	for {
		listProjectsPB.Add(1)

		githubRepos, resp, err := g.Client.Repositories.ListByOrg(g.Ctx, g.OrgName, opt)
		if rl_err, ok := err.(*github.RateLimitError); ok {
			log.Printf("ListByOrg ratelimited. Pausing until %s", rl_err.Rate.Reset.Time.String())
			time.Sleep(time.Until(rl_err.Rate.Reset.Time))
			continue
		} else if err != nil {
			log.Errorf("Failed to get github repositories: %v\n", err)
			return nil, err
		}

		for _, repo := range githubRepos {
			localProjects = append(localProjects, &structs.SyringeProject{
				Id:        *repo.ID,
				Name:      *repo.Name,
				Branch:    *repo.DefaultBranch,
				Lockfiles: []*structs.VcsFile{},
				CiFiles:   []*structs.VcsFile{},
				Hydrated:  false,
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return &localProjects, nil
}

func handleErr(callerName string, err error) bool {
	if rl_err, ok := err.(*github.RateLimitError); ok {
		log.Printf("%v ratelimited. Pausing until %s", callerName, rl_err.Rate.Reset.Time.String())
		time.Sleep(time.Until(rl_err.Rate.Reset.Time))
		return true
	} else {
		return false
	}
}

// unfinished
func (g *GithubClient) SearchOrgForFilename(orgRepo string, filename string, fileExtension string) (*github.CodeSearchResult, error) {
	orgNameQuery := fmt.Sprintf("org:%v", orgRepo)
	//filenameQuery := fmt.Sprintf("filename:%v", filename)
	//filenameQuery := "package-lock.json"
	fileExtenionQuery := fmt.Sprintf("extension:%v", fileExtension)
	//fileContentsQuery := "in:path"

	searchQuery := fmt.Sprintf("%s %s",
		orgNameQuery,
		//filenameQuery,
		fileExtenionQuery,
		//fileContentsQuery,
	)

	result, resp, err := g.Client.Search.Code(g.Ctx, searchQuery, nil)
	if err != nil {
		log.Printf("search failed: %v\n", err)
		return nil, err
	}
	_ = resp

	return result, nil
}

// unfinished
func (g *GithubClient) SearchRepoForFilename(org string, repo string, filename string, fileExtension string) (*github.CodeSearchResult, error) {
	orgNameQuery := fmt.Sprintf("org:%v", org)
	repoQuery := fmt.Sprintf("repo:%v", repo)
	filenameQuery := fmt.Sprintf("filename:%v", filename)
	//filenameQuery := "package-lock.json"
	fileExtenionQuery := fmt.Sprintf("extension:%v", fileExtension)
	//fileContentsQuery := "in:path"

	searchQuery := fmt.Sprintf("%s %s %s %s",
		orgNameQuery,
		repoQuery,
		filenameQuery,
		fileExtenionQuery,
		//fileContentsQuery,
	)

	result, resp, err := g.Client.Search.Code(g.Ctx, searchQuery, nil)
	if err != nil {
		log.Printf("search failed: %v\n", err)
		return nil, err
	}
	_ = resp

	return result, nil
}

// GetTree: only to be used when Truncated is set in ListFiles and we have to do it iteratively
func (g *GithubClient) GetTree(repoName string, commitSHA string, treePath string) (*github.Tree, error) {
	var resultsTree github.Tree

	// note that we're descending
	log.Warnf("GH_GetTree for repo:%v\n", repoName)

	// first try a recurisve request to GetTree
	ghTree, _, err := g.Client.Git.GetTree(g.Ctx, g.OrgName, repoName, commitSHA, true)
	if err != nil {
		if !handleErr("GetTree", err) {
			log.Errorf("Failed to GetTree from %v: %v\n", repoName, err)
			return nil, err
		}
	}

	// If the response is truncated, go iterative
	if *ghTree.Truncated {
		ghTree, _, err = g.Client.Git.GetTree(g.Ctx, g.OrgName, repoName, commitSHA, false)
		if err != nil {
			if !handleErr("GetTree", err) {
				log.Errorf("Failed to GetTree from %v: %v\n", repoName, err)
				return nil, err
			}
		}

	}

	for _, treeEntry := range ghTree.Entries {
		switch *treeEntry.Type {
		case "blob": // file
			tempEntry := new(github.TreeEntry)
			tempEntry = &treeEntry
			var tempNewPath string
			if treePath != "" && *ghTree.Truncated {
				tempNewPath = fmt.Sprintf("%v/%v", treePath, *tempEntry.Path)
			} else {
				tempNewPath = *tempEntry.Path
			}
			tempEntry.Path = &tempNewPath
			resultsTree.Entries = append(resultsTree.Entries, *tempEntry)
		case "tree": // directory
			if *ghTree.Truncated {
				tempTree, err := g.GetTree(repoName, *treeEntry.SHA, *treeEntry.Path)
				if err != nil {
					if !handleErr("GetTree (subtree)", err) {
						log.Errorf("Failed to GetTree (subtree) from %v: %v\n", repoName, err)
						return nil, err
					}
				}
				// next append it
				resultsTree.Entries = append(resultsTree.Entries, tempTree.Entries...)
			}

		default:
			log.Warnf("GH_GetTree: found treeEntry type: %v\n", *treeEntry.Type)
		}
	}

	return &resultsTree, nil
}

func (g *GithubClient) ListFiles(repoName string, branch string) (*github.Tree, error) {
	var resultsTree github.Tree

	commits, resp, err := g.Client.Repositories.ListCommits(g.Ctx, g.OrgName, repoName, &github.CommitsListOptions{})
	if err != nil {
		log.Errorf("GH_ListFiles: failed to ListCommits from %v: %v\n", repoName, err)
		log.Errorf("%v\n", resp.StatusCode)
		return nil, err
	}

	// Get the latest commmit SHA for the repo and branch
	lastCommitSHA := *commits[0].SHA

	// Get the tree of objects based on the commit SHA
	ghTree, resp, err := g.Client.Git.GetTree(g.Ctx, g.OrgName, repoName, lastCommitSHA, true)
	if err != nil {
		log.Errorf("Failed to GetTree from %v: %v\n", repoName, err)
		return nil, err
	}
	//resultsTree.Truncated = ghTree.Truncated

	if *ghTree.Truncated {
		repo, _, err := g.Client.Repositories.Get(g.Ctx, g.OrgName, repoName)
		if err != nil {
			log.Errorf("GH_ListFiles: Failed to Get Repo %v: %v\n", repoName, err)
			return nil, err
		}

		// Check if we're looking at a fork
		if *repo.Fork {
			log.Warnf("GH_ListFiles: Skipping %v as it is a fork AND has an incredibly large GitTree", repoName)
			return nil, nil
		}

		// No, this is just a monster project
		log.Infof("GH_ListFiles: Found an incredibly large GitTree: %v - descending\n", repoName)
		tempTree, err := g.GetTree(repoName, lastCommitSHA, "")
		if err != nil {
			log.Errorf("Failed to GetTree from %v: %v\n", repoName, err)
			return nil, err
		}
		resultsTree = *tempTree

	} else {
		for _, treeEntry := range ghTree.Entries {
			switch *treeEntry.Type {
			case "blob": // file
				resultsTree.Entries = append(resultsTree.Entries, treeEntry)
			case "tree": // directory
				// This is handled by the check for Truncated above
			default:
				log.Warnf("GH_ListFiles: found treeEntry type: %v\n", *treeEntry.Type)
			}
		}
	}

	return &resultsTree, nil
}

// func (g *GithubClient) GetLockfilesByProject(repoName string, mainBranchName string) ([]*structs.VcsFile, error) {
func (g *GithubClient) GetLockfilesByProject(projectId int64, mainBranchName string) ([]*structs.VcsFile, error) {
	var retLockfiles []*structs.VcsFile

	// Get Repo name via ID
	repo, _, err := g.Client.Repositories.GetByID(g.Ctx, projectId)
	if err != nil {
		log.Errorf("Failed to GetRepoByID %v: %v\n", projectId, err)
		return nil, err
	}

	projectTree, err := g.ListFiles(*repo.Name, mainBranchName)
	if err != nil {
		log.Errorf("Failed to ListFiles for %v: %v\n", *repo.Name, err)
		return nil, err
	}
	if projectTree == nil {
		return nil, nil
	}

	supportedLockfiles := utils.GetSupportedLockfiles()

	for _, file := range projectTree.Entries {
		fileName := filepath.Base(*file.Path)
		if slices.Contains(supportedLockfiles, fileName) || strings.HasSuffix(fileName, ".csproj") {
			log.Debugf("Lockfile: %v in %v from project: %v\n", fileName, *file.Path, *repo.Name)
			contentHandle, err := g.Client.Repositories.DownloadContents(g.Ctx, g.OrgName, *repo.Name, *file.Path, &github.RepositoryContentGetOptions{})
			if err != nil {
				log.Errorf("Failed to DownloadContents for %v in repo:%v: %v", *file.Path, *repo.Name, err)
				return nil, err
			}
			b, err := ioutil.ReadAll(contentHandle)
			if err != nil {
				log.Errorf("Failed to read bytes from %v: %v\n", *file.Path, err)
				return nil, err
			}
			retLockfiles = append(retLockfiles, &structs.VcsFile{
				Name:          fileName,
				Path:          *file.Path,
				Id:            *file.SHA,
				Content:       b,
				PhylumProject: nil,
			})

		}
	}
	return retLockfiles, nil
}
