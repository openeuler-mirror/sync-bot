package gitee

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	giteeapi "gitee.com/openeuler/go-gitee/gitee"
	"github.com/antihax/optional"
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
	CreatePullRequest(owner, repo, title, body, head, base string, pruneSourceBranch bool) (int, error)
	ListPullRequestComments(owner, repo string, number int) ([]Comment, error)
	ClosePullRequest(owner, repo string, number int) error
	ListPullRequestCommits(owner, repo string, number int) ([]PullRequestCommit, error)
	ListPullRequestIssues(owner, repo string, number int) ([]Issue, error)
}

// RepositoryClient interface for repository related API actions
type RepositoryClient interface {
	GetBranches(owner, repo string, onlyProtected bool) ([]Branch, error)
	GetBranch(owner, repo, branch string) (Branch, error)
	CreateBranch(owner, repo, branch, ref string) error
	GetTextFile(owner, repo, filepath, ref string) (string, error)
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

	drop_branches := GetDroppedBranches()

	for _, branch := range bs {
		if drop_branches[branch.Name] {
			continue
		}
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

func (c *client) GetBranch(owner, repo, branch string) (Branch, error) {
	var b Branch
	opts := &giteeapi.GetV5ReposOwnerRepoBranchesBranchOpts{}
	b1, _, err := c.giteeAPI.RepositoriesApi.GetV5ReposOwnerRepoBranchesBranch(c.context, owner, repo, branch, opts)
	if err != nil {
		return b, err
	}
	return Branch{
		Name:      b1.Name,
		Protected: b1.Protected,
	}, nil
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

func (c *client) GetTextFile(owner, repo, filepath, ref string) (string, error) {
	param := &giteeapi.GetV5ReposOwnerRepoContentsPathOpts{
		Ref: optional.NewString(ref),
	}
	content, _, err := c.giteeAPI.RepositoriesApi.GetV5ReposOwnerRepoContentsPath(c.context, owner, repo, filepath, param)
	if err != nil {
		return "", errors.New(string(err.(giteeapi.GenericSwaggerError).Body()))
	}

	data, err := base64.StdEncoding.DecodeString(content.Content)
	if err != nil {
		return "", err
	}
	return string(data), nil
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

func (c *client) CreatePullRequest(owner, repo, title, body, head, base string, pruneSourceBranch bool) (int, error) {
	param := giteeapi.CreatePullRequestParam{
		Title:             title,
		Body:              body,
		Head:              head,
		Base:              base,
		PruneSourceBranch: pruneSourceBranch,
	}
	pullRequest, _, err := c.giteeAPI.PullRequestsApi.PostV5ReposOwnerRepoPulls(c.context, owner, repo, param)
	if err != nil {
		return 0, errors.New(string(err.(giteeapi.GenericSwaggerError).Body()))
	}
	number := int(pullRequest.Number)
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
			User: User{
				Email:    c.User.Email,
				HTMLURL:  c.User.HtmlUrl,
				ID:       int(c.User.Id),
				Name:     c.User.Name,
				Username: c.User.Login,
			},
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

func (c *client) ListPullRequestCommits(owner, repo string, number int) ([]PullRequestCommit, error) {
	opts := &giteeapi.GetV5ReposOwnerRepoPullsNumberCommitsOpts{}
	cs, _, err := c.giteeAPI.PullRequestsApi.GetV5ReposOwnerRepoPullsNumberCommits(c.context, owner, repo, int32(number), opts)

	var commits []PullRequestCommit
	if err != nil {
		return nil, err
	}
	for _, c := range cs {
		fmt.Println("[ListPullRequestCommits:200]", c);
		if c.Author == nil {
			c.Author = User{
				Email:    "",
				HtmlUrl:  "",
				Id:       0,
				Name:     "",
				Login:    "",
			}
		}
		
		commit := PullRequestCommit{
			Author: User{
				Email:    c.Author.Email,
				HTMLURL:  c.Author.HtmlUrl,
				ID:       int(c.Author.Id),
				Name:     c.Author.Name,
				Username: c.Author.Login,
			},
			CommentsURL: c.CommentsUrl,
			Commit: GitCommit{
				Author: GitUser{
					Date:  c.Commit.Author.Date,
					Email: c.Commit.Author.Email,
					Name:  c.Commit.Author.Name,
				},
				CommentCount: int(c.Commit.CommentCount),
				Committer: GitUser{
					Date:  c.Commit.Committer.Date,
					Email: c.Commit.Committer.Email,
					Name:  c.Commit.Committer.Name,
				},
				Message: c.Commit.Message,
				URL:     c.Commit.Url,
			},
			Committer: User{
				Email:    c.Committer.Email,
				HTMLURL:  c.Committer.HtmlUrl,
				ID:       int(c.Committer.Id),
				Name:     c.Committer.Name,
				Username: c.Committer.Login,
			},
			HTMLURL: c.HtmlUrl,
			Parents: Parents{
				Sha: c.Parents.Sha,
				URL: c.Parents.Url,
			},
			URL: c.Url,
			Sha: c.Sha,
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

func (c *client) ListPullRequestIssues(owner, repo string, number int) ([]Issue, error) {
	opts := &giteeapi.GetV5ReposOwnerRepoPullsNumberIssuesOpts{}
	is, _, err := c.giteeAPI.PullRequestsApi.GetV5ReposOwnerRepoPullsNumberIssues(c.context, owner, repo, int32(number), opts)
	var issues []Issue
	if err != nil {
		return nil, err
	}
	for _, i := range is {
		issue := Issue{
			Body:      i.Body,
			HTMLURL:   i.HtmlUrl,
			ID:        int(i.Id),
			IssueType: i.IssueType,
			Number:    i.Number,
			State:     i.State,
			Title:     i.Title,
		}
		issues = append(issues, issue)
	}
	return issues, nil
}

//NewClient client to access Gitee
func NewClient(getToken func() []byte) Client {
	// oauth
	oauthSecret := string(getToken())
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: oauthSecret},
	)
	// configuration
	giteeConf := giteeapi.NewConfiguration()
	giteeConf.HTTPClient = oauth2.NewClient(ctx, ts)

	return &client{
		token:    getToken,
		giteeAPI: giteeapi.NewAPIClient(giteeConf),
	}
}
