package ui

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dece2183/go-clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/core"
	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/devicemeta"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/events"
	"github.com/ramonvermeulen/whosthere/internal/ui/routes"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/ramonvermeulen/whosthere/internal/ui/views"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/rivo/tview"
)

const (
	refreshInterval  = 1 * time.Second
	statusMessageTTL = 1500 * time.Millisecond
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
	metaStore     devicemeta.Store
	stopOnce      sync.Once

	// Engine lifecycle management for runtime interface switching.
	// When the user switches network interface, the old engine must stop cleanly
	// and the new one must start without any stale device data leaking through.
	engineCancel context.CancelFunc // cancels the context of the active handleEngineEvents goroutine
	engineMu     sync.RWMutex       // write-locked during interface switch, read-locked during device upsert
	engineGen    atomic.Uint64      // generation counter; old goroutines compare against this before upserting devices
}

func NewApp(cfg *config.Config, logger *slog.Logger, version string) (*App, error) {
	app := tview.NewApplication()
	appState := state.NewAppState(cfg, version)

	if logger == nil {
		logger = slog.Default()
	}

	metaStore, err := devicemeta.OpenDefault()
	if err != nil {
		return nil, fmt.Errorf("open device metadata store: %w", err)
	}

	a := &App{
		Application: app,
		state:       appState,
		cfg:         cfg,
		events:      make(chan events.Event, 100),
		clipboard:   clipboard.New(clipboard.ClipboardOptions{Primary: false}),
		logger:      logger,
		metaStore:   metaStore,
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
		gen := a.engineGen.Add(1)
		a.engine.Start(context.Background())
		go a.handleEngineEvents(ctx, a.engine, gen)
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
	aliasModal := views.NewAliasModalView(a.emit)
	resetAliasesModal := views.NewResetAliasesModalView(a.emit)

	a.pages.AddPage(routes.RouteDashboard, dashboardPage, true, false)
	a.pages.AddPage(routes.RouteDetail, detailPage, true, false)
	a.pages.AddPage(routes.RouteSplash, splashPage, true, false)
	a.pages.AddPage(routes.RouteThemePicker, themePickerModal, true, false)
	a.pages.AddPage(routes.RoutePortScan, portScanModal, true, false)
	a.pages.AddPage(routes.RouteInterfacePicker, interfacePickerModal, true, false)
	a.pages.AddPage(routes.RouteAliasEditor, aliasModal, true, false)
	a.pages.AddPage(routes.RouteAliasReset, resetAliasesModal, true, false)

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
	case tcell.KeyCtrlR:
		a.emit(events.AliasesResetRequested{})
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
// It exits when the context is cancelled, the generation changes, or the channel closes.
func (a *App) handleEngineEvents(ctx context.Context, engine *discovery.Engine, gen uint64) {
	defer func() {
		if r := recover(); r != nil {
			a.logger.Error("panic in handleEngineEvents", "panic", r)
		}
	}()

	for event := range engine.Events {
		if ctx.Err() != nil || a.engineGen.Load() != gen {
			return
		}
		switch event.Type {
		case discovery.EventScanStarted:
			a.emit(events.DiscoveryStarted{})
		case discovery.EventScanCompleted:
			a.emit(events.DiscoveryStopped{})
		case discovery.EventDeviceDiscovered:
			if event.Device != nil {
				a.cacheAliasForDevice(event.Device)
				a.engineMu.RLock()
				if a.engineGen.Load() == gen {
					a.state.UpsertDevice(event.Device)
				}
				a.engineMu.RUnlock()
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
			if front == routes.RouteAliasEditor {
				a.state.ClearAliasEditorDraft()
			}
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
		case events.AliasDraftChanged:
			a.state.SetAliasEditorDraft(event.Alias)
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
		case events.AliasEditRequested:
			a.openAliasEditor()
		case events.AliasSubmitted:
			a.saveAlias(event.Alias)
		case events.AliasCleared:
			a.clearAlias()
		case events.AliasesResetRequested:
			a.emit(events.NavigateTo{Route: routes.RouteAliasReset, Overlay: true})
		case events.AliasesResetConfirmed:
			a.resetAliases()
		}
		a.rerenderVisibleViews()
	}
}

// Stop shuts down app-owned resources before stopping the TUI application.
func (a *App) Stop() {
	a.stopOnce.Do(func() {
		if a.refreshTicker != nil {
			a.refreshTicker.Stop()
		}
		if a.engineCancel != nil {
			a.engineCancel()
		}
		if a.engine != nil {
			go a.engine.Stop()
		}
		if a.metaStore != nil {
			if err := a.metaStore.Close(); err != nil {
				a.logger.Warn("failed to close metadata store", "error", err)
			}
		}
		a.Application.Stop()
	})
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

	gen := a.engineGen.Add(1)

	if a.engine != nil {
		oldEngine := a.engine
		go oldEngine.Stop()
		time.Sleep(50 * time.Millisecond)
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
	go a.handleEngineEvents(ctx, engine, gen)
}

func (a *App) cacheAliasForDevice(device *discovery.Device) {
	if device == nil || a.metaStore == nil {
		return
	}

	normalizedMAC, ok := devicemeta.NormalizeMAC(device.MAC())
	if !ok || a.state.AliasLoaded(normalizedMAC) {
		return
	}

	record, found, err := a.metaStore.Get(normalizedMAC)
	if err != nil {
		a.logger.Warn("failed to load device alias", "mac", normalizedMAC, "error", err)
		return
	}

	if found {
		a.state.SetAlias(normalizedMAC, record.Alias)
		return
	}

	a.state.ClearAlias(normalizedMAC)
}

func (a *App) openAliasEditor() {
	device, ok := a.state.Selected()
	if !ok {
		return
	}

	if _, validMAC := devicemeta.NormalizeMAC(device.MAC()); !validMAC {
		a.logger.Warn("cannot edit alias for device without a valid MAC", "ip", device.IP().String())
		return
	}

	a.state.SetAliasEditorDraft(a.state.AliasFor(device))
	a.emit(events.NavigateTo{Route: routes.RouteAliasEditor, Overlay: true})
}

func (a *App) saveAlias(alias string) {
	device, normalizedMAC, ok := a.selectedDeviceMAC()
	if !ok {
		return
	}

	if err := a.metaStore.SetAlias(normalizedMAC, alias); err != nil {
		a.logger.Error("failed to save device alias", "mac", normalizedMAC, "error", err)
		a.state.SetStatusMessage("Failed to save alias", state.StatusSeverityError, statusMessageTTL)
		return
	}

	if alias == "" {
		a.state.ClearAlias(normalizedMAC)
		a.state.SetStatusMessage("Alias cleared", state.StatusSeveritySuccess, statusMessageTTL)
	} else {
		a.state.SetAlias(normalizedMAC, alias)
		a.state.SetStatusMessage("Alias saved", state.StatusSeveritySuccess, statusMessageTTL)
	}

	a.state.ClearAliasEditorDraft()
	a.emit(events.HideView{})
	a.logger.Info("device alias saved", "mac", normalizedMAC, "ip", device.IP().String())
}

func (a *App) clearAlias() {
	_, normalizedMAC, ok := a.selectedDeviceMAC()
	if !ok {
		return
	}

	if err := a.metaStore.ClearAlias(normalizedMAC); err != nil {
		a.logger.Error("failed to clear device alias", "mac", normalizedMAC, "error", err)
		a.state.SetStatusMessage("Failed to clear alias", state.StatusSeverityError, statusMessageTTL)
		return
	}

	a.state.ClearAlias(normalizedMAC)
	a.state.SetStatusMessage("Alias cleared", state.StatusSeveritySuccess, statusMessageTTL)
}

func (a *App) resetAliases() {
	if a.metaStore == nil {
		return
	}

	if err := a.metaStore.ResetAliases(); err != nil {
		a.logger.Error("failed to reset aliases", "error", err)
		a.state.SetStatusMessage("Failed to reset aliases", state.StatusSeverityError, statusMessageTTL)
		return
	}

	a.state.ResetAliases()
	a.state.SetStatusMessage("All aliases cleared", state.StatusSeveritySuccess, statusMessageTTL)
	a.emit(events.HideView{})
	a.logger.Info("all device aliases cleared")
}

func (a *App) selectedDeviceMAC() (*discovery.Device, string, bool) {
	device, ok := a.state.Selected()
	if !ok {
		return nil, "", false
	}

	normalizedMAC, validMAC := devicemeta.NormalizeMAC(device.MAC())
	if !validMAC {
		a.logger.Warn("selected device has no valid MAC address", "ip", device.IP().String())
		return nil, "", false
	}

	return device, normalizedMAC, true
}
