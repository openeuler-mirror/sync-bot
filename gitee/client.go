package gitee

import (
	"context"

	giteeapi "gitee.com/openeuler/go-gitee/gitee"
	"golang.org/x/oauth2"
)

// CommentClient interface for comment related API actions
type CommentClient interface {
	CreateComment(org, repo string, number int, comment string) error
}

// PullRequestClient interface for pull request related API actions
type PullRequestClient interface {
	GetPullRequests(org, repo string) ([]PullRequest, error)
	GetPullRequest(org, repo string, number int) (*PullRequest, error)
	GetPullRequestChanges(org, repo string, number int) ([]PullRequestChange, error)
	GetPullRequestPatch(org, repo string, number int) ([]byte, error)
	CreatePullRequest(org, repo, title, body, head, base string, canModify bool) (int, error)
	ListPullRequestComments(org, repo string, number int) ([]Comment, error)
	ClosePR(org, repo string, number int) error
}

// RepositoryClient interface for repository related API actions
type RepositoryClient interface {
	GetBranches(org, repo string, onlyProtected bool) ([]Branch, error)
	GetFile(org, repo, filepath, commit string) ([]byte, error)
}

// Client interface for Gitee API
type Client interface {
	PullRequestClient
	CommentClient
	RepositoryClient
}

// client
type client struct {
	token   func() []byte
	gitAPI  *giteeapi.APIClient
	context context.Context
}

func (c *client) GetBranches(org, repo string, onlyProtected bool) ([]Branch, error) {
	panic("implement me")
}

func (c *client) GetFile(org, repo, filepath, commit string) ([]byte, error) {
	panic("implement me")
}

func (c *client) GetPullRequests(org, repo string) ([]PullRequest, error) {
	panic("implement me")
}

func (c *client) GetPullRequest(org, repo string, number int) (*PullRequest, error) {
	panic("implement me")
}

func (c *client) GetPullRequestChanges(org, repo string, number int) ([]PullRequestChange, error) {
	panic("implement me")
}

func (c *client) GetPullRequestPatch(org, repo string, number int) ([]byte, error) {
	panic("implement me")
}

func (c *client) CreatePullRequest(org, repo, title, body, head, base string, canModify bool) (int, error) {
	panic("implement me")
}

func (c *client) ListPullRequestComments(org, repo string, number int) ([]Comment, error) {
	panic("implement me")
}

func (c *client) ClosePR(org, repo string, number int) error {
	panic("implement me")
}

func (c *client) CreateComment(org, repo string, number int, comment string) error {
	panic("implement me")
}

func NewClient(getToken func() []byte) Client {
	// oauth
	oauthSecret := string(getToken())
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: string(oauthSecret)},
	)
	// configuration
	giteeConf := giteeapi.NewConfiguration()
	giteeConf.HTTPClient = oauth2.NewClient(ctx, ts)

	return &client{
		token:  getToken,
		gitAPI: giteeapi.NewAPIClient(giteeConf),
	}
}
