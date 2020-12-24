package hook

import (
	"github.com/sirupsen/logrus"

	"sync-bot/gitee"
)

func (s *Server) OpenPullRequest(e gitee.PullRequestEvent) {
	owner := e.PullRequest.Base.Repo.Owner.Name
	repo := e.PullRequest.Base.Repo.Name
	number := e.PullRequest.Number
	logrus.WithFields(logrus.Fields{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}).Infoln("OpenPullRequest")
}

func (s *Server) MergePullRequest(e gitee.PullRequestEvent) {
	owner := e.PullRequest.Base.Repo.Owner.Name
	repo := e.PullRequest.Base.Repo.Name
	number := e.PullRequest.Number
	logrus.WithFields(logrus.Fields{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}).Infoln("MergePullRequest")
}

func (s *Server) ClosePullRequest(e gitee.PullRequestEvent) {
	owner := e.PullRequest.Base.Repo.Owner.Name
	repo := e.PullRequest.Base.Repo.Name
	number := e.PullRequest.Number
	logrus.WithFields(logrus.Fields{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}).Infoln("ClosePullRequest")
}

func (s *Server) HandlePullRequestEvent(e gitee.PullRequestEvent) {
	switch e.Action {
	case gitee.ActionOpen:
		s.OpenPullRequest(e)
	case gitee.ActionMerge:
		s.MergePullRequest(e)
	case gitee.ActionClose:
		s.ClosePullRequest(e)
	default:
		logrus.Infoln("Ignoring unhandled action:", e.Action)
	}
}
