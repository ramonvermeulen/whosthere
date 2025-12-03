package ui

import (
	"context"
	"math"
	"time"

	"github.com/derailed/tview"
	"github.com/ramonvermeulen/whosthere/internal/config"
	"github.com/ramonvermeulen/whosthere/internal/discovery"
	"github.com/ramonvermeulen/whosthere/internal/discovery/ssdp"
	"github.com/ramonvermeulen/whosthere/internal/ui/components"
)

type App struct {
	*tview.Application
	Main        *Pages
	cfg         *config.Config
	deviceTable *DeviceTable
	engine      *discovery.Engine
	spinner     *components.Spinner
	rescanEvery time.Duration
}

func NewApp(cfg *config.Config) *App {
	a := App{
		Application: tview.NewApplication(),
		Main:        NewPages(),
		cfg:         cfg,
		deviceTable: NewDeviceTable(),
		engine:      &discovery.Engine{Scanners: []discovery.Scanner{&ssdp.Scanner{}}, Timeout: 6 * time.Second},
		spinner:     components.NewSpinner(),
		rescanEvery: 10 * time.Second,
	}
	a.layout()
	if a.cfg != nil && a.cfg.Splash.Enabled {
		a.Main.SwitchToPage("splash")
	} else {
		a.Main.SwitchToPage("main")
	}
	return &a
}

func (a *App) Run() error {
	if a.cfg != nil && a.cfg.Splash.Enabled {
		go func(delaySeconds float32) {
			ms := int64(math.Round(float64(delaySeconds) * 1000.0))
			timer := time.NewTimer(time.Duration(ms) * time.Millisecond)
			<-timer.C
			a.QueueUpdateDraw(func() {
				a.Main.SwitchToPage("main")
			})
			a.startDiscoveryLoop()
		}(a.cfg.Splash.Delay)
	} else {
		a.startDiscoveryLoop()
	}
	return a.Application.Run()
}

func (a *App) startDiscoveryLoop() {
	queue := func(f func()) { a.QueueUpdateDraw(f) }

	go func() {
		for {
			a.spinner.Start(queue)

			ctx := context.Background()
			_, _ = a.engine.Stream(ctx, func(d discovery.Device) {
				a.QueueUpdateDraw(func() { a.deviceTable.Upsert(d) })
			})

			a.spinner.Stop(queue)
			a.QueueUpdateDraw(func() { a.deviceTable.refresh() })

			time.Sleep(a.rescanEvery)
		}
	}()
}

func (a *App) layout() {
	main := tview.NewFlex().SetDirection(tview.FlexRow)
	main.AddItem(tview.NewTextView().SetText("whosthere").SetTextAlign(tview.AlignCenter), 0, 1, false)
	main.AddItem(a.deviceTable, 0, 18, true)

	status := tview.NewFlex().SetDirection(tview.FlexColumn)
	status.AddItem(a.spinner.View(), 0, 1, false)
	status.AddItem(tview.NewTextView().SetText("jK up/down - gG top/bottom").SetTextAlign(tview.AlignRight), 0, 1, false)
	main.AddItem(status, 1, 0, false)

	a.Main.AddPage("main", main, true, false)
	a.Main.AddPage("splash", NewSplash(), true, true)
	a.SetRoot(a.Main, true)
}
