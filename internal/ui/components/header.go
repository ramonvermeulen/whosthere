package components

import (
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

var _ UIComponent = &Header{}

// Header is a simple reusable header bar for pages.
// It renders the app title and, optionally, a version string.
type Header struct {
	*tview.TextView
}

// NewHeader creates a header with a fixed base title.
func NewHeader() *Header {
	const baseTitle = "whosthere"
	tv := tview.NewTextView().
		SetText(baseTitle).
		SetTextAlign(tview.AlignCenter)

	theme.RegisterPrimitive(tv)
	return &Header{TextView: tv}
}

// Render implements UIComponent.
func (h *Header) Render(s state.ReadOnly) {
	const baseTitle = "whosthere"
	text := baseTitle
	if version := s.Version(); version != "" {
		text = baseTitle + " - " + version
	}
	h.SetText(text)
}
