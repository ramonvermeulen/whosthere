package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/components"
	"github.com/ramonvermeulen/whosthere/internal/ui/events"
	"github.com/ramonvermeulen/whosthere/internal/ui/routes"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/ramonvermeulen/whosthere/internal/ui/utils"
	"github.com/rivo/tview"
)

var _ View = &DetailView{}

// DetailView shows detailed information about the currently selected device.
type DetailView struct {
	*tview.Flex
	info      *tview.TextView
	header    *components.Header
	statusBar *components.StatusBar

	emit  func(events.Event)
	queue func(f func())
}

func NewDetailView(emit func(events.Event), queue func(f func())) *DetailView {
	main := tview.NewFlex().SetDirection(tview.FlexRow)
	header := components.NewHeader()

	info := tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	info.SetBorder(true).
		SetTitle(" Details ")

	statusBar := components.NewStatusBar()
	statusBar.SetHelp("Esc/q: Back" + components.Divider + "y/Y: Copy IP/MAC" + components.Divider + "p: Port Scan")

	main.AddItem(header, 1, 0, false)
	main.AddItem(info, 0, 1, true)
	main.AddItem(statusBar, 1, 0, false)

	p := &DetailView{
		Flex:      main,
		info:      info,
		header:    header,
		statusBar: statusBar,
		emit:      emit,
		queue:     queue,
	}

	info.SetInputCapture(handleInput(p))

	theme.RegisterPrimitive(p)
	theme.RegisterPrimitive(p.info)

	return p
}

func handleInput(p *DetailView) func(ev *tcell.EventKey) *tcell.EventKey {
	return func(ev *tcell.EventKey) *tcell.EventKey {
		if ev == nil {
			return nil
		}
		switch {
		case ev.Key() == tcell.KeyEsc || ev.Rune() == 'q':
			p.emit(events.NavigateTo{Route: routes.RouteDashboard, Overlay: true})
			return nil
		case ev.Rune() == 'p':
			p.emit(events.NavigateTo{Route: routes.RoutePortScan, Overlay: true})
			return nil
		case ev.Rune() == 'y':
			p.emit(events.CopyIP{})
			return nil
		case ev.Rune() == 'Y':
			p.emit(events.CopyMac{})
			return nil
		default:
			return ev
		}
	}
}

func (d *DetailView) FocusTarget() tview.Primitive { return d.info }

// Render reloads the text view from the currently selected device, if any.
func (d *DetailView) Render(s state.ReadOnly) {
	d.info.Clear()
	device, ok := s.Selected()
	if !ok {
		_, _ = fmt.Fprintln(d.info, "No device selected.")
		return
	}

	noColor := s.NoColor()
	var labelColor, valueColor string
	if !noColor {
		labelColor = utils.ColorToHexTag(tview.Styles.SecondaryTextColor)
		valueColor = utils.ColorToHexTag(tview.Styles.PrimaryTextColor)
	}

	writeLine := func(label, value string) {
		if noColor {
			_, _ = fmt.Fprintf(d.info, "%s: %s\n", label, value)
		} else {
			_, _ = fmt.Fprintf(d.info, "[%s::b]%s:[-::-] [%s::]%s[-::-]\n", labelColor, label, valueColor, value)
		}
	}

	writeSection := func(label string) {
		if noColor {
			_, _ = fmt.Fprintf(d.info, "%s:\n", label)
		} else {
			_, _ = fmt.Fprintf(d.info, "[%s::b]%s:[-::-]\n", labelColor, label)
		}
	}

	writeProto := func(key string) {
		if noColor {
			_, _ = fmt.Fprintf(d.info, "  %s:\n", strings.ToUpper(key))
		} else {
			_, _ = fmt.Fprintf(d.info, "  [%s::b]%s:[-::-]\n", labelColor, strings.ToUpper(key))
		}
	}

	writeLastScan := func(timeStr string) {
		if noColor {
			_, _ = fmt.Fprintf(d.info, "Last portscan: %s\n", timeStr)
		} else {
			_, _ = fmt.Fprintf(d.info, "[%s::b]Last portscan:[-::-] %s\n", labelColor, timeStr)
		}
	}

	formatTime := func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("2006-01-02 15:04:05")
	}

	writeLine("IP", device.IP.String())
	writeLine("Display Name", device.DisplayName)
	writeLine("MAC", device.MAC)
	writeLine("Manufacturer", device.Manufacturer)
	writeLine("First Seen", formatTime(device.FirstSeen))
	writeLine("Last Seen", formatTime(device.LastSeen))
	_, _ = fmt.Fprintln(d.info)

	writeSection("Sources")
	if len(device.Sources) == 0 {
		_, _ = fmt.Fprintln(d.info, "  (none)")
	} else {
		for _, src := range utils.SortedKeys(device.Sources) {
			_, _ = fmt.Fprintf(d.info, "  %s\n", src)
		}
	}

	_, _ = fmt.Fprintln(d.info)
	writeSection("Open Ports")
	if len(device.OpenPorts) == 0 {
		_, _ = fmt.Fprintln(d.info, "  (no ports scanned yet)")
	} else {
		for _, key := range utils.SortedKeys(device.OpenPorts) {
			ports := device.OpenPorts[key]
			if len(ports) > 0 {
				writeProto(key)
				for _, port := range ports {
					_, _ = fmt.Fprintf(d.info, "    %d\n", port)
				}
				_, _ = fmt.Fprintln(d.info)
			}
		}
		if !device.LastPortScan.IsZero() {
			writeLastScan(device.LastPortScan.Format("2006-01-02 15:04:05"))
		}
	}

	_, _ = fmt.Fprintln(d.info)
	writeSection("Extra Data")
	if len(device.ExtraData) == 0 {
		_, _ = fmt.Fprintln(d.info, "  (none)")
	} else {
		for _, k := range utils.SortedKeys(device.ExtraData) {
			_, _ = fmt.Fprintf(d.info, "  %s: %s\n", k, utils.SanitizeString(device.ExtraData[k]))
		}
	}

	d.header.Render(s)
	d.statusBar.Render(s)

	switch {
	case s.IsPortscanning():
		d.statusBar.Spinner().SetSuffix(" Port scanning...")
		d.statusBar.Spinner().Start(d.queue)
	case s.IsDiscovering():
		d.statusBar.Spinner().SetSuffix(" Discovering Devices...")
		d.statusBar.Spinner().Start(d.queue)
	default:
		d.statusBar.Spinner().Stop(d.queue)
	}
}
