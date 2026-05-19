package state

import (
	"net"
	"testing"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

func TestNewAppState(t *testing.T) {
	cfg := config.DefaultConfig()
	version := "1.0.0"
	state := NewAppState(cfg, version)

	if state.version != version {
		t.Errorf("expected version %s, got %s", version, state.version)
	}
	if state.cfg != cfg {
		t.Errorf("expected config to be set")
	}
	if state.previousTheme != config.DefaultThemeName {
		t.Errorf("expected previous theme %s, got %s", config.DefaultThemeName, state.previousTheme)
	}
}

func TestUpsertDevice(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	ip := net.ParseIP("192.168.1.1")
	device := discovery.NewDevice(ip)
	device.SetDisplayName("test")

	state.UpsertDevice(device)

	devices := state.DevicesSnapshot()
	if len(devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(devices))
	}
	if devices[0].IP().String() != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1, got %s", devices[0].IP().String())
	}
}

func TestDevicesSnapshot(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	ip1 := net.ParseIP("192.168.1.2")
	ip2 := net.ParseIP("192.168.1.1")
	state.UpsertDevice(discovery.NewDevice(ip1))
	state.UpsertDevice(discovery.NewDevice(ip2))

	devices := state.DevicesSnapshot()
	if len(devices) != 2 {
		t.Errorf("expected 2 devices, got %d", len(devices))
	}
	// Should be sorted by IP
	if devices[0].IP().String() != "192.168.1.1" {
		t.Errorf("expected first IP 192.168.1.1, got %s", devices[0].IP().String())
	}
}

func TestDevicesSnapshotNumericSort(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	ips := []string{"192.168.1.1", "192.168.1.100", "192.168.1.2", "192.168.1.200"}
	for _, ip := range ips {
		state.UpsertDevice(discovery.NewDevice(net.ParseIP(ip)))
	}

	devices := state.DevicesSnapshot()
	if len(devices) != 4 {
		t.Errorf("expected 4 devices, got %d", len(devices))
	}

	expected := []string{"192.168.1.1", "192.168.1.2", "192.168.1.100", "192.168.1.200"}
	for i, exp := range expected {
		if devices[i].IP().String() != exp {
			t.Errorf("expected IP at index %d to be %s, got %s", i, exp, devices[i].IP().String())
		}
	}
}

func TestSelected(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	ip := net.ParseIP("192.168.1.1")
	device := discovery.NewDevice(ip)
	state.UpsertDevice(device)

	state.SetSelectedIP("192.168.1.1")
	selected, ok := state.Selected()
	if !ok {
		t.Errorf("expected selected device")
	}
	if selected.IP().String() != "192.168.1.1" {
		t.Errorf("expected selected IP 192.168.1.1, got %s", selected.IP().String())
	}

	state.SetSelectedIP("192.168.1.2")
	_, ok = state.Selected()
	if ok {
		t.Errorf("expected no selected device")
	}
}

func TestCurrentTheme(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	state.SetCurrentTheme("dark")
	if state.CurrentTheme() != "dark" {
		t.Errorf("expected theme dark, got %s", state.CurrentTheme())
	}
}

func TestVersion(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	state.SetVersion("2.0.0")
	if state.Version() != "2.0.0" {
		t.Errorf("expected version 2.0.0, got %s", state.Version())
	}
}

func TestFilterPattern(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	state.SetFilterPattern("test")
	if state.FilterPattern() != "test" {
		t.Errorf("expected filter test, got %s", state.FilterPattern())
	}
}

func TestIsDiscovering(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	state.SetIsDiscovering(true)
	if !state.IsDiscovering() {
		t.Errorf("expected discovering true")
	}
}

func TestIsPortscanning(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	state.SetIsPortscanning(true)
	if !state.IsPortscanning() {
		t.Errorf("expected portscanning true")
	}
}

func TestGetDevice(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	ip := net.ParseIP("192.168.1.1")
	device := discovery.NewDevice(ip)
	state.UpsertDevice(device)

	d, ok := state.GetDevice("192.168.1.1")
	if !ok {
		t.Errorf("expected device")
	}
	if d.IP().String() != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1")
	}
}

func TestSearch(t *testing.T) {
	state := NewAppState(config.DefaultConfig(), "1.0.0")

	state.SetSearchActive(true)
	if !state.SearchActive() {
		t.Errorf("expected search active")
	}

	state.SetSearchError(true)
	if !state.SearchError() {
		t.Errorf("expected search error")
	}

	state.SetFilterPattern("search")
	if state.SearchText() != "search" {
		t.Errorf("expected search text search, got %s", state.SearchText())
	}
}

