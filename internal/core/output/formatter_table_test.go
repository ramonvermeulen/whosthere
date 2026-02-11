package output

import (
	"bytes"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

func TestPrintDevices(t *testing.T) {
	devices := []*discovery.Device{
		func() *discovery.Device {
			d := discovery.NewDevice(net.ParseIP("192.168.1.1"))
			d.SetDisplayName("Router")
			d.SetMAC("AA:BB:CC:DD:EE:FF")
			d.SetManufacturer("Cisco")
			return d
		}(),
		discovery.NewDevice(net.ParseIP("192.168.1.100")),
	}

	results := &discovery.ScanResults{
		Devices: devices,
		Stats: &discovery.ScanStats{
			Count:    len(devices),
			Duration: 1500 * time.Millisecond,
		},
	}

	var buf bytes.Buffer
	err := PrintDevices(&buf, results, FormatTable)
	if err != nil {
		t.Fatalf("PrintDevices failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "IP") {
		t.Error("expected header to contain IP")
	}
	if !strings.Contains(output, "192.168.1.1") {
		t.Error("expected output to contain device IP")
	}
	if !strings.Contains(output, "Router") {
		t.Error("expected output to contain device name")
	}
	if !strings.Contains(output, "AA:BB:CC:DD:EE:FF") {
		t.Error("expected output to contain MAC address")
	}
	if !strings.Contains(output, "Cisco") {
		t.Error("expected output to contain manufacturer")
	}
	if !strings.Contains(output, "2 device(s) found") {
		t.Error("expected output to contain device count")
	}
	if !strings.Contains(output, "1.5s") {
		t.Error("expected output to contain elapsed time")
	}
	if !strings.Contains(output, "-") {
		t.Error("expected empty fields to show '-'")
	}
}

func TestPrintDevices_Empty(t *testing.T) {
	results := &discovery.ScanResults{
		Devices: []*discovery.Device{},
		Stats: &discovery.ScanStats{
			Count:    0,
			Duration: 100 * time.Millisecond,
		},
	}

	var buf bytes.Buffer
	err := PrintDevices(&buf, results, FormatTable)
	if err != nil {
		t.Fatalf("PrintDevices failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "0 device(s) found") {
		t.Error("expected output to show 0 devices")
	}
	if !strings.Contains(output, "0.1s") {
		t.Error("expected output to contain elapsed time")
	}
}
