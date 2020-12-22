package gitee

import (
	"time"
)

type User struct {
	Email    string `json:"email"`
	HTMLURL  string `json:"html_url"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type Repository struct {
	DefaultBranch     string `json:"default_branch"`
	Fork              bool   `json:"fork"`
	GitHTTPURL        string `json:"git_http_url"`
	GitSSHURL         string `json:"git_ssh_url"`
	HTMLURL           string `json:"html_url"`
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
	Owner             User   `json:"owner"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"path_with_namespace"`
	Private           bool   `json:"private"`
}

type Repo struct {
	//Project Repository `json:"project"`
	Repository Repository `json:"repository"`
}

type Label struct {
	Color string `json:"color"`
	ID    int    `json:"id"`
	Name  string `json:"name"`
}

type Enterprise struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Branch contains general branch information.
type Branch struct {
	Name      string `json:"name"`
	Protected bool   `json:"protected"` // only included for ?protection=true requests
}

type Comment struct {
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	HTMLURL   string    `json:"html_url"`
	ID        int       `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
	User      User      `json:"user"`
}

// CommentPullRequestEvent is what Gitee sends us when a comment is create/edited/deleted.
type CommentPullRequestEvent struct {
	Action      string      `json:"action"`
	Comment     Comment     `json:"comment"`
	Enterprise  Enterprise  `json:"enterprise"`
	HookName    string      `json:"hook_name"`
	NotableType string      `json:"noteable_type"` // noteable_type NOT notable_type
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
	Timestamp   string      `json:"timestamp"`
}

// PullRequestChange contains information about what a PR changed.
type PullRequestChange struct {
	SHA              string `json:"sha"`
	Filename         string `json:"filename"`
	Status           string `json:"status"`
	Additions        int    `json:"additions"`
	Deletions        int    `json:"deletions"`
	Changes          int    `json:"changes"`
	Patch            string `json:"patch"`
	BlobURL          string `json:"blob_url"`
	PreviousFilename string `json:"previous_filename"`
}

type PullRequestBranch struct {
	Label string     `json:"label"`
	Ref   string     `json:"ref"`
	Repo  Repository `json:"repo"`
	Sha   string     `json:"sha"`
	User  User       `json:"user"`
}

type PullRequest struct {
	Additions          int               `json:"additions"`
	Base               PullRequestBranch `json:"base"`
	Body               string            `json:"body"`
	ChangedFiles       int               `json:"changed_files"`
	Comments           int               `json:"comments"`
	Commits            int               `json:"commits"`
	CreatedAt          time.Time         `json:"created_at"`
	DiffURL            string            `json:"diff_url"`
	Head               PullRequestBranch `json:"head"`
	HTMLURL            string            `json:"html_url"`
	ID                 int               `json:"id"`
	Labels             []Label           `json:"labels"`
	MergeReferenceName string            `json:"merge_reference_name"`
	MergeStatus        string            `json:"merge_status"`
	Mergeable          bool              `json:"mergeable"`
	Merged             bool              `json:"merged"`
	Number             int               `json:"number"`
	PatchURL           string            `json:"patch_url"`
	State              string            `json:"state"`
	Title              string            `json:"title"`
	UpdatedBy          User              `json:"updated_by"`
	User               User              `json:"user"`
}

// PullRequestEvent is what Gitee sends us when a PR is create/update/merge/close.
type PullRequestEvent struct {
	Action      string      `json:"action"`
	ActionDesc  string      `json:"action_desc"`
	Enterprise  Enterprise  `json:"enterprise"`
	HookName    string      `json:"hook_name"`
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
}

type Commit struct {
	Url     string `json:"url"`
	Sha     string `json:"sha"`
	HtmlUrl string `json:"html_url"`
	Message string `json:"message"`
}
