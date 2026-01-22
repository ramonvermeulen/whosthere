package logging

import (
	"os"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected zapcore.Level
	}{
		{"trace", zapcore.DebugLevel},
		{"debug", zapcore.DebugLevel},
		{"info", zapcore.InfoLevel},
		{"", zapcore.InfoLevel},
		{"warn", zapcore.WarnLevel},
		{"error", zapcore.ErrorLevel},
		{"invalid", zapcore.InfoLevel},
	}

	for _, test := range tests {
		result := ParseLevel(test.input)
		if result != test.expected {
			t.Errorf("ParseLevel(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestLevelFromEnv(t *testing.T) {
	// No env
	_ = os.Unsetenv("WHOSTHERE_LOG")
	_ = os.Unsetenv("WHOSTHERE_DEBUG")
	level := LevelFromEnv(zapcore.InfoLevel)
	if level != zapcore.InfoLevel {
		t.Errorf("expected InfoLevel, got %v", level)
	}

	// Set WHOSTHERE_LOG
	t.Setenv("WHOSTHERE_LOG", "debug")
	level = LevelFromEnv(zapcore.InfoLevel)
	if level != zapcore.DebugLevel {
		t.Errorf("expected DebugLevel, got %v", level)
	}

	// Set WHOSTHERE_DEBUG
	t.Setenv("WHOSTHERE_LOG", "")
	t.Setenv("WHOSTHERE_DEBUG", "1")
	level = LevelFromEnv(zapcore.InfoLevel)
	if level != zapcore.DebugLevel {
		t.Errorf("expected DebugLevel, got %v", level)
	}
}

func TestL(t *testing.T) {
	// Before init, should return nop
	logger := L()
	if logger == nil {
		t.Errorf("expected logger")
	}
	// Can't easily test after init without global state
}

func TestResolveLogPath(t *testing.T) {
	path, err := resolveLogPath()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if path == "" {
		t.Errorf("expected path")
	}
}
