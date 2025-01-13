package vcr

import "testing"

func Test_stripRandomNumbers(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "packer snapshot",
			args: args{"snapshot-packer-1736526252"},
			want: "snapshot-packer",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripRandomNumbers(tt.args.s); got != tt.want {
				t.Errorf("stripRandomNumbers() = %v, want %v", got, tt.want)
			}
		})
	}
}
