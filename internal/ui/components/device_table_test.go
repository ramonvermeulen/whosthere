package components

import (
	"net"
	"testing"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

func TestDeviceTableRenderUsesPreferredName(t *testing.T) {
	t.Parallel()

	appState := state.NewAppState(config.DefaultConfig(), "1.0.0")
	device := discovery.NewDevice(net.ParseIP("192.168.1.10"))
	device.SetMAC("AA:BB:CC:DD:EE:FF")
	device.SetDisplayName("Detected Host")
	appState.UpsertDevice(device)
	appState.SetAlias("aa:bb:cc:dd:ee:ff", "Kitchen Tablet")

	table := NewDeviceTable(nil)
	table.Render(appState.ReadOnly())

	if got := table.GetCell(1, 1).Text; got != "Kitchen Tablet" {
		t.Fatalf("table hostname cell = %q, want %q", got, "Kitchen Tablet")
	}
}

func TestDeviceTableFilterMatchesAliasAndDetectedName(t *testing.T) {
	t.Parallel()

	appState := state.NewAppState(config.DefaultConfig(), "1.0.0")
	device := discovery.NewDevice(net.ParseIP("192.168.1.11"))
	device.SetMAC("AA:BB:CC:DD:EE:11")
	device.SetDisplayName("Detected Printer")
	appState.UpsertDevice(device)
	appState.SetAlias("aa:bb:cc:dd:ee:11", "Office Printer")

	table := NewDeviceTable(nil)
	table.Render(appState.ReadOnly())

	if err := table.SetFilter("Office"); err != nil {
		t.Fatalf("SetFilter(alias) error = %v", err)
	}
	if table.GetRowCount() != 2 {
		t.Fatalf("row count after alias filter = %d, want 2", table.GetRowCount())
	}

	if err := table.SetFilter("Detected"); err != nil {
		t.Fatalf("SetFilter(detected name) error = %v", err)
	}
	if table.GetRowCount() != 2 {
		t.Fatalf("row count after detected-name filter = %d, want 2", table.GetRowCount())
	}
}
