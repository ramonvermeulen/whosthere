package discovery

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery/oui"
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
func (n NoOpLogger) Log(_ context.Context, _ slog.Level, _ string, _ ...any) {}

// Scanner defines a discovery strategy (SSDP, mDNS, ARP, etc.).
// Additional scanners can be implemented by satisfying this interface.
type Scanner interface {
	Name() string
	Scan(ctx context.Context, out chan<- *Device) error
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
	// maybe refactor as part of runtime interface switching?
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

// NewEngine creates a new discovery engine with the provided options.
// Use the provided options to configure the engine's behavior, such as scan intervals,
// timeouts, scanners, sweeper, logger, and network interface.
//
// Returns an error if essential components are missing (e.g. no scanners or interface).
//
// Example:
//
//	engine, err := discovery.NewEngine(
//	    discovery.WithInterface(myInterface),
//	    discovery.WithScanners(myScanner1, myScanner2),
//	    discovery.WithSweeper(mySweeper),
//	    discovery.WithScanInterval(30*time.Second),
//	    discovery.WithLogger(myLogger),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewEngine(opts ...Option) (*Engine, error) {
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

// Start begins continuous network discovery in the background.
// Scans begin immediately and repeat at the interval specified by WithScanInterval
// (default: 20 seconds). If the interval is 0, only a single scan is performed.
//
// Returns the Events channel for monitoring discoveries. Read from this channel
// to receive EventDeviceDiscovered, EventScanCompleted, EventError, and lifecycle events.
//
// Safe to call multiple times - subsequent calls return the same Events channel
// without starting additional background workers.
//
// Call Stop() to halt discovery and close the Events channel. The engine waits
// for all background tasks to complete before Stop() returns.
//
// Example:
//
//	events := engine.Start(context.Background())
//	for event := range events {
//	    switch event.Type {
//	    case discovery.EventDeviceDiscovered:
//	        fmt.Printf("Found: %s\n", event.Device.IP())
//	    case discovery.EventScanCompleted:
//	        fmt.Printf("Scan done: %d devices in %v\n",
//	            event.Stats.DeviceCount, event.Stats.Duration)
//	    case discovery.EventError:
//	        fmt.Printf("Error: %v\n", event.Error)
//	    }
//	}
func (e *Engine) Start(ctx context.Context) <-chan Event {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return e.Events
	}

	ctx, cancel := context.WithCancel(ctx)
	e.cancel = cancel
	e.running = true

	e.emit(NewEngineStartedEvent())

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

// Stop gracefully shuts down the discovery engine.
// Cancels all background scanning, waits for active scans to complete,
// and closes the Events channel. Blocks until all cleanup is finished.
//
// Safe to call multiple times or when the engine is not running.
// After Stop() completes, the Events channel will be closed and no more
// events will be sent.
//
// Example:
//
//	engine.Start(ctx)
//	defer engine.Stop()
//	// ... do work ...
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

	e.emit(NewEngineStoppedEvent())
	close(e.events)
}

// Scan performs a single device discovery scan and returns all found devices.
// Unlike Run(), this is a blocking, synchronous operation that returns when
// the scan completes or the context deadline is reached.
//
// The context timeout defaults to the engine's scan timeout (default: 10 seconds).
// If a sweeper is configured, it runs concurrently during the scan to populate
// the ARP cache.
//
// Returns a slice of discovered devices or an error if the scan fails.
// An empty slice is returned if no devices are found (not an error).
//
// Use this method for one-off discovery. For continuous monitoring, use Run().
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
//	defer cancel()
//	devices, err := engine.Scan(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, dev := range devices {
//	    fmt.Printf("%s - %s\n", dev.IP(), dev.DisplayName())
//	}
func (e *Engine) Scan(ctx context.Context) ([]*Device, error) {
	ctx, cancel := context.WithTimeout(ctx, e.scanTimeout)
	defer cancel()

	if e.sweeper != nil {
		go e.sweeper.Start(ctx)
	}

	return e.performScan(ctx)
}

// runScanLoop runs continuous scans at interval.
//
// Contract:
//   - The first scan starts immediately.
//   - Subsequent scans start on a fixed-rate schedule, i.e. scanInterval is measured from scan start.
//   - Scans never overlap. If a scan takes longer than scanInterval, the next scan starts immediately.
func (e *Engine) runScanLoop(ctx context.Context) {
	defer e.wg.Done()

	if e.scanInterval <= 0 {
		scanCtx, cancel := context.WithTimeout(ctx, e.scanTimeout)
		_, err := e.performScan(scanCtx)
		cancel()
		if err != nil && ctx.Err() == nil {
			e.emit(NewErrorEvent(err))
		}
		return
	}

	nextDue := time.Now()

	for {
		if wait := time.Until(nextDue); wait > 0 {
			t := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				t.Stop()
				return
			case <-t.C:
			}
		}

		if ctx.Err() != nil {
			return
		}

		scanStart := time.Now()
		scanCtx, cancel := context.WithTimeout(ctx, e.scanTimeout)
		_, err := e.performScan(scanCtx)
		cancel()
		if err != nil && ctx.Err() == nil {
			e.emit(NewErrorEvent(err))
		}

		nextDue = scanStart.Add(e.scanInterval)
		if nextDue.Before(time.Now()) {
			nextDue = time.Now()
		}
	}
}

func (e *Engine) performScan(ctx context.Context) ([]*Device, error) {
	e.emit(NewScanStartedEvent())
	start := time.Now()

	scannerOut := make(chan *Device, e.maxDevices)
	var scannerWg sync.WaitGroup

	for _, scanner := range e.scanners {
		scannerWg.Add(1)
		go func(s Scanner) {
			defer scannerWg.Done()
			if err := s.Scan(ctx, scannerOut); err != nil {
				e.emit(NewErrorEvent(fmt.Errorf("scanner failed: %w", err)))
			}
		}(scanner)
	}

	// close channel when scanners done
	go func() {
		scannerWg.Wait()
		close(scannerOut)
	}()

	// process until channel closes
	devices := make(map[string]*Device)
	var mu sync.Mutex
	for device := range scannerOut {
		mu.Lock()
		e.processDevice(device, devices)
		mu.Unlock()
	}

	stats := &ScanStats{
		DeviceCount: len(devices),
		Duration:    time.Since(start),
	}
	e.emit(NewScanCompletedEvent(stats))

	return mapToSlicePtr(devices), nil
}

// processDevice handles a single discovered device
func (e *Engine) processDevice(d *Device, devices map[string]*Device) {
	if d == nil {
		return
	}

	if d.IP() == nil {
		return
	}

	key := d.IP().String()
	if key == "" {
		return
	}

	if existing, found := devices[key]; found {
		existing.Merge(d)
		e.fillManufacturer(existing)
		d = existing
	} else {
		if d.FirstSeen().IsZero() {
			d.SetFirstSeen(time.Now())
		}
		e.fillManufacturer(d)
		devices[key] = d
	}

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
	if d == nil || e.ouiRegistry == nil || d.Manufacturer() != "" || d.MAC() == "" {
		return
	}
	if org, ok := e.ouiRegistry.Lookup(d.MAC()); ok {
		d.SetManufacturer(org)
	}
}

func mapToSlicePtr(m map[string]*Device) []*Device {
	res := make([]*Device, 0, len(m))
	for _, v := range m {
		res = append(res, v)
	}
	return res
}
