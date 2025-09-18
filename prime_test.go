package tools

import (
	"reflect"
	"slices"
	"testing"
)

func TestKSetIn(t *testing.T) {
	set := NewKSet(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	tests := []struct {
		in  []int
		out []int
	}{
		{in: []int{}, out: nil},
		{in: []int{11, 200, 66}, out: nil},
		{in: []int{1, 3, 5, 7, 9, 11}, out: []int{1, 3, 5, 7, 9}},
		{in: []int{8, 6, 10, 20}, out: []int{8, 6, 10}},
	}

	for _, test := range tests {
		out := slices.Collect(set.In(slices.Values(test.in)))
		if !reflect.DeepEqual(out, test.out) {
			t.Errorf("got %v, want %v", out, test.out)
		}
	}
}
