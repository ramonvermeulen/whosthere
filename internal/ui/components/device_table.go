package components

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/events"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/ramonvermeulen/whosthere/internal/ui/utils"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/rivo/tview"
)

var _ UIComponent = &DeviceTable{}

// DeviceTable wraps a tview.Table for displaying discovered devices.
type DeviceTable struct {
	*tview.Table
	devices     []*discovery.Device
	filterRE    *regexp.Regexp
	searching   bool
	searchInput string

	emit func(events.Event)
}

func NewDeviceTable(emit func(events.Event)) *DeviceTable {
	t := &DeviceTable{Table: tview.NewTable(), devices: []*discovery.Device{}, emit: emit}
	t.
		SetBorder(true).
		SetTitle(" Devices ")
	t.SetFixed(1, 0)
	t.SetSelectable(true, false)

	theme.RegisterPrimitive(t.Table)

	return t
}

// HandleInput processes vim-style search input and table shortcuts. It returns the
// event to continue default handling, or nil if consumed.
func (dt *DeviceTable) HandleInput(ev *tcell.EventKey) *tcell.EventKey {
	if ev == nil {
		return nil
	}

	if dt.searching {
		return dt.handleSearchKey(ev)
	}
	return dt.handleNormalKey(ev)
}

// handleNormalKey processes keys while in normal mode.
func (dt *DeviceTable) handleNormalKey(ev *tcell.EventKey) *tcell.EventKey {
	switch {
	case ev.Key() == tcell.KeyEsc:
		if dt.filterRE != nil {
			dt.applySearch("")
			return nil
		}
		return ev
	case ev.Rune() == '/':
		dt.searching = true
		dt.searchInput = ""
		dt.emit(events.SearchStarted{})
		return nil
	case ev.Rune() == 'g':
		dt.SelectFirst()
		return nil
	case ev.Rune() == 'G':
		dt.SelectLast()
		return nil
	case ev.Rune() == 'y':
		ip := dt.SelectedIP()
		if ip != "" {
			dt.emit(events.CopyIP{IP: ip})
		}
		return nil
	case ev.Rune() == 'Y':
		mac := dt.SelectedMAC()
		if mac != "" {
			dt.emit(events.CopyMac{MAC: mac})
		}
		return nil
	default:
		return ev
	}
}

// handleSearchKey processes keys while in search mode.
func (dt *DeviceTable) handleSearchKey(ev *tcell.EventKey) *tcell.EventKey {
	switch ev.Key() {
	case tcell.KeyEnter:
		dt.searching = false
		dt.emit(events.SearchFinished{})
		return nil
	case tcell.KeyEsc:
		dt.searching = false
		dt.applySearch("")
		dt.emit(events.SearchFinished{})
		return nil
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if dt.searchInput != "" {
			dt.searchInput = dt.searchInput[:len(dt.searchInput)-1]
			dt.applySearch(dt.searchInput)
			return nil
		}
		dt.searching = false
		dt.searchInput = ""
		_ = dt.SetFilter("")
		dt.emit(events.SearchFinished{})
		return nil
	default:
		if r := ev.Rune(); r != 0 {
			dt.searchInput += string(r)
			dt.applySearch(dt.searchInput)
			return nil
		}
	}
	return nil
}

// SetFilter compiles and applies a regex filter across visible columns (case-insensitive).
func (dt *DeviceTable) SetFilter(pattern string) error {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		dt.filterRE = nil
		dt.refresh()
		return nil
	}
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return err
	}
	dt.filterRE = re
	dt.refresh()
	return nil
}

// applySearch applies an incremental search pattern, keeping the previous filter on errors.
func (dt *DeviceTable) applySearch(pattern string) {
	pattern = strings.TrimSpace(pattern)
	if dt.emit != nil {
		dt.emit(events.FilterChanged{Pattern: pattern})
	}
	if pattern == "" {
		dt.emit(events.SearchError{Error: false})
		_ = dt.SetFilter("")
		return
	}
	if err := dt.SetFilter(pattern); err != nil {
		dt.emit(events.SearchError{Error: true})
		return
	}
	dt.emit(events.SearchError{Error: false})
}

// Render updates the table with the latest devices from state.
func (dt *DeviceTable) Render(st state.ReadOnly) {
	dt.devices = st.DevicesSnapshot()
	_ = dt.SetFilter(st.FilterPattern())
}

