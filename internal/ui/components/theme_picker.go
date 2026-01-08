package components

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

// ThemePicker is a component for selecting and previewing themes.
// It's just a themed list that handles theme selection logic.
type ThemePicker struct {
	list          *tview.List
	themes        []string
	currentIndex  int
	originalTheme string
	onSelect      func(themeName string)
	onSave        func(themeName string)
	onCancel      func()
	themeManager  *theme.Manager
}

// NewThemePicker creates a new theme picker list component.
func NewThemePicker(tm *theme.Manager) *ThemePicker {
	list := tview.NewList()
	list.ShowSecondaryText(false)

	tp := &ThemePicker{
		list:         list,
		themes:       theme.Names(),
		themeManager: tm,
	}

	tp.buildList()
	theme.RegisterPrimitive(tp.list)

	return tp
}

// buildList populates the list with available themes.
func (tp *ThemePicker) buildList() {
	tp.list.Clear()
	tp.list.SetBorder(true).
		SetTitle(" Theme Picker - Preview themes live ").
		SetTitleAlign(tview.AlignCenter).
		SetTitleColor(tview.Styles.TitleColor).
		SetBorderColor(tview.Styles.BorderColor).
		SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	tp.list.ShowSecondaryText(false)

	currentTheme := tp.themeManager.Current()

	for i, themeName := range tp.themes {
		displayName := themeName
		if themeName == currentTheme {
			displayName = fmt.Sprintf("● %s (current)", themeName)
			tp.currentIndex = i
		}

		name := themeName
		tp.list.AddItem(displayName, "", 0, func() {
			if tp.onSelect != nil {
				tp.onSelect(name)
			}
		})
	}

	tp.list.SetCurrentItem(tp.currentIndex)
	tp.setupInputHandling()
}

// setupInputHandling configures vim-style navigation and preview.
func (tp *ThemePicker) setupInputHandling() {
	tp.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Rune() == 'j' || event.Key() == tcell.KeyDown:
			nextIdx := tp.list.GetCurrentItem() + 1
			if nextIdx < len(tp.themes) {
				tp.list.SetCurrentItem(nextIdx)
				tp.previewTheme(tp.themes[nextIdx])
			}
			return nil
		case event.Rune() == 'k' || event.Key() == tcell.KeyUp:
			prevIdx := tp.list.GetCurrentItem() - 1
			if prevIdx >= 0 {
				tp.list.SetCurrentItem(prevIdx)
				tp.previewTheme(tp.themes[prevIdx])
			}
			return nil
		case event.Key() == tcell.KeyEnter && event.Modifiers()&tcell.ModShift != 0:
			currentIdx := tp.list.GetCurrentItem()
			if currentIdx >= 0 && currentIdx < len(tp.themes) {
				if tp.onSave != nil {
					tp.onSave(tp.themes[currentIdx])
				}
			}
			return nil
		case event.Key() == tcell.KeyEnter:
			currentIdx := tp.list.GetCurrentItem()
			if currentIdx >= 0 && currentIdx < len(tp.themes) {
				if tp.onSelect != nil {
					tp.onSelect(tp.themes[currentIdx])
				}
			}
			return nil
		case event.Key() == tcell.KeyEsc || event.Rune() == 'q':
			if tp.themeManager != nil && tp.originalTheme != "" {
				tp.themeManager.SetTheme(tp.originalTheme)
			}
			if tp.onCancel != nil {
				tp.onCancel()
			}
			return nil
		}
		return event
	})
}

// previewTheme temporarily applies a theme for preview.
func (tp *ThemePicker) previewTheme(themeName string) {
	if tp.themeManager != nil {
		tp.themeManager.SetTheme(themeName)
	}
}

// OnSelect registers a callback for when a theme is selected (Enter).
func (tp *ThemePicker) OnSelect(fn func(themeName string)) {
	tp.onSelect = fn
}

// OnSave registers a callback for when a theme should be saved to config (Shift+Enter).
func (tp *ThemePicker) OnSave(fn func(themeName string)) {
	tp.onSave = fn
}

// OnCancel registers a callback for when the picker is cancelled (Esc).
func (tp *ThemePicker) OnCancel(fn func()) {
	tp.onCancel = fn
}

// Show displays the theme picker and stores the current theme for potential rollback.
func (tp *ThemePicker) Show() {
	if tp.themeManager != nil {
		tp.originalTheme = tp.themeManager.Current()
	}
	tp.buildList()
}

// GetList returns the list primitive for rendering.
func (tp *ThemePicker) GetList() *tview.List {
	return tp.list
}
