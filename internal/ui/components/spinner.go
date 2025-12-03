package components

import (
	"time"

	"github.com/derailed/tview"
)

type Spinner struct {
	view    *tview.TextView
	stop    chan struct{}
	running bool
}

func NewSpinner() *Spinner {
	tv := tview.NewTextView().SetText(" ").SetTextAlign(tview.AlignLeft)
	return &Spinner{view: tv, stop: make(chan struct{}, 1)}
}

func (s *Spinner) View() *tview.TextView { return s.view }

func (s *Spinner) Start(queue func(f func())) {
	if s.running {
		return
	}
	s.running = true

	frames := []rune{'|', '/', '-', '\\'}
	interval := 120 * time.Millisecond

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
				queue(func() { s.view.SetText(ch + " Scanning...") })
			}
		}
	}()
}

func (s *Spinner) Stop(queue func(f func())) {
	select {
	case s.stop <- struct{}{}:
	default:
	}
	if s.running {
		queue(func() { s.view.SetText("") })
	}
}
