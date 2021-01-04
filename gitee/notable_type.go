package gitee

// NotableType which object the note for
type NotableType string

// NotableType enum
const (
	NotableTypeComment     = "Comment" // Comment for Repository
	NotableTypeCommit      = "Commit"
	NotableTypeIssue       = "Issue"
	NotableTypePullRequest = "PullRequest"
)
