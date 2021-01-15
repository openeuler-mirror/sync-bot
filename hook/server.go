package hook

import (
	"crypto/hmac"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/sirupsen/logrus"

	"sync-bot/git"
	"sync-bot/gitee"
)

type Server struct {
	// Client for git operation
	GitClient *git.Client
	// Client for access Gitee OpenAPI
	GiteeClient gitee.Client
	// function to get Gitee webhook secret
	Secret func() []byte
}

func (s *Server) demuxEvent(eventType string, payload []byte, h http.Header) error {
	switch eventType {
	case gitee.MergeRequestHook:
		var e gitee.PullRequestEvent
		if err := json.Unmarshal(payload, &e); err != nil {
			return err
		}
		go s.HandlePullRequestEvent(e)
	case gitee.NoteHook:
		var e gitee.CommentPullRequestEvent
		if err := json.Unmarshal(payload, &e); err != nil {
			return err
		}
		go s.HandleNoteEvent(e)
	default:
		logrus.Infoln("Ignoring unhandled event type:", eventType)
	}
	return nil
}

func (s *Server) hook(req *restful.Request, resp *restful.Response) {
	eventType, isPingEvent, payload, err := ValidateWebhook(req, resp)
	if err != nil {
		logrus.Errorln(err)
		return
	}

	_, err = resp.Write([]byte(eventType + ": event received."))
	if err != nil {
		logrus.Errorln("Response to webhook:", err)
		return
	}

	if isPingEvent {
		logrus.Infoln("Receive the Ping Event:", eventType)
		return
	}

	err = s.demuxEvent(eventType, payload, req.Request.Header)
	if err != nil {
		logrus.Errorln("demuxEvent:", err)
	}
}

func auth(secret func() []byte) func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		xGiteeToken := req.Request.Header.Get("X-Gitee-Token")
		if !hmac.Equal([]byte(xGiteeToken), secret()) {
			logrus.Errorln("Authorized failed from:", req.Request.RemoteAddr)
			resp.AddHeader("WWW-Authenticate", "Basic realm=Protected Area")
			_ = resp.WriteErrorString(401, "401: Not Authorized")
			return
		}
		chain.ProcessFilter(req, resp)
	}
}

func (s *Server) WebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	ws.Route(ws.POST("/hook").To(s.hook))
	ws.Filter(auth(s.Secret))
	return ws
}

func ValidateWebhook(req *restful.Request, resp *restful.Response) (string, bool, []byte, error) {
	defer req.Request.Body.Close()
	giteePing := req.Request.Header.Get("X-Gitee-Ping")
	isPingEvent := giteePing == "true"
	eventType := req.Request.Header.Get("X-Gitee-Event")
	if eventType == "" {
		_ = resp.WriteErrorString(http.StatusBadRequest, "400 Bad Request: Missing X-GitHub-Event Header")
		return "", isPingEvent, []byte{}, errors.New("400 Bad Request: Missing X-GitHub-Event Header")
	}
	//
	switch eventType {
	case gitee.PushHook:
	case gitee.TagPushHook:
	case gitee.IssueHook:
	case gitee.MergeRequestHook:
	case gitee.NoteHook:
	default:
		_, _ = resp.Write([]byte("invalid event type"))
		return "", isPingEvent, []byte{}, errors.New("invalid event type")
	}

	payload, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		return eventType, isPingEvent, nil, err
	}

	return eventType, isPingEvent, payload, nil
}
