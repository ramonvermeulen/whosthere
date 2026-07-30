package components

import (
	"strings"
	"testing"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
)

func TestHeaderRenderIncludesVersionInterfaceAndAskQuestion(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Mode = config.ModePersistent
	st := state.NewAppState(cfg, "1.2.3")
	st.SetCurrentInterface("en0")

	header := NewHeader(nil)
	header.Render(st.ReadOnly())

	if got := header.title.GetText(false); got != "whosthere - v1.2.3" {
		t.Fatalf("unexpected header title: %q", got)
	}

	metaLabel := header.metaLabel.GetText(false)
	if !strings.Contains(metaLabel, "mode: persistent") {
		t.Fatalf("expected mode in header label view, got %q", metaLabel)
	}
	if !strings.Contains(metaLabel, "interface: en0") {
		t.Fatalf("expected interface in header label view, got %q", metaLabel)
	}

	actionLink := header.actionLink.GetText(false)
	if !strings.Contains(actionLink, "Ask Question") {
		t.Fatalf("expected ask question label in header link view, got %q", actionLink)
	}
}

func TestRenderHeaderMetaNoColor(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Theme.NoColor = true
	cfg.Mode = config.ModeSession

	st := state.NewAppState(cfg, "1.2.3")
	st.SetCurrentInterface("very-long-interface-name")

	metaLabel, actionLink := renderHeaderMeta(st.ReadOnly())
	if !strings.Contains(metaLabel, "mode: session") || !strings.Contains(metaLabel, "interface:") || !strings.Contains(metaLabel, Divider) {
		t.Fatalf("expected plain interface label, got %q", metaLabel)
	}
	if !strings.Contains(actionLink, "Ask Question") {
		t.Fatalf("expected plain ask question label, got %q", actionLink)
	}
	if strings.Contains(metaLabel, "#") || strings.Contains(actionLink, "#") {
		t.Fatalf("expected plain text without color tags, got meta=%q link=%q", metaLabel, actionLink)
	}
}
