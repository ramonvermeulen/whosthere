package components

import (
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

// Header is a simple reusable header bar for pages.
// It renders the app title and, optionally, a version string.
type Header struct {
	*tview.TextView
}

// NewHeader creates a header with a fixed base title and optional version.
// If version is non-empty, the header text will be "whosthere - v<version>".
func NewHeader(version string) *Header {
	const baseTitle = "whosthere"
	text := baseTitle
	if version != "" {
		text = baseTitle + " - v" + version
	}
	tv := tview.NewTextView().
		SetText(text).
		SetTextAlign(tview.AlignCenter)

	theme.RegisterPrimitive(tv)
	return &Header{TextView: tv}
}
