package views

import (
	"fmt"
	"strings"

	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/events"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

var _ View = (*SplashView)(nil)

var LogoBig = []string{
	`Knock Knock..                                                     `,
	`                _               _   _                   ___       `,
	`      __      _| |__   ___  ___| |_| |__   ___ _ __ ___/ _ \      `,
	`      \ \ /\ / / '_ \ / _ \/ __| __| '_ \ / _ \ '__/ _ \// /      `,
	`       \ V  V /| | | | (_) \__ \ |_| | | |  __/ | |  __/ \/       `,
	`        \_/\_/ |_| |_|\___/|___/\__|_| |_|\___|_|  \___| ()       `,
	"\n",
	"\n",
	"\n",
}

// SplashView adapts the splash logo into a View.
type SplashView struct {
	*tview.Flex
	footer *tview.TextView

	emit func(events.Event)
}

func NewSplashView(emit func(events.Event)) *SplashView {
	root := tview.NewFlex().SetDirection(tview.FlexRow)

	logo := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)
	logo.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	logoText := strings.Join(LogoBig, "\n")
	_, err := fmt.Fprint(logo, logoText)
	if err != nil {
		return nil
	}
	logoLines := len(strings.Split(logoText, "\n"))

	topSpacer := tview.NewTextView()

	bottomSpacer := tview.NewTextView()

	centeredLogo := tview.NewFlex().SetDirection(tview.FlexRow)
	centeredLogo.AddItem(topSpacer, 0, 1, false)
	centeredLogo.AddItem(logo, logoLines, 0, false)
	centeredLogo.AddItem(bottomSpacer, 0, 1, false)

	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	root.AddItem(centeredLogo, 0, 1, false)
	root.AddItem(footer, 1, 0, false)

	s := &SplashView{
		Flex:   root,
		emit:   emit,
		footer: footer,
	}

	theme.RegisterPrimitive(s)
	theme.RegisterPrimitive(logo)
	theme.RegisterPrimitive(footer)

	footer.SetTextColor(tview.Styles.SecondaryTextColor)
	return s
}

func (p *SplashView) Render(s state.ReadOnly) {
	version := s.Version()
	if version != "" {
		p.footer.SetText(fmt.Sprintf("v%s", version))
	}
}

func (p *SplashView) FocusTarget() tview.Primitive { return p }
