package hook

import (
	"testing"
)

func Test_matchTitle(t *testing.T) {
	type args struct {
		title string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"match",
			args{
				"[sync-bot] title with a specific prefix",
			},
			true,
		},
		{
			"match_with_whitespace",
			args{
				" [sync-bot] title with preceding whitespace character and a specific prefix",
			},
			true,
		},
		{
			"not_match",
			args{
				"normal title",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchTitle(tt.args.title); got != tt.want {
				t.Errorf("matchTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_matchSyc(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"match",
			args{
				"/sync branch1",
			},
			true,
		},
		{
			"not_match",
			args{
				"/sync",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchSync(tt.args.content); got != tt.want {
				t.Errorf("matchSync() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_matchSyncCheck(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"match",
			args{
				"/sync-check",
			},
			true,
		},
		{
			"not_match",
			args{
				"/sync-check-",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchSyncCheck(tt.args.content); got != tt.want {
				t.Errorf("matchSyncCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}
