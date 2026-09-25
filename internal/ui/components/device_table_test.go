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
	"github.com/stretchr/testify/require"
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

	require.Equal(t, "Kitchen Tablet", table.GetCell(1, 1).Text, "table hostname cell")
}

func TestDeviceTableRenderUsesAliasNameHeader(t *testing.T) {
	t.Parallel()

	table := NewDeviceTable(nil)
	table.Render(state.NewAppState(config.DefaultConfig(), "1.0.0").ReadOnly())

	require.Equal(t, "Alias/Name", table.GetCell(0, 1).Text, "table header cell")
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

	require.NoError(t, table.SetFilter("Office"), "SetFilter(alias)")
	require.Equal(t, 2, table.GetRowCount(), "row count after alias filter")

	require.NoError(t, table.SetFilter("Detected"), "SetFilter(detected name)")
	require.Equal(t, 2, table.GetRowCount(), "row count after detected-name filter")
}

func TestLastSeenColorUsesFreshnessBuckets(t *testing.T) {
	t.Parallel()

	require.Equal(t, tview.Styles.ContrastSecondaryTextColor, lastSeenColor(20*time.Second, false), "fresh last-seen color")

	require.Equal(t, tview.Styles.TertiaryTextColor, lastSeenColor(2*time.Minute, false), "normal last-seen color")

	require.Equal(t, tview.Styles.TertiaryTextColor, lastSeenColor(10*time.Minute, false), "stale last-seen color")

	require.Equal(t, tview.Styles.PrimaryTextColor, lastSeenColor(20*time.Second, true), "no-color last-seen color")
}

func TestDeviceTableSelectedRowUsesThemeAccentStyle(t *testing.T) {
	device := discovery.NewDevice(net.ParseIP("192.168.1.10"))

	table := NewDeviceTable(nil)
	table.devices = []*discovery.Device{device}
	table.applyThemeStyles()
	table.refresh()
	table.SetRect(0, 0, 80, 10)

	screen := tcell.NewSimulationScreen("UTF-8")
	require.NoError(t, screen.Init(), "init screen")
	defer screen.Fini()

	table.Draw(screen)

	x, y, _ := table.GetCell(1, 0).GetLastPosition()
	_, style, _ := screen.Get(x, y)
	foreground, background, attrs := style.Decompose()

	require.Equal(t, tview.Styles.InverseTextColor, foreground, "selected row foreground")
	require.Equal(t, tview.Styles.SecondaryTextColor, background, "selected row background")
	require.NotEqual(t, 0, attrs&tcell.AttrBold, "selected row attrs, want bold")
}
