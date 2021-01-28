package hook

import (
	"bytes"
	"text/template"
)

const (
	replySyncCheck = `
当前仓库包含以下 __保护分支__
| Protected Branch | Version | Release |
|---|---|---|
{{- range .}}
|{{.Name}}| | |
{{- end}}

评论 ` + "`/sync <branch1> <branch2> ...`" + ` 可以将当前 PR 的修改应用到其它分支(创建同步操作 PR)：
a) 如果当前 PR 是 Open 状态，同步操作将延迟到 PR 被合并时执行；
b) 如果当前 PR 已经 Merged，将立即执行同步操作。
(/sync 命令可以同时指定多个分支)
`

	replySync = `
In response to [this]({{.URL}}):
> {{.Command}}

@{{.User}}
一旦当前 PR 被合入，以下同步操作将会执行:

| Branch | Status |
|---|---|
{{- range .Branches}}
|{{print .Name}}|{{print .Status}}|
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

	syncResult = `
In response to [this]({{.URL}}):
> {{.Command}}

@{{.User}}

同步操作执行结果:

| Branch | Status | Pull Request |
|---|---|---|
{{- range .SyncStatus}}
|{{print .Name}}|{{print .Status}}|{{print .PR}}|
{{- end}}
`

	replyClose = `
In response to [this]({{.URL}}):
> {{.Command}}

@{{.User}}

{{.Status}}
`
)

var (
	replySyncCheckTmpl = template.Must(template.New("replySyncCheck").Parse(replySyncCheck))
	replySyncTmpl      = template.Must(template.New("replySync").Parse(replySync))
	syncPRBodyTmpl     = template.Must(template.New("syncPRBody").Parse(syncPRBody))
	syncResultTmpl     = template.Must(template.New("syncPRBody").Parse(syncResult))
	replyCloseTmpl     = template.Must(template.New("syncPRBody").Parse(replyClose))
)

const (
	branchExist    = "sync operation will be performed"
	branchNonExist = "branch not found, ignored"
)

type branchStatus struct {
	Name   string
	Status string
}

type syncStatus struct {
	Name   string
	Status string
	PR     string
}

func executeTemplate(tmpl *template.Template, data interface{}) (string, error) {
	var buffer bytes.Buffer
	err := tmpl.Execute(&buffer, data)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
