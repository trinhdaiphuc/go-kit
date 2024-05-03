package version

import "testing"

func TestCompareVersion(t *testing.T) {
	type args struct {
		version      string
		pivotVersion string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "2 empty version",
			args: args{
				version:      "",
				pivotVersion: "",
			},
			want: false,
		},
		{
			name: "pivot version is empty",
			args: args{
				version:      "1.1.1",
				pivotVersion: "",
			},
			want: false,
		},
		{
			name: "version is less than pivot version",
			args: args{
				version:      "1.1.1",
				pivotVersion: "1.1.2",
			},
			want: false,
		},
		{
			name: "version is equal pivot version",
			args: args{
				version:      "1.1.1",
				pivotVersion: "1.1.1",
			},
			want: true,
		},
		{
			name: "version is greater than pivot version",
			args: args{
				version:      "2.1.1",
				pivotVersion: "1.1.1",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompareVersion(tt.args.version, tt.args.pivotVersion); got != tt.want {
				t.Errorf("CompareVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
