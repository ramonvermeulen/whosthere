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
	headerMetaWidth      = 42
	askQuestionLabel     = "Ask Question"
	headerActionWidth    = len(askQuestionLabel)
	modeLabelPrefix      = "mode: "
	interfaceLabelPrefix = "interface: "
)

var _ UIComponent = &Header{}

// Header renders the page title and right-aligned metadata/actions.
type Header struct {
	*tview.Flex
	title      *tview.TextView
	metaLabel  *tview.TextView
	actionLink *tview.TextView
	rightSide  *tview.Flex
}

// NewHeader creates a reusable page header.
func NewHeader(emit func(events.Event)) *Header {
	title := tview.NewTextView().
		SetTextAlign(tview.AlignLeft)
	metaLabel := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight)
	actionLink := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight)

	if emit != nil {
		actionLink.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
			if action == tview.MouseLeftClick {
				emit(events.AskQuestionRequested{})
				return action, nil
			}
			return action, event
		})
	}

	rightSide := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(metaLabel, headerMetaWidth, 0, false).
		AddItem(actionLink, headerActionWidth, 0, false)

	row := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(title, 0, 1, false).
		AddItem(rightSide, headerMetaWidth+headerActionWidth, 0, false)

	theme.RegisterPrimitive(title)
	theme.RegisterPrimitive(metaLabel)
	theme.RegisterPrimitive(actionLink)
	theme.RegisterPrimitive(rightSide)
	theme.RegisterPrimitive(row)

	return &Header{
		Flex:       row,
		title:      title,
		metaLabel:  metaLabel,
		actionLink: actionLink,
		rightSide:  rightSide,
	}
}

// Render implements UIComponent.
func (h *Header) Render(s state.ReadOnly) {
	text := baseTitle
	if version := s.Version(); version != "" {
		text = baseTitle + " - v" + version
	}
	h.title.SetText(text)
	metaText, actionText := renderHeaderMeta(s)
	h.metaLabel.SetText(metaText)
	h.actionLink.SetText(actionText)
}

func renderHeaderMeta(s state.ReadOnly) (metaText, actionText string) {
	if s == nil {
		actionText = askQuestionLabel
		return
	}

	modeLabel := strings.TrimSpace(s.Config().Mode.String())
	if modeLabel != "" {
		modeLabel = modeLabelPrefix + modeLabel
	}

	interfaceLabel := strings.TrimSpace(s.CurrentInterface())
	if interfaceLabel != "" {
		interfaceLabel = utils.Truncate(interfaceLabel, 18)
		interfaceLabel = interfaceLabelPrefix + interfaceLabel
	}

	switch {
	case modeLabel != "" && interfaceLabel != "":
		metaText = modeLabel + " " + Divider + " " + interfaceLabel
	case modeLabel != "":
		metaText = modeLabel
	default:
		metaText = interfaceLabel
	}

	if s.NoColor() {
		actionText = askQuestionLabel
		if metaText == "" {
			return
		}
		metaText = metaText + " " + Divider
		return
	}

	actionColor := utils.ColorToHexTag(tview.Styles.PrimaryTextColor)
	actionText = "[" + actionColor + "::bu]" + askQuestionLabel + "[-:-:-]"
	if metaText == "" {
		return
	}

	metaColor := utils.ColorToHexTag(tview.Styles.SecondaryTextColor)
	metaText = "[" + metaColor + "::]" + metaText + "[-:-:-] " + Divider
	return
}
