package discovery

import (
	"context"
	"sync"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/oui"
	"go.uber.org/zap"
)

// Scanner defines a discovery strategy (SSDP, mDNS, ARP, etc.).
type Scanner interface {
	Name() string
	Scan(ctx context.Context, out chan<- Device) error
}

// Engine coordinates multiple scanners and merges device results.
type Engine struct {
	Scanners    []Scanner
	Sweeper     *Sweeper
	Timeout     time.Duration
	OUIRegistry *oui.Registry
}

type EngineOption func(*Engine)

func WithTimeout(d time.Duration) EngineOption {
	return func(e *Engine) { e.Timeout = d }
}

func WithOUIRegistry(r *oui.Registry) EngineOption {
	return func(e *Engine) { e.OUIRegistry = r }
}

func NewEngine(scanners []Scanner, opts ...EngineOption) *Engine {
	e := &Engine{
		Scanners: scanners,
		Timeout:  config.DefaultScanDuration,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// fillManufacturerIfEmpty fills the Manufacturer field of the device using OUI lookup if it's empty.
func (e *Engine) fillManufacturerIfEmpty(d *Device) {
	if d == nil {
		return
	}
	if e.OUIRegistry == nil {
		zap.L().Debug("OUI: registry is nil; skipping manufacturer lookup")
		return
	}
	if d.Manufacturer != "" {
		return
	}
	if d.MAC == "" {
		zap.L().Debug("OUI: device has no MAC; skipping", zap.String("ip", d.IP.String()))
		return
	}
	if org, ok := e.OUIRegistry.Lookup(d.MAC); ok {
		zap.L().Debug("OUI: setting manufacturer from OUI", zap.String("ip", d.IP.String()), zap.String("mac", d.MAC), zap.String("org", org))
		d.Manufacturer = org
	}
}

// Stream runs scanners and invokes onDevice for each incremental merged device observed.
func (e *Engine) Stream(ctx context.Context, onDevice func(*Device)) ([]Device, error) {
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
			e.handleDevice(&d, devices, onDevice)
		}
	}
}

// handleDevice processes a discovered device, merging with existing or adding new.
func (e *Engine) handleDevice(d *Device, devices map[string]*Device, onDevice func(*Device)) {
	if d.IP == nil || d.IP.String() == "" {
		return
	}
	key := d.IP.String()
	if existing, found := devices[key]; found {
		existing.Merge(d)
		e.fillManufacturerIfEmpty(existing)
		if onDevice != nil {
			onDevice(existing)
		}
	} else {
		dev := d
		if dev.FirstSeen.IsZero() {
			dev.FirstSeen = time.Now()
		}
		e.fillManufacturerIfEmpty(dev)
		devices[key] = dev
		if onDevice != nil {
			onDevice(dev)
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
