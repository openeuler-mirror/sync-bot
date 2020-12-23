package hook

import (
	"flag"
	"regexp"
	"strings"
)

type Strategy int

// sync strategy
const (
	Pick Strategy = iota
	Merge
	Overwrite
)

type SyncOption struct {
	strategy Strategy
	branches []string
}

func parse(cmd string) (SyncOption, error) {
	var opt SyncOption
	f := flag.NewFlagSet("/sync", flag.ContinueOnError)
	sep := regexp.MustCompile("[ \t]+")
	str := sep.Split(strings.TrimSpace(cmd), -1)
	err := f.Parse(str[1:])
	if err != nil {
		return opt, err
	}
	opt.strategy = Pick
	opt.branches = append(opt.branches, f.Args()...)
	return opt, nil
}
