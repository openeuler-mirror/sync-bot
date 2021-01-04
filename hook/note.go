package hook

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"sync-bot/gitee"
)

func (s *Server) replySyncCheck(owner string, repo string, number int, targetBranch string) {
	branches, err := s.GiteeClient.GetBranches(owner, repo, true)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Errorln("GetBranches failed")
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

	opt, err := parse(comment)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"opt": opt,
		}).Errorln("Parse /sync command failed:", err)
		comment := fmt.Sprintf("Receive comment look like /sync command, but parse failed: %v", err)
		err = s.GiteeClient.CreateComment(owner, repo, number, comment)
		if err != nil {
			logrus.Errorln("Create NotableTypeComment failed:", err)
		}
		return
	}

	// TODO: need to check if branch exist in repository
	var branches []gitee.Branch
	for _, b := range opt.branches {
		branches = append(branches, gitee.Branch{
			Name: b,
		})
	}

	data := struct {
		URL      string
		Command  string
		User     string
		Branches []gitee.Branch
	}{
		URL:      url,
		Command:  comment,
		User:     user,
		Branches: branches,
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
	targetBranch := e.PullRequest.Base.Ref
	state := e.PullRequest.State
	logrus.WithFields(logrus.Fields{
		"owner":   owner,
		"repo":    repo,
		"number":  number,
		"comment": comment,
	}).Infoln("NotePullRequest")

	if matchSyncCheck(comment) {
		logrus.Infoln("Receive /sync-check command")
		s.replySyncCheck(owner, repo, number, targetBranch)
		return
	}

	if matchSync(comment) {
		logrus.Infoln("Receive /sync command")
		if state == gitee.StateOpen {
			s.replySync(e)
		} else {
			s.sync(owner, repo, e.PullRequest, comment)
		}
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
