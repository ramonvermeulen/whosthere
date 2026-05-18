package ui

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dece2183/go-clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/core"
	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/events"
	"github.com/ramonvermeulen/whosthere/internal/ui/routes"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/ramonvermeulen/whosthere/internal/ui/views"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/rivo/tview"
)

const (
	refreshInterval = 1 * time.Second
)

// App represents the main TUI application.
type App struct {
	*tview.Application
	pages         *tview.Pages
	engine        *discovery.Engine
	state         *state.AppState
	refreshTicker *time.Ticker
	cfg           *config.Config
	events        chan events.Event
	emit          func(events.Event)
	portScanner   *discovery.PortScanner
	isReady       bool
	clipboard     *clipboard.Clipboard
	logger        *slog.Logger
	engineCancel  context.CancelFunc
	engineMu      sync.Mutex
}

func NewApp(cfg *config.Config, logger *slog.Logger, version string) (*App, error) {
	app := tview.NewApplication()
	appState := state.NewAppState(cfg, version)

	if logger == nil {
		logger = slog.Default()
	}

	a := &App{
		Application: app,
		state:       appState,
		cfg:         cfg,
		events:      make(chan events.Event, 100),
		clipboard:   clipboard.New(clipboard.ClipboardOptions{Primary: false}),
		logger:      logger,
	}
	a.setupSignalHandler()

	a.emit = func(e events.Event) {
		a.events <- e
	}
	a.pages = tview.NewPages()

	a.applyTheme(appState.CurrentTheme())
	a.setupPages(cfg)

	engine, err := core.BuildEngine(a.cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("build engine: %w", err)
	}
	a.engine = engine
	// todo(ramon) handle in BuildEngine -> WithPortScanner(...)
	a.portScanner = discovery.NewPortScanner(100, engine.Iface)
	a.state.SetCurrentInterface(engine.Iface.Interface.Name)

	app.SetRoot(a.pages, true)
	app.SetInputCapture(a.handleGlobalKeys)
	app.EnableMouse(true)

	return a, nil
}

func (a *App) setupSignalHandler() {
	defer func() {
		if r := recover(); r != nil {
			a.logger.Error("panic in signal handler", "panic", r)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				a.logger.Error("panic in signal handler goroutine", "panic", r)
				os.Exit(1)
			}
		}()

		sig := <-sigChan
		a.logger.Info("received signal, shutting down", "signal", sig)
		a.Stop()
	}()
}

func (a *App) Run() error {
	a.logger.Debug("App Run started")
	go a.handleEvents()
	a.startUIRefreshLoop()

	if a.cfg != nil && a.cfg.Splash.Enabled {
		go func(delay time.Duration) {
			time.Sleep(delay)
			a.emit(events.NavigateTo{Route: routes.RouteDashboard})
			a.isReady = true
		}(a.cfg.Splash.Delay)
	} else {
		a.isReady = true
	}

	if a.engine != nil && a.cfg != nil {
		ctx, cancel := context.WithCancel(context.Background())
		a.engineCancel = cancel
		a.engine.Start(context.Background())
		go a.handleEngineEvents(ctx, a.engine)
	}

	return a.Application.Run()
}

func (a *App) setupPages(cfg *config.Config) {
	dashboardPage := views.NewDashboardView(a.emit, a.QueueUpdateDraw)
	detailPage := views.NewDetailView(a.emit, a.QueueUpdateDraw)
	splashPage := views.NewSplashView(a.emit)
	themePickerModal := views.NewThemeModalView(a.emit)
	portScanModal := views.NewPortScanModalView(a.emit)
	interfacePickerModal := views.NewInterfaceModalView(a.emit)

	a.pages.AddPage(routes.RouteDashboard, dashboardPage, true, false)
	a.pages.AddPage(routes.RouteDetail, detailPage, true, false)
	a.pages.AddPage(routes.RouteSplash, splashPage, true, false)
	a.pages.AddPage(routes.RouteThemePicker, themePickerModal, true, false)
	a.pages.AddPage(routes.RoutePortScan, portScanModal, true, false)
	a.pages.AddPage(routes.RouteInterfacePicker, interfacePickerModal, true, false)

	initialPage := routes.RouteDashboard
	if cfg != nil && cfg.Splash.Enabled {
		initialPage = routes.RouteSplash
	}
	a.pages.SwitchToPage(initialPage)
}

