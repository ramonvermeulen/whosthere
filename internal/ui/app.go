package ui

import (
	"context"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/config"
	"github.com/ramonvermeulen/whosthere/internal/discovery"
	"github.com/ramonvermeulen/whosthere/internal/discovery/arp"
	"github.com/ramonvermeulen/whosthere/internal/discovery/mdns"
	"github.com/ramonvermeulen/whosthere/internal/discovery/ssdp"
	"github.com/ramonvermeulen/whosthere/internal/oui"
	"github.com/ramonvermeulen/whosthere/internal/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/navigation"
	"github.com/ramonvermeulen/whosthere/internal/ui/pages"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

const (
	// refreshInterval frequency of UI refreshes for redrawing tables/spinners/etc.
	refreshInterval = 1 * time.Second
)

// App is the public interface for running the TUI.
type App interface {
	Run() error
}

// tui is the concrete implementation of the App interface.
type tui struct {
	app          *tview.Application
	cfg          *config.Config
	router       *navigation.Router
	engine       *discovery.Engine
	state        *state.AppState
	version      string
	themeManager *theme.Manager
}

// NewApp constructs a new TUI instance.
func NewApp(cfg *config.Config, ouiDB *oui.Registry, version string) App {
	app := tview.NewApplication()

	appState := state.NewAppState()

	t := &tui{
		app:          app,
		cfg:          cfg,
		version:      version,
		themeManager: theme.Init(app, cfg),
		state:        appState,
		router:       navigation.NewRouter(appState),
	}

	themeName := config.DefaultThemeName
	if cfg != nil && cfg.Theme.Name != "" {
		themeName = cfg.Theme.Name
	}
	t.themeManager.SetTheme(themeName)

	if cfg != nil {
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

		t.engine = discovery.NewEngine(
			scanners,
			discovery.WithTimeout(cfg.ScanDuration),
			discovery.WithOUIRegistry(ouiDB),
			discovery.WithSubnetHook(sweeper.Trigger),
		)
	}

	dashboardPage := pages.NewDashboardPage(appState, t.router, version)
	detailPage := pages.NewDetailPage(appState, t.router, version)
	splashPage := pages.NewSplashPage(version)
	themePickerPage := pages.NewThemePickerModalPage(t.router)

	t.router.Register(dashboardPage)
	t.router.Register(detailPage)
	t.router.Register(splashPage)
	t.router.Register(themePickerPage)

	if cfg != nil && cfg.Splash.Enabled {
		t.router.NavigateTo(navigation.RouteSplash)
	} else {
		t.router.NavigateTo(navigation.RouteDashboard)
	}

	app.SetRoot(t.router, true)
	t.router.FocusCurrent(app)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlT {
			t.router.ShowOverlay(navigation.RouteThemePicker)
			return nil
		}
		return event
	})

	return t
}

// UIQueue returns a helper suitable for components that need to queue UI updates.
func (t *tui) UIQueue() func(func()) {
	return func(f func()) { t.app.QueueUpdateDraw(f) }
}

// Run starts the TUI event loop and background workers.
func (t *tui) Run() error {
	t.startBackgroundTasks()
	if t.cfg != nil && t.cfg.Splash.Enabled {
		go func(delay time.Duration) {
			time.Sleep(delay)
			t.app.QueueUpdateDraw(func() {
				if t.router != nil {
					t.router.NavigateTo(navigation.RouteDashboard)
					t.router.FocusCurrent(t.app)
				}
			})
		}(t.cfg.Splash.Delay)
	}
	return t.app.Run()
}

// startBackgroundTasks launches app-wide background workers (UI refresh, discovery scanning).
func (t *tui) startBackgroundTasks() {
	t.startUIRefreshLoop()
	t.startDiscoveryScanLoop()
}

// startUIRefreshLoop periodically refreshes the current page to show updated state.
func (t *tui) startUIRefreshLoop() {
	if t.router == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()
		for range ticker.C {
			currentRoute := t.router.Current()
			currentPage := t.router.Page(currentRoute)

			// Refresh the current page if it implements the Refresh method
			if refreshable, ok := currentPage.(interface{ Refresh() }); ok {
				t.app.QueueUpdateDraw(func() { refreshable.Refresh() })
			}
		}
	}()
}

// startDiscoveryScanLoop runs periodic network discovery and controls the spinner around scans.
func (t *tui) startDiscoveryScanLoop() {
	if t.cfg == nil || t.engine == nil || t.router == nil || t.state == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(t.cfg.ScanInterval)
		defer ticker.Stop()

		doScan := func() {
			mp, _ := t.router.Page(navigation.RouteDashboard).(*pages.DashboardPage)
			if mp == nil {
				return
			}
			mp.Spinner().Start(t.UIQueue())
			ctx := context.Background()
			cctx, cancel := context.WithTimeout(ctx, t.cfg.ScanDuration)
			_, _ = t.engine.Stream(cctx, func(d discovery.Device) {
				t.state.UpsertDevice(&d)
			})
			cancel()
			mp.Spinner().Stop(t.UIQueue())
		}

		doScan()

		for range ticker.C {
			doScan()
		}
	}()
}
