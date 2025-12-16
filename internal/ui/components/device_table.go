package components

import (
	"fmt"
	"sort"
	"time"

	"github.com/derailed/tview"
	"github.com/ramonvermeulen/whosthere/internal/discovery"
	"go.uber.org/zap"
)

// DeviceTable wraps a tview.Table for displaying discovered devices.
type DeviceTable struct {
	*tview.Table
	devices map[string]discovery.Device
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

// Upsert merges device and refreshes table UI.
func (dt *DeviceTable) Upsert(d discovery.Device) {
	key := ""
	if d.IP != nil {
		key = d.IP.String()
	}
	if key == "" {
		zap.L().Debug("skipping device with no IP", zap.Any("device", d))
		return
	}
	if existing, ok := dt.devices[key]; ok {
		existing.Merge(d)
		dt.devices[key] = existing
	} else {
		dt.devices[key] = d
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
	ip, hostname, mac, manufacturer, model, lastSeen string
}

func (dt *DeviceTable) buildRows() []tableRow {
	rows := make([]tableRow, 0, len(dt.devices))
	for _, d := range dt.devices {
		rows = append(rows, tableRow{
			ip:           d.IP.String(),
			hostname:     d.DisplayName,
			mac:          d.MAC,
			manufacturer: d.Manufacturer,
			model:        d.Model,
			lastSeen:     fmtDuration(time.Since(d.LastSeen)),
		})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ip < rows[j].ip })
	return rows
}

func (dt *DeviceTable) refresh() {
	dt.Clear()
	const maxColWidth = 30

	headers := []string{"IP", "DisplayName", "MAC", "Manufacturer", "Model", "Last Seen"}

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
		modelText := truncate(rowData.model, maxColWidth)
		seenText := truncate(rowData.lastSeen, maxColWidth)

		dt.SetCell(r, 0, tview.NewTableCell(ipText).SetExpansion(1))
		dt.SetCell(r, 1, tview.NewTableCell(hostText).SetExpansion(1))
		dt.SetCell(r, 2, tview.NewTableCell(macText).SetExpansion(1))
		dt.SetCell(r, 3, tview.NewTableCell(manuText).SetExpansion(1))
		dt.SetCell(r, 4, tview.NewTableCell(modelText).SetExpansion(1))
		dt.SetCell(r, 5, tview.NewTableCell(seenText).SetExpansion(1))
	}
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

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 1 {
		return s[:max]
	}
	return s[:max-1] + "…"
}