func (a *App) handleGlobalKeys(event *tcell.EventKey) *tcell.EventKey {
	if !a.isReady {
		return event
	}
	switch event.Key() {
	case tcell.KeyCtrlT:
		a.emit(events.NavigateTo{Route: routes.RouteThemePicker, Overlay: true})
		return nil
	case tcell.KeyCtrlI:
		a.emit(events.NavigateTo{Route: routes.RouteInterfacePicker, Overlay: true})
		return nil
	case tcell.KeyRune:
		if event.Rune() == 'q' || event.Rune() == 'Q' {
			a.Stop()
			return nil
		}
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

// handleEngineEvents processes events from a specific engine instance.
// It exits when the context is cancelled or the engine's Events channel is closed.
func (a *App) handleEngineEvents(ctx context.Context, engine *discovery.Engine) {
	defer func() {
		if r := recover(); r != nil {
			a.logger.Error("panic in handleEngineEvents", "panic", r)
		}
	}()

	for event := range engine.Events {
		if ctx.Err() != nil {
			return
		}
		switch event.Type {
		case discovery.EventScanStarted:
			a.emit(events.DiscoveryStarted{})
		case discovery.EventScanCompleted:
			a.emit(events.DiscoveryStopped{})
		case discovery.EventDeviceDiscovered:
			if event.Device != nil {
				a.state.UpsertDevice(event.Device)
			}
		case discovery.EventError:
			a.emit(events.DiscoveryStopped{})
			if event.Error != nil {
				a.logger.Error("scan failed", "error", event.Error)
			}
		}
	}
}

func (a *App) QueueUpdateDraw(f func()) {
	if a.Application == nil {
		return
	}
	go func() {
		a.Application.QueueUpdateDraw(f)
	}()
}

// applyTheme applies a theme by name, updates state, applies to primitives, and renders all pages.
func (a *App) applyTheme(name string) {
	a.cfg.Theme.Name = name

	var th tview.Theme
	switch {
	case theme.IsNoColor() || a.cfg.Theme.NoColor:
		th = theme.NoColorTheme()
	case !a.cfg.Theme.Enabled:
		th = theme.TviewDefaultTheme()
	default:
		th = theme.Resolve(&a.cfg.Theme)
	}

	tview.Styles = th
	theme.ApplyThemeToAllRegisteredPrimitives()
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
	defer func() {
		if r := recover(); r != nil {
			a.logger.Error("panic in handleEvents", "panic", r)
		}
	}()

	for e := range a.events {
		a.logger.Debug("handling event", "event", e)
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
			a.applyTheme(event.Name)
			a.state.SetCurrentTheme(event.Name)
		case events.ThemeSaved:
			if err := theme.SaveToConfig(event.Name, a.cfg); err != nil {
				a.logger.Error("failed to save theme", "error", err)
			}
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
		case events.PortScanStarted:
			a.state.SetIsPortscanning(true)
			a.emit(events.HideView{})
			go a.startPortscan()
		case events.PortScanStopped:
			a.state.SetIsPortscanning(false)
		case events.SearchStarted:
			a.state.SetSearchActive(true)
		case events.SearchError:
			a.state.SetSearchError(event.Error)
		case events.SearchFinished:
			a.state.SetSearchActive(false)
		case events.CopyIP:
			var ip string
			if event.IP != "" {
				ip = event.IP
			} else {
				device, ok := a.state.Selected()
				if ok {
					ip = device.IP().String()
				}
			}
			if ip != "" {
				if err := a.clipboard.CopyText(ip); err != nil {
					a.logger.Warn("failed to copy to clipboard", "error", err)
				}
			}
		case events.CopyMac:
			var mac string
			if event.MAC != "" {
				mac = event.MAC
			} else {
				device, ok := a.state.Selected()
				if ok {
					mac = device.MAC()
				}
			}
			if mac != "" {
				if err := a.clipboard.CopyText(mac); err != nil {
					a.logger.Warn("failed to copy to clipboard", "error", err)
				}
			}
		case events.InterfaceSelected:
			go a.switchInterface(event.Name)
		}
		a.rerenderVisibleViews()
	}
}

func (a *App) startPortscan() {
	device, ok := a.state.Selected()
	if !ok {
		a.emit(events.PortScanStopped{})
		return
	}
	ip := device.IP().String()
	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.ScanTimeout)
	defer cancel()

	openPorts := make(map[string][]int)
	device.SetOpenPorts(openPorts)
	device.SetLastPortScan(time.Now())

	var mu sync.Mutex
	_ = a.portScanner.Stream(ctx, ip, a.cfg.PortScanner.TCP, a.cfg.PortScanner.Timeout, func(port int) {
		mu.Lock()
		defer mu.Unlock()
		openPorts["tcp"] = append(openPorts["tcp"], port)
	})

	device.SetOpenPorts(openPorts)
	a.emit(events.PortScanStopped{})
}

func (a *App) switchInterface(name string) {
	a.engineMu.Lock()
	defer a.engineMu.Unlock()

	if a.engineCancel != nil {
		a.engineCancel()
	}

	if a.engine != nil {
		oldEngine := a.engine
		go oldEngine.Stop()
	}

	if !a.cfg.AllInterfaces {
		a.state.ClearDevices()
	}
	a.state.SetIsDiscovering(false)
	a.rerenderVisibleViews()

	a.cfg.NetworkInterface = name
	engine, err := core.BuildEngine(a.cfg, a.logger)
	if err != nil {
		a.logger.Error("failed to build engine for interface", "interface", name, "error", err)
		return
	}

	a.engine = engine
	a.portScanner = discovery.NewPortScanner(100, engine.Iface)
	a.state.SetCurrentInterface(name)

	ctx, cancel := context.WithCancel(context.Background())
	a.engineCancel = cancel

	a.engine.Start(context.Background())
	go a.handleEngineEvents(ctx, engine)
}

