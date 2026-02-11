package cmd

import (
	"context"
	"os"

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
  whosthere scan --sweeper=false
  whosthere scan --mdns=false --ssdp=false
  whosthere scan --timeout 15s
  whosthere scan --json
`,
		RunE: runScan,
	}

	cmd.Flags().Bool("json", false, "Output results in JSON format.")
	cmd.Flags().Bool("pretty", false, "Pretty print output.")

	return cmd
}

func runScan(cmd *cobra.Command, _ []string) error {
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

	results, err := eng.Scan(ctx)

	spinner.Stop()

	if err != nil {
		return err
	}

	format, opts := parseScanSpecificFlags(cmd)
	out, err := output.NewOutput(format, opts...)
	if err != nil {
		return err
	}

	return out.PrintDevices(os.Stdout, results)
}

func parseScanSpecificFlags(cmd *cobra.Command) (output.Format, []output.Option) {
	var opts []output.Option

	jsonFlag, _ := cmd.Flags().GetBool("json")
	format := output.FormatTable
	if jsonFlag {
		format = output.FormatJSON
	}

	prettyFlag, _ := cmd.Flags().GetBool("pretty")
	if prettyFlag {
		opts = append(opts, output.WithPretty())
	}

	return format, opts
}
