package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

var _ UIComponent = &FilterBar{}

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

// Render implements UIComponent.
func (f *FilterBar) Render(s state.ReadOnly) {
	if s.SearchActive() {
		color := tview.Styles.PrimaryTextColor
		if s.SearchError() {
			color = tcell.ColorRed
		}
		f.SetTextColor(color)
		f.SetText("Regex search: /" + s.SearchText())
	} else {
		f.SetText("")
	}
}
