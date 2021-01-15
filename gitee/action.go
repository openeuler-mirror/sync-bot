package gitee

// Action action of pull request or issue
type Action string

// Action enum
const (
	ActionApproved    Action = "approved"
	ActionAssign      Action = "assign"
	ActionClose       Action = "close"
	ActionComment     Action = "comment"
	ActionDelete      Action = "delete"
	ActionDeleted     Action = "deleted"
	ActionEdited      Action = "edited"
	ActionMerge       Action = "merge"
	ActionOpen        Action = "open"
	ActionStateChange Action = "state_change"
	ActionTest        Action = "test"
	ActionTested      Action = "tested"
	ActionUnAssign    Action = "unassign"
	ActionUnTest      Action = "untest"
	ActionUpdate      Action = "update"
)

//// Action Desc
//const (
//	SourceBranchChanged = "source_branch_changed"
//	targetBranchChanged = "target_branch_changed"
//	UpdateLabel         = "update_label"
//)
