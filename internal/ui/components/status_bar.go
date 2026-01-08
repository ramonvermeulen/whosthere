package components

import (
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

// StatusBar combines a Spinner with a right-aligned help text into a single flex row.
type StatusBar struct {
	root    *tview.Flex
	spinner *Spinner
	help    *tview.TextView
}

func NewStatusBar() *StatusBar {
	sp := NewSpinner()
	help := tview.NewTextView().
		SetTextAlign(tview.AlignRight)
	row := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(sp.View(), 0, 1, false).
		AddItem(help, 0, 2, false)

	theme.RegisterPrimitive(help)
	theme.RegisterPrimitive(row)

	return &StatusBar{
		root:    row,
		spinner: sp,
		help:    help,
	}
}

func (s *StatusBar) Primitive() tview.Primitive { return s.root }

func (s *StatusBar) Spinner() *Spinner { return s.spinner }

func (s *StatusBar) SetHelp(text string) {
	if s == nil || s.help == nil {
		return
	}
	s.help.SetText(text)
}

// HelpText exposes the underlying help TextView for callers that need direct access.
func (s *StatusBar) HelpText() *tview.TextView { return s.help }
