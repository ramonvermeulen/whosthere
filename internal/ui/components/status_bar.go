package components

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

const statusBarGap = 1

var _ UIComponent = &StatusBar{}

// StatusBar renders spinner/status text followed by the footer help line.
type StatusBar struct {
	*tview.Box
	spinner      *Spinner
	helpText     string
	displayText  string
	displayColor tcell.Color
	noColor      bool
}

func NewStatusBar() *StatusBar {
	sb := &StatusBar{
		Box:          tview.NewBox(),
		spinner:      NewSpinner(),
		displayColor: tview.Styles.PrimaryTextColor,
	}

	theme.RegisterPrimitive(sb)
	return sb
}

func (s *StatusBar) Spinner() *Spinner { return s.spinner }

func (s *StatusBar) SetHelp(text string) {
	if s == nil {
		return
	}
	s.helpText = text
}

// Draw implements tview.Primitive.
func (s *StatusBar) Draw(screen tcell.Screen) {
	if s == nil || s.Box == nil {
		return
	}

	s.Box.Draw(screen)
	x, y, width, height := s.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}

	spinnerText := s.spinner.Text()
	spinnerWidth := tview.TaggedStringWidth(spinnerText)
	spinnerColor := tview.Styles.SecondaryTextColor
	if s.noColor {
		spinnerColor = tview.Styles.PrimaryTextColor
	}

	if spinnerText != "" {
		tview.Print(screen, spinnerText, x, y, width, tview.AlignLeft, spinnerColor)
	}

	textX := x
	textWidth := width
	if spinnerText != "" {
		textX += spinnerWidth + statusBarGap
		textWidth -= spinnerWidth + statusBarGap
	}
	if textWidth <= 0 || s.displayText == "" {
		return
	}

	tview.Print(screen, truncateToWidth(s.displayText, textWidth), textX, y, textWidth, tview.AlignLeft, s.displayColor)
}

// Render implements UIComponent.
func (s *StatusBar) Render(st state.ReadOnly) {
	if s == nil {
		return
	}

	s.displayText = s.helpText
	s.displayColor = tview.Styles.PrimaryTextColor
	s.noColor = false

	if st == nil {
		return
	}

	s.noColor = st.NoColor()
	if message := st.StatusMessage(); message != "" {
		s.displayText = message
		s.displayColor = statusSeverityColor(st)
		return
	}

	s.displayText = s.helpText
	s.displayColor = tview.Styles.PrimaryTextColor
}

func truncateToWidth(text string, width int) string {
	if width <= 0 || text == "" {
		return ""
	}
	if tview.TaggedStringWidth(text) <= width {
		return text
	}
	if width == 1 {
		return "…"
	}

	runes := []rune(text)
	for len(runes) > 0 {
		candidate := strings.TrimRight(string(runes), " ") + " …"
		if tview.TaggedStringWidth(candidate) <= width {
			return candidate
		}
		runes = runes[:len(runes)-1]
	}

	return "…"
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
