package gitee

// NotableType which object the note for
type NotableType string

// NotableType enum
const (
	NotableTypeComment     NotableType = "Comment" // Comment for Repository
	NotableTypeCommit      NotableType = "Commit"
	NotableTypeIssue       NotableType = "Issue"
	NotableTypePullRequest NotableType = "PullRequest"
)
