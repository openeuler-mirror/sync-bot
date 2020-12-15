package hook

import (
	"github.com/sirupsen/logrus"

	"sync-bot/gitee"
	"sync-bot/gitee/event/action"
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
	case action.Open:
		s.OpenPullRequest(e)
	case action.Merge:
		s.MergePullRequest(e)
	case action.Close:
		s.ClosePullRequest(e)
	default:
		logrus.Infoln("Ignoring unhandled action:", e.Action)
	}
}
