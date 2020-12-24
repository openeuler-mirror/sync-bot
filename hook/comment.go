package hook

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/sirupsen/logrus"

	"sync-bot/gitee"
)

const (
	syncCheckContent = `
This repository has the following protected branches:
| Protected Branch | Version | Release |
|---|---|---|
{{range . -}}
|{{print .Name -}}|
{{end}}
Use ` + "`/sync <branch>`" + ` command to register the branch that the current PR changes will synchronize to.
Once the current PR is merged, the synchronization operation will be performed.
(Only the last comment which include valid /sync command will be processed.)
`
	syncContent = `
In response to [this]({{.URL}}):
> {{.Command}}

@{{.User}}
Receive the synchronization command. The synchronization operation will be applied to the following branches, once the current PR is merged:
{{range .Branches -}}
__{{print .Name}}__
{{end -}}
`
)

var (
	syncCheckTmpl = template.Must(template.New("comment").Parse(syncCheckContent))
	syncTmpl      = template.Must(template.New("comment").Parse(syncContent))
)

func executeTemplate(tmpl *template.Template, data interface{}) (string, error) {
	var buffer bytes.Buffer
	err := tmpl.Execute(&buffer, data)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (s *Server) replySyncCheck(owner string, repo string, number int) {
	protectedBranches, err := s.GiteeClient.GetBranches(owner, repo, true)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Errorln("GetBranches failed")
		return
	}

	// convert branch to branch URL
	for i, branch := range protectedBranches {
		protectedBranches[i].Name = fmt.Sprintf("[%s](https://gitee.com/%s/%s/tree/%s)",
			branch.Name, owner, repo, branch.Name)
	}

	branchListComment, err := executeTemplate(syncCheckTmpl, protectedBranches)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tmpl":              syncCheckTmpl,
			"protectedBranches": protectedBranches,
		}).Errorln("Execute template failed:", err)
		return
	}

	err = s.GiteeClient.CreateComment(owner, repo, number, branchListComment)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"org":               owner,
			"repo":              repo,
			"number":            number,
			"branchListComment": branchListComment,
		}).Errorln("Create comment failed:", err)
		return
	}
}

func (s *Server) replySync(owner string, repo string, number int, url string, user string, command string) {
	opt, err := parse(command)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"command": command,
			"opt":     opt,
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
		Command:  command,
		User:     user,
		Branches: branches,
	}
	comment, err := executeTemplate(syncTmpl, data)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tmpl": syncTmpl,
			"data": data,
		}).Errorln("Execute template failed:", err)
		return
	}
	err = s.GiteeClient.CreateComment(owner, repo, number, comment)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"org":               owner,
			"repo":              repo,
			"number":            number,
			"branchListComment": comment,
		}).Errorln("Create comment failed:", err)
	}
}

func (s *Server) CommentPullRequest(e gitee.CommentPullRequestEvent) {
	owner := e.Repository.Namespace
	repo := e.Repository.Path
	number := e.PullRequest.Number
	comment := e.Comment.Body
	user := e.Comment.User.Username
	logrus.WithFields(logrus.Fields{
		"org":     owner,
		"repo":    repo,
		"number":  number,
		"comment": comment,
	}).Infoln("CommentPullRequest")

	if matchSyncCheck(comment) {
		logrus.Println("Receive /sync-check command")
		s.replySyncCheck(owner, repo, number)
		return
	}

	if matchSync(comment) {
		logrus.Println("Receive /sync command")
		s.replySync(owner, repo, number, e.Comment.HTMLURL, user, comment)
		return
	}
	logrus.Println("Ignoring unhandled comment.")
}

func (s *Server) HandleNoteEvent(e gitee.CommentPullRequestEvent) {
	switch e.Action {
	case gitee.ActionComment:
		switch e.NotableType {
		case gitee.NotableTypePullRequest:
			s.CommentPullRequest(e)
		default:
			logrus.Println("Ignoring unhandled notable type:", e.NotableType)
		}
	default:
		logrus.WithFields(logrus.Fields{
			"comment": e.Comment.Body,
		}).Println("Ignoring unhandled action:", e.Action)
	}
}
