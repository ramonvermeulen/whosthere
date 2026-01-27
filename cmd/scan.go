package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/ramonvermeulen/whosthere/internal/core"
	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
)

func NewScanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Run network scanners standalone for debugging/experimentation",
		Long: `Run one or more scanners directly (mdns, ssdp, arp).

Examples:
 whosthere scan -s mdns
 whosthere scan -s "arp,ssdp" -t 30
`,
		RunE: runScan,
	}

	cmd.Flags().StringP("scanner", "s", "", "Comma-separated scanners to run, this overrides the config")
	cmd.Flags().IntP("timeout", "t", 0, "Timeout in seconds for the scan")
	return cmd
}

func runScan(cmd *cobra.Command, args []string) error {
	scannerNames, _ := cmd.Flags().GetString("scanner")
	timeoutSec, _ := cmd.Flags().GetInt("timeout")
	scanDuration := time.Duration(timeoutSec) * time.Second

	result, err := InitComponents("", whosthereFlags.NetworkInterface, true)
	if err != nil {
		return err
	}

	err = applyFlagOverrides(result.Config, scannerNames, scanDuration)
	if err != nil {
		return err
	}

	eng := core.BuildEngine(result.Interface, result.OuiDB, result.Config)

	if eng.Sweeper != nil {
		go eng.Sweeper.Start(context.Background())
	}

	ctx, cancel := context.WithTimeout(context.Background(), result.Config.ScanDuration)
	defer cancel()
	devices, err := eng.Stream(ctx, func(_ *discovery.Device) {})
	if err != nil {
		return err
	}

	zap.L().Info("scan complete", zap.Int("devices", len(devices)))
	for _, d := range devices {
		zap.L().Info("device",
			zap.String("ip", d.IP.String()),
			zap.String("hostname", d.DisplayName),
			zap.String("mac", d.MAC),
			zap.String("manufacturer", d.Manufacturer),
		)
	}
	return nil
}

// todo(ramon): find a better pattern for flags -> env vars -> config
// would be nice to have a single pattern
// e.g. flags always taking precedence over env vars, which take precedence over config
// with a single source of truth somehow, always resulting in a *config.Config state
func applyFlagOverrides(cfg *config.Config, scannerNames string, duration time.Duration) error {
	if duration > 0 {
		cfg.ScanDuration = duration
	}

	if scannerNames == "" {
		return nil
	}

	requested := strings.Split(scannerNames, ",")

	cfg.Scanners = config.ScannerConfig{
		SSDP: config.ScannerToggle{Enabled: false},
		ARP:  config.ScannerToggle{Enabled: false},
		MDNS: config.ScannerToggle{Enabled: false},
	}

	for _, r := range requested {
		r = strings.TrimSpace(strings.ToLower(r))
		switch r {
		case "ssdp":
			cfg.Scanners.SSDP.Enabled = true
		case "arp":
			cfg.Scanners.ARP.Enabled = true
		case "mdns":
			cfg.Scanners.MDNS.Enabled = true
		case "all":
			cfg.Scanners.SSDP.Enabled = true
			cfg.Scanners.ARP.Enabled = true
			cfg.Scanners.MDNS.Enabled = true
		default:
			return fmt.Errorf("invalid scanner '%s'. Allowed: mdns, arp, ssdp, all", r)
		}
	}

	return nil
}
