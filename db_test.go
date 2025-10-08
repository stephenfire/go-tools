package tools

import (
	"testing"
	"time"
)

func TestDate(t *testing.T) {
	t.Log(time.Local.String())
	tests := []struct {
		input  Date
		want   string
		output Date
	}{
		{Date(NewDate(2025, 10, 8, 21, 12, 0)), "2025-10-08", NewADate(2025, 10, 8)},
	}

	for _, test := range tests {
		s := test.input.String()
		if s != test.want {
			t.Fatal(test.input, ":", s)
		}
		ss := "\"" + s + "\""
		var targetDate Date
		if err := targetDate.UnmarshalJSON([]byte(ss)); err != nil {
			t.Fatal(err)
		}
		if test.input.Formalize() != targetDate {
			t.Fatal(test.input.Formalize(), ":", targetDate)
		}
		if targetDate != test.output {
			t.Fatal(Time(test.input).String(), ":", Time(targetDate).String(), ":", Time(test.output).String())
		}
		t.Log(Time(test.input).String(), ":", Time(targetDate).String(), ":", Time(test.output).String())
	}
}
