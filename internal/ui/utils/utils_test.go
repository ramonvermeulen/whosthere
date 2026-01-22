package utils

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

func TestColorToHexTag(t *testing.T) {
	color := tcell.NewRGBColor(255, 128, 0)
	tag := ColorToHexTag(color)
	expected := "#ff8000"
	if tag != expected {
		t.Errorf("expected %s, got %s", expected, tag)
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]int{"b": 2, "a": 1, "c": 3}
	keys := SortedKeys(m)
	expected := []string{"a", "b", "c"}
	if len(keys) != len(expected) {
		t.Errorf("expected %v, got %v", expected, keys)
	}
	for i, k := range keys {
		if k != expected[i] {
			t.Errorf("expected %v, got %v", expected, keys)
		}
	}
}

func TestFmtDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{500 * time.Millisecond, "<1s"},
		{30 * time.Second, "30s"},
		{2 * time.Minute, "2m"},
	}
	for _, test := range tests {
		result := FmtDuration(test.duration)
		if result != test.expected {
			t.Errorf("FmtDuration(%v) = %s, expected %s", test.duration, result, test.expected)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello", 3, "he…"},
		{"hello", 1, "h"},
		{"hello", 0, "hello"},
	}
	for _, test := range tests {
		result := Truncate(test.input, test.maxLen)
		if result != test.expected {
			t.Errorf("Truncate(%s, %d) = %s, expected %s", test.input, test.maxLen, result, test.expected)
		}
	}
}
