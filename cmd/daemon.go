package cmd

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/ramonvermeulen/whosthere/internal/core"
	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/logging"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/core/version"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/spf13/cobra"
)

func NewDaemonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Run whosthere in daemon mode with an HTTP API",
		Long: `Run whosthere in daemon mode, continuously scanning the network and providing live device data via HTTP API.
` + magenta + `
Examples:` + reset + `
  whosthere daemon --port=8080
`,
		RunE: runDaemon,
	}

	cmd.Flags().StringP("port", "p", "", "Port for the HTTP API server")
	return cmd
}

func runDaemon(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()
	logger, err := logging.New(true)
	if err != nil {
		return err
	}

	port, _ := cmd.Flags().GetString("port")
	if port == "" {
		port = "8080"
		logger.Log(ctx, slog.LevelInfo, "no port specified, defaulting to 8080")
	}

	cfg, err := config.LoadForMode(config.ModeApp, whosthereFlags)
	if err != nil {
		return err
	}

	appState := state.NewAppState(cfg, version.Version)
	eng, err := core.BuildEngine(cfg, logger)
	if err != nil {
		return err
	}

	http.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		logger.Log(ctx, slog.LevelDebug, "received request", "method", r.Method, "path", r.URL.Path)
		handleDevices(w, r, appState)
	})
	http.HandleFunc("/devices/", func(w http.ResponseWriter, r *http.Request) {
		logger.Log(ctx, slog.LevelDebug, "received request", "method", r.Method, "path", r.URL.Path)
		handleDeviceByIP(w, r, appState)
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		logger.Log(ctx, slog.LevelDebug, "received request", "method", r.Method, "path", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	go func() {
		logger.Log(context.Background(), slog.LevelInfo, "starting HTTP server", "port", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			logger.Log(context.Background(), slog.LevelError, "HTTP server failed", "error", err)
		}
	}()

	go func() {
		for event := range eng.Events {
			switch event.Type {
			case discovery.EventScanStarted:
			case discovery.EventScanCompleted:
			case discovery.EventDeviceDiscovered:
				if event.Device != nil {
					appState.UpsertDevice(event.Device)
				}
			case discovery.EventError:
			default:
			}
		}
	}()

	eng.Start(context.Background())

	select {}
}

func handleDevices(w http.ResponseWriter, _ *http.Request, appState *state.AppState) {
	devices := appState.DevicesSnapshot()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(devices); err != nil {
		http.Error(w, "Failed to encode devices", http.StatusInternalServerError)
		return
	}
}

func handleDeviceByIP(w http.ResponseWriter, r *http.Request, appState *state.AppState) {
	ipStr := strings.TrimPrefix(r.URL.Path, "/devices/")
	if ipStr == "" {
		http.NotFound(w, r)
		return
	}
	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil {
		http.Error(w, "Invalid IP address", http.StatusBadRequest)
		return
	}
	device, ok := appState.GetDevice(ipStr)
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(device); err != nil {
		http.Error(w, "Failed to encode device", http.StatusInternalServerError)
		return
	}
}
