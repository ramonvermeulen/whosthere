package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ramonvermeulen/whosthere/internal/core/paths"
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
// enableStdout: if true, also logs to stdout with console format.
func Init(level zapcore.Level, enableStdout bool) (*zap.Logger, string, error) {
	var initErr error
	once.Do(func() {
		path, err := resolveLogPath()
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
		jsonEncoder := zapcore.NewJSONEncoder(encCfg)

		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			initErr = fmt.Errorf("open log file: %w", err)
			return
		}
		wsFile := zapcore.AddSync(f)

		fileCore := zapcore.NewCore(jsonEncoder, wsFile, level)

		var core zapcore.Core
		if enableStdout {
			wsStdout := zapcore.AddSync(os.Stdout)
			stdoutCore := zapcore.NewCore(jsonEncoder, wsStdout, level)
			core = zapcore.NewTee(fileCore, stdoutCore)
		} else {
			core = fileCore
		}

		logger = zap.New(core, zap.AddCaller())
		zap.ReplaceGlobals(logger)
	})
	return logger, cachedPath, initErr
}

// L returns the global logger. If Init hasn't been called yet it
// returns a no-op logger without mutating the global zap logger.
//
// This makes accidental logging before initialization cheap and
// predictable, while keeping Init as the only place that installs
// a real global logger.
func L() *zap.Logger {
	if logger == nil {
		return zap.NewNop()
	}
	return logger
}

func resolveLogPath() (string, error) {
	dir, err := paths.StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "app.log"), nil
}
