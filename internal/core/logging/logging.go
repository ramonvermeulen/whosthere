// pkg/logging/logger.go
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ramonvermeulen/whosthere/internal/core/paths"
)

var (
	slogLogger *slog.Logger
	once       sync.Once
)

// parseLevel converts a string like "DEBUG" to slog.Level.
// Supports TRACE (mapped to DEBUG), DEBUG, INFO, WARN, ERROR.
func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "trace", "debug":
		return slog.LevelDebug
	case "info", "":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// levelFromEnv returns the level from WHOSTHERE_LOG if set, else default.
func levelFromEnv(defaultLevel slog.Level) slog.Level {
	if v := os.Getenv("WHOSTHERE_LOG"); v != "" {
		return parseLevel(v)
	}
	return defaultLevel
}

// New sets up a new slog logger instance
func New(enableStdout bool) (*slog.Logger, error) {
	var initErr error
	once.Do(func() {
		path, err := resolveLogPath()
		if err != nil {
			initErr = err
			return
		}

		level := levelFromEnv(slog.LevelInfo)

		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			initErr = fmt.Errorf("open log file: %w", err)
			return
		}

		var w io.Writer = f
		if enableStdout {
			w = io.MultiWriter(f, os.Stdout)
		}

		h := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
		slogLogger = slog.New(h)
		slog.SetDefault(slogLogger)
	})

	return slogLogger, initErr
}

func resolveLogPath() (string, error) {
	dir, err := paths.StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "app.log"), nil
}
