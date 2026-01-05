package pages

import (
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/ramonvermeulen/whosthere/internal/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/components"
	"github.com/ramonvermeulen/whosthere/internal/ui/navigation"
)

var _ navigation.Page = &DashboardPage{}

// DashboardPage is the dashboard showing discovered devices.
type DashboardPage struct {
	*tview.Flex
	deviceTable *components.DeviceTable
	spinner     *components.Spinner
	state       *state.AppState

	navigate func(route string)

	filterView *tview.TextView
	statusRow  tview.Primitive
	helpText   *tview.TextView
}

func NewDashboardPage(s *state.AppState, navigate func(route string)) *DashboardPage {
	t := components.NewDeviceTable()
	spinner := components.NewSpinner()
	spinner.SetSuffix(" Scanning...")

	main := tview.NewFlex().SetDirection(tview.FlexRow)
	main.AddItem(
		tview.NewTextView().
			SetText("whosthere").
			SetTextAlign(tview.AlignCenter),
		0, 1, false,
	)
	main.AddItem(t, 0, 18, true)

	filterView := tview.NewTextView().SetTextAlign(tview.AlignLeft)
	status := tview.NewFlex().SetDirection(tview.FlexColumn)
	helpText := tview.NewTextView().
		SetText("j/k: up/down - g/G: top/bottom - Enter: details").
		SetTextAlign(tview.AlignRight)
	status.AddItem(spinner.View(), 0, 1, false)
	status.AddItem(helpText, 0, 2, false)

	dp := &DashboardPage{
		Flex:        main,
		deviceTable: t,
		spinner:     spinner,
		state:       s,
		navigate:    navigate,
		filterView:  filterView,
		statusRow:   status,
		helpText:    helpText,
	}
	dp.updateFooter(false)

	t.OnSearchStatus(dp.handleSearchStatus)
	t.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey { return t.HandleInput(ev) })
	t.SetSelectedFunc(func(row, col int) {
		ip := t.SelectedIP()
		if ip == "" {
			return
		}
		s.SetSelectedIP(ip)
		if dp.navigate != nil {
			dp.navigate(navigation.RouteDetail)
		}
	})

	return dp
}

func (p *DashboardPage) GetName() string { return navigation.RouteDashboard }

func (p *DashboardPage) GetPrimitive() tview.Primitive { return p }

func (p *DashboardPage) FocusTarget() tview.Primitive { return p.deviceTable }

func (p *DashboardPage) Spinner() *components.Spinner { return p.spinner }

func (p *DashboardPage) RefreshFromState() {
	devices := p.state.DevicesSnapshot()
	p.deviceTable.ReplaceAll(devices)
}

func (p *DashboardPage) Refresh() {
	p.RefreshFromState()
}

func (p *DashboardPage) updateFooter(showFilter bool) {
	if p.Flex == nil || p.statusRow == nil || p.filterView == nil {
		return
	}
	p.RemoveItem(p.filterView)
	p.RemoveItem(p.statusRow)
	if showFilter {
		p.AddItem(p.filterView, 1, 0, false)
	}
	p.AddItem(p.statusRow, 1, 0, false)
}

// handleSearchStatus updates footer visibility and help text based on table search state.
func (p *DashboardPage) handleSearchStatus(status components.SearchStatus) {
	if p.filterView != nil {
		p.filterView.SetTextColor(status.Color)
		p.filterView.SetText(status.Text)
	}
	p.updateFooter(status.Showing)
}
