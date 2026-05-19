package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

var _ UIComponent = &StatusBar{}

// StatusBar combines a Spinner with a right-aligned help text into a single flex row.
type StatusBar struct {
	*tview.Flex
	spinner  *Spinner
	help     *tview.TextView
	helpText string
}

func NewStatusBar() *StatusBar {
	sp := NewSpinner()
	help := tview.NewTextView().
		SetTextAlign(tview.AlignRight)
	row := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(sp, 0, 1, false).
		AddItem(help, 0, 2, false)

	theme.RegisterPrimitive(help)
	theme.RegisterPrimitive(row)

	return &StatusBar{
		Flex:    row,
		spinner: sp,
		help:    help,
	}
}

func (s *StatusBar) Spinner() *Spinner { return s.spinner }

func (s *StatusBar) SetHelp(text string) {
	if s == nil || s.help == nil {
		return
	}
	s.helpText = text
	s.help.SetText(text)
}

// Render implements UIComponent.
func (s *StatusBar) Render(st state.ReadOnly) {
	if s == nil || s.help == nil {
		return
	}

	if st == nil {
		s.help.SetText(s.helpText)
		s.help.SetTextColor(tview.Styles.PrimaryTextColor)
		return
	}

	if message := st.StatusMessage(); message != "" {
		s.help.SetText(message)
		s.help.SetTextColor(statusSeverityColor(st))
		return
	}

	s.help.SetText(s.helpText)
	s.help.SetTextColor(tview.Styles.PrimaryTextColor)
}

func statusSeverityColor(st state.ReadOnly) tcell.Color {
	if st == nil || st.NoColor() {
		return tview.Styles.PrimaryTextColor
	}

	switch st.StatusSeverity() {
	case state.StatusSeverityError:
		return tcell.ColorRed
	case state.StatusSeveritySuccess:
		return tview.Styles.ContrastSecondaryTextColor
	default:
		return tview.Styles.SecondaryTextColor
	}
}
