package cmd

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/ramonvermeulen/whosthere/internal/core"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/core/version"
)

func NewDaemonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Run whosthere in daemon mode with an HTTP API",
		Long: `Run whosthere in daemon mode, continuously scanning the network and providing live device data via HTTP API.

Examples:
 whosthere daemon --port 8080
`,
		RunE: runDaemon,
	}

	cmd.Flags().StringP("port", "p", "", "Port for the HTTP API server")
	return cmd
}

func runDaemon(cmd *cobra.Command, _ []string) error {
	port, _ := cmd.Flags().GetString("port")
	if port == "" {
		port = "8080"
		zap.L().Info("no port specified, using default port", zap.String("port", port))
	}

	result, err := InitComponents("", whosthereFlags.NetworkInterface, true)
	if err != nil {
		return err
	}

	appState := state.NewAppState(result.Config, version.Version)
	eng := core.BuildEngine(result.Interface, result.OuiDB, result.Config)

	http.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		handleDevices(w, r, appState)
	})
	http.HandleFunc("/devices/", func(w http.ResponseWriter, r *http.Request) {
		handleDeviceByIP(w, r, appState)
	})
	http.HandleFunc("/health", handleHealth)

	go func() {
		zap.L().Info("starting HTTP server", zap.String("port", port))
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			zap.L().Error("HTTP server failed", zap.Error(err))
		}
	}()

	if eng.Sweeper != nil {
		go eng.Sweeper.Start(context.Background())
	}

	for {
		zap.L().Info("starting scan cycle")
		_, err := eng.Stream(context.Background(), func(d *discovery.Device) {
			appState.UpsertDevice(d)
		})
		if err != nil {
			zap.L().Error("scan failed", zap.Error(err))
		}
		time.Sleep(result.Config.ScanInterval)
	}
}

func handleDevices(w http.ResponseWriter, r *http.Request, appState *state.AppState) {
	zap.L().Info("incoming request", zap.String("method", r.Method), zap.String("path", r.URL.Path))
	devices := appState.DevicesSnapshot()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(devices); err != nil {
		http.Error(w, "Failed to encode devices", http.StatusInternalServerError)
		return
	}
}

func handleDeviceByIP(w http.ResponseWriter, r *http.Request, appState *state.AppState) {
	zap.L().Info("incoming request", zap.String("method", r.Method), zap.String("path", r.URL.Path))
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

func handleHealth(w http.ResponseWriter, r *http.Request) {
	zap.L().Info("incoming request", zap.String("method", r.Method), zap.String("path", r.URL.Path))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
