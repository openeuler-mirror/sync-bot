package hook

import (
	"html/template"
	"testing"

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
			"ok",
			args{
				template.Must(template.New("test").Parse(`foo`)),
				nil,
			},
			"foo",
			false,
		},
		{
			"ok",
			args{
				syncTmpl,
				struct {
					URL      string
					Command  string
					User     string
					Branches []gitee.Branch
				}{
					URL:     "https://example",
					Command: "/sync hello",
					User:    "me",
					Branches: []gitee.Branch{
						{Name: "branch1"},
						{Name: "branch2"},
					},
				},
			},
			`
In response to [this](https://example):
> /sync hello

@me
Receive the synchronization command. The synchronization operation will be applied to the following branches, once the current PR is merged:
__branch1__
__branch2__
`,
			false,
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
