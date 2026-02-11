package output

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

// Format defines the output format type
type Format int

const (
	FormatTable Format = iota
	FormatJSON
)

var DefaultSortFunc = func(a, b *discovery.Device) bool {
	return discovery.CompareIPs(a.IP(), b.IP())
}

const DefaultPretty = false

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.1fs", d.Seconds())
}

// Output handles device output formatting
type Output struct {
	formatter Formatter
	sortFunc  func(a, b *discovery.Device) bool
	pretty    bool
}

// Formatter defines the interface for output formatters
type Formatter interface {
	Format(w io.Writer, results *discovery.ScanResults) error
}

// NewOutput creates a new output handler with the given options
func NewOutput(format Format, opts ...Option) (*Output, error) {
	o := &Output{
		sortFunc: DefaultSortFunc,
		pretty:   DefaultPretty,
	}

	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}

	var formatter Formatter
	switch format {
	case FormatJSON:
		formatter = NewJSONFormatter(o.pretty)
	default:
		formatter = NewTableFormatter()
	}

	o.formatter = formatter
	return o, nil
}

// PrintDevices prints devices to the writer
func (o *Output) PrintDevices(w io.Writer, results *discovery.ScanResults) error {
	sort.Slice(results.Devices, func(i, j int) bool {
		return o.sortFunc(results.Devices[i], results.Devices[j])
	})

	return o.formatter.Format(w, results)
}

// PrintDevices is a convenience function to print devices with the given format and options
func PrintDevices(w io.Writer, results *discovery.ScanResults, format Format, opts ...Option) error {
	o, err := NewOutput(format, opts...)
	if err != nil {
		return err
	}
	return o.PrintDevices(w, results)
}
