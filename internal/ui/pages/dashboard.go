package pages

import (
	"strings"

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

	searchInput  string
	searching    bool
	filterView   *tview.TextView
	statusRow    tview.Primitive
	helpText     *tview.TextView
	baseHelp     string
	filterActive bool
	lastFilter   string
	filterError  bool
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
	helpMsg := "j/k: up/down - g/G: top/bottom - Enter: details"
	helpText := tview.NewTextView().SetText(helpMsg).SetTextAlign(tview.AlignRight)
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
		baseHelp:    helpMsg,
	}

	// Base layout: header + table already added; footer managed dynamically.
	dp.updateFooter(false)

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

	t.SetInputCapture(dp.handleTableInput())

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

func (p *DashboardPage) handleTableInput() func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event == nil {
			return nil
		}

		if p.searching {
			switch {
			case event.Key() == tcell.KeyEnter:
				p.searching = false
				p.applyFilter(p.searchInput)
				p.updateFooter(false)
				p.updateHelp()
				return nil
			case event.Key() == tcell.KeyEsc:
				p.searching = false
				p.updateFooter(false)
				p.updateHelp()
				return nil
			case event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2:
				if len(p.searchInput) > 0 {
					p.searchInput = p.searchInput[:len(p.searchInput)-1]
					p.applyFilter(p.searchInput)
					return nil
				}
				p.searching = false
				p.searchInput = ""
				p.applyFilter("")
				p.updateFooter(false)
				p.updateHelp()
				return nil
			default:
				if r := event.Rune(); r != 0 {
					p.searchInput += string(r)
					p.applyFilter(p.searchInput)
					return nil
				}
			}
			return nil
		}

		// Normal mode shortcuts.
		switch {
		case event.Key() == tcell.KeyEsc:
			if p.filterActive {
				p.searching = false
				p.searchInput = ""
				p.applyFilter("")
				p.filterActive = false
				p.lastFilter = ""
				p.filterError = false
				p.updateHelp()
				return nil
			}
			return event
		case event.Rune() == '/':
			p.searching = true
			p.searchInput = ""
			p.filterError = false
			p.setFilterText("/")
			return nil
		case event.Rune() == 'g':
			p.deviceTable.SelectFirst()
			return nil
		case event.Rune() == 'G':
			p.deviceTable.SelectLast()
			return nil
		}
		return event
	}
}

func (p *DashboardPage) applyFilter(pattern string) {
	pattern = strings.TrimSpace(pattern)
	if err := p.deviceTable.SetFilter(pattern); err != nil {
		p.filterError = true
		p.setFilterText(pattern)
		p.updateHelp()
		return
	}
	p.filterError = false
	if pattern == "" {
		p.setFilterText("/")
		p.filterActive = false
		p.lastFilter = ""
		p.updateHelp()
		return
	}
	p.setFilterText("/" + pattern)
	p.filterActive = true
	p.lastFilter = pattern
	p.updateHelp()
}

// setFilterText updates the filter bar text view with a prefix label and toggles visibility.
func (p *DashboardPage) setFilterText(text string) {
	if p.filterView == nil {
		return
	}
	if !p.searching {
		p.updateFooter(false)
		return
	}
	display := text
	if display == "" {
		display = "/"
	}
	color := tview.Styles.PrimaryTextColor
	if p.filterError {
		color = tcell.ColorRed
	}
	p.filterView.SetTextColor(color)
	p.filterView.SetText("Regex Search: " + display)
	p.updateFooter(true)
}

// updateFooter rebuilds the footer rows to show/hide the filter bar without shifting the table when hidden.
func (p *DashboardPage) updateFooter(showFilter bool) {
	if p.Flex == nil || p.statusRow == nil || p.filterView == nil {
		return
	}
	p.Flex.RemoveItem(p.filterView)
	p.Flex.RemoveItem(p.statusRow)
	if showFilter {
		p.Flex.AddItem(p.filterView, 1, 0, false)
	}
	p.Flex.AddItem(p.statusRow, 1, 0, false)
}

// updateHelp shows the active filter inline with the status help without moving layout.
func (p *DashboardPage) updateHelp() {
	if p.helpText == nil {
		return
	}
	if p.filterActive && p.lastFilter != "" {
		p.helpText.SetText(p.baseHelp + " | Filter: /" + p.lastFilter)
		return
	}
	p.helpText.SetText(p.baseHelp)
}

// navigateSelected replicates the table's selected handler for Enter.
func (p *DashboardPage) navigateSelected() {
	ip := p.deviceTable.SelectedIP()
	if ip == "" {
		return
	}
	p.state.SetSelectedIP(ip)
	if p.navigate != nil {
		p.navigate(navigation.RouteDetail)
	}
}
