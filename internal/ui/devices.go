package ui

import (
	"fmt"
	"sort"
	"strings"
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
	t.SetBorder(true).SetTitle("Devices")
	t.SetFixed(1, 0)
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

func (dt *DeviceTable) refresh() {
	dt.Clear()

	headers := []string{"IP", "Hostname", "MAC", "Manufacturer", "Model", "Services", "Sources", "Last Seen"}
	for i, h := range headers {
		dt.SetCell(0, i, tview.NewTableCell(h).
			SetSelectable(false).
			SetTextColor(tview.Styles.SecondaryTextColor).
			SetExpansion(1))
	}

	list := make([]discovery.Device, 0, len(dt.devices))
	for _, v := range dt.devices {
		list = append(list, v)
	}
	// TODO(ramon): maybe this can be simplified? Sorting directly while iterating?
	sort.Slice(list, func(i, j int) bool { return list[i].IP.String() < list[j].IP.String() })

	for row, d := range list {
		r := row + 1
		ip := ""
		if d.IP != nil {
			ip = d.IP.String()
		}
		seen := fmtDuration(time.Since(d.LastSeen))
		dt.SetCell(r, 0, tview.NewTableCell(ip).SetExpansion(1))
		dt.SetCell(r, 1, tview.NewTableCell(d.Hostname).SetExpansion(1))
		dt.SetCell(r, 2, tview.NewTableCell(d.MAC).SetExpansion(1))
		dt.SetCell(r, 3, tview.NewTableCell(d.Manufacturer).SetExpansion(1))
		dt.SetCell(r, 4, tview.NewTableCell(d.Model).SetExpansion(1))
		dt.SetCell(r, 5, tview.NewTableCell(formatServices(d.Services)).SetExpansion(1))
		dt.SetCell(r, 6, tview.NewTableCell(formatSources(d.Sources)).SetExpansion(1))
		dt.SetCell(r, 7, tview.NewTableCell(seen).SetExpansion(1))
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

func formatServices(svc map[string]int) string {
	if len(svc) == 0 {
		return ""
	}
	parts := make([]string, 0, len(svc))
	for name, port := range svc {
		if port > 0 {
			parts = append(parts, fmt.Sprintf("%s:%d", name, port))
		} else {
			parts = append(parts, name)
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

func formatSources(src map[string]struct{}) string {
	if len(src) == 0 {
		return ""
	}
	parts := make([]string, 0, len(src))
	for k := range src {
		parts = append(parts, k)
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}
