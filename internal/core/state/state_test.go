package state

import (
	"net"
	"testing"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppState(t *testing.T) {
	cfg := config.DefaultConfig()
	version := "1.0.0"
	state := NewAppState(cfg, version)

	assert.Equal(t, version, state.version, "expected version %s", version)
	assert.Same(t, cfg, state.cfg, "expected config to be set")
	assert.Equal(t, config.DefaultThemeName, state.previousTheme, "expected previous theme %s", config.DefaultThemeName)
}

func TestUpsertDevice(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	ip := net.ParseIP("192.168.1.1")
	device := discovery.NewDevice(ip)
	device.SetDisplayName("test")

	state.UpsertDevice(device)

	devices := state.DevicesSnapshot()
	assert.Equal(t, 1, len(devices), "expected 1 device")
	assert.Equal(t, "192.168.1.1", devices[0].IP().String(), "expected IP 192.168.1.1")
}

func TestDevicesSnapshot(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	ip1 := net.ParseIP("192.168.1.2")
	ip2 := net.ParseIP("192.168.1.1")
	state.UpsertDevice(discovery.NewDevice(ip1))
	state.UpsertDevice(discovery.NewDevice(ip2))

	devices := state.DevicesSnapshot()
	assert.Equal(t, 2, len(devices), "expected 2 devices")
	// Should be sorted by IP
	assert.Equal(t, "192.168.1.1", devices[0].IP().String(), "expected first IP 192.168.1.1")
}

func TestDevicesSnapshotNumericSort(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	ips := []string{"192.168.1.1", "192.168.1.100", "192.168.1.2", "192.168.1.200"}
	for _, ip := range ips {
		state.UpsertDevice(discovery.NewDevice(net.ParseIP(ip)))
	}

	devices := state.DevicesSnapshot()
	assert.Equal(t, 4, len(devices), "expected 4 devices")

	expected := []string{"192.168.1.1", "192.168.1.2", "192.168.1.100", "192.168.1.200"}
	for i, exp := range expected {
		assert.Equal(t, exp, devices[i].IP().String(), "expected IP at index %d", i)
	}
}

func TestSelected(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	ip := net.ParseIP("192.168.1.1")
	device := discovery.NewDevice(ip)
	state.UpsertDevice(device)

	state.SetSelectedIP("192.168.1.1")
	selected, ok := state.Selected()
	assert.True(t, ok, "expected selected device")
	assert.Equal(t, "192.168.1.1", selected.IP().String(), "expected selected IP 192.168.1.1")

	state.SetSelectedIP("192.168.1.2")
	_, ok = state.Selected()
	assert.False(t, ok, "expected no selected device")
}

func TestCurrentTheme(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	state.SetCurrentTheme("dark")
	assert.Equal(t, "dark", state.CurrentTheme(), "expected theme dark")
}

func TestVersion(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	state.SetVersion("2.0.0")
	assert.Equal(t, "2.0.0", state.Version(), "expected version 2.0.0")
}

func TestFilterPattern(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	state.SetFilterPattern("test")
	assert.Equal(t, "test", state.FilterPattern(), "expected filter test")
}

func TestIsDiscovering(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	state.SetIsDiscovering(true)
	assert.True(t, state.IsDiscovering(), "expected discovering true")
}

func TestIsPortscanning(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	state.SetIsPortscanning(true)
	assert.True(t, state.IsPortscanning(), "expected portscanning true")
}

func TestGetDevice(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	ip := net.ParseIP("192.168.1.1")
	device := discovery.NewDevice(ip)
	state.UpsertDevice(device)

	d, ok := state.GetDevice("192.168.1.1")
	assert.True(t, ok, "expected device")
	assert.Equal(t, "192.168.1.1", d.IP().String(), "expected IP 192.168.1.1")
}

func TestSearch(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	state.SetSearchActive(true)
	assert.True(t, state.SearchActive(), "expected search active")

	state.SetSearchError(true)
	assert.True(t, state.SearchError(), "expected search error")

	state.SetFilterPattern("search")
	assert.Equal(t, "search", state.SearchText(), "expected search text search")
}

func TestAliasOrDetectedNameForAliasPrecedence(t *testing.T) {
	t.Parallel()

	appState := NewAppState(config.DefaultConfig(), "1.0.0")

	device := discovery.NewDevice(net.ParseIP("192.168.1.42"))
	device.SetMAC("AA:BB:CC:DD:EE:FF")
	device.SetDisplayName("Detected Device")
	device.SetManufacturer("Acme")

	require.Equal(t, "Detected Device", appState.AliasOrDetectedNameFor(device), "AliasOrDetectedNameFor() without alias")

	appState.UpsertDevice(device)
	appState.SetAliasForMAC("aa:bb:cc:dd:ee:ff", "Desk Speaker")

	require.Equal(t, "Desk Speaker", appState.AliasFor(device), "AliasFor()")
	require.Equal(t, "Desk Speaker", appState.AliasOrDetectedNameFor(device), "AliasOrDetectedNameFor() with alias")
}

func TestAliasOrDetectedNameForFallsBackToManufacturerAndIP(t *testing.T) {
	t.Parallel()

	appState := NewAppState(config.DefaultConfig(), "1.0.0")

	device := discovery.NewDevice(net.ParseIP("192.168.1.99"))
	device.SetManufacturer("Vendor")
	require.Equal(t, "Vendor", appState.AliasOrDetectedNameFor(device), "AliasOrDetectedNameFor() manufacturer fallback")

	device.SetManufacturer("")
	require.Equal(t, "192.168.1.99", appState.AliasOrDetectedNameFor(device), "AliasOrDetectedNameFor() IP fallback")
}

func TestDetectedNameForIgnoresAliasAndFallsBackToManufacturerAndIP(t *testing.T) {
	t.Parallel()

	t.Run("manufacturer fallback", func(t *testing.T) {
		t.Parallel()

		appState := NewAppState(config.DefaultConfig(), "1.0.0")
		device := discovery.NewDevice(net.ParseIP("192.168.1.99"))
		device.SetMAC("AA:BB:CC:DD:EE:99")
		device.SetManufacturer("Vendor")
		appState.UpsertDevice(device)
		appState.SetAliasForMAC("aa:bb:cc:dd:ee:99", "Desk Speaker")

		require.Equal(t, "Vendor", appState.DetectedNameFor(device), "DetectedNameFor() manufacturer fallback")
	})

	t.Run("ip fallback", func(t *testing.T) {
		t.Parallel()

		appState := NewAppState(config.DefaultConfig(), "1.0.0")
		device := discovery.NewDevice(net.ParseIP("192.168.1.99"))
		device.SetMAC("AA:BB:CC:DD:EE:99")
		appState.UpsertDevice(device)
		appState.SetAliasForMAC("aa:bb:cc:dd:ee:99", "Desk Speaker")

		require.Equal(t, "192.168.1.99", appState.DetectedNameFor(device), "DetectedNameFor() IP fallback")
	})

	t.Run("detected display name", func(t *testing.T) {
		t.Parallel()

		appState := NewAppState(config.DefaultConfig(), "1.0.0")
		device := discovery.NewDevice(net.ParseIP("192.168.1.99"))
		device.SetMAC("AA:BB:CC:DD:EE:99")
		device.SetDisplayName("Detected Device")
		appState.UpsertDevice(device)
		appState.SetAliasForMAC("aa:bb:cc:dd:ee:99", "Desk Speaker")

		require.Equal(t, "Detected Device", appState.DetectedNameFor(device), "DetectedNameFor() detected name")
	})
}

func TestSetAliasMarksAliasLoaded(t *testing.T) {
	t.Parallel()

	appState := NewAppState(config.DefaultConfig(), "1.0.0")
	const mac = "aa:bb:cc:dd:ee:ff"
	device := discovery.NewDevice(net.ParseIP("192.168.1.10"))
	device.SetMAC("AA:BB:CC:DD:EE:FF")
	appState.UpsertDevice(device)

	require.False(t, appState.HasAliasMetadataForMAC(mac), "HasAliasMetadataForMAC() = true before caching, want false")

	appState.ClearAliasForMAC(mac)
	require.True(t, appState.HasAliasMetadataForMAC(mac), "HasAliasMetadataForMAC() = false after ClearAliasForMAC, want true")
}

func TestResetAliasesClearsCache(t *testing.T) {
	t.Parallel()

	appState := NewAppState(config.DefaultConfig(), "1.0.0")
	const mac = "aa:bb:cc:dd:ee:ff"

	device := discovery.NewDevice(net.ParseIP("192.168.1.10"))
	device.SetMAC("AA:BB:CC:DD:EE:FF")
	appState.UpsertDevice(device)
	appState.SetAliasForMAC(mac, "Laptop")

	require.Equal(t, "Laptop", appState.AliasFor(device), "AliasFor() before reset")

	appState.ResetAliases()

	require.Empty(t, appState.AliasFor(device), "AliasFor() after reset")
	require.False(t, appState.HasAliasMetadataForMAC(mac), "HasAliasMetadataForMAC() after reset = true, want false")
}

func TestAliasEditorDraft(t *testing.T) {
	t.Parallel()

	appState := NewAppState(config.DefaultConfig(), "1.0.0")

	appState.SetAliasEditorDraft("Living Room TV")
	require.Equal(t, "Living Room TV", appState.AliasEditorDraft(), "AliasEditorDraft()")

	appState.ClearAliasEditorDraft()
	require.Empty(t, appState.AliasEditorDraft(), "AliasEditorDraft() after clear")
}

func TestStatusMessageLifecycle(t *testing.T) {
	t.Parallel()

	appState := NewAppState(config.DefaultConfig(), "1.0.0")

	appState.SetStatusMessage("alias saved", StatusSeveritySuccess, 50*time.Millisecond)
	require.Equal(t, "alias saved", appState.StatusMessage(), "StatusMessage()")
	require.Equal(t, StatusSeveritySuccess, appState.StatusSeverity(), "StatusSeverity()")

	time.Sleep(75 * time.Millisecond)
	require.Empty(t, appState.StatusMessage(), "StatusMessage() after expiry")
	require.Equal(t, StatusSeverityInfo, appState.StatusSeverity(), "StatusSeverity() after expiry")
}
