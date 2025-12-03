package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger     *zap.Logger
	once       sync.Once
	cachedPath string
)

// ParseLevel converts a string like "DEBUG" to zapcore.Level.
// Supports TRACE (mapped to DEBUG), DEBUG, INFO, WARN, ERROR, DPANIC, PANIC, FATAL.
func ParseLevel(s string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "trace":
		return zapcore.DebugLevel
	case "debug":
		return zapcore.DebugLevel
	case "info", "":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// LevelFromEnv returns the level from WHOSTHERE_LOG if set, else default.
// Backward-compat: WHOSTHERE_DEBUG=1 forces DEBUG.
func LevelFromEnv(defaultLevel zapcore.Level) zapcore.Level {
	if v := os.Getenv("WHOSTHERE_LOG"); v != "" {
		return ParseLevel(v)
	}
	if os.Getenv("WHOSTHERE_DEBUG") == "1" {
		return zapcore.DebugLevel
	}
	return defaultLevel
}

// Init sets up a global Zap logger writing JSON to a file.
// level: zapcore.InfoLevel for prod, zapcore.DebugLevel for dev.
func Init(appName string, level zapcore.Level) (*zap.Logger, string, error) {
	var initErr error
	once.Do(func() {
		path, err := resolveLogPath(appName)
		if err != nil {
			initErr = err
			return
		}
		cachedPath = path

		encCfg := zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stack",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
		encoder := zapcore.NewJSONEncoder(encCfg)

		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			initErr = fmt.Errorf("open log file: %w", err)
			return
		}
		ws := zapcore.AddSync(f)

		core := zapcore.NewCore(encoder, ws, level)
		logger = zap.New(core, zap.AddCaller())
		zap.ReplaceGlobals(logger)
	})
	return logger, cachedPath, initErr
}

// L returns the global logger (no-op if Init wasn't called).
func L() *zap.Logger {
	if logger == nil {
		logger = zap.NewNop()
		zap.ReplaceGlobals(logger)
	}
	return logger
}

// LogPath returns the current log file path.
func LogPath() string { return cachedPath }

func resolveLogPath(appName string) (string, error) {
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "state")
	}
	dir := filepath.Join(base, appName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "app.log"), nil
}
