package output

import (
	"fmt"
	"io"
	"time"
)

// Spinner struct for displaying spinner animation
type Spinner struct {
	w       io.Writer
	message string
	total   time.Duration
	stop    chan struct{}
	done    chan struct{}
}

// NewSpinner creates a new Spinner
func NewSpinner(w io.Writer, message string, total time.Duration) *Spinner {
	return &Spinner{
		w:       w,
		message: message,
		total:   total,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	go s.run()
}

// Stop ends the spinner animation
func (s *Spinner) Stop() {
	close(s.stop)
	<-s.done
	_, _ = fmt.Fprint(s.w, "\r\033[K")
}

func (s *Spinner) run() {
	defer close(s.done)
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	start := time.Now()
	i := 0
	// Initial print to show the spinner immediately
	_, _ = fmt.Fprintf(s.w, "\r%s %s (%s/%.0fs)", frames[i%len(frames)], s.message, formatDuration(time.Since(start)), s.total.Seconds())
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			elapsed := time.Since(start)
			_, _ = fmt.Fprintf(s.w, "\r%s %s (%s/%.0fs)", frames[i%len(frames)], s.message, formatDuration(elapsed), s.total.Seconds())
			i++
		}
	}
}
