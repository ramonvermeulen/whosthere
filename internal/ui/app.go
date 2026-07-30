package ui

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
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
	askQuestionURL   = "https://github.com/ramonvermeulen/whosthere/discussions/new/choose"
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
	syncer        *devicemeta.Syncer
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
		syncer:      devicemeta.NewSyncer(cfg.Mode, metaStore),
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
	if err := a.loadStoredDevices(engine.Iface); err != nil {
		return nil, err
	}

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
		if event.Rune() == 'M' {
			a.emit(events.ToggleModeRequested{})
			return nil
		}
		if event.Rune() == 'q' || event.Rune() == 'Q' {
			// If focus is currently on an input field (e.g. alias editor),
			// let the focused primitive handle the rune instead of treating
			// it as a global quit. This prevents typing 'q' while editing
			// from exiting the application.
			if p := a.GetFocus(); p != nil {
				if _, ok := p.(*tview.InputField); ok {
					return event
				}
			}
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
				a.engineMu.RLock()
				if a.engineGen.Load() == gen {
					if err := a.syncer.SyncDevice(event.Device, a.scopeForInterface(engine.Iface)); err != nil {
						a.logger.Warn("failed to sync device metadata", "mac", event.Device.MAC(), "error", err)
					}
					a.state.UpsertDevice(event.Device)
					a.cacheAliasForDevice(event.Device)
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
		case events.ToggleModeRequested:
			if err := a.toggleMode(); err != nil {
				a.logger.Error("failed to toggle runtime mode", "error", err)
				a.state.SetStatusMessage("Failed to switch mode", state.StatusSeverityError, statusMessageTTL)
			}
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
		case events.AskQuestionRequested:
			if err := openURL(askQuestionURL); err != nil {
				a.logger.Warn("failed to open GitHub Discussions", "url", askQuestionURL, "error", err)
			}
			a.resetFocus()
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
	if err := a.loadStoredDevices(engine.Iface); err != nil {
		a.logger.Warn("failed to reload stored devices after interface switch", "error", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	a.engineCancel = cancel

	a.engine.Start(context.Background())
	go a.handleEngineEvents(ctx, engine, gen)
}

func openURL(target string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", target)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}

	return cmd.Start()
}

func (a *App) cacheAliasForDevice(device *discovery.Device) {
	if device == nil || a.metaStore == nil {
		return
	}

	normalizedMAC, hasValidMAC := devicemeta.NormalizeMAC(device.MAC())
	if !hasValidMAC || a.state.HasAliasMetadataForMAC(normalizedMAC) {
		return
	}

	record, found, err := a.metaStore.Get(a.scopeForDevice(device).ScopeID(), normalizedMAC)
	if err != nil {
		a.logger.Warn("failed to load device alias", "mac", normalizedMAC, "error", err)
		return
	}

	if found {
		a.state.SetAliasForMAC(normalizedMAC, record.Alias)
		return
	}

	a.state.ClearAliasForMAC(normalizedMAC)
}

func (a *App) loadStoredDevices(iface *discovery.InterfaceInfo) error {
	if a.syncer == nil || !a.syncer.Enabled() {
		return nil
	}

	devices, err := a.syncer.StoredDevicesForScope(a.scopeForInterface(iface))
	if err != nil {
		return err
	}
	for _, device := range devices {
		a.state.UpsertPersistedDevice(device)
		a.cacheAliasForDevice(device)
	}

	return nil
}

func (a *App) scopeForInterface(iface *discovery.InterfaceInfo) devicemeta.Scope {
	return devicemeta.ScopeFromInterfaceInfo(iface, a.cfg != nil && a.cfg.AllInterfaces)
}

func (a *App) scopeForDevice(device *discovery.Device) devicemeta.Scope {
	return a.scopeForInterface(a.currentInterfaceInfo()).ScopeForDevice(device)
}

func (a *App) toggleMode() error {
	a.engineMu.Lock()
	defer a.engineMu.Unlock()

	if a.cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	currentMode := a.cfg.Mode
	nextMode := config.ModePersistent
	if currentMode == config.ModePersistent {
		nextMode = config.ModeSession
	}

	a.cfg.Mode = nextMode
	a.syncer = devicemeta.NewSyncer(nextMode, a.metaStore)

	switch nextMode {
	case config.ModePersistent:
		if err := a.persistCurrentDevices(); err != nil {
			return err
		}
		if err := a.loadStoredDevices(a.currentInterfaceInfo()); err != nil {
			return err
		}
		a.state.SetStatusMessage("Persistent mode enabled", state.StatusSeveritySuccess, statusMessageTTL)
	case config.ModeSession:
		a.state.RemovePersistedOnlyDevices()
		a.state.SetStatusMessage("Session mode enabled", state.StatusSeveritySuccess, statusMessageTTL)
	}

	if err := config.Save(a.cfg, ""); err != nil {
		a.logger.Error("failed to save runtime mode change", "mode", nextMode, "error", err)
		a.state.SetStatusMessage("Mode switched, but failed to save config", state.StatusSeverityError, statusMessageTTL)
		return nil
	}

	a.logger.Info("runtime mode switched", "from", currentMode, "to", nextMode)
	return nil
}

func (a *App) persistCurrentDevices() error {
	if a.syncer == nil || !a.syncer.Enabled() {
		return nil
	}

	scope := a.scopeForInterface(a.currentInterfaceInfo())
	for _, device := range a.state.DevicesSnapshot() {
		if err := a.syncer.SyncDevice(device, scope); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) currentInterfaceInfo() *discovery.InterfaceInfo {
	if a.engine == nil {
		return nil
	}
	return a.engine.Iface
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

	scopeID := a.scopeForDevice(device).ScopeID()
	if err := a.metaStore.SetAlias(scopeID, normalizedMAC, alias); err != nil {
		a.logger.Error("failed to save device alias", "scope_id", scopeID, "mac", normalizedMAC, "error", err)
		a.state.SetStatusMessage("Failed to save alias", state.StatusSeverityError, statusMessageTTL)
		return
	}

	if alias == "" {
		a.state.ClearAliasForMAC(normalizedMAC)
		a.state.SetStatusMessage("Alias cleared", state.StatusSeveritySuccess, statusMessageTTL)
	} else {
		a.state.SetAliasForMAC(normalizedMAC, alias)
		a.state.SetStatusMessage("Alias saved", state.StatusSeveritySuccess, statusMessageTTL)
	}

	a.state.ClearAliasEditorDraft()
	a.emit(events.HideView{})
	a.logger.Info("device alias saved", "mac", normalizedMAC, "ip", device.IP().String())
}

func (a *App) clearAlias() {
	device, normalizedMAC, ok := a.selectedDeviceMAC()
	if !ok {
		return
	}

	scopeID := a.scopeForDevice(device).ScopeID()
	if err := a.metaStore.ClearAlias(scopeID, normalizedMAC); err != nil {
		a.logger.Error("failed to clear device alias", "scope_id", scopeID, "mac", normalizedMAC, "error", err)
		a.state.SetStatusMessage("Failed to clear alias", state.StatusSeverityError, statusMessageTTL)
		return
	}

	a.state.ClearAliasForMAC(normalizedMAC)
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
