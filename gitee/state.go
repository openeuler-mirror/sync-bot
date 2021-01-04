package gitee

// State state of pull request or issue
type State string

// State enum
const (
	StateOpen        = "open"
	StateMerged      = "merged"
	StateClosed      = "closed"
	StateProgressing = "progressing"
	StateRejected    = "rejected"
)
