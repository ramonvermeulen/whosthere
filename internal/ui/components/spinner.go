package components

import (
	"time"

	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

type Spinner struct {
	view    *tview.TextView
	stop    chan struct{}
	running bool
	suffix  string
}

func NewSpinner() *Spinner {
	tv := tview.NewTextView().SetText(" ").SetTextAlign(tview.AlignLeft)
	theme.RegisterPrimitive(tv)
	return &Spinner{view: tv, stop: make(chan struct{}, 1), suffix: ""}
}

func (s *Spinner) View() *tview.TextView { return s.view }

func (s *Spinner) SetSuffix(suf string) { s.suffix = suf }

func (s *Spinner) Start(queue func(f func())) {
	if s.running {
		return
	}
	s.running = true

	frames := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
	interval := 100 * time.Millisecond

	select {
	case <-s.stop:
	default:
	}

	go func() {
		idx := 0
		for {
			select {
			case <-s.stop:
				s.running = false
				queue(func() { s.view.SetText("") })
				return
			case <-time.After(interval):
				ch := string(frames[idx%len(frames)])
				idx++
				queue(func() { s.view.SetText(ch + s.suffix) })
			}
		}
	}()
}

func (s *Spinner) Stop(queue func(f func())) {
	select {
	case s.stop <- struct{}{}:
	default:
	}
	queue(func() { s.view.SetText("") })
	s.running = false
}
