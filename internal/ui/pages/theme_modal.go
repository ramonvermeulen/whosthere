package pages

import (
	"github.com/ramonvermeulen/whosthere/internal/ui/components"
	"github.com/ramonvermeulen/whosthere/internal/ui/navigation"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

var _ navigation.Page = &ThemeModalPage{}

// ThemeModalPage is a modal overlay page for selecting themes.
// It uses a centered flex layout to create a modal-like appearance.
type ThemeModalPage struct {
	root         *tview.Flex
	picker       *components.ThemePicker
	footer       *tview.TextView
	router       *navigation.Router
	app          *tview.Application
	themeManager *theme.Manager
}

// NewThemePickerModalPage creates a new theme picker modal page.
func NewThemePickerModalPage(tm *theme.Manager, router *navigation.Router, app *tview.Application) *ThemeModalPage {
	picker := components.NewThemePicker(tm)

	footer := tview.NewTextView()
	footer.SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("j/k: navigate | Enter: apply | Shift+Enter: save to config | Esc: cancel")
	footer.SetTextColor(tview.Styles.SecondaryTextColor)
	footer.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	content := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(picker.GetList(), 0, 1, true).
		AddItem(footer, 1, 0, false)

	modalWidth := len(footer.GetText(false))

	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(content, modalWidth, 0, true).
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 1, false)

	p := &ThemeModalPage{
		root:         root,
		picker:       picker,
		footer:       footer,
		router:       router,
		app:          app,
		themeManager: tm,
	}

	theme.RegisterPrimitive(content)
	theme.RegisterPrimitive(footer)

	picker.OnSelect(func(themeName string) {
		p.router.HideOverlay(navigation.RouteThemePicker)
		p.router.FocusCurrent(p.app)
	})

	picker.OnSave(func(themeName string) {
		_ = p.themeManager.SaveThemeToConfig(themeName)
		p.router.HideOverlay(navigation.RouteThemePicker)
		p.router.FocusCurrent(p.app)
	})

	picker.OnCancel(func() {
		p.router.HideOverlay(navigation.RouteThemePicker)
		p.router.FocusCurrent(p.app)
	})

	return p
}

func (p *ThemeModalPage) GetName() string               { return navigation.RouteThemePicker }
func (p *ThemeModalPage) GetPrimitive() tview.Primitive { return p.root }
func (p *ThemeModalPage) FocusTarget() tview.Primitive  { return p.picker.GetList() }

// Refresh prepares the picker for display by resetting it to the current theme.
func (p *ThemeModalPage) Refresh() {
	p.picker.Show()
}
