package cmd

import (
	"context"
	"os"

	"github.com/ramonvermeulen/whosthere/internal/config"
	"github.com/ramonvermeulen/whosthere/internal/logging"
	"github.com/ramonvermeulen/whosthere/internal/oui"
	"github.com/ramonvermeulen/whosthere/internal/ui"
	"github.com/ramonvermeulen/whosthere/internal/version"
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

	whosthereFlags = &config.Flags{}
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
	ctx := context.Background()

	level := logging.LevelFromEnv(zapcore.InfoLevel)
	logger, logPath, err := logging.Init(appName, level, false)
	if err != nil {
		return err
	} else {
		logger.Info("logger initialized", zap.String("path", logPath), zap.String("level", level.String()))
	}

	cfg, err := config.Load(whosthereFlags.ConfigFile)
	if err != nil {
		zap.L().Error("failed to load or create config", zap.Error(err))
		return err
	}

	ouiDB, err := oui.Init(ctx)
	if err != nil {
		zap.L().Warn("failed to initialize OUI database; manufacturer lookups will be disabled", zap.Error(err))
	}

	tui := ui.NewApp(cfg, ouiDB, version.Version)
	if err := tui.Run(); err != nil {
		zap.L().Error("tui run failed", zap.Error(err))
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
