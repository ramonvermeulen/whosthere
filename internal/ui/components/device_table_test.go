package components

import (
	"net"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/rivo/tview"
)

func TestDeviceTableRenderUsesPreferredName(t *testing.T) {
	t.Parallel()

	appState := state.NewAppState(config.DefaultConfig(), "1.0.0")
	device := discovery.NewDevice(net.ParseIP("192.168.1.10"))
	device.SetMAC("AA:BB:CC:DD:EE:FF")
	device.SetDisplayName("Detected Host")
	appState.UpsertDevice(device)
	appState.SetAliasForMAC("aa:bb:cc:dd:ee:ff", "Kitchen Tablet")

	table := NewDeviceTable(nil)
	table.Render(appState.ReadOnly())

	if got := table.GetCell(1, 1).Text; got != "Kitchen Tablet" {
		t.Fatalf("table hostname cell = %q, want %q", got, "Kitchen Tablet")
	}
}

func TestDeviceTableRenderUsesAliasNameHeader(t *testing.T) {
	t.Parallel()

	table := NewDeviceTable(nil)
	table.Render(state.NewAppState(config.DefaultConfig(), "1.0.0").ReadOnly())

	if got := table.GetCell(0, 1).Text; got != "Alias/Name" {
		t.Fatalf("table header cell = %q, want %q", got, "Alias/Name")
	}
}

func TestDeviceTableFilterMatchesAliasAndDetectedName(t *testing.T) {
	t.Parallel()

	appState := state.NewAppState(config.DefaultConfig(), "1.0.0")
	device := discovery.NewDevice(net.ParseIP("192.168.1.11"))
	device.SetMAC("AA:BB:CC:DD:EE:11")
	device.SetDisplayName("Detected Printer")
	appState.UpsertDevice(device)
	appState.SetAliasForMAC("aa:bb:cc:dd:ee:11", "Office Printer")

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

func TestLastSeenColorUsesFreshnessBuckets(t *testing.T) {
	t.Parallel()

	if got := lastSeenColor(20*time.Second, false); got != tview.Styles.ContrastSecondaryTextColor {
		t.Fatalf("fresh last-seen color = %v, want %v", got, tview.Styles.ContrastSecondaryTextColor)
	}

	if got := lastSeenColor(2*time.Minute, false); got != tview.Styles.TertiaryTextColor {
		t.Fatalf("normal last-seen color = %v, want %v", got, tview.Styles.TertiaryTextColor)
	}

	if got := lastSeenColor(10*time.Minute, false); got != tview.Styles.TertiaryTextColor {
		t.Fatalf("stale last-seen color = %v, want %v", got, tview.Styles.TertiaryTextColor)
	}

	if got := lastSeenColor(20*time.Second, true); got != tview.Styles.PrimaryTextColor {
		t.Fatalf("no-color last-seen color = %v, want %v", got, tview.Styles.PrimaryTextColor)
	}
}

func TestDeviceTableSelectedRowUsesThemeAccentStyle(t *testing.T) {
	device := discovery.NewDevice(net.ParseIP("192.168.1.10"))

	table := NewDeviceTable(nil)
	table.devices = []*discovery.Device{device}
	table.applyThemeStyles()
	table.refresh()
	table.SetRect(0, 0, 80, 10)

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("init screen: %v", err)
	}
	defer screen.Fini()

	table.Draw(screen)

	x, y, _ := table.GetCell(1, 0).GetLastPosition()
	_, style, _ := screen.Get(x, y)
	foreground, background, attrs := style.Decompose()

	if foreground != tview.Styles.InverseTextColor {
		t.Fatalf("selected row foreground = %v, want %v", foreground, tview.Styles.InverseTextColor)
	}
	if background != tview.Styles.SecondaryTextColor {
		t.Fatalf("selected row background = %v, want %v", background, tview.Styles.SecondaryTextColor)
	}
	if attrs&tcell.AttrBold == 0 {
		t.Fatalf("selected row attrs = %v, want bold", attrs)
	}
}
