package components

import (
	"strings"
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
	text    string
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

func (s *Spinner) Text() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.text
}

func (s *Spinner) setText(text string) {
	s.mu.Lock()
	s.text = text
	s.mu.Unlock()
	s.TextView.SetText(text)
}

func formatSpinnerText(frame, suffix string) string {
	label := strings.TrimSpace(suffix)
	if label == "" {
		return frame
	}
	if frame == "" {
		return label
	}
	return frame + " " + label
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
				queue(func() { s.setText("") })
				return
			case <-time.After(interval):
				ch := string(frames[idx%len(frames)])
				idx++
				s.mu.Lock()
				suffix := s.suffix
				s.mu.Unlock()
				queue(func() { s.setText(formatSpinnerText(ch, suffix)) })
			}
		}
	}()
}

func (s *Spinner) Stop(queue func(f func())) {
	select {
	case s.stop <- struct{}{}:
	default:
	}
	queue(func() { s.setText("") })
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
}

func (s *Spinner) Render(_ state.ReadOnly) {
}
