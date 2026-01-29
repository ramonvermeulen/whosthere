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

	var buf bytes.Buffer
	PrintDevices(&buf, devices, 1500*time.Millisecond)

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
	var buf bytes.Buffer
	PrintDevices(&buf, []*discovery.Device{}, 100*time.Millisecond)

	output := buf.String()

	if !strings.Contains(output, "0 device(s) found") {
		t.Error("expected output to show 0 devices")
	}
	if !strings.Contains(output, "0.1s") {
		t.Error("expected output to contain elapsed time")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{500 * time.Millisecond, "0.5s"},
		{1500 * time.Millisecond, "1.5s"},
		{3 * time.Second, "3.0s"},
	}

	for _, tc := range tests {
		result := formatDuration(tc.input)
		if result != tc.expected {
			t.Errorf("formatDuration(%v) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}
