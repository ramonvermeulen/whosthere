package pages

import (
	"fmt"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/ramonvermeulen/whosthere/internal/state"
)

// DetailPage shows detailed information about the currently selected device.
type DetailPage struct {
	root  *tview.Flex
	state *state.AppState
	info  *tview.TextView

	onBack func()
}

func NewDetailPage(s *state.AppState, onBack func()) *DetailPage {
	main := tview.NewFlex().SetDirection(tview.FlexRow)
	main.AddItem(
		tview.NewTextView().
			SetText("whosthere").
			SetTextAlign(tview.AlignCenter),
		0, 1, false,
	)
	info := tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	info.
		SetBorder(true).
		SetTitle("Details").
		SetBorderColor(tview.Styles.BorderColor).
		SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	main.AddItem(info, 0, 18, true)

	status := tview.NewFlex().SetDirection(tview.FlexColumn)
	status.AddItem(
		tview.NewTextView().
			SetText("j/k: up/down  g/G: top/bottom  Enter: details").
			SetTextAlign(tview.AlignRight),
		0, 1, false,
	)
	main.AddItem(status, 1, 0, false)

	p := &DetailPage{
		root:   main,
		state:  s,
		info:   info,
		onBack: onBack,
	}

	info.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev == nil {
			return ev
		}
		if ev.Rune() == 'q' || ev.Key() == tcell.KeyEsc {
			if p.onBack != nil {
				p.onBack()
			}
			return nil
		}
		return ev
	})

	p.Refresh()
	return p
}

func (p *DetailPage) GetName() string { return "detail" }

func (p *DetailPage) GetPrimitive() tview.Primitive { return p.root }

// FocusTarget returns the main text view so it receives input (for q/Esc) when this page is active.
func (p *DetailPage) FocusTarget() tview.Primitive { return p.info }

// tview dynamic color tags.
func colorToHexTag(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// Refresh reloads the text view from the currently selected device, if any.
func (p *DetailPage) Refresh() {
	p.info.Clear()
	d, ok := p.state.Selected()
	if !ok {
		_, _ = fmt.Fprintln(p.info, "No device selected.")
		return
	}

	labelColor := colorToHexTag(tview.Styles.SecondaryTextColor)
	valueColor := colorToHexTag(tview.Styles.PrimaryTextColor)

	_, _ = fmt.Fprintf(p.info, "[%s::b]IP:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, d.IP)
	_, _ = fmt.Fprintf(p.info, "[%s::b]Hostname:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, d.Hostname)
	_, _ = fmt.Fprintf(p.info, "[%s::b]MAC:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, d.MAC)
	_, _ = fmt.Fprintf(p.info, "[%s::b]Manufacturer:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, d.Manufacturer)
	_, _ = fmt.Fprintf(p.info, "[%s::b]Model:[-::-] [%s::]%s[-::-]\n\n", labelColor, valueColor, d.Model)

	_, _ = fmt.Fprintf(p.info, "[%s::b]Services:[-::-]\n", labelColor)
	if len(d.Services) == 0 {
		_, _ = fmt.Fprintln(p.info, "  (none)")
	} else {
		for name, port := range d.Services {
			_, _ = fmt.Fprintf(p.info, "  %s:%d\n", name, port)
		}
	}

	_, _ = fmt.Fprintf(p.info, "\n[%s::b]Sources:[-::-]\n", labelColor)
	if len(d.Sources) == 0 {
		_, _ = fmt.Fprintln(p.info, "  (none)")
	} else {
		for src := range d.Sources {
			_, _ = fmt.Fprintf(p.info, "  %s\n", src)
		}
	}
}
