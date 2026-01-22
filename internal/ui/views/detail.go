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
		SetTitle(" Details ").
		SetTitleColor(tview.Styles.TitleColor).
		SetBorderColor(tview.Styles.BorderColor).
		SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	statusBar := components.NewStatusBar()
	statusBar.SetHelp("Esc/q: Back" + components.Divider + "y: Copy IP" + components.Divider + "p: Port Scan")

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

	labelColor := utils.ColorToHexTag(tview.Styles.SecondaryTextColor)
	valueColor := utils.ColorToHexTag(tview.Styles.PrimaryTextColor)
	formatTime := func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("2006-01-02 15:04:05")
	}

	_, _ = fmt.Fprintf(d.info, "[%s::b]IP:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, device.IP)
	_, _ = fmt.Fprintf(d.info, "[%s::b]Display Name:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, device.DisplayName)
	_, _ = fmt.Fprintf(d.info, "[%s::b]MAC:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, device.MAC)
	_, _ = fmt.Fprintf(d.info, "[%s::b]Manufacturer:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, device.Manufacturer)
	_, _ = fmt.Fprintf(d.info, "[%s::b]First Seen:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, formatTime(device.FirstSeen))
	_, _ = fmt.Fprintf(d.info, "[%s::b]Last Seen:[-::-] [%s::]%s[-::-]\n\n", labelColor, valueColor, formatTime(device.LastSeen))

	_, _ = fmt.Fprintf(d.info, "[%s::b]Sources:[-::-]\n", labelColor)
	if len(device.Sources) == 0 {
		_, _ = fmt.Fprintln(d.info, "  (none)")
	} else {
		for _, src := range utils.SortedKeys(device.Sources) {
			_, _ = fmt.Fprintf(d.info, "  %s\n", src)
		}
	}

	_, _ = fmt.Fprintf(d.info, "\n[%s::b]Open Ports:[-::-]\n", labelColor)
	if len(device.OpenPorts) == 0 {
		_, _ = fmt.Fprintln(d.info, "  (none)")
	} else {
		for _, key := range utils.SortedKeys(device.OpenPorts) {
			ports := device.OpenPorts[key]
			if len(ports) > 0 {
				_, _ = fmt.Fprintf(d.info, "  [%s::b]%s:[-::-]\n", labelColor, strings.ToUpper(key))
				for _, port := range ports {
					_, _ = fmt.Fprintf(d.info, "    %d\n", port)
				}
				_, _ = fmt.Fprintf(d.info, "\n")
			}
		}
		if !device.LastPortScan.IsZero() {
			_, _ = fmt.Fprintf(d.info, "[%s::b]Last portscan:[-::-] %s\n", labelColor, device.LastPortScan.Format("2006-01-02 15:04:05"))
		}
	}

	_, _ = fmt.Fprintf(d.info, "\n[%s::b]Extra Data:[-::-]\n", labelColor)
	if len(device.ExtraData) == 0 {
		_, _ = fmt.Fprintln(d.info, "  (none)")
	} else {
		for _, k := range utils.SortedKeys(device.ExtraData) {
			_, _ = fmt.Fprintf(d.info, "  %s: %s\n", k, device.ExtraData[k])
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
