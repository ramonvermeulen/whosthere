package views

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/components"
	"github.com/ramonvermeulen/whosthere/internal/ui/events"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

var _ View = &AliasModalView{}

// AliasModalView is a modal overlay for editing the selected device alias.
type AliasModalView struct {
	*tview.Flex
	input  *tview.InputField
	footer *tview.TextView
}

func NewAliasModalView(emit func(events.Event)) *AliasModalView {
	input := tview.NewInputField().
		SetLabel("Alias: ").
		SetFieldWidth(36)
	input.SetBorder(true).SetTitle(" Device Alias ")
	input.SetChangedFunc(func(text string) {
		emit(events.AliasDraftChanged{Alias: text})
	})
	input.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			emit(events.AliasSubmitted{Alias: input.GetText()})
		case tcell.KeyEsc:
			emit(events.HideView{})
		}
	})

	footer := tview.NewTextView()
	footer.SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("Enter: save" + components.Divider + "Esc: cancel" + components.Divider + "empty alias clears")
	footer.SetTextColor(tview.Styles.SecondaryTextColor)
	footer.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	content := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(input, 3, 0, true).
		AddItem(footer, 1, 0, false)

	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(content, 60, 0, true).
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 1, false)

	v := &AliasModalView{
		Flex:   root,
		input:  input,
		footer: footer,
	}

	theme.RegisterPrimitive(content)
	theme.RegisterPrimitive(input)
	theme.RegisterPrimitive(footer)

	return v
}

func (v *AliasModalView) FocusTarget() tview.Primitive { return v.input }

func (v *AliasModalView) Render(s state.ReadOnly) {
	device, ok := s.Selected()
	if !ok {
		v.input.SetTitle(" Device Alias ")
		return
	}

	title := " Device Alias "
	if mac := device.MAC(); mac != "" {
		title = fmt.Sprintf(" Device Alias (%s) ", mac)
	}
	v.input.SetTitle(title)
	if draft := s.AliasEditorDraft(); v.input.GetText() != draft {
		v.input.SetText(draft)
	}
}
