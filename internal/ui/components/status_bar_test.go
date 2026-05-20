package components

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/rivo/tview"
)

func TestStatusBarDrawsHelpFromLeftWhenSpinnerHidden(t *testing.T) {
	screen := newSimulationScreen(t)

	bar := NewStatusBar()
	bar.SetHelp("q: quit | Enter: details")
	bar.Render(state.NewAppState(nil, ""))
	bar.SetRect(0, 0, 40, 1)
	bar.Draw(screen)

	if got := strings.TrimRight(readScreenLine(screen, 0, 40), " "); !strings.HasPrefix(got, "q: quit | Enter: details") {
		t.Fatalf("expected help text to start at the left edge, got %q", got)
	}
}

func TestStatusBarPlacesHelpImmediatelyAfterSpinnerText(t *testing.T) {
	screen := newSimulationScreen(t)

	bar := NewStatusBar()
	bar.SetHelp("q: quit")
	bar.spinner.setText("x Discovering Devices")
	bar.Render(state.NewAppState(nil, ""))
	bar.SetRect(0, 0, 40, 1)
	bar.Draw(screen)

	got := strings.TrimRight(readScreenLine(screen, 0, 40), " ")
	wantPrefix := "x Discovering Devices q: quit"
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("expected spinner and help to share one left-aligned line, got %q", got)
	}
}

func TestStatusBarTruncatesHelpWithEllipsisOnSmallWidths(t *testing.T) {
	screen := newSimulationScreen(t)

	bar := NewStatusBar()
	bar.SetHelp("j/k: up/down | Enter: details | q: quit")
	bar.spinner.setText("x Discovering Devices")
	bar.Render(state.NewAppState(nil, ""))
	bar.SetRect(0, 0, 30, 1)
	bar.Draw(screen)

	got := strings.TrimRight(readScreenLine(screen, 0, 30), " ")
	if !strings.HasSuffix(got, " …") {
		t.Fatalf("expected truncated help to end with ellipsis, got %q", got)
	}
}

func TestStatusBarUsesSecondaryColorForSpinner(t *testing.T) {
	screen := newSimulationScreen(t)

	bar := NewStatusBar()
	bar.SetHelp("q: quit")
	bar.spinner.setText("x Discovering Devices")
	bar.Render(nil)
	bar.SetRect(0, 0, 40, 1)
	bar.Draw(screen)

	_, style, _ := screen.Get(0, 0)
	foreground, _, _ := style.Decompose()
	if foreground != tview.Styles.SecondaryTextColor {
		t.Fatalf("expected spinner to use secondary text color, got %v", foreground)
	}
}

func newSimulationScreen(t *testing.T) tcell.SimulationScreen {
	t.Helper()

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("init screen: %v", err)
	}
	t.Cleanup(screen.Fini)
	return screen
}

func readScreenLine(screen tcell.SimulationScreen, y, width int) string {
	var b strings.Builder
	for x := 0; x < width; x++ {
		text, _, _ := screen.Get(x, y)
		if text == "" {
			text = " "
		}
		r, _ := utf8.DecodeRuneInString(text)
		b.WriteRune(r)
	}
	return b.String()
}
