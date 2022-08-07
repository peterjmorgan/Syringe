package Syringe

type ClientType int

const (
	GithubType = iota
	GitlabType
)

type Client interface {
	listProjects()
}

func NewClient(clientType ClientType, token string, baseUrl string) (Client, error) {
	var c Client
	var err error

	switch clientType {
	case GithubType: // github
		c = NewGitlabClient(token, baseUrl)
	case GitlabType: // gitlab
		c = NewGithubClient(token, baseUrl)
	}

	return c, err
}
