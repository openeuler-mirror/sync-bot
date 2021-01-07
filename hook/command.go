package hook

import (
	"flag"
	"regexp"
	"strings"
)

// Strategy strategy of sync
type Strategy int

// three types strategy
const (
	Pick Strategy = iota
	Merge
	Overwrite
)

// SyncCmdOption /sync command option
type SyncCmdOption struct {
	strategy Strategy
	branches []string
}

func commandParse(command string) (SyncCmdOption, error) {
	var opt SyncCmdOption
	f := flag.NewFlagSet("/sync", flag.ContinueOnError)
	sep := regexp.MustCompile("[ \t]+")
	str := sep.Split(strings.TrimSpace(command), -1)
	err := f.Parse(str[1:])
	if err != nil {
		return opt, err
	}
	// Todo: default is Merge now, will change to Pick
	opt.strategy = Merge
	opt.branches = append(opt.branches, f.Args()...)
	return opt, nil
}
