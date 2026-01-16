package ui

import (
	"context"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/arp"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/mdns"
	"github.com/ramonvermeulen/whosthere/internal/core/discovery/ssdp"
	"github.com/ramonvermeulen/whosthere/internal/core/oui"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/events"
	"github.com/ramonvermeulen/whosthere/internal/ui/routes"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/ramonvermeulen/whosthere/internal/ui/views"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

const (
	refreshInterval = 1 * time.Second
)

type App struct {
	*tview.Application
	pages         *tview.Pages
	engine        *discovery.Engine
	state         *state.AppState
	scanTicker    *time.Ticker
	refreshTicker *time.Ticker
	cfg           *config.Config
	primitives    []tview.Primitive
	events        chan events.Event
	emit          func(events.Event)
}

func NewApp(cfg *config.Config, ouiDB *oui.Registry, version string) *App {
	app := tview.NewApplication()
	appState := state.NewAppState()
	appState.SetVersion(version)
	themeName := config.DefaultThemeName
	if cfg != nil && cfg.Theme.Name != "" {
		themeName = cfg.Theme.Name
	}
	appState.SetCurrentTheme(themeName)
	appState.SetPreviousTheme(themeName)

	a := &App{
		Application: app,
		state:       appState,
		cfg:         cfg,
		events:      make(chan events.Event, 100),
	}

	a.emit = func(e events.Event) {
		select {
		case a.events <- e:
		default:
			// drop if full
		}
	}
	a.pages = tview.NewPages()

	theme.SetRegisterFunc(a.RegisterPrimitive)

	a.applyTheme(themeName)
	a.setupPages(cfg)

	if cfg != nil {
		a.setupEngine(cfg, ouiDB)
	}

	app.SetRoot(a.pages, true)
	app.SetInputCapture(a.handleGlobalKeys)
	app.EnableMouse(true)

	return a
}

func (a *App) Run() error {
	zap.L().Debug("App Run started")
	go a.handleEvents()
	a.startUIRefreshLoop()

	if a.cfg != nil && a.cfg.Splash.Enabled {
		go func(delay time.Duration) {
			defer func() {
				if r := recover(); r != nil {
					zap.L().Error("Panic in splash delay", zap.Any("panic", r))
				}
			}()
			time.Sleep(delay)
			a.emit(events.NavigateTo{Route: routes.RouteDashboard})
		}(a.cfg.Splash.Delay)
	}

	if a.engine != nil && a.cfg != nil {
		a.startDiscoveryScanLoop()
	}

	return a.Application.Run()
}

func (a *App) setupPages(cfg *config.Config) {
	dashboardPage := views.NewDashboardView(a.emit, a.QueueUpdateDraw)
	detailPage := views.NewDetailView(a.emit)
	splashPage := views.NewSplashView(a.emit)
	themePickerModal := views.NewThemeModalView(a.emit)

	a.pages.AddPage(routes.RouteDashboard, dashboardPage, true, false)
	a.pages.AddPage(routes.RouteDetail, detailPage, true, false)
	a.pages.AddPage(routes.RouteSplash, splashPage, true, false)
	a.pages.AddPage(routes.RouteThemePicker, themePickerModal, true, false)

	initialPage := routes.RouteDashboard
	if cfg != nil && cfg.Splash.Enabled {
		initialPage = routes.RouteSplash
	}
	a.pages.SwitchToPage(initialPage)
}

func (a *App) setupEngine(cfg *config.Config, ouiDB *oui.Registry) {
	sweeper := arp.NewSweeper(5*time.Minute, time.Minute)
	var scanners []discovery.Scanner

	if cfg.Scanners.SSDP.Enabled {
		scanners = append(scanners, &ssdp.Scanner{})
	}
	if cfg.Scanners.ARP.Enabled {
		scanners = append(scanners, arp.NewScanner(sweeper))
	}
	if cfg.Scanners.MDNS.Enabled {
		scanners = append(scanners, &mdns.Scanner{})
	}

	a.engine = discovery.NewEngine(
		scanners,
		discovery.WithTimeout(cfg.ScanDuration),
		discovery.WithOUIRegistry(ouiDB),
		discovery.WithSubnetHook(sweeper.Trigger),
	)
}

func (a *App) handleGlobalKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlT:
		a.emit(events.NavigateTo{Route: routes.RouteThemePicker, Overlay: true})
		return nil
	case tcell.KeyCtrlC:
		a.Stop()
		return nil
	}

	return event
}

