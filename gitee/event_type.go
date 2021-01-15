package gitee

// EventType hook event type
type EventType string

// EventType enum
const (
	PushHook         EventType = "Push Hook"
	TagPushHook      EventType = "Tag Push Hook"
	IssueHook        EventType = "Issue Hook"
	MergeRequestHook EventType = "Merge Request Hook"
	NoteHook         EventType = "Note Hook"
)
