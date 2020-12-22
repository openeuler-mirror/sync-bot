package gitee

import (
	"context"

	giteeapi "gitee.com/openeuler/go-gitee/gitee"
	"github.com/sirupsen/logrus"
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
	CreatePullRequest(org, repo, title, body, head, base string) (int, error)
	ListPullRequestComments(org, repo string, number int) ([]Comment, error)
	ClosePullRequest(org, repo string, number int) error
	ListPullRequestCommits(org, repo string, number int) ([]Commit, error)
}

// RepositoryClient interface for repository related API actions
type RepositoryClient interface {
	GetBranches(org, repo string, onlyProtected bool) ([]Branch, error)
	CreateBranch(org, repo, branch, ref string) error
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
	token    func() []byte
	giteeAPI *giteeapi.APIClient
	context  context.Context
}

func (c *client) GetBranches(org, repo string, onlyProtected bool) ([]Branch, error) {
	logrus.Infoln("GetBranches", org, repo, onlyProtected)
	opts := &giteeapi.GetV5ReposOwnerRepoBranchesOpts{}
	branches_, _, err := c.giteeAPI.RepositoriesApi.GetV5ReposOwnerRepoBranches(c.context, org, repo, opts)
	if err != nil {
		return nil, err
	}
	branches := make([]Branch, 0)
	for _, branch := range branches_ {
		if onlyProtected && !branch.Protected {
			continue
		}
		branches = append(branches, Branch{
			Name:      branch.Name,
			Protected: branch.Protected,
		})
	}
	return branches, nil
}

func (c *client) CreateBranch(org, repo, branchName, ref string) error {
	param := giteeapi.CreateBranchParam{
		Refs:       ref,
		BranchName: branchName,
	}
	_, _, err := c.giteeAPI.RepositoriesApi.PostV5ReposOwnerRepoBranches(c.context, org, repo, param)
	if err != nil {
		return err
	}
	return nil
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

func (c *client) CreatePullRequest(org, repo, title, body, head, base string) (int, error) {
	param := giteeapi.CreatePullRequestParam{
		Title: title,
		Body:  body,
		Head:  head,
		Base:  base,
	}
	pullRequest, _, err := c.giteeAPI.PullRequestsApi.PostV5ReposOwnerRepoPulls(c.context, org, repo, param)
	if err != nil {
		logrus.Errorln("CreatePullRequest failed.")
		return -1, err
	}
	number := int(pullRequest.Id)
	return number, nil
}

func (c *client) ListPullRequestComments(org, repo string, number int) ([]Comment, error) {
	opts := &giteeapi.GetV5ReposOwnerRepoPullsNumberCommentsOpts{}
	result, _, err := c.giteeAPI.PullRequestsApi.GetV5ReposOwnerRepoPullsNumberComments(c.context, org, repo, int32(number), opts)
	if err != nil {
		return nil, err
	}
	comments := make([]Comment, 0)
	for _, c := range result {
		comment := Comment{
			Body:    c.Body,
			HTMLURL: c.HtmlUrl,
			ID:      int(c.Id),
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

func (c *client) ClosePullRequest(org, repo string, number int) error {
	panic("implement me")
}

func (c *client) CreateComment(org, repo string, number int, comment string) error {
	body := giteeapi.PullRequestCommentPostParam{
		Body: comment,
	}
	_, _, err := c.giteeAPI.PullRequestsApi.PostV5ReposOwnerRepoPullsNumberComments(c.context, org, repo, int32(number), body)
	return err
}

func (c *client) ListPullRequestCommits(org, repo string, number int) ([]Commit, error) {
	opts := &giteeapi.GetV5ReposOwnerRepoPullsNumberCommitsOpts{}
	commits_, _, err := c.giteeAPI.PullRequestsApi.GetV5ReposOwnerRepoPullsNumberCommits(c.context, org, repo, int32(number), opts)

	commits := make([]Commit, 0)
	if err != nil {
		return nil, err
	}
	for _, c := range commits_ {
		commit := Commit{Url: c.Url,
			Sha:     c.Sha,
			HtmlUrl: c.HtmlUrl,
			Message: c.Commit.Message,
		}
		commits = append(commits, commit)
	}
	return commits, nil
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
		token:    getToken,
		giteeAPI: giteeapi.NewAPIClient(giteeConf),
	}
}
