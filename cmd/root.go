package cmd

import (
	"os"

	"github.com/ramonvermeulen/whosthere/internal/config"
	"github.com/ramonvermeulen/whosthere/internal/logging"
	"github.com/ramonvermeulen/whosthere/internal/ui"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	appName      = "whosthere"
	shortAppDesc = "Local network discovery tool with a modern TUI interface."
	longAppDesc  = `Local network discovery tool with a modern TUI interface written in Go.
Discover, explore, and understand your Local Area Network in an intuitive way.

Knock Knock... who's there? 🚪`
)

var (
	rootCmd = &cobra.Command{
		Use:   appName,
		Short: shortAppDesc,
		Long:  longAppDesc,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		RunE: run,
	}
	whosthereFlags = config.NewFlags()
)

func init() {
	initWhosthereFlags()
}

// Execute is the entrypoint for the CLI application
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func run(*cobra.Command, []string) error {
	level := logging.LevelFromEnv(zapcore.InfoLevel)
	logger, logPath, err := logging.Init(appName, level)
	if err != nil {
		return err
	} else {
		logger.Info("logger initialized", zap.String("path", logPath), zap.String("level", level.String()))
	}

	cfg, _, err := config.Load(whosthereFlags.ConfigFile)
	if err != nil {
		zap.L().Error("failed to load or create config", zap.Error(err))
		return err
	}

	app := ui.NewApp(cfg)
	if err := app.Run(); err != nil {
		zap.L().Error("ui run failed", zap.Error(err))
		return err
	}

	return nil
}

func initWhosthereFlags() {
	rootCmd.Flags().StringVarP(
		&whosthereFlags.ConfigFile,
		"config-file", "c",
		"",
		"Path to config file.",
	)
}
