package cmd

import (
	"context"
	"os"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core"
	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/output"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/spf13/cobra"
)

func NewScanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Run a single discovery scan and output results to the console",
		Long: `Run exactly one discovery scan.

By default, all scanners (mDNS, SSDP, ARP) and the sweeper are enabled.
Use --no-xxx flags to disable specific scanners.` + magenta + `

Examples:` + reset + `
  whosthere scan
  whosthere scan --no-sweeper
  whosthere scan --no-mdns --no-ssdp
  whosthere scan --timeout 15s
`,
		RunE: runScan,
	}

	return cmd
}

func runScan(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	cfg, err := config.LoadForMode(config.ModeCLI, whosthereFlags)
	if err != nil {
		return err
	}

	eng, err := core.BuildEngine(cfg, discovery.NoOpLogger{})
	if err != nil {
		return err
	}

	spinner := output.NewSpinner(os.Stdout, "Scanning network...", cfg.ScanTimeout)
	spinner.Start()

	start := time.Now()
	devices, err := eng.Scan(ctx)
	elapsed := time.Since(start)

	spinner.Stop()

	if err != nil {
		return err
	}

	output.PrintDevices(os.Stdout, devices, elapsed)
	return nil
}
