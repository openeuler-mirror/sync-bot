package hook

import (
	"bytes"
	"text/template"
)

const (
	replySyncCheck = `
This repository has the following protected branches:
| Protected Branch | Version | Release |
|---|---|---|
{{- range .}}
|{{.Name}}| | |
{{- end}}

Use ` + "`/sync <branch>`" + ` command to register the branch that the current PR changes will synchronize to.
Once the current PR is merged, the synchronization operation will be performed.
(Only the last comment which include valid /sync command will be processed.)
`

	replySync = `
In response to [this]({{.URL}}):
> {{.Command}}

@{{.User}}
Receive the synchronization command. The synchronization operation will be applied to the following branches, once the current PR is merged:
{{- range .Branches}}
__{{print .Name}}__
{{- end}}
`

	syncPRBody = `
### 1. Origin pull request:
{{.PR}}

### 2. Original pull request related issue(s):
{{- range .Issues}}
{{.HTMLURL}}
{{- end}}

### 3. Original pull request related commit(s):
| Sha | Datetime | Message |
|---|---|---|
{{- range .Commits}}
|[{{slice .Sha 0 8}}]({{.HTMLURL}})|{{.Commit.Author.Date}}|{{.Commit.Message}}|
{{- end}}
`
)

var (
	replySyncCheckTmpl = template.Must(template.New("replySyncCheck").Parse(replySyncCheck))
	replySyncTmpl      = template.Must(template.New("replySync").Parse(replySync))
	syncPRBodyTmpl     = template.Must(template.New("syncPRBody").Parse(syncPRBody))
)

func executeTemplate(tmpl *template.Template, data interface{}) (string, error) {
	var buffer bytes.Buffer
	err := tmpl.Execute(&buffer, data)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
