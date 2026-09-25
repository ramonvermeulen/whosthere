package views

import (
	"net"
	"testing"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/stretchr/testify/require"
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
	require.Contains(t, text, "Name: Living Room TV", "detail text missing detected name")
	require.Contains(t, text, "Alias: TV Upstairs", "detail text missing alias")
	require.NotContains(t, text, "Name: TV Upstairs", "detail text used alias as name")
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
	require.Contains(t, text, "Name: Sony", "detail text missing manufacturer fallback name")
	require.Contains(t, text, "Alias: Playstation", "detail text missing alias")
	require.NotContains(t, text, "Name: Playstation", "detail text used alias as name fallback")
}
