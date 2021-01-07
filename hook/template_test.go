package hook

import (
	"testing"
	"text/template"
	"time"

	"sync-bot/gitee"
)

func Test_executeTemplate(t *testing.T) {
	type args struct {
		tmpl *template.Template
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "replySyncCheck",
			args: args{
				tmpl: replySyncCheckTmpl,
				data: []struct {
					Name    string
					Version string
					Release string
				}{
					{
						Name:    "__* branch1__",
						Version: "1.0",
						Release: "2",
					},
					{
						Name:    "branch2",
						Version: "1.0",
						Release: "2",
					},
				},
			},
			want: `
This repository has the following protected branches:
| Protected Branch | Version | Release |
|---|---|---|
|__* branch1__| | |
|branch2| | |

Use ` + "`/sync <branch> ...`" + ` command to register the branch that the current PR changes will synchronize to.
Once the current PR is merged, the synchronization operation will be performed.
(Only the last comment which include valid /sync command will be processed.)
`,
			wantErr: false,
		},
		{
			name: "replySync",
			args: args{
				tmpl: replySyncTmpl,
				data: struct {
					URL      string
					Command  string
					User     string
					Branches []branchStatus
				}{
					URL:     "https://example.com",
					Command: "/sync hello",
					User:    "me",
					Branches: []branchStatus{
						{
							Name:   "branch1",
							Status: branchExist,
						},
						{
							Name:   "branch2",
							Status: branchNonExist,
						},
					},
				},
			},
			want: `
In response to [this](https://example.com):
> /sync hello

@me
Receive the synchronization command.
Sync operation will be applied to the following branch(es), if the current PR is merged:

| Branch | Status |
|---|---|
|branch1|sync operation will be performed|
|branch2|branch not found, ignored|
`,
			wantErr: false,
		},
		{
			name: "syncPRBody",
			args: args{
				tmpl: syncPRBodyTmpl,
				data: struct {
					PR      string
					Issues  []gitee.Issue
					Commits []gitee.PullRequestCommit
				}{
					PR: "http://example.com",
					Issues: []gitee.Issue{
						{
							HTMLURL: "http://example.com/issue1",
						},
						{
							HTMLURL: "http://example.com/issue2",
						},
						{
							HTMLURL: "http://example.com/issue3",
						},
					},
					Commits: []gitee.PullRequestCommit{
						{
							Sha:     "1234567890",
							HTMLURL: "http://example.com/commit1",
							Commit: gitee.GitCommit{
								Author: gitee.GitUser{
									Date: time.Date(0, 1, 2, 3, 4, 5, 6, time.UTC),
								},
								Message: "commit1",
							},
						},
						{
							Sha:     "1234567890",
							HTMLURL: "http://example.com/commit1",
							Commit: gitee.GitCommit{
								Author: gitee.GitUser{
									Date: time.Date(0, 1, 2, 3, 4, 5, 6, time.UTC),
								},
								Message: "commit1",
							},
						},
					},
				},
			},

			want: `
### 1. Origin pull request:
http://example.com

### 2. Original pull request related issue(s):
http://example.com/issue1
http://example.com/issue2
http://example.com/issue3

### 3. Original pull request related commit(s):
| Sha | Datetime | Message |
|---|---|---|
|[12345678](http://example.com/commit1)|0000-01-02 03:04:05.000000006 +0000 UTC|commit1|
|[12345678](http://example.com/commit1)|0000-01-02 03:04:05.000000006 +0000 UTC|commit1|
`,
			wantErr: false,
		},
		{
			name: "syncResult",
			args: args{
				tmpl: syncResultTmpl,
				data: struct {
					URL        string
					Command    string
					User       string
					SyncStatus []syncStatus
				}{
					URL:     "https://example.com",
					Command: "/sync hello",
					User:    "me",
					SyncStatus: []syncStatus{
						{
							Name:   "branch1",
							Status: branchNonExist,
							PR:     "",
						},
						{
							Name:   "hello",
							Status: "Create pull request",
							PR:     "https://example.com/pr/1",
						},
					},
				},
			},
			want: `
In response to [this](https://example.com):
> /sync hello

@me

The following sync operations have been performed:

| Branch | Status | Pull Request |
|---|---|---|
|branch1|branch not found, ignored||
|hello|Create pull request|https://example.com/pr/1|
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := executeTemplate(tt.args.tmpl, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("executeTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("executeTemplate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
