package hook

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"sync-bot/gitee"
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

func (s *Server) doClosePullRequest(e gitee.CommentPullRequestEvent) error {
	owner := e.Repository.Namespace
	repo := e.Repository.Path
	title := e.PullRequest.Title
	number := e.PullRequest.Number
	state := e.PullRequest.State
	url := e.Comment.HTMLURL
	user := e.Comment.User.Username
	comment := e.Comment.Body

	logger := logrus.WithFields(logrus.Fields{
		"owner":   owner,
		"repo":    repo,
		"title":   title,
		"number":  number,
		"state":   state,
		"url":     url,
		"user":    user,
		"comment": comment,
	})

	if !matchTitle(title) {
		logger.Infoln("Pull request not create by sync-bot, ignoring it.")
		return nil
	}
	if state != gitee.StateOpen {
		logger.Infoln("Pull request state is not open, ignoring it.")
		return nil
	}

	r, err := s.GitClient.Clone(owner, repo)
	if err != nil {
		logger.Errorf("Clone repo failed: %v", err)
		return err
	}
	sourceBranch := e.PullRequest.Head.Ref
	logger.Infoln("Close request pull by delete source branch.")

	var status string
	err = r.DeleteRemoteBranch(sourceBranch)
	if err != nil {
		status = fmt.Sprintf("Close the current pull request failed: %v", err)
	} else {
		status = "Close the current pull request by removing the source branch."
	}

	reply, _ := executeTemplate(replyCloseTmpl, struct {
		URL     string
		User    string
		Command string
		Status  string
	}{
		URL:     url,
		User:    user,
		Command: strings.TrimSpace(comment),
		Status:  status,
	})

	return s.GiteeClient.CreateComment(owner, repo, number, reply)

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

	if matchSyncCheck(comment) {
		logrus.Infoln("Receive /sync-check command")
		s.replySyncCheck(owner, repo, number, targetBranch)
		return
	}

	if matchSync(comment) {
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

	if matchClose(comment) {
		_ = s.doClosePullRequest(e)
		return
	}

	logrus.Infoln("Ignoring unhandled comment.")
}

func (s *Server) HandleNoteEvent(e gitee.CommentPullRequestEvent) {
	switch e.Action {
	case gitee.ActionComment:
		switch e.NotableType {
		case gitee.NotableTypePullRequest:
			s.NotePullRequest(e)
		default:
			logrus.Infoln("Ignoring unhandled notable type:", e.NotableType)
		}
	default:
		logrus.Infoln("Ignoring unhandled action:", e.Action)
	}
}
