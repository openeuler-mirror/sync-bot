package hook

import (
	"bytes"
	"text/template"
)

const (
	replySyncCheck = `
当前仓库存在以下 __保护分支__ ：
| Protected Branch | Version | Release |
|---|---|---|
{{- range .}}
|{{.Name}}|{{.Version}}|{{.Release}}|
{{- end}}

评论 ` + "`/sync <branch1> <branch2> ...`" + ` 可将当前 Pull Request 修改同步到其它分支（通过创建同步 Pull Request）：
a) 如果当前 PR 是 Open 状态，同步操作将延迟到 PR 被合并时执行；
b) 如果当前 PR 已经 Merged，将立即执行同步操作。

> 注意：/sync 命令可以指定同步多个分支，仅最后一个 /sync 命令生效
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
	replySyncCheckTmpl = template.Must(template.New("greeting").Parse(replySyncCheck))
	replySyncTmpl      = template.Must(template.New("replySync").Parse(replySync))
	syncPRBodyTmpl     = template.Must(template.New("syncPRBody").Parse(syncPRBody))
	syncResultTmpl     = template.Must(template.New("syncPRBody").Parse(syncResult))
	replyCloseTmpl     = template.Must(template.New("syncPRBody").Parse(replyClose))
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
