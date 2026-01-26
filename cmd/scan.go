package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/ramonvermeulen/whosthere/internal/core"
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

		var enabled []string
		requested := strings.Split(scannerNames, ",")
		for _, r := range requested {
			r = strings.TrimSpace(strings.ToLower(r))
			switch r {
			case "", "all":
				enabled = []string{"ssdp", "arp", "mdns"}
			case "ssdp", "arp", "mdns":
				enabled = append(enabled, r)
			default:
				return fmt.Errorf("unknown scanner: %s", r)
			}
		}

		if len(enabled) == 0 {
			return fmt.Errorf("no scanners selected")
		}

		eng := core.BuildEngine(result.Interface, result.OuiDB, enabled, scanDuration)

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
	scanCmd.Flags().StringP("scanner", "s", "all", "Comma-separated scanners to run (mdns,ssdp,arp,all)")
	scanCmd.Flags().IntP("timeout", "t", 10, "Timeout in seconds for the scan")
	rootCmd.AddCommand(scanCmd)
}
