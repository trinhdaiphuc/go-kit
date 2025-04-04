package collection

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInRange(t *testing.T) {
	type args[N Number] struct {
		number N
		from   N
		to     N
	}
	type testCase struct {
		name string
		args args[int]
		want bool
	}
	tests := []testCase{
		{
			name: "Number=3,From=0,To=10",
			args: args[int]{
				number: 3,
				from:   0,
				to:     10,
			},
			want: true,
		},
		{
			name: "Number=3,From=5,To=10",
			args: args[int]{
				number: 3,
				from:   5,
				to:     10,
			},
			want: false,
		},
		{
			name: "Number=3,From=3,To=3",
			args: args[int]{
				number: 3,
				from:   3,
				to:     3,
			},
			want: true,
		},
		{
			name: "Number=3,From=5,To=3",
			args: args[int]{
				number: 3,
				from:   3,
				to:     5,
			},
			want: true,
		},
		{
			name: "Number=3,From=5,To=3",
			args: args[int]{
				number: 5,
				from:   3,
				to:     5,
			},
			want: true,
		},
		{
			name: "Number=3,From=5,To=3",
			args: args[int]{
				number: 5,
				from:   5,
				to:     -1,
			},
			want: false,
		},
		{
			name: "Number=3,From=5,To=3",
			args: args[int]{
				number: 6,
				from:   5,
				to:     -1,
			},
			want: false,
		},
		{
			name: "Number=3,From=5,To=3",
			args: args[int]{
				number: -2,
				from:   5,
				to:     -1,
			},
			want: false,
		},
		{
			name: "Number=3,From=5,To=3",
			args: args[int]{
				number: 3,
				from:   5,
				to:     -1,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InRange(tt.args.number, tt.args.from, tt.args.to); got != tt.want {
				t.Errorf("InRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	type args[T comparable] struct {
		a []T
		x T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want bool
	}
	tests := []testCase[string]{
		{
			name: "Test contains x",
			args: args[string]{
				a: []string{"a", "b", "c"},
				x: "a",
			},
			want: true,
		},
		{
			name: "Test not contains x",
			args: args[string]{
				a: []string{"a", "b", "c"},
				x: "d",
			},
			want: false,
		},
		{
			name: "Test not array is empty",
			args: args[string]{
				a: []string{},
				x: "d",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.args.a, tt.args.x); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFind(t *testing.T) {
	type args[T comparable] struct {
		a []T
		x T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want int
	}
	tests := []testCase[string]{
		{
			name: "Find element with index",
			args: args[string]{
				a: []string{"a", "b", "c", "d"},
				x: "b",
			},
			want: 1,
		},
		{
			name: "Not find element with index",
			args: args[string]{
				a: []string{"a", "b", "c", "d"},
				x: "e",
			},
			want: -1,
		},
		{
			name: "Array input is empty",
			args: args[string]{
				a: []string{},
				x: "e",
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Find(tt.args.a, tt.args.x); got != tt.want {
				t.Errorf("Find() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToArrayString(t *testing.T) {
	type StructType struct {
		Name string
	}
	type args[T comparable] struct {
		a  []T
		fn func(ele StructType) string
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want []string
	}
	tests := []testCase[StructType]{
		{
			name: "Convert to array string successfully",
			args: args[StructType]{
				a: []StructType{
					{Name: "foo"},
					{Name: "bar"},
				},
				fn: func(ele StructType) string {
					return ele.Name
				},
			},
			want: []string{"foo", "bar"},
		},
		{
			name: "Input array is empty",
			args: args[StructType]{
				a: []StructType{},
				fn: func(ele StructType) string {
					return ele.Name
				},
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToArrayString(tt.args.a, tt.args.fn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToArrayString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToNumber(t *testing.T) {
	type StructNumber struct {
		Number int64
	}
	type args[T comparable, N Number] struct {
		a  []T
		fn func(ele T) (N, error)
	}
	type testCase[T comparable, N Number] struct {
		name    string
		args    args[T, N]
		want    []N
		wantErr bool
	}
	tests := []testCase[StructNumber, int64]{
		{
			name: "Convert number array success",
			args: args[StructNumber, int64]{
				a: []StructNumber{
					{Number: 1},
					{Number: 2},
					{Number: 3},
				},
				fn: func(ele StructNumber) (int64, error) {
					return ele.Number, nil
				},
			},
			want:    []int64{1, 2, 3},
			wantErr: false,
		},
		{
			name: "Input array empty",
			args: args[StructNumber, int64]{
				a: []StructNumber{},
				fn: func(ele StructNumber) (int64, error) {
					return ele.Number, nil
				},
			},
			want:    []int64{},
			wantErr: false,
		},
		{
			name: "Function parse number get error",
			args: args[StructNumber, int64]{
				a: []StructNumber{
					{Number: 1},
					{Number: 2},
					{Number: 3},
				},
				fn: func(ele StructNumber) (int64, error) {
					return ele.Number, errors.New("parse number failed")
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToArrayNumber(tt.args.a, tt.args.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToNumber() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeDuplicate(t *testing.T) {
	type args[T comparable] struct {
		a []T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want []T
	}
	testInt32 := []testCase[int32]{
		{
			name: "Array items are unique",
			args: args[int32]{
				a: []int32{1, 2, 3, 4, 5},
			},
			want: []int32{1, 2, 3, 4, 5},
		},
		{
			name: "All array items are duplicated",
			args: args[int32]{
				a: []int32{1, 1, 1, 1, 1},
			},
			want: []int32{1},
		},
		{
			name: "Empty array",
			args: args[int32]{
				a: []int32{},
			},
			want: []int32{},
		},
		{
			name: "Array items has duplicated",
			args: args[int32]{
				a: []int32{1, 1, 2, 1, 2, 5, 10, 10},
			},
			want: []int32{1, 2, 5, 10},
		},
	}
	for _, tt := range testInt32 {
		t.Run(tt.name, func(t *testing.T) {
			got := DeDuplicate(tt.args.a)
			sort.Slice(got, func(i, j int) bool { // Sort out put arrays for easy comparison
				return got[i] < got[j]
			})
			assert.Equalf(t, tt.want, got, "DeDuplicate() = %v, want %v", got, tt.want)
		})
	}
}

type StructType struct {
	Name string
	Age  int
}

func TestTransformArray(t *testing.T) {
	type args[A comparable, E comparable] struct {
		a  []A
		fn func(ele A) E
	}
	type testCase[A comparable, E comparable] struct {
		name string
		args args[A, E]
		want []E
	}
	tests := []testCase[StructType, string]{
		{
			name: "Transform to array name successfully",
			args: args[StructType, string]{
				a: []StructType{
					{Name: "John", Age: 1},
					{Name: "Alice", Age: 2},
				},
				fn: func(ele StructType) string {
					return ele.Name
				},
			},
			want: []string{"John", "Alice"},
		},
		{
			name: "Transform array empty",
			args: args[StructType, string]{
				a: []StructType{},
				fn: func(ele StructType) string {
					return ele.Name
				},
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, TransformArray(tt.args.a, tt.args.fn), "TransformArray(%v, %v)", tt.args.a, tt.args.fn)
		})
	}
}

func TestContainsFunc(t *testing.T) {
	tests := []struct {
		name string
		a    []int
		fn   func(ele int) bool
		want bool
	}{
		{
			name: "Element exists",
			a:    []int{1, 2, 3, 4, 5},
			fn: func(ele int) bool {
				return ele == 3
			},
			want: true,
		},
		{
			name: "Element does not exist",
			a:    []int{1, 2, 3, 4, 5},
			fn: func(ele int) bool {
				return ele == 6
			},
			want: false,
		},
		{
			name: "Empty slice",
			a:    []int{},
			fn: func(ele int) bool {
				return ele == 1
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsFunc(tt.a, tt.fn)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFindFunc(t *testing.T) {
	type args[T comparable] struct {
		a  []T
		fn func(ele T) bool
	}
	type testCase[T comparable] struct {
		name  string
		args  args[T]
		wantT T
	}
	tests := []testCase[*StructType]{
		{
			name: "Find element successfully",
			args: args[*StructType]{
				a: []*StructType{
					{Name: "John", Age: 1},
					{Name: "Alice", Age: 2},
					{Name: "Bob", Age: 3},
				},
				fn: func(ele *StructType) bool {
					return ele.Name == "Alice"
				},
			},
			wantT: &StructType{Name: "Alice", Age: 2},
		},
		{
			name: "Element does not exist",
			args: args[*StructType]{
				a: []*StructType{
					{Name: "John", Age: 1},
					{Name: "Alice", Age: 2},
					{Name: "Bob", Age: 3},
				},
				fn: func(ele *StructType) bool {
					return ele.Name == "Tom"
				},
			},
			wantT: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantT, FindFunc(tt.args.a, tt.args.fn), "FindFunc(%v, %v)", tt.args.a, tt.args.fn)
		})
	}
}

func TestDeDuplicateFunc(t *testing.T) {
	type args[T comparable] struct {
		a  []T
		fn func(ele T) string
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[StructType]{
		{
			name: "De-duplicate successfully",
			args: args[StructType]{
				a: []StructType{
					{Name: "John", Age: 1},
					{Name: "Alice", Age: 2},
					{Name: "John", Age: 3},
				},
				fn: func(ele StructType) string { return ele.Name },
			},
			want: []StructType{
				{Name: "John", Age: 1},
				{Name: "Alice", Age: 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, DeDuplicateFunc(tt.args.a, tt.args.fn), "DeDuplicateFunc(%v, %v)", tt.args.a, tt.args.fn)
		})
	}
}

func TestFilter(t *testing.T) {
	type args[T comparable] struct {
		a  []T
		fn func(ele T) bool
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[StructType]{
		{
			name: "Filter successfully",
			args: args[StructType]{
				a: []StructType{
					{Name: "John", Age: 1},
					{Name: "Alice", Age: 2},
					{Name: "Bob", Age: 3},
				},
				fn: func(ele StructType) bool {
					return ele.Age > 1
				},
			},
			want: []StructType{
				{Name: "Alice", Age: 2},
				{Name: "Bob", Age: 3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Filter(tt.args.a, tt.args.fn), "Filter(%v, %v)", tt.args.a, tt.args.fn)
		})
	}
}