func TestPreferredNameForAliasPrecedence(t *testing.T) {
	t.Parallel()

	appState := NewAppState(config.DefaultConfig(), "1.0.0")

	device := discovery.NewDevice(net.ParseIP("192.168.1.42"))
	device.SetMAC("AA:BB:CC:DD:EE:FF")
	device.SetDisplayName("Detected Device")
	device.SetManufacturer("Acme")

	if got := appState.PreferredNameFor(device); got != "Detected Device" {
		t.Fatalf("PreferredNameFor() without alias = %q, want %q", got, "Detected Device")
	}

	appState.UpsertDevice(device)
	appState.SetAliasForMAC("aa:bb:cc:dd:ee:ff", "Desk Speaker")

	if got := appState.AliasFor(device); got != "Desk Speaker" {
		t.Fatalf("AliasFor() = %q, want %q", got, "Desk Speaker")
	}
	if got := appState.PreferredNameFor(device); got != "Desk Speaker" {
		t.Fatalf("PreferredNameFor() with alias = %q, want %q", got, "Desk Speaker")
	}
}

func TestPreferredNameForFallsBackToManufacturerAndIP(t *testing.T) {
	t.Parallel()

	appState := NewAppState(config.DefaultConfig(), "1.0.0")

	device := discovery.NewDevice(net.ParseIP("192.168.1.99"))
	device.SetManufacturer("Vendor")
	if got := appState.PreferredNameFor(device); got != "Vendor" {
		t.Fatalf("PreferredNameFor() manufacturer fallback = %q, want %q", got, "Vendor")
	}

	device.SetManufacturer("")
	if got := appState.PreferredNameFor(device); got != "192.168.1.99" {
		t.Fatalf("PreferredNameFor() IP fallback = %q, want %q", got, "192.168.1.99")
	}
}

func TestSetAliasMarksAliasLoaded(t *testing.T) {
	t.Parallel()

	appState := NewAppState(config.DefaultConfig(), "1.0.0")
	const mac = "aa:bb:cc:dd:ee:ff"
	device := discovery.NewDevice(net.ParseIP("192.168.1.10"))
	device.SetMAC("AA:BB:CC:DD:EE:FF")
	appState.UpsertDevice(device)

	if appState.HasAliasMetadataForMAC(mac) {
		t.Fatal("HasAliasMetadataForMAC() = true before caching, want false")
	}

	appState.ClearAliasForMAC(mac)
	if !appState.HasAliasMetadataForMAC(mac) {
		t.Fatal("HasAliasMetadataForMAC() = false after ClearAliasForMAC, want true")
	}
}

func TestResetAliasesClearsCache(t *testing.T) {
	t.Parallel()

	appState := NewAppState(config.DefaultConfig(), "1.0.0")
	const mac = "aa:bb:cc:dd:ee:ff"

	device := discovery.NewDevice(net.ParseIP("192.168.1.10"))
	device.SetMAC("AA:BB:CC:DD:EE:FF")
	appState.UpsertDevice(device)
	appState.SetAliasForMAC(mac, "Laptop")

	if got := appState.AliasFor(device); got != "Laptop" {
		t.Fatalf("AliasFor() before reset = %q, want %q", got, "Laptop")
	}

	appState.ResetAliases()

	if got := appState.AliasFor(device); got != "" {
		t.Fatalf("AliasFor() after reset = %q, want empty", got)
	}
	if appState.HasAliasMetadataForMAC(mac) {
		t.Fatal("HasAliasMetadataForMAC() after reset = true, want false")
	}
}

func TestAliasEditorDraft(t *testing.T) {
	t.Parallel()

	appState := NewAppState(config.DefaultConfig(), "1.0.0")

	appState.SetAliasEditorDraft("Living Room TV")
	if got := appState.AliasEditorDraft(); got != "Living Room TV" {
		t.Fatalf("AliasEditorDraft() = %q, want %q", got, "Living Room TV")
	}

	appState.ClearAliasEditorDraft()
	if got := appState.AliasEditorDraft(); got != "" {
		t.Fatalf("AliasEditorDraft() after clear = %q, want empty", got)
	}
}

func TestStatusMessageLifecycle(t *testing.T) {
	t.Parallel()

	appState := NewAppState(config.DefaultConfig(), "1.0.0")

	appState.SetStatusMessage("alias saved", StatusSeveritySuccess, 50*time.Millisecond)
	if got := appState.StatusMessage(); got != "alias saved" {
		t.Fatalf("StatusMessage() = %q, want %q", got, "alias saved")
	}
	if got := appState.StatusSeverity(); got != StatusSeveritySuccess {
		t.Fatalf("StatusSeverity() = %q, want %q", got, StatusSeveritySuccess)
	}

	time.Sleep(75 * time.Millisecond)
	if got := appState.StatusMessage(); got != "" {
		t.Fatalf("StatusMessage() after expiry = %q, want empty", got)
	}
	if got := appState.StatusSeverity(); got != StatusSeverityInfo {
		t.Fatalf("StatusSeverity() after expiry = %q, want %q", got, StatusSeverityInfo)
	}
}
