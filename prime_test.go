package tools

import (
	"reflect"
	"slices"
	"strings"
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

func normalizeNonEmpty(input string) (string, bool) {
	v := strings.ToLower(strings.TrimSpace(input))
	if v == "" {
		return "", false
	}
	return v, true
}

func TestSSDedupAndDedupIn_DefaultTransfer(t *testing.T) {
	var nilSS SS
	if got := nilSS.Dedup(); got != nil {
		t.Fatalf("Dedup(nil) got %v, want nil", got)
	}
	if got := nilSS.DedupIn(); got != nil {
		t.Fatalf("DedupIn(nil) got %v, want nil", got)
	}

	empty := SS{}
	if got := empty.Dedup(); got == nil || len(got) != 0 {
		t.Fatalf("Dedup(empty) got %v, want empty non-nil slice", got)
	}
	if got := empty.DedupIn(); got == nil || len(got) != 0 {
		t.Fatalf("DedupIn(empty) got %v, want empty non-nil slice", got)
	}

	input := SS{"a", " ", "", " b", "\t"}
	expected := SS{"a", "b"}

	if got := input.Dedup(); !slices.Equal(got, expected) {
		t.Fatalf("Dedup(default) got %v, want %v", got, expected)
	}
	inPlace := input.Clone()
	if got := inPlace.DedupIn(); !slices.Equal(got, expected) {
		t.Fatalf("DedupIn(default) got %v, want %v", got, expected)
	}
}

func TestSSDedupAndDedupIn_WithCustomTransfer(t *testing.T) {
	origin := SS{" A ", "a", "", "B ", " b", "A"}
	expected := SS{"a", "b"}

	gotDedup := origin.Dedup(normalizeNonEmpty)
	if !slices.Equal(gotDedup, expected) {
		t.Fatalf("Dedup got %v, want %v", gotDedup, expected)
	}
	if !slices.Equal(origin, SS{" A ", "a", "", "B ", " b", "A"}) {
		t.Fatalf("Dedup should not mutate input, got %v", origin)
	}
	if len(gotDedup) > 0 && &gotDedup[0] == &origin[0] {
		t.Fatalf("Dedup should return a distinct result slice")
	}

	inPlace := origin.Clone()
	gotDedupIn := inPlace.DedupIn(normalizeNonEmpty)
	if !slices.Equal(gotDedupIn, expected) {
		t.Fatalf("DedupIn got %v, want %v", gotDedupIn, expected)
	}
	if len(gotDedupIn) > 0 && &gotDedupIn[0] != &inPlace[0] {
		t.Fatalf("DedupIn should reuse the original slice backing array")
	}
	if !slices.Equal(inPlace[:len(expected)], expected) {
		t.Fatalf("DedupIn should write results in place, got prefix %v", inPlace[:len(expected)])
	}
}