// SelectedIP returns the IP for the currently selected row, if any.
func (dt *DeviceTable) SelectedIP() string {
	row, _ := dt.GetSelection()
	if row <= 0 {
		return ""
	}
	cell := dt.GetCell(row, 0)
	if cell == nil {
		return ""
	}
	return cell.Text
}

// SelectedMAC returns the MAC for the currently selected row, if any.
func (dt *DeviceTable) SelectedMAC() string {
	row, _ := dt.GetSelection()
	if row <= 0 {
		return ""
	}
	cell := dt.GetCell(row, 2)
	return cell.Text
}

// SelectFirst selects the first data row below the header, if any.
func (dt *DeviceTable) SelectFirst() {
	if dt.GetRowCount() > 1 {
		dt.Select(1, 0)
	}
}

// SelectLast selects the last data row.
func (dt *DeviceTable) SelectLast() {
	rows := dt.GetRowCount()
	if rows > 1 {
		dt.Select(rows-1, 0)
	}
}

type tableRow struct {
	ip, hostname, mac, manufacturer, lastSeen string
}

func (dt *DeviceTable) buildRows() []tableRow {
	rows := make([]tableRow, 0, len(dt.devices))
	for _, d := range dt.devices {
		row := tableRow{
			ip:           d.IP().String(),
			hostname:     d.DisplayName(),
			mac:          d.MAC(),
			manufacturer: d.Manufacturer(),
			lastSeen:     utils.FmtDuration(time.Since(d.LastSeen())),
		}
		if dt.filterRE != nil && !dt.rowMatches(&row) {
			continue
		}
		rows = append(rows, row)
	}
	return rows
}

func (dt *DeviceTable) refresh() {
	selectedIP := dt.SelectedIP()
	dt.Clear()
	const maxColWidth = 30

	headers := []string{"IP", "Display Name", "MAC", "Manufacturer", "Last Seen"}

	for i, h := range headers {
		text := utils.Truncate(h, maxColWidth)
		dt.SetCell(0, i, tview.NewTableCell(text).
			SetSelectable(false).
			SetTextColor(tview.Styles.SecondaryTextColor).
			SetExpansion(1))
	}

	rows := dt.buildRows()

	title := fmt.Sprintf(" Devices (%v) ", len(rows))
	if dt.filterRE != nil {
		title += fmt.Sprintf(" [%s]<%s>[-] ", utils.ColorToHexTag(tview.Styles.SecondaryTextColor), dt.filterRE.String())
	}
	dt.SetTitle(title)

	for rowIndex, rowData := range rows {
		r := rowIndex + 1

		ipText := utils.Truncate(rowData.ip, maxColWidth)
		hostText := utils.Truncate(rowData.hostname, maxColWidth)
		macText := utils.Truncate(rowData.mac, maxColWidth)
		manuText := utils.Truncate(rowData.manufacturer, maxColWidth)
		seenText := utils.Truncate(rowData.lastSeen, maxColWidth)

		dt.SetCell(r, 0, tview.NewTableCell(ipText).SetExpansion(1))
		dt.SetCell(r, 1, tview.NewTableCell(hostText).SetExpansion(1))
		dt.SetCell(r, 2, tview.NewTableCell(macText).SetExpansion(1))
		dt.SetCell(r, 3, tview.NewTableCell(manuText).SetExpansion(1))
		dt.SetCell(r, 4, tview.NewTableCell(seenText).SetExpansion(1))
	}
	// Restore selection if possible, otherwise select first.
	if dt.GetRowCount() > 1 {
		selectedRow := -1
		for i, row := range rows {
			if utils.Truncate(row.ip, maxColWidth) == selectedIP {
				selectedRow = i + 1 // +1 for header
				break
			}
		}
		if selectedRow > 0 && selectedRow < dt.GetRowCount() {
			dt.Select(selectedRow, 0)
		} else {
			dt.Select(1, 0)
		}
	}
}

func (dt *DeviceTable) rowMatches(r *tableRow) bool {
	if dt.filterRE == nil {
		return true
	}
	return dt.filterRE.MatchString(r.ip) ||
		dt.filterRE.MatchString(r.hostname) ||
		dt.filterRE.MatchString(r.mac) ||
		dt.filterRE.MatchString(r.manufacturer) ||
		dt.filterRE.MatchString(r.lastSeen)
}
