package components

import (
	"testing"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/stretchr/testify/require"
)

func TestHeaderRenderIncludesVersionInterfaceAndAskQuestion(t *testing.T) {
	cfg := config.DefaultConfig()
	st := state.NewAppState(cfg, "1.2.3")
	st.SetCurrentInterface("en0")

	header := NewHeader(nil)
	header.Render(st.ReadOnly())

	require.Equal(t, "whosthere - v1.2.3", header.title.GetText(false), "unexpected header title")

	interfaceLabel := header.interfaceLabel.GetText(false)
	require.Contains(t, interfaceLabel, "interface: en0", "expected interface in header label view, got %q", interfaceLabel)

	link := header.link.GetText(false)
	require.Contains(t, link, "Ask Question", "expected ask question label in header link view, got %q", link)
}

func TestRenderHeaderMetaNoColor(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Theme.NoColor = true

	st := state.NewAppState(cfg, "1.2.3")
	st.SetCurrentInterface("very-long-interface-name")

	interfaceLabel, link := renderHeaderMeta(st.ReadOnly())
	require.Contains(t, interfaceLabel, "interface:", "expected plain interface label, got %q", interfaceLabel)
	require.Contains(t, interfaceLabel, Divider, "expected plain interface label with divider, got %q", interfaceLabel)
	require.Contains(t, link, "Ask Question", "expected plain ask question label, got %q", link)
	require.NotContains(t, interfaceLabel, "#", "expected plain text without color tags, got interface=%q", interfaceLabel)
	require.NotContains(t, link, "#", "expected plain text without color tags, got link=%q", link)
}
