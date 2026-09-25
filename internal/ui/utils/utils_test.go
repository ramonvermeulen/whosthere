package utils

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestColorToHexTag(t *testing.T) {
	color := tcell.NewRGBColor(255, 128, 0)
	tag := ColorToHexTag(color)
	expected := "#ff8000"
	assert.Equal(t, expected, tag)
}

func TestSortedKeys(t *testing.T) {
	m := map[string]int{"b": 2, "a": 1, "c": 3}
	keys := SortedKeys(m)
	expected := []string{"a", "b", "c"}
	assert.Equal(t, expected, keys)
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
		assert.Equal(t, test.expected, result, "FmtDuration(%v)", test.duration)
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
		assert.Equal(t, test.expected, result, "Truncate(%s, %d)", test.input, test.maxLen)
	}
}
