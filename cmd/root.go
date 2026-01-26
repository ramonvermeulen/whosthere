package cmd

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/version"
	"github.com/ramonvermeulen/whosthere/internal/ui"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	appName      = "whosthere"
	shortAppDesc = "Local network discovery tool with a modern TUI interface."
	cyan         = "\033[36m"
	reset        = "\033[0m"
	longAppDesc  string
)

var (
	rootCmd = &cobra.Command{
		Use:          appName,
		Short:        shortAppDesc,
		Long:         longAppDesc,
		SilenceUsage: true,
		RunE:         run,
	}

	whosthereFlags = &config.Flags{}
)

func init() {
	if theme.IsNoColor() {
		cyan = ""
		reset = ""
	}
	longAppDesc = cyan + "whosthere [global flags] <subcommand> [args]\n" + reset + `
Knock Knock..
          _               _   _                   ___
__      _| |__   ___  ___| |_| |__   ___ _ __ ___/ _ \
\ \ /\ / / '_ \ / _ \/ __| __| '_ \ / _ \ '__/ _ \// /
 \ V  V /| | | | (_) \__ \ |_| | | |  __/ | |  __/ \/
  \_/\_/ |_| |_|\___/|___/\__|_| |_|\___|_|  \___| ()


Local Area Network discovery tool with a modern Terminal User Interface (TUI) written in Go.
Discover, explore, and understand your LAN in an intuitive way.

Knock Knock... who's there? 🚪`
	rootCmd.Long = longAppDesc
	initWhosthereFlags()
	setCobraUsageTemplate()
}

// Execute is the entrypoint for the CLI application
func Execute() {
	cobra.MousetrapHelpText = ""
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func SetVersion(v string) {
	rootCmd.Version = v
}

func run(*cobra.Command, []string) error {
	result, err := InitComponents(whosthereFlags.ConfigFile, whosthereFlags.NetworkInterface, false)
	if err != nil {
		return err
	}

	logger := result.Logger
	logPath := result.LogPath
	cfg := result.Config
	ouiDB := result.OuiDB

	logger.Info("logger initialized", zap.String("path", logPath), zap.String("level", logger.Level().String()))

	if ouiDB == nil {
		logger.Warn("OUI database is not initialized; manufacturer lookups will be disabled")
	}

	if whosthereFlags.PprofPort != "" {
		go func() {
			logger.Info("starting pprof server", zap.String("port", whosthereFlags.PprofPort))
			if err := http.ListenAndServe(":"+whosthereFlags.PprofPort, nil); err != nil {
				logger.Error("pprof server failed", zap.Error(err))
			}
		}()
	}

	app, err := ui.NewApp(cfg, ouiDB, version.Version)
	if err != nil {
		logger.Error("failed to create app", zap.Error(err))
		return err
	}

	if err := app.Run(); err != nil {
		logger.Error("app run failed", zap.Error(err))
		return err
	}

	return nil
}

func initWhosthereFlags() {
	rootCmd.PersistentFlags().StringVarP(
		&whosthereFlags.ConfigFile,
		"config-file", "c",
		"",
		"Path to config file (overrides default).",
	)
	rootCmd.PersistentFlags().StringVar(
		&whosthereFlags.PprofPort,
		"pprof-port",
		"",
		"Pprof HTTP server port for debugging and profiling purposes (e.g., 6060)",
	)
	rootCmd.PersistentFlags().StringVarP(
		&whosthereFlags.NetworkInterface,
		"interface", "i",
		"",
		"Network interface to use for scanning (overrides config).",
	)
}

func setCobraUsageTemplate() {
	cobra.AddTemplateFunc("StyleHeading", func(s string) string { return cyan + s + reset })
	usageTemplate := rootCmd.UsageTemplate()
	usageTemplate = strings.NewReplacer(
		`Usage:`, `{{StyleHeading "Usage:"}}`,
		`Available Commands:`, `{{StyleHeading "Available Commands:"}}`,
		`Flags:`, `{{StyleHeading "Flags:"}}`,
		`Global Flags:`, `{{StyleHeading "Global Flags:"}}`,
	).Replace(usageTemplate)
	rootCmd.SetUsageTemplate(usageTemplate)
}
