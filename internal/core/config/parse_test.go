package config

import (
	"testing"
	"time"
)

func TestParseBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		wantErr  bool
	}{
		{"true", true, false},
		{"false", false, false},
		{"TRUE", true, false},
		{"FALSE", false, false},
		{"1", true, false},
		{"0", false, false},
		{"yes", true, false},
		{"no", false, false},
		{"y", true, false},
		{"n", false, false},
		{"on", true, false},
		{"off", false, false},
		{"t", true, false},
		{"f", false, false},
		{"  true  ", true, false},
		{"invalid", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseBool(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBool(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"1s", time.Second, false},
		{"5m", 5 * time.Minute, false},
		{"2h", 2 * time.Hour, false},
		{"500ms", 500 * time.Millisecond, false},
		{"  10s  ", 10 * time.Second, false},
		{"15", 15 * time.Second, false},
		{"  30  ", 30 * time.Second, false},
		{"0", 0, false},
		{"", 0, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("parseDuration(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		wantErr  bool
	}{
		{"0", 0, false},
		{"42", 42, false},
		{"-10", -10, false},
		{"  100  ", 100, false},
		{"not a number", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseInt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInt(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("parseInt(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseIntSlice(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
		wantErr  bool
	}{
		{"", []int{}, false},
		{"22", []int{22}, false},
		{"22,80,443", []int{22, 80, 443}, false},
		{"22, 80, 443", []int{22, 80, 443}, false},
		{"  22 , 80 , 443  ", []int{22, 80, 443}, false},
		{"22,,80", []int{22, 80}, false},
		{"invalid", nil, true},
		{"22,invalid,80", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseIntSlice(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIntSlice(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equalIntSlices(got, tt.expected) {
				t.Errorf("parseIntSlice(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func equalIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
