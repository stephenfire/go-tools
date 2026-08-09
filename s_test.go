package tools

import (
	"encoding/json"
	"math"
	"reflect"
	"slices"
	"testing"
)

func TestS_IsValid(t *testing.T) {
	tests := []struct {
		input S
		want  bool
	}{
		{S(""), false},
		{S("hello"), true},
		{S(" "), true},
		{S("0"), true},
	}
	for _, tt := range tests {
		got := tt.input.IsValid()
		if got != tt.want {
			t.Errorf("S(%q).IsValid() = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestS_Trim(t *testing.T) {
	tests := []struct {
		input S
		want  S
	}{
		{S("  hello  "), S("hello")},
		{S("\t\n world \n"), S("world")},
		{S("no-space"), S("no-space")},
		{S(""), S("")},
		{S("   "), S("")},
	}
	for _, tt := range tests {
		got := tt.input.Trim()
		if got != tt.want {
			t.Errorf("S(%q).Trim() = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestS_ToLower(t *testing.T) {
	tests := []struct {
		input S
		want  S
	}{
		{S("HELLO"), S("hello")},
		{S("Hello World"), S("hello world")},
		{S("already lower"), S("already lower")},
		{S(""), S("")},
		{S("MIXED Case"), S("mixed case")},
	}
	for _, tt := range tests {
		got := tt.input.ToLower()
		if got != tt.want {
			t.Errorf("S(%q).ToLower() = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestS_ToUpper(t *testing.T) {
	tests := []struct {
		input S
		want  S
	}{
		{S("hello"), S("HELLO")},
		{S("Hello World"), S("HELLO WORLD")},
		{S("ALREADY UPPER"), S("ALREADY UPPER")},
		{S(""), S("")},
		{S("Mixed CASE"), S("MIXED CASE")},
	}
	for _, tt := range tests {
		got := tt.input.ToUpper()
		if got != tt.want {
			t.Errorf("S(%q).ToUpper() = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestS_Replace(t *testing.T) {
	tests := []struct {
		input        S
		source, targ string
		want         S
	}{
		{S("hello world"), "world", "golang", S("hello golang")},
		{S("aaa"), "a", "b", S("bbb")},
		{S("no match"), "x", "y", S("no match")},
		{S(""), "a", "b", S("")},
		{S("foo bar foo"), "foo", "baz", S("baz bar baz")},
	}
	for _, tt := range tests {
		got := tt.input.Replace(tt.source, tt.targ)
		if got != tt.want {
			t.Errorf("S(%q).Replace(%q, %q) = %q, want %q", tt.input, tt.source, tt.targ, got, tt.want)
		}
	}
}

func TestS_Formalize(t *testing.T) {
	tests := []struct {
		input S
		want  S
	}{
		{S("  Hello World  "), S("hello world")},
		{S("UPPER"), S("upper")},
		{S("\t\n MIXED \n"), S("mixed")},
		{S(""), S("")},
		{S("   "), S("")},
	}
	for _, tt := range tests {
		got := tt.input.Formalize()
		if got != tt.want {
			t.Errorf("S(%q).Formalize() = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestS_CamelToSnake(t *testing.T) {
	tests := []struct {
		input S
		want  S
	}{
		{S("CamelCase"), S("camel_case")},
		{S("HTMLElement"), S("h_t_m_l_element")},
		{S("already_snake"), S("already_snake")},
		{S(""), S("")},
		{S("lowercase"), S("lowercase")},
		{S("Simple"), S("simple")},
		{S("JSONParser"), S("j_s_o_n_parser")},
		{S("MyVarName"), S("my_var_name")},
	}
	for _, tt := range tests {
		got := tt.input.CamelToSnake()
		if got != tt.want {
			t.Errorf("S(%q).CamelToSnake() = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestS_Includes(t *testing.T) {
	tests := []struct {
		input S
		args  []string
		want  bool
	}{
		{S("hello world"), []string{"hello"}, true},
		{S("hello world"), []string{"hello", "world"}, true},
		{S("hello world"), []string{"hello", "missing"}, false},
		{S(""), []string{"a"}, false},
		{S("hello world"), []string{}, true},
		{S("hello world"), []string{""}, true}, // empty string skipped
		{S("foo"), []string{"f", "o"}, true},
		{S("foo"), []string{"f", "x"}, false},
	}
	for _, tt := range tests {
		got := tt.input.Includes(tt.args...)
		if got != tt.want {
			t.Errorf("S(%q).Includes(%v) = %v, want %v", tt.input, tt.args, got, tt.want)
		}
	}
}

func TestS_SplitBy(t *testing.T) {
	tests := []struct {
		input S
		sep   string
		want  []string
	}{
		{S("a,b,c"), ",", []string{"a", "b", "c"}},
		{S("hello world"), " ", []string{"hello", "world"}},
		{S("no-sep"), ",", []string{"no-sep"}},
		{S(""), ",", []string{""}},
		{S("a::b::c"), "::", []string{"a", "b", "c"}},
	}
	for _, tt := range tests {
		got := tt.input.SplitBy(tt.sep)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("S(%q).SplitBy(%q) = %v, want %v", tt.input, tt.sep, got, tt.want)
		}
	}
}

func TestS_RegexpSplit(t *testing.T) {
	tests := []struct {
		input    S
		expr     string
		want     []string
		wantErr  bool
	}{
		{S("a,b,c"), `,`, []string{"a", "b", "c"}, false},
		{S("a, b, c"), `,\s*`, []string{"a", "b", "c"}, false},
		{S("hello  world"), `\s+`, []string{"hello", "world"}, false},
		{S(""), `,`, []string{""}, false},
		{S("abc"), `[invalid`, nil, true},
	}
	for _, tt := range tests {
		got, err := tt.input.RegexpSplit(tt.expr)
		if (err != nil) != tt.wantErr {
			t.Errorf("S(%q).RegexpSplit(%q) error = %v, wantErr %v", tt.input, tt.expr, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
			t.Errorf("S(%q).RegexpSplit(%q) = %v, want %v", tt.input, tt.expr, got, tt.want)
		}
	}
}

func TestS_CSVSplit(t *testing.T) {
	tests := []struct {
		input S
		want  []string
	}{
		{S("a,b,c"), []string{"a", "b", "c"}},
		{S("a，b、c"), []string{"a", "b", "c"}}, // Chinese comma and enumeration comma
		{S("a,b，c、d"), []string{"a", "b", "c", "d"}},
		{S("single"), []string{"single"}},
		{S(""), []string{""}},
	}
	for _, tt := range tests {
		got := tt.input.CSVSplit()
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("S(%q).CSVSplit() = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestS_String(t *testing.T) {
	tests := []struct {
		input S
		want  string
	}{
		{S("hello"), "hello"},
		{S(""), ""},
		{S("world"), "world"},
	}
	for _, tt := range tests {
		got := tt.input.String()
		if got != tt.want {
			t.Errorf("S(%q).String() = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestS_Bytes(t *testing.T) {
	tests := []struct {
		input S
		want  []byte
	}{
		{S("hello"), []byte("hello")},
		{S(""), []byte("")},
		{S("abc"), []byte("abc")},
	}
	for _, tt := range tests {
		got := tt.input.Bytes()
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("S(%q).Bytes() = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestS_FirstByte(t *testing.T) {
	tests := []struct {
		input S
		want  byte
	}{
		{S("hello"), 'h'},
		{S("abc"), 'a'},
		{S(""), 0},
		{S("0"), '0'},
	}
	for _, tt := range tests {
		got := tt.input.FirstByte()
		if got != tt.want {
			t.Errorf("S(%q).FirstByte() = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestS_Int64(t *testing.T) {
	tests := []struct {
		input    S
		want     int64
		wantErr  bool
	}{
		{S("123"), 123, false},
		{S("-456"), -456, false},
		{S("0"), 0, false},
		{S(" 789 "), 789, false},
		{S(""), 0, true},
		{S("abc"), 0, true},
		{S("12.34"), 0, true},
		{S("9223372036854775807"), 9223372036854775807, false}, // max int64
	}
	for _, tt := range tests {
		got, err := tt.input.Int64()
		if (err != nil) != tt.wantErr {
			t.Errorf("S(%q).Int64() error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("S(%q).Int64() = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestS_Float64(t *testing.T) {
	tests := []struct {
		input   S
		want    float64
		wantErr bool
	}{
		{S("123.45"), 123.45, false},
		{S("-456.78"), -456.78, false},
		{S("0"), 0, false},
		{S(" 3.14 "), 3.14, false},
		{S(""), 0, true},
		{S("abc"), 0, true},
		{S("1e10"), 1e10, false},
	}
	for _, tt := range tests {
		got, err := tt.input.Float64()
		if (err != nil) != tt.wantErr {
			t.Errorf("S(%q).Float64() error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("S(%q).Float64() = %f, want %f", tt.input, got, tt.want)
		}
	}
}

func TestS_CSVLike(t *testing.T) {
	tests := []struct {
		input S
		want  string
	}{
		{S("hello"), "%,hello,%"},
		{S(""), "%,,%"},
		{S("a,b"), "%,a,b,%"},
	}
	for _, tt := range tests {
		got := tt.input.CSVLike()
		if got != tt.want {
			t.Errorf("S(%q).CSVLike() = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestS_Like(t *testing.T) {
	tests := []struct {
		input S
		want  string
	}{
		{S("hello"), "%hello%"},
		{S(""), "%%"},
		{S("foo%bar"), "%foo%bar%"},
	}
	for _, tt := range tests {
		got := tt.input.Like()
		if got != tt.want {
			t.Errorf("S(%q).Like() = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestS_ID(t *testing.T) {
	tests := []struct {
		input   S
		want    ID
		wantErr bool
	}{
		{S("123"), ID(123), false},
		{S("0"), ID(0), false},
		{S(" 456 "), ID(456), false},
		{S(""), ID(0), false}, // empty after trim: no iteration → returns 0 without error
		{S("abc"), ID(0), true},
		{S("12a34"), ID(0), true},
		{S("000"), ID(0), false},
	}
	for _, tt := range tests {
		got, err := tt.input.ID()
		if (err != nil) != tt.wantErr {
			t.Errorf("S(%q).ID() error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("S(%q).ID() = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestS_MustID(t *testing.T) {
	tests := []struct {
		input S
		want  ID
	}{
		{S("123"), ID(123)},
		{S("0"), ID(0)},
		{S("abc"), ID(0)}, // invalid returns 0
		{S(""), ID(0)},    // invalid returns 0
		{S(" 456 "), ID(456)},
	}
	for _, tt := range tests {
		got := tt.input.MustID()
		if got != tt.want {
			t.Errorf("S(%q).MustID() = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestS_JSON(t *testing.T) {
	tests := []struct {
		input   S
		want    JSON
		wantErr bool
	}{
		{S(`{"key":"value"}`), JSON(`{"key":"value"}`), false},
		{S(`"just a string"`), JSON(`"just a string"`), false},
		{S(`123`), JSON(`123`), false},
		{S(`null`), nil, false},
		{S(`invalid json`), nil, true},
		{S(``), nil, true},
	}
	for _, tt := range tests {
		got, err := tt.input.JSON()
		if (err != nil) != tt.wantErr {
			t.Errorf("S(%q).JSON() error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr {
			if tt.want == nil {
				if got != nil {
					t.Errorf("S(%q).JSON() = %v, want nil", tt.input, got)
				}
			} else {
				// Compare by re-serializing to normalize
				var gotJSON, wantJSON any
				json.Unmarshal(got, &gotJSON)
				json.Unmarshal(tt.want, &wantJSON)
				if !reflect.DeepEqual(gotJSON, wantJSON) {
					t.Errorf("S(%q).JSON() = %s, want %s", tt.input, string(got), string(tt.want))
				}
			}
		}
	}
}

func TestS_SplitLines(t *testing.T) {
	tests := []struct {
		input   S
		want    SS
		wantErr bool
	}{
		{S("line1\nline2\nline3"), SS{"line1", "line2", "line3"}, false},
		{S("single line"), SS{"single line"}, false},
		{S(""), nil, false},
		{S("line1\r\nline2"), SS{"line1", "line2"}, false},
	}
	for _, tt := range tests {
		got, err := tt.input.SplitLines()
		if (err != nil) != tt.wantErr {
			t.Errorf("S(%q).SplitLines() error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr {
			if got == nil && tt.want != nil {
				t.Errorf("S(%q).SplitLines() = nil, want %v", tt.input, tt.want)
			} else if got != nil && !slices.Equal(got, tt.want) {
				t.Errorf("S(%q).SplitLines() = %v, want %v", tt.input, got, tt.want)
			}
		}
	}
}

func TestS_FirstRune(t *testing.T) {
	tests := []struct {
		input S
		want  rune
	}{
		{S("hello"), 'h'},
		{S("世界"), '世'},
		{S(""), 0},
		{S("a"), 'a'},
	}
	for _, tt := range tests {
		got := tt.input.FirstRune()
		if got != tt.want {
			t.Errorf("S(%q).FirstRune() = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestS_FirstRuneString(t *testing.T) {
	tests := []struct {
		input S
		want  S
	}{
		{S("hello"), S("h")},
		{S("世界"), S("世")},
		{S(""), S("")},
		{S("a"), S("a")},
	}
	for _, tt := range tests {
		got := tt.input.FirstRuneString()
		if got != tt.want {
			t.Errorf("S(%q).FirstRuneString() = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestS_SafeSlice(t *testing.T) {
	tests := []struct {
		input  S
		maxLen int
		want   string
	}{
		{S("hello"), 3, "hel"},
		{S("hello"), 10, "hello"},
		{S("hello"), 0, ""},
		{S("hello"), 1, "h"},   // 'h'=1 byte, 1>1? No → continue; 'e'=2, 2>1? Yes → runes[:1]
		{S("hello"), 2, "he"},  // 'h'=1, 'e'=2, 2>2? No → continue; 'l'=3, 3>2? Yes → runes[:2]
		{S("世界"), 3, "世"},     // 世=3 bytes, 3>3? No → continue; 界=6, 6>3? Yes → runes[:1]
		{S("世hello"), 3, "世"},  // 世=3, 3>3? No → continue; 'h'=4, 4>3? Yes → runes[:1]
		{S("世hello"), 4, "世h"}, // 世=3, 3>4? No → continue; 'h'=4, 4>4? No → continue; 'e'=5, 5>4? Yes → runes[:2]
		{S("你好世界"), 4, "你"}, // 你=3, 3>4? No; 好=6, 6>4? Yes → runes[:1] (4 falls mid-char in 好)
		{S("你好世界"), 5, "你"}, // 你=3, 3>5? No; 好=6, 6>5? Yes → runes[:1] (5 still inside 好's 3 bytes)
		{S(""), 5, ""},
	}
	for _, tt := range tests {
		got := tt.input.SafeSlice(tt.maxLen)
		if got != tt.want {
			t.Errorf("S(%q).SafeSlice(%d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

// TestS_Trim_Idempotent verifies Trim is idempotent
func TestS_Trim_Idempotent(t *testing.T) {
	s := S("  hello  ")
	once := s.Trim()
	twice := once.Trim()
	if once != twice {
		t.Errorf("Trim is not idempotent: first=%q, second=%q", once, twice)
	}
}

// TestS_Formalize_Idempotent verifies Formalize is idempotent
func TestS_Formalize_Idempotent(t *testing.T) {
	s := S("  Hello World  ")
	once := s.Formalize()
	twice := once.Formalize()
	if once != twice {
		t.Errorf("Formalize is not idempotent: first=%q, second=%q", once, twice)
	}
}

// TestS_Replace_NoMatchReturnsOriginal verifies Replace returns the same content when no match
func TestS_Replace_NoMatchReturnsOriginal(t *testing.T) {
	s := S("hello")
	got := s.Replace("x", "y")
	if got != s {
		t.Errorf("Replace on no match should return original: got %q, want %q", got, s)
	}
}

// TestS_Includes_AllSubstrings verifies Includes with multiple args
func TestS_Includes_AllSubstrings(t *testing.T) {
	s := S("the quick brown fox")
	if !s.Includes("quick", "brown") {
		t.Error("Includes should return true when all substrings are present")
	}
	if s.Includes("quick", "missing") {
		t.Error("Includes should return false when any substring is missing")
	}
}

// TestS_MustID_PanicsNever verifies MustID never panics
func TestS_MustID_PanicsNever(t *testing.T) {
	inputs := []S{"123", "", "abc", " 456 ", "12a34"}
	for _, input := range inputs {
		// MustID should never panic
		_ = input.MustID()
	}
}

// TestS_CamelToSnake_ConsecutiveUppercase verifies consecutive uppercase letters
func TestS_CamelToSnake_ConsecutiveUppercase(t *testing.T) {
	s := S("XMLParser")
	want := S("x_m_l_parser")
	got := s.CamelToSnake()
	if got != want {
		t.Errorf("CamelToSnake(%q) = %q, want %q", s, got, want)
	}
}
