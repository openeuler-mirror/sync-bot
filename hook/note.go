package hook

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"sync-bot/gitee"
	"sync-bot/util"
)

func (s *Server) replySyncCheck(owner string, repo string, number int, targetBranch string) {
	branches, err := s.GiteeClient.GetBranches(owner, repo, true)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Errorln("Get Branches failed")
		return
	}

	// convert branch to branch URL
	for i, branch := range branches {
		if branches[i].Name == targetBranch {
			// mark target branch of current pull request
			branches[i].Name = fmt.Sprintf("__*__ [%s](https://gitee.com/%s/%s/tree/%s)",
				branch.Name, owner, repo, branch.Name)
		} else {
			branches[i].Name = fmt.Sprintf("[%s](https://gitee.com/%s/%s/tree/%s)",
				branch.Name, owner, repo, branch.Name)
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

	logrus.WithFields(logrus.Fields{
		"owner":   owner,
		"repo":    repo,
		"number":  number,
		"comment": comment,
		"url":     url,
	}).Infoln("NotePullRequest")

	if util.MatchSyncCheck(comment) {
		logrus.Infoln("Receive /sync-check command")
		s.replySyncCheck(owner, repo, number, targetBranch)
		return
	}

	if util.MatchSync(comment) {
		logrus.Infoln("Receive /sync command")
		switch state {
		case gitee.StateOpen:
			logrus.Infoln("Pull request is open, just replay sync.")
			s.replySync(e)
		case gitee.StateMerged:
			logrus.Infoln("Pull request is merge, perform sync operation.")
			_ = s.sync(owner, repo, e.PullRequest, user, url, comment)
		default:
			logrus.WithFields(logrus.Fields{
				"comment": comment,
				"state":   state,
			}).Infoln("Ignoring unhandled pull request state.")
		}
		return
	}

	if util.MatchClose(comment) {
		s.ClosePullRequest(owner, repo, e.PullRequest)
		return
	}

	logrus.Infoln("Ignoring unhandled comment.")
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
	if owner == "openeuler" {
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
