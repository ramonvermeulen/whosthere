package views

import (
	"net"
	"strings"
	"testing"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

func TestDetailViewRenderShowsAliasAndDetectedNameSeparately(t *testing.T) {
	t.Parallel()

	appState := state.NewAppState(config.DefaultConfig(), "1.0.0")
	device := discovery.NewDevice(net.ParseIP("192.168.1.10"))
	device.SetMAC("AA:BB:CC:DD:EE:FF")
	device.SetDisplayName("Living Room TV")
	appState.UpsertDevice(device)
	appState.SetAliasForMAC("aa:bb:cc:dd:ee:ff", "TV Upstairs")
	appState.SetSelectedIP("192.168.1.10")

	view := NewDetailView(nil, func(func()) {})
	view.Render(appState.ReadOnly())

	text := view.info.GetText(true)
	if !strings.Contains(text, "Name: Living Room TV") {
		t.Fatalf("detail text missing detected name, got %q", text)
	}
	if !strings.Contains(text, "Alias: TV Upstairs") {
		t.Fatalf("detail text missing alias, got %q", text)
	}
	if strings.Contains(text, "Name: TV Upstairs") {
		t.Fatalf("detail text used alias as name, got %q", text)
	}
}

func TestDetailViewRenderDoesNotUseAliasAsNameFallback(t *testing.T) {
	t.Parallel()

	appState := state.NewAppState(config.DefaultConfig(), "1.0.0")
	device := discovery.NewDevice(net.ParseIP("192.168.1.11"))
	device.SetMAC("AA:BB:CC:DD:EE:11")
	device.SetManufacturer("Sony")
	appState.UpsertDevice(device)
	appState.SetAliasForMAC("aa:bb:cc:dd:ee:11", "Playstation")
	appState.SetSelectedIP("192.168.1.11")

	view := NewDetailView(nil, func(func()) {})
	view.Render(appState.ReadOnly())

	text := view.info.GetText(true)
	if !strings.Contains(text, "Name: Sony") {
		t.Fatalf("detail text missing manufacturer fallback name, got %q", text)
	}
	if !strings.Contains(text, "Alias: Playstation") {
		t.Fatalf("detail text missing alias, got %q", text)
	}
	if strings.Contains(text, "Name: Playstation") {
		t.Fatalf("detail text used alias as name fallback, got %q", text)
	}
}
