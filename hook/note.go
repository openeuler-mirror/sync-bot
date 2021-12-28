package hook

import (
	"fmt"
	"strings"

	"sync-bot/gitee"
	"sync-bot/util"
	"sync-bot/util/rpm"

	"github.com/sirupsen/logrus"
)

func (s *Server) greeting(owner string, repo string, number int, targetBranch string) {
	logger := logrus.WithFields(logrus.Fields{
		"owner":        owner,
		"repo":         repo,
		"number":       number,
		"targetBranch": targetBranch,
	})
	branches, err := s.GiteeClient.GetBranches(owner, repo, true)
	if err != nil {
		logger.Errorln("Get Branches failed:", err)
		return
	}
	for i, branch := range branches {
		// convert branch to branch URL
		if branches[i].Name == targetBranch {
			// mark target branch of current pull request
			branches[i].Name = fmt.Sprintf("__*__ [%s](https://gitee.com/%s/%s/tree/%s)",
				branch.Name, owner, repo, branch.Name)
		} else {
			branches[i].Name = fmt.Sprintf("[%s](https://gitee.com/%s/%s/tree/%s)",
				branch.Name, owner, repo, branch.Name)
		}
		// extract Version and Release from spec file
		spec, err1 := s.GiteeClient.GetTextFile(owner, repo, repo+".spec", branch.Name)
		if err1 != nil {
			logger.Errorln("Get spec file failed:", err)
			continue
		}
		s := rpm.NewSpec(spec)
		if s != nil {
			branches[i].Version = s.Version()
			branches[i].Release = s.Release()
		}
	}

	replyContent, err := executeTemplate(replySyncCheckTmpl, branches)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tmpl":     replySyncCheckTmpl,
			"branches": branches,
		}).Errorln("Execute template failed:", err)
		return
	}

	err = s.GiteeClient.CreateComment(owner, repo, number, replyContent)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"owner":        owner,
			"repo":         repo,
			"number":       number,
			"replyContent": replyContent,
		}).Errorln("Create comment failed:", err)
	} else {
		logrus.WithFields(logrus.Fields{
			"owner":        owner,
			"repo":         repo,
			"number":       number,
			"replyContent": replyContent,
		}).Infoln("Reply sync-check.")
	}
}

func (s *Server) replySync(e gitee.CommentPullRequestEvent) {
	owner := e.Repository.Namespace
	repo := e.Repository.Path
	number := e.PullRequest.Number
	comment := e.Comment.Body
	user := e.Comment.User.Username
	url := e.Comment.HTMLURL

	opt, err := parseSyncCommand(comment)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"opt": opt,
		}).Errorln("Parse /sync command failed:", err)
		comment := fmt.Sprintf("Receive comment look like /sync command, but parseSyncCommand failed: %v", err)
		logrus.Errorln(comment)
		err = s.GiteeClient.CreateComment(owner, repo, number, comment)
		if err != nil {
			logrus.Errorln("Create Comment failed:", err)
		}
		return
	}

	// retrieve all branches
	allBranches, err := s.GiteeClient.GetBranches(owner, repo, false)
	if err != nil {
		comment := fmt.Sprintf("List branches failed: %v", err)
		logrus.Errorln(comment)
		err = s.GiteeClient.CreateComment(owner, repo, number, comment)
		if err != nil {
			logrus.Errorln("Create Comment failed:", err)
		}
		return
	}
	branchSet := make(map[string]bool)
	for _, b := range allBranches {
		branchSet[b.Name] = true
	}

	var synBranches []branchStatus
	for _, b := range opt.branches {
		if ok := branchSet[b]; ok {
			synBranches = append(synBranches, branchStatus{
				Name:   b,
				Status: branchExist,
			})
		} else {
			synBranches = append(synBranches, branchStatus{
				Name:   b,
				Status: branchNonExist,
			})
		}
	}

	data := struct {
		URL      string
		Command  string
		User     string
		Branches []branchStatus
	}{
		URL:      url,
		Command:  strings.TrimSpace(comment),
		User:     user,
		Branches: synBranches,
	}

	replyComment, err := executeTemplate(replySyncTmpl, data)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tmpl": replySyncTmpl,
			"data": data,
		}).Errorln("Execute template failed:", err)
		return
	}
	err = s.GiteeClient.CreateComment(owner, repo, number, replyComment)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"owner":        owner,
			"repo":         repo,
			"number":       number,
			"replyComment": replyComment,
		}).Errorln("Create comment failed:", err)
	} else {
		logrus.WithFields(logrus.Fields{
			"owner":        owner,
			"repo":         repo,
			"number":       number,
			"replyComment": replyComment,
		}).Infoln("Reply sync.")
	}
}

func (s *Server) NotePullRequest(e gitee.CommentPullRequestEvent) {
	owner := e.Repository.Namespace
	repo := e.Repository.Path
	number := e.PullRequest.Number
	comment := e.Comment.Body
	user := e.Comment.User.Username
	url := e.Comment.HTMLURL
	targetBranch := e.PullRequest.Base.Ref
	state := e.PullRequest.State
	title := e.PullRequest.Title

	logger := logrus.WithFields(logrus.Fields{
		"owner":        owner,
		"repo":         repo,
		"number":       number,
		"comment":      comment,
		"uer":          user,
		"url":          url,
		"targetBranch": targetBranch,
		"state":        state,
		"title":        title,
	})
	logger.Infoln("NotePullRequest")

	if util.MatchSyncCheck(comment) {
		logger.Infoln("Receive /sync-check command")
		s.greeting(owner, repo, number, targetBranch)
		return
	}

	if util.MatchSync(comment) {
		logger.Infoln("Receive /sync command")
		switch state {
		case gitee.StateOpen:
			logger.Infoln("Pull request is open, just replay sync.")
			s.replySync(e)
		case gitee.StateMerged:
			logger.Infoln("Pull request is merge, perform sync operation.")
			_ = s.sync(owner, repo, e.PullRequest, user, url, comment)
		default:
			logger.Infoln("Ignoring unhandled pull request state.")
		}
		return
	}

	if util.MatchClose(comment) {
		logger.Infoln("Receive /close command")
		if util.MatchTitle(title) {
			s.ClosePullRequest(owner, repo, e.PullRequest)
		} else {
			logger.Infoln("Pull request not created by sync-bot, ignoring /close.")
		}
		return
	}

	logger.Infoln("Ignoring unhandled comment.")
}

func (s *Server) HandleNoteEvent(e gitee.CommentPullRequestEvent) {
	owner := e.Repository.Namespace
	repo := e.Repository.Path

	logger := logrus.WithFields(logrus.Fields{
		"owner": owner,
		"repo":  repo,
	})

	// TODO: need to be configurable
	// ignore repo in openeuler
	if owner == "openeuler" && repo != "docs" {
		logger.Infoln("Ignore repo in openeuler")
		return
	}

	switch e.Action {
	case gitee.ActionComment:
		switch e.NotableType {
		case gitee.NotableTypePullRequest:
			s.NotePullRequest(e)
		default:
			logger.Infoln("Ignoring unhandled notable type:", e.NotableType)
		}
	default:
		logger.Infoln("Ignoring unhandled action:", e.Action)
	}
}
