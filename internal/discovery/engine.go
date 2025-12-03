package discovery

import (
	"context"
	"sync"
	"time"
)

// Engine coordinates multiple scanners and merges device results.
type Engine struct {
	Scanners []Scanner
	Timeout  time.Duration
}

// Stream runs scanners and invokes onDevice for each incremental merged device observed.
func (e *Engine) Stream(ctx context.Context, onDevice func(Device)) ([]Device, error) {
	ctx, cancel := context.WithTimeout(ctx, e.Timeout)
	defer cancel()

	out := make(chan Device, 128)
	var wg sync.WaitGroup

	// TODO(ramon): currently all scanners share the same channel, might be worth to have a channel per scanner
	// and launch a separate goroutine to merge results from all channels into 'out'.
	// tip: look into the "fan-in" concurrency pattern.
	for _, s := range e.Scanners {
		wg.Add(1)
		go func(sc Scanner) {
			defer wg.Done()
			_ = sc.Scan(ctx, out)
		}(s)
	}

	// Launched as goroutine to close out channel when all scanners are done
	// So that it is non-blocking for the main loop and can start processing devices immediately
	go func() {
		wg.Wait()
		close(out)
	}()

	devices := map[string]*Device{}
	for {
		select {
		case <-ctx.Done():
			return mapToSlice(devices), nil
		case d, ok := <-out:
			if !ok {
				// channel closed, all scanners are done
				return mapToSlice(devices), nil
			}
			if d.IP == nil || d.IP.String() == "" {
				// skip devices with no IP
				continue
			}
			key := d.IP.String()
			if existing, found := devices[key]; found {
				existing.Merge(d)
				if onDevice != nil {
					onDevice(*existing)
				}
			} else {
				dev := d
				devices[key] = &dev
				if onDevice != nil {
					onDevice(dev)
				}
			}
		}
	}
}

func mapToSlice(m map[string]*Device) []Device {
	res := make([]Device, 0, len(m))
	for _, v := range m {
		res = append(res, *v)
	}
	return res
}
