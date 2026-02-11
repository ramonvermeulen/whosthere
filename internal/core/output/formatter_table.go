package output

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

var _ Formatter = (*TableFormatter)(nil)

// TableFormatter implements Formatter for table output
type TableFormatter struct{}

func NewTableFormatter() *TableFormatter {
	return &TableFormatter{}
}

func (f *TableFormatter) Format(w io.Writer, results *discovery.ScanResults) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	_, _ = fmt.Fprintln(tw, "IP\tDISPLAY NAME\tMAC\tMANUFACTURER")
	_, _ = fmt.Fprintln(tw, "──\t────────────\t───\t────────────")

	for _, d := range results.Devices {
		ip := d.IP().String()
		name := d.DisplayName()
		mac := d.MAC()
		manufacturer := d.Manufacturer()

		if name == "" {
			name = "-"
		}
		if mac == "" {
			mac = "-"
		}
		if manufacturer == "" {
			manufacturer = "-"
		}

		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", ip, name, mac, manufacturer)
	}

	if err := tw.Flush(); err != nil {
		return err
	}

	_, err := fmt.Fprintf(w, "\nScan completed: %d device(s) found in %s\n", len(results.Devices), formatDuration(results.Stats.Duration))
	return err
}
