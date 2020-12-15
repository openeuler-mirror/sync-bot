package hook

import (
	"regexp"

	"github.com/sirupsen/logrus"

	"sync-bot/gitee"
	"sync-bot/gitee/event/action"
	"sync-bot/gitee/event/notable_type"
)

var (
	syncCommandRegex      = regexp.MustCompile(`(?m)^/sync\s+[^\s]+$`)
	syncCheckCommandRegex = regexp.MustCompile(`(?m)^/sync-check\s*$`)
)

func (s *Server) CommentPullRequest(e gitee.CommentPullRequestEvent) {
	org := e.Repository.Namespace
	repo := e.Repository.Path
	number := e.PullRequest.Number
	logrus.WithFields(logrus.Fields{
		"org":    org,
		"repo":   repo,
		"number": number,
	}).Infoln("CommentPullRequest")
}

func (s *Server) HandleNoteEvent(e gitee.CommentPullRequestEvent) {
	switch e.Action {
	case action.Comment:
		switch e.NotableType {
		case notable_type.PullRequest:
			s.CommentPullRequest(e)
		default:
			logrus.Println("Ignoring unhandled notable type:", e.NotableType)
		}
	default:
		logrus.Println("Ignoring unhandled action:", e.Action)
	}
}
