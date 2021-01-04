package hook

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"sync-bot/gitee"
)

func (s *Server) OpenPullRequest(e gitee.PullRequestEvent) {
	owner := e.Repository.Namespace
	repo := e.Repository.Path
	number := e.PullRequest.Number
	title := e.PullRequest.Title
	targetBranch := e.PullRequest.Base.Ref
	logrus.WithFields(logrus.Fields{
		"owner":  owner,
		"repo":   repo,
		"number": number,
		"title":  title,
	}).Infoln("OpenPullRequest")
	if matchTitle(title) {
		logrus.Infoln("Ignore PullRequest created by sync-bot")
		return
	}
	s.replySyncCheck(owner, repo, number, targetBranch)
}

func (s *Server) MergePullRequest(e gitee.PullRequestEvent) {
	owner := e.Repository.Namespace
	repo := e.Repository.Path
	number := e.PullRequest.Number
	logrus.WithFields(logrus.Fields{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}).Infoln("MergePullRequest")

	comments, err := s.GiteeClient.ListPullRequestComments(owner, repo, number)
	if err != nil {
		logrus.Errorln("List PullRequest comments failed", err)
		return
	}

	// find the last /sync command
	for i := range comments {
		comment := comments[len(comments)-1-i]
		body := comment.Body
		if matchSync(body) {
			logrus.WithFields(logrus.Fields{
				"comment": body,
			}).Infoln("match /sync command")
			s.sync(owner, repo, e.PullRequest, body)
			return
		}
	}
	logrus.WithFields(logrus.Fields{
		"comments": comments,
	}).Warnln("Not found valid /sync command in pr comments")
}

func (s *Server) pick() bool {
	panic("implement me")
}

func (s *Server) merge(owner string, repo string, opt SyncCmdOption, pr gitee.PullRequest, title string, body string) bool {
	number := pr.Number
	ref := pr.Head.Sha

	for _, branch := range opt.branches {
		// create temp branch
		tempBranch := fmt.Sprintf("sync-pr%v-to-%v", number, branch)
		err := s.GiteeClient.CreateBranch(owner, repo, tempBranch, ref)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tempBranch": tempBranch,
			}).Errorln("Create temp branch failed:", err)
			// TODO: check if branch exist
		} else {
			logrus.Infoln("Create temp branch:", branch)
		}
		// create pull request
		num, err := s.GiteeClient.CreatePullRequest(owner, repo, title, body, tempBranch, branch, true)
		if err != nil {
			logrus.Errorln("Create PullRequest failed:", err)
		} else {
			logrus.Infoln("Create PullRequest:", num)
		}
	}
	return true
}

func (s *Server) overwrite() bool {
	panic("implement me")
}

func (s *Server) sync(owner string, repo string, pr gitee.PullRequest, command string) bool {
	number := pr.Number

	opt, err := parse(command)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"opt": opt,
		}).Errorln("Parse /sync command failed:", err)
		return false
	}

	issues, err := s.GiteeClient.ListPullRequestIssues(owner, repo, number)
	if err != nil {
		logrus.Errorln("List issues in pull request failed:", err)
		return false
	}

	commits, err := s.GiteeClient.ListPullRequestCommits(owner, repo, number)
	if err != nil {
		logrus.Errorln("List commits failed:", err)
		return false
	}
	for i := range commits {
		commits[i].Commit.Message = strings.ReplaceAll(commits[i].Commit.Message, "\n", "<br>")
	}

	title := fmt.Sprintf("[sync-bot] from PR-%v: %v", number, pr.Title)

	bodyStruct := struct {
		PR      string
		Issues  []gitee.Issue
		Commits []gitee.PullRequestCommit
	}{
		PR:      pr.HTMLURL,
		Issues:  issues,
		Commits: commits,
	}

	body, err := executeTemplate(syncPRBodyTmpl, bodyStruct)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tmpl":       syncPRBodyTmpl,
			"bodyStruct": bodyStruct,
		}).Errorln("Execute template failed:", err)
		return false
	}

	switch opt.strategy {
	case Pick:
		return s.pick()
	case Merge:
		return s.merge(owner, repo, opt, pr, title, body)
	case Overwrite:
		return s.overwrite()
	default:
	}
	return false
}

func (s *Server) ClosePullRequest(e gitee.PullRequestEvent) {
	owner := e.Repository.Namespace
	repo := e.Repository.Path
	number := e.PullRequest.Number

	logrus.WithFields(logrus.Fields{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}).Infoln("ClosePullRequest")

	// TODO: close issue created by sync-bot, delete temp branch

}

func (s *Server) HandlePullRequestEvent(e gitee.PullRequestEvent) {
	title := e.PullRequest.Title
	switch e.Action {
	case gitee.ActionOpen:
		if matchTitle(title) {
			logrus.WithFields(logrus.Fields{
				"title": title,
			}).Infoln("StateOpen Pull Request which created by sync-bot, ignore it.")
		} else {
			s.OpenPullRequest(e)
		}
	case gitee.ActionMerge:
		if matchTitle(title) {
			logrus.WithFields(logrus.Fields{
				"title": title,
			}).Infoln("Merge Pull Request which created by sync-bot, ignore it.")
		} else {
			s.MergePullRequest(e)
		}
	case gitee.ActionClose:
		if !matchTitle(title) {
			logrus.WithFields(logrus.Fields{
				"title": title,
			}).Infoln("Close Pull Request which not created by sync-bot, ignore it.")
		} else {
			s.ClosePullRequest(e)
		}
	default:
		logrus.Infoln("Ignoring unhandled action:", e.Action)
	}
}
