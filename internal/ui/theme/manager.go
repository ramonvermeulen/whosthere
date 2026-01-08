package theme

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/config"
	"github.com/ramonvermeulen/whosthere/internal/ui/navigation"
	"github.com/rivo/tview"
)

// Manager coordinates theme changes using a global registry pattern.
// Primitives register themselves, and theme changes are applied to all registered primitives.
type Manager struct {
	mu         sync.RWMutex
	current    string
	app        *tview.Application
	primitives []tview.Primitive
	cfg        *config.Config // Config for saving theme preference
}

var (
	globalManager *Manager
	once          sync.Once
)

// NewManager creates or returns the singleton theme manager.
// This ensures only one manager instance exists throughout the application lifecycle.
func NewManager(app *tview.Application, cfg *config.Config) *Manager {
	once.Do(func() {
		globalManager = &Manager{
			app:        app,
			cfg:        cfg,
			primitives: make([]tview.Primitive, 0),
		}
	})
	return globalManager
}

// Register adds a primitive to be theme-aware. When themes change, it will be updated.
func (m *Manager) Register(p tview.Primitive) {
	if p == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.primitives = append(m.primitives, p)

	if m.current != "" {
		ApplyToPrimitive(p)
	}
}

// SetTheme applies a theme by name and updates all registered primitives.
func (m *Manager) SetTheme(name string) tview.Theme {
	th := Get(name)
	tview.Styles = th

	m.mu.Lock()
	m.current = name
	primitives := append([]tview.Primitive{}, m.primitives...)
	m.mu.Unlock()

	for _, p := range primitives {
		ApplyToPrimitive(p)
		if page, ok := p.(navigation.Page); ok {
			page.Refresh()
		}
	}

	return th
}

// Current returns the currently active theme name.
func (m *Manager) Current() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

// SaveThemeToConfig saves the current theme to the config file.
func (m *Manager) SaveThemeToConfig(themeName string) error {
	if m.cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	m.cfg.Theme.Name = themeName

	if err := config.Save(m.cfg, ""); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// RegisterPrimitive is a convenience function to register a primitive with the global manager.
func RegisterPrimitive(p tview.Primitive) {
	if globalManager != nil {
		globalManager.Register(p)
	}
}

// ApplyToPrimitive applies theme colors to any tview primitive.
// Supports all official tview primitives with proper styling methods.
func ApplyToPrimitive(p tview.Primitive) {
	if p == nil {
		return
	}

	switch v := p.(type) {
	case *tview.TextView:
		v.SetTextColor(tview.Styles.PrimaryTextColor)
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.TextArea:
		v.SetTextStyle(tcell.StyleDefault.
			Foreground(tview.Styles.PrimaryTextColor).
			Background(tview.Styles.PrimitiveBackgroundColor))
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.Table:
		v.SetBordersColor(tview.Styles.BorderColor)
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.TreeView:
		v.SetGraphicsColor(tview.Styles.GraphicsColor)
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.List:
		v.SetMainTextStyle(tcell.StyleDefault.
			Foreground(tview.Styles.PrimaryTextColor).
			Background(tview.Styles.PrimitiveBackgroundColor))
		v.SetSecondaryTextStyle(tcell.StyleDefault.
			Foreground(tview.Styles.SecondaryTextColor).
			Background(tview.Styles.PrimitiveBackgroundColor))
		v.SetShortcutStyle(tcell.StyleDefault.
			Foreground(tview.Styles.TertiaryTextColor).
			Background(tview.Styles.PrimitiveBackgroundColor))
		v.SetSelectedStyle(tcell.StyleDefault.
			Foreground(tview.Styles.InverseTextColor).
			Background(tview.Styles.SecondaryTextColor))
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.InputField:
		v.SetFieldTextColor(tview.Styles.PrimaryTextColor)
		v.SetFieldBackgroundColor(tview.Styles.ContrastBackgroundColor)
		v.SetLabelColor(tview.Styles.SecondaryTextColor)
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.DropDown:
		v.SetFieldTextColor(tview.Styles.PrimaryTextColor)
		v.SetFieldBackgroundColor(tview.Styles.ContrastBackgroundColor)
		v.SetLabelColor(tview.Styles.SecondaryTextColor)
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.Checkbox:
		v.SetLabelColor(tview.Styles.SecondaryTextColor)
		v.SetFieldTextColor(tview.Styles.PrimaryTextColor)
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.Image:
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.Button:
		v.SetLabelColor(tview.Styles.PrimaryTextColor)
		v.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.Form:
		v.SetFieldBackgroundColor(tview.Styles.ContrastBackgroundColor)
		v.SetFieldTextColor(tview.Styles.PrimaryTextColor)
		v.SetLabelColor(tview.Styles.SecondaryTextColor)
		v.SetButtonBackgroundColor(tview.Styles.ContrastBackgroundColor)
		v.SetButtonTextColor(tview.Styles.PrimaryTextColor)
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.Modal:
		v.SetTextColor(tview.Styles.PrimaryTextColor)
		v.SetButtonBackgroundColor(tview.Styles.ContrastBackgroundColor)
		v.SetButtonTextColor(tview.Styles.PrimaryTextColor)
		v.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)

	case *tview.Grid:
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBordersColor(tview.Styles.BorderColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.Flex:
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.Pages:
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	default:
		if box, ok := p.(interface{ SetBackgroundColor(tcell.Color) *tview.Box }); ok {
			box.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		}
		if bordered, ok := p.(interface{ SetBorderColor(tcell.Color) *tview.Box }); ok {
			bordered.SetBorderColor(tview.Styles.BorderColor)
		}
		if titled, ok := p.(interface{ SetTitleColor(tcell.Color) *tview.Box }); ok {
			titled.SetTitleColor(tview.Styles.TitleColor)
		}
	}
}
