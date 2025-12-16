package pages

import (
	"fmt"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/ramonvermeulen/whosthere/internal/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/navigation"
)

var _ navigation.Page = &DetailPage{}

// DetailPage shows detailed information about the currently selected device.
type DetailPage struct {
	*tview.Flex
	info  *tview.TextView
	state *state.AppState

	navigate func(route string)
}

func NewDetailPage(s *state.AppState, navigate func(route string), uiQueue func(func())) *DetailPage {
	main := tview.NewFlex().SetDirection(tview.FlexRow)
	main.AddItem(tview.NewTextView().SetText("whosthere").SetTextAlign(tview.AlignCenter), 0, 1, false)

	info := tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	info.SetBorder(true).SetTitle("Details").SetBorderColor(tview.Styles.BorderColor).SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	main.AddItem(info, 0, 18, true)

	statusFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	statusFlex.AddItem(tview.NewTextView().SetText("Esc/q: Back").SetTextAlign(tview.AlignRight), 0, 4, false)
	main.AddItem(statusFlex, 1, 0, false)

	p := &DetailPage{Flex: main, state: s, info: info, navigate: navigate}

	info.SetInputCapture(handleInput(p))

	p.Refresh()
	return p
}

func handleInput(p *DetailPage) func(ev *tcell.EventKey) *tcell.EventKey {
	return func(ev *tcell.EventKey) *tcell.EventKey {
		if ev == nil {
			return ev
		}
		switch {
		case ev.Key() == tcell.KeyEsc || ev.Rune() == 'q':
			if p.navigate != nil {
				p.navigate(navigation.RouteDashboard)
			}
			return nil
		default:
			return ev
		}
	}
}

func (p *DetailPage) GetName() string               { return navigation.RouteDetail }
func (p *DetailPage) GetPrimitive() tview.Primitive { return p }
func (p *DetailPage) FocusTarget() tview.Primitive  { return p.info }

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
	_, _ = fmt.Fprintf(p.info, "[%s::b]DisplayName:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, d.DisplayName)
	_, _ = fmt.Fprintf(p.info, "[%s::b]MAC:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, d.MAC)
	_, _ = fmt.Fprintf(p.info, "[%s::b]Manufacturer:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, d.Manufacturer)
	_, _ = fmt.Fprintf(p.info, "[%s::b]Model:[-::-] [%s::]%s[-::-]\n\n", labelColor, valueColor, d.Model)

	_, _ = fmt.Fprintf(p.info, "[%s::b]Sources:[-::-]\n", labelColor)
	if len(d.Sources) == 0 {
		_, _ = fmt.Fprintln(p.info, "  (none)")
	} else {
		for src := range d.Sources {
			_, _ = fmt.Fprintf(p.info, "  %s\n", src)
		}
	}
}

// colorToHexTag converts a tcell.Color to a tview dynamic color hex tag.
func colorToHexTag(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
