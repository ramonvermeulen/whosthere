// pkg/logging/logger.go
package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ramonvermeulen/whosthere/internal/core/paths"
	"github.com/samber/slog-zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	zapLogger  *zap.Logger
	slogLogger *slog.Logger
	once       sync.Once
)

// parseLevel converts a string like "DEBUG" to zapcore.Level.
// Supports TRACE (mapped to DEBUG), DEBUG, INFO, WARN, ERROR, DPANIC, PANIC, FATAL.
func parseLevel(s string) zapcore.Level {
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

// levelFromEnv returns the level from WHOSTHERE_LOG if set, else default.
func levelFromEnv(defaultLevel zapcore.Level) zapcore.Level {
	if v := os.Getenv("WHOSTHERE_LOG"); v != "" {
		return parseLevel(v)
	}
	return defaultLevel
}

// Init sets up both zap and slog loggers
func Init(enableStdout bool) (*slog.Logger, error) {
	var initErr error
	once.Do(func() {
		path, err := resolveLogPath()
		if err != nil {
			initErr = err
			return
		}
		level := levelFromEnv(zapcore.InfoLevel)

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

		zapLogger = zap.New(core, zap.AddCaller())
		zap.ReplaceGlobals(zapLogger)

		// Create slog logger that uses zap
		slogLogger = slog.New(slogzap.Option{
			Logger: zapLogger,
			Level:  convertZapLevelToSlog(level),
		}.NewZapHandler())
	})

	return slogLogger, initErr
}

// L returns the global zap logger (for backward compatibility)
func L() *zap.Logger {
	if zapLogger == nil {
		return zap.NewNop()
	}
	return zapLogger
}

// S returns the global slog logger
func S() *slog.Logger {
	if slogLogger == nil {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}
	return slogLogger
}

func resolveLogPath() (string, error) {
	dir, err := paths.StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "app.log"), nil
}

func convertZapLevelToSlog(level zapcore.Level) slog.Level {
	switch level {
	case zapcore.DebugLevel:
		return slog.LevelDebug
	case zapcore.InfoLevel:
		return slog.LevelInfo
	case zapcore.WarnLevel:
		return slog.LevelWarn
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
