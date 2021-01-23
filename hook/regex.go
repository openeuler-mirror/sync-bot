package hook

import (
	"regexp"
)

var (
	// title start with [sync-bot]
	titleRegex = regexp.MustCompile(`^(\[sync-bot\]|\[sync\])`)
	// just /sync-check
	syncCheckRegex = regexp.MustCompile(`^\s*/sync-check\s*$`)
	// like "/sync new_branch branch-1.0 foo/bar"
	syncRegex = regexp.MustCompile(`^\s*/sync([ \t]+[\w\./_-]+)+\s*$`)
	// /close
	closeRegex = regexp.MustCompile(`^\s*/close\s*$`)
	// sync branch name like "sync-pr103-master-to-openEuler-20.03-LTS"
	syncBranchRegex = regexp.MustCompile(`^sync-pr[\d]+-.+-to-.+$`)
)

// match Pull Request created by sync-bot
func matchTitle(title string) bool {
	return titleRegex.MatchString(title)
}

// match Sync command
func matchSync(content string) bool {
	return syncRegex.MatchString(content)
}

// match SyncCheck command
func matchSyncCheck(content string) bool {
	return syncCheckRegex.MatchString(content)
}

// match close command
func matchClose(content string) bool {
	return closeRegex.MatchString(content)
}

// match branch name
func matchSyncBranch(content string) bool {
	return syncBranchRegex.MatchString(content)
}