func (a *App) startUIRefreshLoop() {
	a.refreshTicker = time.NewTicker(refreshInterval)

	go func() {
		for range a.refreshTicker.C {
			a.rerenderVisibleViews()
		}
	}()
}

func (a *App) startDiscoveryScanLoop() {
	if a.cfg == nil {
		return
	}

	a.scanTicker = time.NewTicker(a.cfg.ScanInterval)

	go func() {
		a.performScan()

		for range a.scanTicker.C {
			a.performScan()
		}
	}()
}

func (a *App) QueueUpdateDraw(f func()) {
	if a.Application == nil {
		return
	}
	go func() {
		a.Application.QueueUpdateDraw(f)
	}()
}

func (a *App) performScan() {
	if a.cfg == nil || a.engine == nil {
		return
	}

	a.emit(events.DiscoveryStarted{})

	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.ScanDuration)
	_, _ = a.engine.Stream(ctx, func(d discovery.Device) {
		a.state.UpsertDevice(&d)
	})
	cancel()

	a.emit(events.DiscoveryStopped{})
}

// RegisterPrimitive registers a primitive for theme updates.
func (a *App) RegisterPrimitive(p tview.Primitive) {
	a.primitives = append(a.primitives, p)
}

// applyTheme applies a theme by name, updates state, applies to primitives, and renders all pages.
func (a *App) applyTheme(name string) {
	th := theme.Get(name)
	tview.Styles = th
	for _, p := range a.primitives {
		theme.ApplyToPrimitive(p)
	}
	a.rerenderVisibleViews()
}

func (a *App) resetFocus() {
	_, item := a.pages.GetFrontPage()
	if item == nil {
		return
	}
	if view, ok := item.(views.View); ok {
		if ft := view.FocusTarget(); ft != nil {
			a.SetFocus(ft)
		} else {
			a.SetFocus(view)
		}
	}
}

func (a *App) rerenderVisibleViews() {
	a.QueueUpdateDraw(func() {
		for _, name := range a.pages.GetPageNames(true) {
			if pageItem := a.pages.GetPage(name); pageItem != nil {
				if view, ok := pageItem.(views.View); ok {
					view.Render(a.state.ReadOnly())
				}
			}
		}
	})
}

func (a *App) handleEvents() {
	for e := range a.events {
		zap.L().Debug("Handling event: ", zap.Any("event", e))
		switch event := e.(type) {
		case events.DeviceSelected:
			a.state.SetSelectedIP(event.IP)
		case events.FilterChanged:
			a.state.SetFilterPattern(event.Pattern)
		case events.NavigateTo:
			if event.Overlay {
				a.pages.SendToFront(event.Route)
				a.pages.ShowPage(event.Route)
			} else {
				a.pages.SwitchToPage(event.Route)
			}
			a.resetFocus()
		case events.ThemeSelected:
			a.state.SetCurrentTheme(event.Name)
			a.applyTheme(event.Name)
		case events.ThemeSaved:
			_ = theme.SaveToConfig(event.Name, a.cfg)
		case events.ThemeConfirmed:
			a.state.SetPreviousTheme(a.state.CurrentTheme())
		case events.HideView:
			front, _ := a.pages.GetFrontPage()
			a.pages.HidePage(front)
			a.resetFocus()
		case events.DiscoveryStarted:
			a.state.SetIsDiscovering(true)
		case events.DiscoveryStopped:
			a.state.SetIsDiscovering(false)
		}

		a.rerenderVisibleViews()
	}
}
