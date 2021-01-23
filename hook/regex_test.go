package hook

import (
	"testing"
)

func TestMatchTitle(t *testing.T) {
	type args struct {
		title string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"exact match",
			args{
				"[sync-bot] title with a specific prefix",
			},
			true,
		},
		{
			"not match",
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

func TestMatchSync(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"one branch",
			args{
				"/sync branch1",
			},
			true,
		},
		{
			"two branch",
			args{
				"/sync branch1 branch1",
			},
			true,
		},
		{
			"special character branch name",
			args{
				"/sync foo.bar foo_bar foo-bar foo/bar",
			},
			true,
		},
		{
			"no branch",
			args{
				"/sync",
			},
			false,
		},
		{
			"middle newline",
			args{
				"/sync a\n/sync b",
			},
			false,
		},
		{
			"multi-line",
			args{
				"\n\t /sync a b\n\t ",
			},
			true,
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

func TestMatchSyncCheck(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"exact match",
			args{
				"/sync-check",
			},
			true,
		},
		{
			"not match",
			args{
				"/sync-check-",
			},
			false,
		},
		{
			"multi-line",
			args{
				"/sync-check\n",
			},
			true,
		},
		{
			"include whitespace",
			args{
				" \t/sync-check \n ",
			},
			true,
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

func TestMatchSyncBranch(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "sync-pr103-master-to-openEuler-20.03-LTS-Next",
			args: args{
				content: "sync-pr103-master-to-openEuler-20.03-LTS-Next",
			},
			want: true,
		},
		{
			name: "sync-pr103-master-to-openEuler-20.03-LTS-SP1",
			args: args{
				content: "sync-pr103-master-to-openEuler-20.03-LTS-SP1",
			},
			want: true,
		},
		{
			name: "openEuler-20.03-LTS-SP1",
			args: args{
				content: "openEuler-20.03-LTS-SP1",
			},
			want: false,
		},
		{
			name: "sync-pr1-to-",
			args: args{
				content: "sync-pr1-to-",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchSyncBranch(tt.args.content); got != tt.want {
				t.Errorf("matchSyncBranch() = %v, want %v", got, tt.want)
			}
		})
	}
}
