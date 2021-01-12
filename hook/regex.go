package hook

import (
	"regexp"
)

var (
	// title start with [sync-bot]
	titleRegex = regexp.MustCompile(`^\[sync-bot\]`)
	// just /sync-check
	syncCheckRegex = regexp.MustCompile(`^\s*/sync-check\s*$`)
	// like "/sync new_branch branch-1.0 foo/bar"
	syncRegex = regexp.MustCompile(`^\s*/sync([ \t]+[\w\./_-]+)+\s*$`)
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
