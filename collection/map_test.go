package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayToMap(t *testing.T) {
	type args[T comparable] struct {
		a  []T
		fn func(ele T) string
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want map[string]T
	}
	tests := []testCase[*StructType]{
		{
			name: "Test ArrayToMap",
			args: args[*StructType]{
				a: []*StructType{{Name: "a"}, {Name: "b"}, {Name: "c"}, {Name: "d"}},
				fn: func(ele *StructType) string {
					return ele.Name
				},
			},
			want: map[string]*StructType{
				"a": {Name: "a"},
				"b": {Name: "b"},
				"c": {Name: "c"},
				"d": {Name: "d"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ArrayToMap(tt.args.a, tt.args.fn), "ArrayToMap(%v, %v)", tt.args.a, tt.args.fn)
		})
	}
}
