package gitee

import (
	"context"
	"errors"

	giteeapi "gitee.com/openeuler/go-gitee/gitee"
	"golang.org/x/oauth2"
)

// CommentClient interface for comment related API actions
type CommentClient interface {
	CreateComment(owner, repo string, number int, comment string) error
}

// PullRequestClient interface for pull request related API actions
type PullRequestClient interface {
	GetPullRequests(owner, repo string) ([]PullRequest, error)
	GetPullRequest(owner, repo string, number int) (*PullRequest, error)
	GetPullRequestChanges(owner, repo string, number int) ([]PullRequestChange, error)
	GetPullRequestPatch(owner, repo string, number int) ([]byte, error)
	CreatePullRequest(owner, repo, title, body, head, base string) (int, error)
	ListPullRequestComments(owner, repo string, number int) ([]Comment, error)
	ClosePullRequest(owner, repo string, number int) error
	ListPullRequestCommits(owner, repo string, number int) ([]Commit, error)
}

// RepositoryClient interface for repository related API actions
type RepositoryClient interface {
	GetBranches(owner, repo string, onlyProtected bool) ([]Branch, error)
	CreateBranch(owner, repo, branch, ref string) error
	GetFile(owner, repo, filepath, commit string) ([]byte, error)
}

// Client interface for Gitee API
type Client interface {
	PullRequestClient
	CommentClient
	RepositoryClient
}

// client Gitee API implementation
type client struct {
	token    func() []byte
	giteeAPI *giteeapi.APIClient
	context  context.Context
}

func (c *client) GetBranches(owner, repo string, onlyProtected bool) ([]Branch, error) {
	opts := &giteeapi.GetV5ReposOwnerRepoBranchesOpts{}
	bs, _, err := c.giteeAPI.RepositoriesApi.GetV5ReposOwnerRepoBranches(c.context, owner, repo, opts)
	if err != nil {
		return nil, err
	}
	branches := make([]Branch, 0)
	for _, branch := range bs {
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

func (c *client) CreateBranch(owner, repo, branchName, ref string) error {
	param := giteeapi.CreateBranchParam{
		Refs:       ref,
		BranchName: branchName,
	}

	_, _, err := c.giteeAPI.RepositoriesApi.PostV5ReposOwnerRepoBranches(c.context, owner, repo, param)
	if err != nil {
		return errors.New(string(err.(giteeapi.GenericSwaggerError).Body()))
	}
	return nil
}

func (c *client) GetFile(owner, repo, filepath, commit string) ([]byte, error) {
	panic("implement me")
}

func (c *client) GetPullRequests(owner, repo string) ([]PullRequest, error) {
	panic("implement me")
}

func (c *client) GetPullRequest(owner, repo string, number int) (*PullRequest, error) {
	panic("implement me")
}

func (c *client) GetPullRequestChanges(owner, repo string, number int) ([]PullRequestChange, error) {
	panic("implement me")
}

func (c *client) GetPullRequestPatch(owner, repo string, number int) ([]byte, error) {
	panic("implement me")
}

func (c *client) CreatePullRequest(owner, repo, title, body, head, base string) (int, error) {
	param := giteeapi.CreatePullRequestParam{
		Title: title,
		Body:  body,
		Head:  head,
		Base:  base,
	}
	pullRequest, _, err := c.giteeAPI.PullRequestsApi.PostV5ReposOwnerRepoPulls(c.context, owner, repo, param)
	if err != nil {
		return -1, errors.New(string(err.(giteeapi.GenericSwaggerError).Body()))
	}
	number := int(pullRequest.Id)
	return number, nil
}

func (c *client) ListPullRequestComments(owner, repo string, number int) ([]Comment, error) {
	opts := &giteeapi.GetV5ReposOwnerRepoPullsNumberCommentsOpts{}
	result, _, err := c.giteeAPI.PullRequestsApi.GetV5ReposOwnerRepoPullsNumberComments(c.context, owner, repo, int32(number), opts)
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

func (c *client) ClosePullRequest(owner, repo string, number int) error {
	panic("implement me")
}

func (c *client) CreateComment(owner, repo string, number int, comment string) error {
	body := giteeapi.PullRequestCommentPostParam{
		Body: comment,
	}
	_, _, err := c.giteeAPI.PullRequestsApi.PostV5ReposOwnerRepoPullsNumberComments(c.context, owner, repo, int32(number), body)
	return err
}

func (c *client) ListPullRequestCommits(owner, repo string, number int) ([]Commit, error) {
	opts := &giteeapi.GetV5ReposOwnerRepoPullsNumberCommitsOpts{}
	cs, _, err := c.giteeAPI.PullRequestsApi.GetV5ReposOwnerRepoPullsNumberCommits(c.context, owner, repo, int32(number), opts)

	commits := make([]Commit, 0)
	if err != nil {
		return nil, err
	}
	for _, c := range cs {
		commit := Commit{URL: c.Url,
			Sha:     c.Sha,
			HTMLURL: c.HtmlUrl,
			Message: c.Commit.Message,
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

//NewClient client to access Gitee
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
