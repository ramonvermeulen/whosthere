package views

import (
	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/components"
	"github.com/ramonvermeulen/whosthere/internal/ui/events"
	"github.com/ramonvermeulen/whosthere/internal/ui/routes"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

var _ View = &DashboardView{}

// DashboardView is the dashboard showing discovered devices.
type DashboardView struct {
	*tview.Flex
	header      *components.Header
	deviceTable *components.DeviceTable
	filterBar   *components.FilterBar
	statusBar   *components.StatusBar

	emit  func(e events.Event)
	queue func(f func())
}

func NewDashboardView(emit func(events.Event), queue func(f func())) *DashboardView {
	header := components.NewHeader()
	t := components.NewDeviceTable(emit)

	main := tview.NewFlex().SetDirection(tview.FlexRow)
	main.AddItem(header, 1, 0, false)
	main.AddItem(t, 0, 1, true)

	statusBar := components.NewStatusBar()
	statusBar.Spinner().SetSuffix(" Discovering Devices...")
	statusBar.SetHelp(
		"j/k: up/down" + components.Divider +
		"Enter: details" + components.Divider +
		"y: copy" + components.Divider +
		"Ctrl+T: theme" + components.Divider +
		"q: quit",
	)


	filterBar := components.NewFilterBar()

	d := &DashboardView{
		Flex:        main,
		header:      header,
		deviceTable: t,
		filterBar:   filterBar,
		statusBar:   statusBar,
		emit:        emit,
		queue:       queue,
	}

	theme.RegisterPrimitive(d)

	d.updateFooter(false)
	t.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey { return t.HandleInput(ev) })
	t.SetSelectedFunc(func(row, col int) {
		ip := t.SelectedIP()
		if ip == "" {
			return
		}
		d.emit(events.DeviceSelected{IP: ip})
		d.emit(events.NavigateTo{Route: routes.RouteDetail})
	})

	return d
}

func (d *DashboardView) FocusTarget() tview.Primitive { return d.deviceTable }

func (d *DashboardView) Render(s state.ReadOnly) {
	d.deviceTable.Render(s)
	d.header.Render(s)
	d.filterBar.Render(s)
	d.statusBar.Render(s)

	d.updateFooter(s.SearchActive())

	if s.IsDiscovering() {
		d.statusBar.Spinner().Start(d.queue)
	} else {
		d.statusBar.Spinner().Stop(d.queue)
	}
}

func (d *DashboardView) updateFooter(showFilter bool) {
	if d.Flex == nil || d.statusBar == nil || d.filterBar == nil {
		return
	}
	d.RemoveItem(d.filterBar)
	d.RemoveItem(d.statusBar)
	if showFilter {
		d.AddItem(d.filterBar, 1, 0, false)
	}
	d.AddItem(d.statusBar, 1, 0, false)
}
