package cmd

import (
	"context"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/ramonvermeulen/whosthere/internal/core"
	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Run network scanners standalone for debugging/experimentation",
	Long: `Run one or more scanners directly (mdns, ssdp, arp).

Examples:
 whosthere scan -s mdns
 whosthere scan -s "arp,ssdp" -t 30
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		scannerNames, _ := cmd.Flags().GetString("scanner")
		timeoutSec, _ := cmd.Flags().GetInt("timeout")
		scanDuration := time.Duration(timeoutSec) * time.Second

		result, err := InitComponents("", whosthereFlags.NetworkInterface, true)
		if err != nil {
			return err
		}

		ctx := context.Background()

		applyFlagOverrides(result.Config, scannerNames, scanDuration)
		eng := core.BuildEngine(result.Interface, result.OuiDB, result.Config)

		ctx, cancel := context.WithTimeout(ctx, scanDuration)
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
	},
}

func init() {
	scanCmd.Flags().StringP("scanner", "s", "", "Comma-separated scanners to run, this overrides the config")
	scanCmd.Flags().IntP("timeout", "t", 10, "Timeout in seconds for the scan")
	rootCmd.AddCommand(scanCmd)
}

func applyFlagOverrides(cfg *config.Config, scannerNames string, duration time.Duration) {
	cfg.ScanDuration = duration
	cfg.Sweeper.Enabled = false

	if scannerNames == "" {
		return
	}

	cfg.Scanners = config.ScannerConfig{
		SSDP: config.ScannerToggle{Enabled: false},
		ARP:  config.ScannerToggle{Enabled: false},
		MDNS: config.ScannerToggle{Enabled: false},
	}

	requested := strings.Split(scannerNames, ",")
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
		}
	}
}
