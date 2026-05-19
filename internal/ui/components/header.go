package components

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/events"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/ramonvermeulen/whosthere/internal/ui/utils"
	"github.com/rivo/tview"
)

const (
	baseTitle            = "whosthere"
	headerInterfaceWidth = 24
	askQuestionLabel     = "Ask Question"
	headerLinkWidth      = len(askQuestionLabel)
	interfaceLabelPrefix = "interface: "
)

var _ UIComponent = &Header{}

// Header renders the page title and right-aligned metadata/actions.
type Header struct {
	*tview.Flex
	title          *tview.TextView
	interfaceLabel *tview.TextView
	link           *tview.TextView
	rightSide      *tview.Flex
}

// NewHeader creates a reusable page header.
func NewHeader(emit func(events.Event)) *Header {
	title := tview.NewTextView().
		SetTextAlign(tview.AlignLeft)
	interfaceLabel := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight)
	link := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight)

	if emit != nil {
		link.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
			if action == tview.MouseLeftClick {
				emit(events.AskQuestionRequested{})
				return action, nil
			}
			return action, event
		})
	}

	rightSide := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(interfaceLabel, headerInterfaceWidth, 0, false).
		AddItem(link, headerLinkWidth, 0, false)

	row := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(title, 0, 1, false).
		AddItem(rightSide, headerInterfaceWidth+headerLinkWidth, 0, false)

	theme.RegisterPrimitive(title)
	theme.RegisterPrimitive(interfaceLabel)
	theme.RegisterPrimitive(link)
	theme.RegisterPrimitive(rightSide)
	theme.RegisterPrimitive(row)

	return &Header{
		Flex:           row,
		title:          title,
		interfaceLabel: interfaceLabel,
		link:           link,
		rightSide:      rightSide,
	}
}

// Render implements UIComponent.
func (h *Header) Render(s state.ReadOnly) {
	text := baseTitle
	if version := s.Version(); version != "" {
		text = baseTitle + " - v" + version
	}
	h.title.SetText(text)
	interfaceText, linkText := renderHeaderMeta(s)
	h.interfaceLabel.SetText(interfaceText)
	h.link.SetText(linkText)
}

func renderHeaderMeta(s state.ReadOnly) (interfaceText string, linkText string) {
	if s == nil {
		linkText = askQuestionLabel
		return
	}

	interfaceText = strings.TrimSpace(s.CurrentInterface())
	if interfaceText != "" {
		interfaceText = utils.Truncate(interfaceText, 18)
		interfaceText = interfaceLabelPrefix + interfaceText
	}

	if s.NoColor() {
		linkText = askQuestionLabel
		if interfaceText == "" {
			return
		}
		interfaceText = interfaceText + " " + Divider
		return
	}

	linkColor := utils.ColorToHexTag(tview.Styles.PrimaryTextColor)
	linkText = "[" + linkColor + "::bu]" + askQuestionLabel + "[-:-:-]"
	if interfaceText == "" {
		return
	}

	interfaceColor := utils.ColorToHexTag(tview.Styles.SecondaryTextColor)
	interfaceText = "[" + interfaceColor + "::]" + interfaceText + "[-:-:-] " + Divider
	return
}
