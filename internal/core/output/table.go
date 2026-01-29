package output

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

type Spinner struct {
	w       io.Writer
	message string
	total   time.Duration
	stop    chan struct{}
	done    chan struct{}
}

func NewSpinner(w io.Writer, message string, total time.Duration) *Spinner {
	return &Spinner{
		w:       w,
		message: message,
		total:   total,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
}

func (s *Spinner) Start() {
	go s.run()
}

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

func PrintDevices(w io.Writer, devices []*discovery.Device, elapsed time.Duration) {
	sort.Slice(devices, func(i, j int) bool {
		return discovery.CompareIPs(devices[i].IP(), devices[j].IP())
	})

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	_, _ = fmt.Fprintln(tw, "IP\tDISPLAY NAME\tMAC\tMANUFACTURER")
	_, _ = fmt.Fprintln(tw, "──\t────────────\t───\t────────────")

	for _, d := range devices {
		ip := d.IP().String()
		name := d.DisplayName()
		mac := d.MAC()
		manufacturer := d.Manufacturer()

		if name == "" {
			name = "-"
		}
		if mac == "" {
			mac = "-"
		}
		if manufacturer == "" {
			manufacturer = "-"
		}

		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", ip, name, mac, manufacturer)
	}

	_ = tw.Flush()

	_, _ = fmt.Fprintf(w, "\nScan completed: %d device(s) found in %s\n", len(devices), formatDuration(elapsed))
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.1fs", d.Seconds())
}
