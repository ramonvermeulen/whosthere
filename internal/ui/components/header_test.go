package components

import (
	"strings"
	"testing"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
)

func TestHeaderRenderIncludesVersionInterfaceAndAskQuestion(t *testing.T) {
	cfg := config.DefaultConfig()
	st := state.NewAppState(cfg, "1.2.3")
	st.SetCurrentInterface("en0")

	header := NewHeader(nil)
	header.Render(st.ReadOnly())

	if got := header.title.GetText(false); got != "whosthere - v1.2.3" {
		t.Fatalf("unexpected header title: %q", got)
	}

	interfaceLabel := header.interfaceLabel.GetText(false)
	if !strings.Contains(interfaceLabel, "interface: en0") {
		t.Fatalf("expected interface in header label view, got %q", interfaceLabel)
	}

	link := header.link.GetText(false)
	if !strings.Contains(link, "Ask Question") {
		t.Fatalf("expected ask question label in header link view, got %q", link)
	}
}

func TestRenderHeaderMetaNoColor(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Theme.NoColor = true

	st := state.NewAppState(cfg, "1.2.3")
	st.SetCurrentInterface("very-long-interface-name")

	interfaceLabel, link := renderHeaderMeta(st.ReadOnly())
	if !strings.Contains(interfaceLabel, "interface:") || !strings.Contains(interfaceLabel, Divider) {
		t.Fatalf("expected plain interface label, got %q", interfaceLabel)
	}
	if !strings.Contains(link, "Ask Question") {
		t.Fatalf("expected plain ask question label, got %q", link)
	}
	if strings.Contains(interfaceLabel, "#") || strings.Contains(link, "#") {
		t.Fatalf("expected plain text without color tags, got interface=%q link=%q", interfaceLabel, link)
	}
}
