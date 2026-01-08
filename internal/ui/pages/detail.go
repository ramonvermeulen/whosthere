package pages

import (
	"fmt"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/components"
	"github.com/ramonvermeulen/whosthere/internal/ui/navigation"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

var _ navigation.Page = &DetailPage{}

// DetailPage shows detailed information about the currently selected device.
type DetailPage struct {
	*tview.Flex
	info      *tview.TextView
	state     *state.AppState
	header    *components.Header
	statusBar *components.StatusBar

	navigate func(route string)
}

func NewDetailPage(s *state.AppState, navigate func(route string), version string) *DetailPage {
	main := tview.NewFlex().SetDirection(tview.FlexRow)
	header := components.NewHeader(version)

	info := tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	info.SetBorder(true).
		SetTitle("Details").
		SetTitleColor(tview.Styles.TitleColor).
		SetBorderColor(tview.Styles.BorderColor).
		SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	statusBar := components.NewStatusBar()
	statusBar.SetHelp("Esc/q: Back")

	main.AddItem(header, 0, 1, false)
	main.AddItem(info, 0, 18, true)
	main.AddItem(statusBar.Primitive(), 1, 0, false)

	p := &DetailPage{
		Flex:      main,
		state:     s,
		info:      info,
		header:    header,
		statusBar: statusBar,
		navigate:  navigate,
	}

	info.SetInputCapture(handleInput(p))

	theme.RegisterPrimitive(p)
	theme.RegisterPrimitive(p.info)

	p.Refresh()
	return p
}

func handleInput(p *DetailPage) func(ev *tcell.EventKey) *tcell.EventKey {
	return func(ev *tcell.EventKey) *tcell.EventKey {
		if ev == nil {
			return nil
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
	formatTime := func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("2006-01-02 15:04:05")
	}

	_, _ = fmt.Fprintf(p.info, "[%s::b]IP:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, d.IP)
	_, _ = fmt.Fprintf(p.info, "[%s::b]Display Name:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, d.DisplayName)
	_, _ = fmt.Fprintf(p.info, "[%s::b]MAC:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, d.MAC)
	_, _ = fmt.Fprintf(p.info, "[%s::b]Manufacturer:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, d.Manufacturer)
	_, _ = fmt.Fprintf(p.info, "[%s::b]First Seen:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, formatTime(d.FirstSeen))
	_, _ = fmt.Fprintf(p.info, "[%s::b]Last Seen:[-::-] [%s::]%s[-::-]\n\n", labelColor, valueColor, formatTime(d.LastSeen))

	_, _ = fmt.Fprintf(p.info, "[%s::b]Sources:[-::-]\n", labelColor)
	if len(d.Sources) == 0 {
		_, _ = fmt.Fprintln(p.info, "  (none)")
	} else {
		for _, src := range sortedKeys(d.Sources) {
			_, _ = fmt.Fprintf(p.info, "  %s\n", src)
		}
	}

	_, _ = fmt.Fprintf(p.info, "\n[%s::b]Extra Data:[-::-]\n", labelColor)
	if len(d.ExtraData) == 0 {
		_, _ = fmt.Fprintln(p.info, "  (none)")
	} else {
		for _, k := range sortedKeys(d.ExtraData) {
			_, _ = fmt.Fprintf(p.info, "  %s: %s\n", k, d.ExtraData[k])
		}
	}
}

// colorToHexTag converts a tcell.Color to a tview dynamic color hex tag.
func colorToHexTag(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// sortedKeys is a helper to return asc sorted map keys.
func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
