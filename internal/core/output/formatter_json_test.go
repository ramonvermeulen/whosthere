package output

import (
	"bytes"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

func TestPrintDevices_JSON(t *testing.T) {
	devices := []*discovery.Device{
		discovery.NewDevice(net.ParseIP("192.168.1.1")),
	}

	results := &discovery.ScanResults{
		Devices: devices,
		Stats: &discovery.ScanStats{
			Count:    len(devices),
			Duration: 1000 * time.Millisecond,
		},
	}

	var buf bytes.Buffer
	err := PrintDevices(&buf, results, FormatJSON)
	if err != nil {
		t.Fatalf("PrintDevices failed: %v", err)
	}

	output := buf.String()

	if strings.Contains(output, "  ") {
		t.Error("expected minified JSON, but contains extra spaces")
	}
	if !strings.Contains(output, "\n") {
		t.Error("expected trailing newline")
	}
	if !strings.Contains(output, `"count":1`) {
		t.Error("expected to contain count")
	}
}

func TestPrintDevices_JSON_Pretty(t *testing.T) {
	devices := []*discovery.Device{
		discovery.NewDevice(net.ParseIP("192.168.1.1")),
	}

	results := &discovery.ScanResults{
		Devices: devices,
		Stats: &discovery.ScanStats{
			Count:    len(devices),
			Duration: 1000 * time.Millisecond,
		},
	}

	var buf bytes.Buffer
	err := PrintDevices(&buf, results, FormatJSON, WithPretty())
	if err != nil {
		t.Fatalf("PrintDevices failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "\n") || !strings.Contains(output, "  ") {
		t.Error("expected pretty JSON with newlines and indentation")
	}
	if !strings.Contains(output, `"count": 1`) {
		t.Error("expected to contain count with space")
	}
}
