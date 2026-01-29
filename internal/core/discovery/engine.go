package discovery

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/discovery/oui"
)

const (
	DefaultScanInterval  = 20 * time.Second
	DefaultScanTimeout   = 10 * time.Second
	DefaultSweepInterval = 5 * time.Minute
	DefaultSweepTimeout  = 20 * time.Second
	DefaultEventBuf      = 512
)

var (
	ErrNoScannersOrSweeper = errors.New("no scanners or sweeper configured; at least one is required")
	ErrNoInterface         = errors.New("no network interface provided")
)

// Logger defines a simple logging interface for the engine.
// This allows plugging in different loggers as long they are compatible with slog.
type Logger interface {
	Log(ctx context.Context, level slog.Level, msg string, args ...any)
}

// NoOpLogger is a logger that does nothing.
// Useful as default logger to avoid nil checks.
type NoOpLogger struct{}

// Log implements the Logger interface but does nothing.
// This is useful as a default logger to avoid nil checks.
func (n NoOpLogger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {}

// Scanner defines a discovery strategy (SSDP, mDNS, ARP, etc.).
// Additional scanners can be implemented by satisfying this interface.
type Scanner interface {
	Name() string
	Scan(ctx context.Context, out chan<- Device) error
}

// Sweeper defines a sweeper that triggers ARP requests to populate the ARP cache.
// This is typically done by sending packets to IPs in the target subnet.
// The sweeper can be overridden by providing a custom implementation.
type Sweeper interface {
	Start(ctx context.Context)
}

// Engine coordinates multiple scanners and merges device results.
// It exposes two read-only channels: Devices for discovered devices and Events for scan lifecycle.
type Engine struct {
	// Events is a read-only channel for all events
	Events <-chan Event
	// Private write channel
	events chan Event

	scanners []Scanner
	sweeper  Sweeper
	// todo: what to do with this public field?
	Iface         *InterfaceInfo
	sweepInterval time.Duration
	sweepTimeout  time.Duration
	scanInterval  time.Duration
	scanTimeout   time.Duration
	ouiRegistry   *oui.Registry
	logger        Logger
	maxDevices    int

	mu      sync.RWMutex
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running bool
}

func New(opts ...Option) (*Engine, error) {
	e := &Engine{
		scanInterval:  DefaultScanInterval,
		scanTimeout:   DefaultScanTimeout,
		sweepInterval: DefaultSweepInterval,
		sweepTimeout:  DefaultSweepTimeout,
		logger:        &NoOpLogger{},
	}

	for _, opt := range opts {
		if err := opt(e); err != nil {
			return nil, err
		}
	}

	// these are essential components, so when missing we return an error
	if len(e.scanners) == 0 && e.sweeper == nil {
		return nil, ErrNoScannersOrSweeper
	}
	if e.Iface == nil {
		return nil, ErrNoInterface
	}

	e.events = make(chan Event, DefaultEventBuf)
	e.Events = e.events

	return e, nil
}

// Run starts continuous discovery
func (e *Engine) Run(ctx context.Context) <-chan Event {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return e.Events
	}

	ctx, cancel := context.WithCancel(ctx)
	e.cancel = cancel
	e.running = true

	e.emit(Event{Type: EventEngineStarted})

	if e.sweeper != nil {
		e.wg.Add(1)
		go func() {
			defer e.wg.Done()
			e.sweeper.Start(ctx)
		}()
	}

	e.wg.Add(1)
	go e.runScanLoop(ctx)

	return e.Events
}

// Stop gracefully stops the engine
func (e *Engine) Stop() {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return
	}

	cancel := e.cancel
	e.running = false
	e.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	e.wg.Wait()

	e.emit(Event{Type: EventEngineStopped})
	close(e.events)
}

// Scan performs a one-time scan
func (e *Engine) Scan(ctx context.Context) ([]Device, error) {
	return e.performScan(ctx, false)
}

// runScanLoop runs continuous scans at interval
func (e *Engine) runScanLoop(ctx context.Context) {
	defer e.wg.Done()

	e.performScan(ctx, true)

	if e.scanInterval <= 0 {
		return
	}

	ticker := time.NewTicker(e.scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.performScan(ctx, true)
		}
	}
}

func (e *Engine) performScan(ctx context.Context, continuous bool) ([]Device, error) {
	e.emit(NewScanStartedEvent())
	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, e.scanTimeout)
	defer cancel()

	scannerOut := make(chan Device, e.maxDevices)
	var scannerWg sync.WaitGroup

	// Start scanners
	for _, scanner := range e.scanners {
		scannerWg.Add(1)
		go func(s Scanner) {
			defer scannerWg.Done()
			_ = s.Scan(ctx, scannerOut) // Ignore errors for now
		}(scanner)
	}

	// Close channel when scanners done
	go func() {
		scannerWg.Wait()
		close(scannerOut)
	}()

	// Process until channel closes
	devices := make(map[string]*Device)
	var mu sync.Mutex
	for device := range scannerOut {
		e.processDevice(&device, devices, &mu, continuous)
	}

	stats := &ScanStats{
		DeviceCount: len(devices),
		Duration:    time.Since(start),
	}
	e.emit(NewScanCompletedEvent(stats))

	return mapToSlice(devices), ctx.Err()
}

// processDevice handles a single discovered device
func (e *Engine) processDevice(d *Device, devices map[string]*Device, mu *sync.Mutex, continuous bool) {
	if d.IP == nil || d.IP.String() == "" {
		return
	}

	key := d.IP.String()

	// Lock for map access if in continuous mode
	if continuous {
		mu.Lock()
		defer mu.Unlock()
	}

	// Merge with existing or add new
	if existing, found := devices[key]; found {
		existing.Merge(d)
		e.fillManufacturer(existing)
		d = existing
	} else {
		if d.FirstSeen.IsZero() {
			d.FirstSeen = time.Now()
		}
		e.fillManufacturer(d)
		devices[key] = d
	}

	// Emit immediately (streaming pattern)
	e.emit(Event{
		Type:   EventDeviceDiscovered,
		Device: d,
	})
}

// emit sends an event non-blocking
func (e *Engine) emit(event Event) {
	select {
	case e.events <- event:
		// Success
	default:
		// Channel full
		if event.Type == EventError && event.Error != nil {
			e.logger.Log(context.Background(), slog.LevelWarn, "event channel full, dropping error")
		}
	}
}

// fillManufacturer fills the Manufacturer field using OUI lookup if empty.
func (e *Engine) fillManufacturer(d *Device) {
	if d == nil || e.ouiRegistry == nil || d.Manufacturer != "" || d.MAC == "" {
		return
	}
	if org, ok := e.ouiRegistry.Lookup(d.MAC); ok {
		d.Manufacturer = org
	}
}

func mapToSlice(m map[string]*Device) []Device {
	res := make([]Device, 0, len(m))
	for _, v := range m {
		res = append(res, *v)
	}
	return res
}
