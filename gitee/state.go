package gitee

// State state of pull request or issue
type State string

// State enum
const (
	StateOpen        State = "open"
	StateMerged      State = "merged"
	StateClosed      State = "closed"
	StateProgressing State = "progressing"
	StateRejected    State = "rejected"
)
