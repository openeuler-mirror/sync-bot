package gitee

// EventType hook event type
type EventType string

// EventType enum
const (
	PushHook         = "Push Hook"
	TagPushHook      = "Tag Push Hook"
	IssueHook        = "Issue Hook"
	MergeRequestHook = "Merge Request Hook"
	NoteHook         = "Note Hook"
)
