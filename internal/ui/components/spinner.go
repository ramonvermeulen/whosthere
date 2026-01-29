package components

import (
	"sync"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

var _ UIComponent = &Spinner{}

type Spinner struct {
	*tview.TextView
	mu      sync.Mutex
	stop    chan struct{}
	running bool
	suffix  string
}

func NewSpinner() *Spinner {
	tv := tview.NewTextView().SetText(" ").SetTextAlign(tview.AlignLeft)
	theme.RegisterPrimitive(tv)
	return &Spinner{TextView: tv, stop: make(chan struct{}, 1), suffix: ""}
}

func (s *Spinner) SetSuffix(suf string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.suffix = suf
}

func (s *Spinner) Start(queue func(f func())) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

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
				s.mu.Lock()
				s.running = false
				s.mu.Unlock()
				queue(func() { s.SetText("") })
				return
			case <-time.After(interval):
				ch := string(frames[idx%len(frames)])
				idx++
				s.mu.Lock()
				suffix := s.suffix
				s.mu.Unlock()
				queue(func() { s.SetText(ch + suffix) })
			}
		}
	}()
}

func (s *Spinner) Stop(queue func(f func())) {
	select {
	case s.stop <- struct{}{}:
	default:
	}
	queue(func() { s.SetText("") })
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
}

func (s *Spinner) Render(_ state.ReadOnly) {
}
