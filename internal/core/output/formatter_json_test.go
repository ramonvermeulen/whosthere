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
	require.NoError(t, err)

	output := buf.String()

	assert.NotContains(t, output, "  ", "expected minified JSON, but contains extra spaces")
	assert.Contains(t, output, "\n", "expected trailing newline")
	assert.Contains(t, output, `"count":1`, "expected to contain count")
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
	require.NoError(t, err)

	output := buf.String()

	assert.Contains(t, output, "\n", "expected pretty JSON with newlines")
	assert.Contains(t, output, "  ", "expected pretty JSON with indentation")
	assert.Contains(t, output, `"count": 1`, "expected to contain count with space")
}
