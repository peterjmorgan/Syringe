package Syringe

import (
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

type GitlabClient struct {
	Client *gitlab.Client
}

func NewGitlabClient(gitlabToken string, gitlabBaseUrl string) *GitlabClient {
	gitlabClient, err := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabBaseUrl))
	if err != nil {
		log.Fatalf("Failed to create gitlab client: %v\n", err)
	}

	return &GitlabClient{Client: gitlabClient}
}

func (g *GitlabClient) listProjects() {

	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 10,
			Page:    0,
		},
		Owned: gitlab.Bool(s.MineOnly),
	}

	// ch := make(chan []*gitlab.Project)
	var localProjects []*gitlab.Project

	_, resp, err := s.Gitlab.Projects.ListProjects(opt)
	if err != nil {
		log.Errorf("Failed to list gitlab projects: %v\n", err)
		return err
	}
	count := resp.TotalPages
	// listProjectsBar := uiprogress.AddBar(count).AppendCompleted().PrependElapsed()
	lPbar := progressbar.New64(int64(count))

	for {
		// listProjectsBar.Incr()
		lPbar.Add(1)

		temp, resp, err := s.Gitlab.Projects.ListProjects(opt)
		if err != nil {
			log.Errorf("Failed to list gitlab projects: %v\n", err)
			return err
		}
		localProjects = append(localProjects, temp...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		log.Debugf("ListProjects() paging to page #%v\n", opt.Page)

	}

	log.Debugf("Len of gitlab projects: %v\n", len(localProjects))
	*projects = &localProjects
	return nil
}
