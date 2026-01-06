package components

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/ramonvermeulen/whosthere/internal/discovery"
	"go.uber.org/zap"
)

// DeviceTable wraps a tview.Table for displaying discovered devices.
type DeviceTable struct {
	*tview.Table
	devices       map[string]discovery.Device
	filterPattern string
	filterRE      *regexp.Regexp

	// live search state
	searching   bool
	searchInput string
	filterError bool

	onSearchStatus func(SearchStatus)
}

// SearchStatus describes the current regex search UI state.
type SearchStatus struct {
	Showing bool        // whether the search bar should be shown
	Text    string      // text to render in the search bar
	Color   tcell.Color // text color (e.g., red on error)
	Active  bool        // whether a filter is applied
	Filter  string      // last applied filter
	Error   bool        // whether the current input is invalid
}

func NewDeviceTable() *DeviceTable {
	t := &DeviceTable{Table: tview.NewTable(), devices: map[string]discovery.Device{}}
	t.
		SetBorder(true).
		SetTitle("Devices").
		SetBorderColor(tview.Styles.BorderColor).
		SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	t.SetFixed(1, 0)
	t.SetSelectable(true, false)
	t.Select(0, 0)
	t.refresh()
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

// handleSearchKey processes keys while in regex search mode.
func (dt *DeviceTable) handleSearchKey(ev *tcell.EventKey) *tcell.EventKey {
	switch ev.Key() {
	case tcell.KeyEnter:
		dt.searching = false
		// this allows the Enter key to propagate and trigger selection
		return ev
	case tcell.KeyEsc:
		dt.searching = false
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

// handleNormalKey processes keys while in normal mode.
func (dt *DeviceTable) handleNormalKey(ev *tcell.EventKey) *tcell.EventKey {
	switch {
	case ev.Key() == tcell.KeyEsc:
		if dt.filterPattern != "" {
			_ = dt.SetFilter("")
			return nil
		}
		return ev
	case ev.Rune() == '/':
		dt.searching = true
		if dt.filterPattern != "" {
			dt.searchInput = dt.filterPattern
		} else {
			dt.searchInput = ""
		}
		dt.filterError = false
		dt.emitStatus()
		return nil
	case ev.Rune() == 'g':
		dt.SelectFirst()
		return nil
	case ev.Rune() == 'G':
		dt.SelectLast()
		return nil
	default:
		return ev
	}
}

// OnSearchStatus registers a callback for search status changes.
func (dt *DeviceTable) OnSearchStatus(cb func(SearchStatus)) { dt.onSearchStatus = cb }

// SetFilter compiles and applies a regex filter across visible columns (case-insensitive).
func (dt *DeviceTable) SetFilter(pattern string) error {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		dt.filterPattern = ""
		dt.filterRE = nil
		dt.refresh()
		dt.SelectFirst()
		dt.filterError = false
		dt.emitStatus()
		return nil
	}
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		dt.filterError = true
		dt.emitStatus()
		return err
	}
	dt.filterPattern = pattern
	dt.filterRE = re
	dt.filterError = false
	dt.refresh()
	dt.SelectFirst()
	dt.emitStatus()
	return nil
}

// FilterPattern returns the current filter pattern, if any.
func (dt *DeviceTable) FilterPattern() string { return dt.filterPattern }

// applySearch applies an incremental search pattern, keeping the previous filter on errors.
func (dt *DeviceTable) applySearch(pattern string) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		dt.filterError = false
		_ = dt.SetFilter("")
		return
	}
	if err := dt.SetFilter(pattern); err != nil {
		// Keep last good filter, only mark error state.
		dt.filterError = true
		dt.emitStatusWith(pattern)
		return
	}
	dt.filterError = false
	dt.emitStatus()
}

