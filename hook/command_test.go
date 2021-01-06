package hook

import (
	"reflect"
	"testing"
)

func Test_parse(t *testing.T) {
	type args struct {
		cmd string
	}
	tests := []struct {
		name    string
		args    args
		want    SyncCmdOption
		wantErr bool
	}{
		{
			name: "No branch",
			args: args{
				"/sync",
			},
			want: SyncCmdOption{
				strategy: Merge,
				branches: nil,
			},
			wantErr: false,
		},

		{
			name: "One branch",
			args: args{
				"/sync branch1",
			},
			want: SyncCmdOption{
				strategy: Merge,
				branches: []string{"branch1"},
			},
			wantErr: false,
		},
		{
			name: "Two branch",
			args: args{
				"/sync branch1 branch2",
			},
			want: SyncCmdOption{
				strategy: Merge,
				branches: []string{"branch1", "branch2"},
			},
			wantErr: false,
		},
		{
			name: "Failed",
			args: args{
				"/sync --force --ignore \tx.spec openEuler-20.03-LTS make_build openEuler-20.09",
			},
			want:    SyncCmdOption{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parse(tt.args.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}
