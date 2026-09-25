package output

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	output := buf.String()

	assert.Contains(t, output, "IP", "expected header to contain IP")
	assert.Contains(t, output, "192.168.1.1", "expected output to contain device IP")
	assert.Contains(t, output, "Router", "expected output to contain device name")
	assert.Contains(t, output, "AA:BB:CC:DD:EE:FF", "expected output to contain MAC address")
	assert.Contains(t, output, "Cisco", "expected output to contain manufacturer")
	assert.Contains(t, output, "2 device(s) found", "expected output to contain device count")
	assert.Contains(t, output, "1.5s", "expected output to contain elapsed time")
	assert.Contains(t, output, "-", "expected empty fields to show '-'")
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
	require.NoError(t, err)

	output := buf.String()

	assert.Contains(t, output, "0 device(s) found", "expected output to show 0 devices")
	assert.Contains(t, output, "0.1s", "expected output to contain elapsed time")
}
