package views

import (
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/components"
	"github.com/ramonvermeulen/whosthere/internal/ui/events"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

var _ View = &InterfaceModalView{}

type InterfaceModalView struct {
	*tview.Flex
	picker *components.InterfacePicker
	footer *tview.TextView

	emit func(events.Event)
}

func NewInterfaceModalView(emit func(events.Event)) *InterfaceModalView {
	picker := components.NewInterfacePicker(emit)
	footer := tview.NewTextView()
	footer.SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("j/k: navigate" + components.Divider + "Enter: select" + components.Divider + "Esc: cancel")
	footer.SetTextColor(tview.Styles.SecondaryTextColor)
	footer.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	content := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(picker, 0, 1, true).
		AddItem(footer, 1, 0, false)

	modalWidth := len(footer.GetText(false)) + 10

	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(content, modalWidth, 0, true).
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 1, false)

	v := &InterfaceModalView{
		Flex:   root,
		picker: picker,
		footer: footer,
		emit:   emit,
	}

	theme.RegisterPrimitive(content)
	theme.RegisterPrimitive(footer)

	return v
}

func (v *InterfaceModalView) FocusTarget() tview.Primitive { return v.picker }

func (v *InterfaceModalView) Render(s state.ReadOnly) {
	v.picker.Render(s)
}

