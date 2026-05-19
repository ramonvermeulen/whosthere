package views

import (
	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/events"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

var _ View = &ResetAliasesModalView{}

// ResetAliasesModalView confirms clearing all persisted aliases.
type ResetAliasesModalView struct {
	*tview.Modal
}

func NewResetAliasesModalView(emit func(events.Event)) *ResetAliasesModalView {
	modal := tview.NewModal().
		SetText("").
		AddButtons([]string{"Reset Aliases", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, _ string) {
			if buttonIndex == 0 {
				emit(events.AliasesResetConfirmed{})
				return
			}
			emit(events.HideView{})
		})

	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			emit(events.HideView{})
			return nil
		}
		return event
	})

	theme.RegisterPrimitive(modal)

	return &ResetAliasesModalView{Modal: modal}
}

func (v *ResetAliasesModalView) FocusTarget() tview.Primitive { return v.Modal }

func (v *ResetAliasesModalView) Render(state.ReadOnly) {
	v.SetTitle(" Reset Aliases ")
	v.SetText("Delete all saved device aliases from local storage?\n\nDetected names from scanners will remain visible.\nThis action cannot be undone.")
}
