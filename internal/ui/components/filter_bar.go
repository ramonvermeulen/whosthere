package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

// FilterBar wraps a TextView used to display live search/filter status in the footer.
type FilterBar struct {
	*tview.TextView
}

func NewFilterBar() *FilterBar {
	fv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	theme.RegisterPrimitive(fv)
	return &FilterBar{TextView: fv}
}

// Show updates the filter bar text and color.
func (f *FilterBar) Show(text string, color tcell.Color) {
	if f == nil {
		return
	}
	f.SetTextColor(color)
	f.SetText(text)
}

// Clear removes any text from the filter bar.
func (f *FilterBar) Clear() {
	if f == nil {
		return
	}
	f.SetText("")
}