// Upsert merges device and refreshes table UI.
func (dt *DeviceTable) Upsert(d *discovery.Device) {
	key := ""
	if d == nil || d.IP == nil {
		zap.L().Debug("skipping device with no IP", zap.Any("device", d))
		return
	}
	key = d.IP.String()
	if key == "" {
		zap.L().Debug("skipping device with empty IP", zap.Any("device", d))
		return
	}
	if existing, ok := dt.devices[key]; ok {
		existing.Merge(d)
		dt.devices[key] = existing
	} else {
		dt.devices[key] = *d
	}
	dt.refresh()
}

// Refresh forces a full redraw; kept for external callers like MainPage.
func (dt *DeviceTable) Refresh() { dt.refresh() }

// ReplaceAll clears the table and replaces its contents with the
// provided devices slice.
func (dt *DeviceTable) ReplaceAll(list []discovery.Device) {
	dt.devices = make(map[string]discovery.Device, len(list))
	for _, d := range list {
		if d.IP == nil || d.IP.String() == "" {
			continue
		}
		dt.devices[d.IP.String()] = d
	}
	dt.refresh()
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
			ip:           d.IP.String(),
			hostname:     d.DisplayName,
			mac:          d.MAC,
			manufacturer: d.Manufacturer,
			lastSeen:     fmtDuration(time.Since(d.LastSeen)),
		}
		if dt.filterRE != nil && !dt.rowMatches(&row) {
			continue
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ip < rows[j].ip })
	return rows
}

func (dt *DeviceTable) refresh() {
	dt.Clear()
	const maxColWidth = 30

	headers := []string{"IP", "Display Name", "MAC", "Manufacturer", "Last Seen"}

	for i, h := range headers {
		text := truncate(h, maxColWidth)
		dt.SetCell(0, i, tview.NewTableCell(text).
			SetSelectable(false).
			SetTextColor(tview.Styles.SecondaryTextColor).
			SetExpansion(1))
	}

	rows := dt.buildRows()

	for rowIndex, rowData := range rows {
		r := rowIndex + 1

		ipText := truncate(rowData.ip, maxColWidth)
		hostText := truncate(rowData.hostname, maxColWidth)
		macText := truncate(rowData.mac, maxColWidth)
		manuText := truncate(rowData.manufacturer, maxColWidth)
		seenText := truncate(rowData.lastSeen, maxColWidth)

		dt.SetCell(r, 0, tview.NewTableCell(ipText).SetExpansion(1))
		dt.SetCell(r, 1, tview.NewTableCell(hostText).SetExpansion(1))
		dt.SetCell(r, 2, tview.NewTableCell(macText).SetExpansion(1))
		dt.SetCell(r, 3, tview.NewTableCell(manuText).SetExpansion(1))
		dt.SetCell(r, 4, tview.NewTableCell(seenText).SetExpansion(1))
	}
	// Ensure selection remains within bounds after filtering.
	if dt.GetRowCount() > 1 {
		row, _ := dt.GetSelection()
		if row <= 0 || row >= dt.GetRowCount() {
			dt.Select(1, 0)
		}
	}
}

// emitStatus reports the current search status to any subscriber.
func (dt *DeviceTable) emitStatus() { dt.emitStatusWith(dt.searchInput) }

func (dt *DeviceTable) emitStatusWith(input string) {
	if dt.onSearchStatus == nil {
		return
	}

	status := SearchStatus{
		Showing: dt.searching,
		Active:  dt.filterPattern != "",
		Filter:  dt.filterPattern,
		Error:   dt.filterError,
		Color:   tview.Styles.PrimaryTextColor,
	}

	if status.Error {
		status.Color = tcell.ColorRed
	}

	if status.Showing {
		prefix := "/"
		if input != "" {
			prefix += input
		}
		status.Text = "Regex Search: " + prefix
	}

	dt.onSearchStatus(status)
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

func fmtDuration(d time.Duration) string {
	if d < time.Second {
		return "<1s"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d/time.Second))
	}
	return fmt.Sprintf("%dm", int(d/time.Minute))
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return s[:maxLen]
	}
	return s[:maxLen-1] + "…"
}
