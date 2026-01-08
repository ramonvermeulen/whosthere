package pages

import (
	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/components"
	"github.com/ramonvermeulen/whosthere/internal/ui/navigation"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

var _ navigation.Page = &DashboardPage{}

// DashboardPage is the dashboard showing discovered devices.
type DashboardPage struct {
	*tview.Flex
	deviceTable *components.DeviceTable
	spinner     *components.Spinner
	state       *state.AppState
	router      *navigation.Router

	header    *components.Header
	filterBar *components.FilterBar
	statusBar *components.StatusBar
	version   string
}

func NewDashboardPage(s *state.AppState, router *navigation.Router, version string) *DashboardPage {
	header := components.NewHeader(version)
	t := components.NewDeviceTable()

	main := tview.NewFlex().SetDirection(tview.FlexRow)
	main.AddItem(header, 1, 0, false)
	main.AddItem(t, 0, 1, true)

	statusBar := components.NewStatusBar()
	statusBar.Spinner().SetSuffix(" Scanning...")
	statusBar.SetHelp("j/k: up/down - g/G: top/bottom - Enter: details - Ctrl+T: theme")

	filterBar := components.NewFilterBar()

	dp := &DashboardPage{
		Flex:        main,
		deviceTable: t,
		spinner:     statusBar.Spinner(),
		state:       s,
		router:      router,
		header:      header,
		filterBar:   filterBar,
		statusBar:   statusBar,
		version:     version,
	}

	theme.RegisterPrimitive(dp)

	dp.updateFooter(false)
	t.OnSearchStatus(dp.handleSearchStatus)
	t.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey { return t.HandleInput(ev) })
	t.SetSelectedFunc(func(row, col int) {
		ip := t.SelectedIP()
		if ip == "" {
			return
		}
		s.SetSelectedIP(ip)
		if dp.router != nil {
			dp.router.NavigateTo(navigation.RouteDetail)
		}
	})

	return dp
}

func (p *DashboardPage) GetName() string { return navigation.RouteDashboard }

func (p *DashboardPage) GetPrimitive() tview.Primitive { return p }

func (p *DashboardPage) FocusTarget() tview.Primitive { return p.deviceTable }

func (p *DashboardPage) Spinner() *components.Spinner { return p.spinner }

func (p *DashboardPage) Refresh() {
	devices := p.state.DevicesSnapshot()
	p.deviceTable.ReplaceAll(devices)
}

func (p *DashboardPage) updateFooter(showFilter bool) {
	if p.Flex == nil || p.statusBar == nil || p.filterBar == nil {
		return
	}
	p.RemoveItem(p.filterBar)
	p.RemoveItem(p.statusBar.Primitive())
	if showFilter {
		p.AddItem(p.filterBar, 1, 0, false)
	}
	p.AddItem(p.statusBar.Primitive(), 1, 0, false)
}

// handleSearchStatus updates footer visibility and filter bar based on table search state.
func (p *DashboardPage) handleSearchStatus(status components.SearchStatus) {
	if p.filterBar != nil {
		if status.Showing {
			p.filterBar.Show(status.Text, status.Color)
		} else {
			p.filterBar.Clear()
		}
	}
	p.updateFooter(status.Showing)
}
