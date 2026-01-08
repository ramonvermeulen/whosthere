package pages

import (
	"fmt"
	"strings"

	"github.com/ramonvermeulen/whosthere/internal/ui/navigation"
	"github.com/rivo/tview"
)

var _ navigation.Page = (*SplashPage)(nil)

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

// SplashPage adapts the splash logo into a Page.
type SplashPage struct {
	root    *tview.Flex
	version string
}

func (p *SplashPage) Refresh() {}

func NewSplashPage(version string) *SplashPage {
	root := tview.NewFlex().SetDirection(tview.FlexRow)
	root.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	logo := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tview.Styles.SecondaryTextColor)
	logo.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	logoText := strings.Join(LogoBig, "\n")
	_, err := fmt.Fprint(logo, logoText)
	if err != nil {
		return nil
	}
	logoLines := len(strings.Split(logoText, "\n"))

	topSpacer := tview.NewTextView()
	topSpacer.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	bottomSpacer := tview.NewTextView()
	bottomSpacer.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	centeredLogo := tview.NewFlex().SetDirection(tview.FlexRow)
	centeredLogo.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	centeredLogo.AddItem(topSpacer, 0, 1, false)    // Top spacer (flexible)
	centeredLogo.AddItem(logo, logoLines, 0, false) // Logo (fixed height)
	centeredLogo.AddItem(bottomSpacer, 0, 1, false) // Bottom spacer (flexible)

	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tview.Styles.SecondaryTextColor)
	footer.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	if version != "" {
		_, _ = fmt.Fprintf(footer, "whosthere - v%s", version)
	}

	root.AddItem(centeredLogo, 0, 1, false) // Centered logo takes all available space
	root.AddItem(footer, 1, 0, false)       // Footer is 1 line at bottom

	return &SplashPage{root: root, version: version}
}

func (p *SplashPage) GetName() string { return navigation.RouteSplash }

func (p *SplashPage) GetPrimitive() tview.Primitive { return p.root }

func (p *SplashPage) FocusTarget() tview.Primitive { return p.root }

func (p *SplashPage) RefreshFromState() {}
