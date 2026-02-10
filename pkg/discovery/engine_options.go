package discovery

import (
	"errors"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery/oui"
)

// Option configures an Engine during construction with NewEngine.
// Options return an error if invalid values are provided.
type Option func(*Engine) error

// WithScanTimeout sets the maximum duration for each scan cycle.
// After this timeout, the scan is canceled even if scanners haven't finished.
// Must be positive.
//
// Default: 10 seconds (DefaultScanTimeout)
func WithScanTimeout(timeout time.Duration) Option {
	return func(e *Engine) error {
		if timeout <= 0 {
			return errors.New("timeout must be positive")
		}
		e.scanTimeout = timeout
		return nil
	}
}

// WithScanInterval sets the interval between scan cycles when using Start().
// Scans start on a fixed schedule, measured from scan start time.
// If a scan takes longer than the interval, the next scan starts immediately.
//
// Set to 0 to perform only a single scan when Start() is called, then stop.
// This differs from Scan(): Start(ctx) with interval 0 still uses the event
// channel and background goroutines, while Scan(ctx) is fully synchronous
// and returns devices directly. Use Start() with interval 0 when you want
// event-driven notifications; use Scan() when you want blocking behavior.
//
// Negative values are rejected with an error.
//
// Default: 20 seconds (DefaultScanInterval)
func WithScanInterval(interval time.Duration) Option {
	return func(e *Engine) error {
		if interval < 0 {
			return errors.New("interval must be >= 0")
		}
		e.scanInterval = interval
		return nil
	}
}

// WithSweeper configures the engine to use an ARP cache sweeper.
// The sweeper sends network packets to populate the OS ARP cache before
// ARP-based scanning. Highly recommended when using the ARP scanner.
func WithSweeper(sweeper Sweeper) Option {
	return func(e *Engine) error {
		e.sweeper = sweeper
		return nil
	}
}

// WithLogger sets a custom logger for the engine and its scanners.
// Any slog-compatible logger works. Use this to integrate with your
// application's logging infrastructure.
//
// Default: NoOpLogger (discards all logs)
func WithLogger(logger Logger) Option {
	return func(e *Engine) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		e.logger = logger
		return nil
	}
}

// WithInterface sets the network interface used for discovery.
// This is required - NewEngine returns ErrNoInterface if not provided.
//
// Use NewInterfaceInfo() to create an InterfaceInfo from an interface name,
// or pass an empty string to auto-detect the default interface.
func WithInterface(iface *InterfaceInfo) Option {
	return func(e *Engine) error {
		if iface == nil {
			return errors.New("interface cannot be nil")
		}
		e.Iface = iface
		return nil
	}
}

// WithScanners configures the engine with one or more discovery scanners.
// At least one scanner or sweeper is required - NewEngine returns
// ErrNoScannersOrSweeper if neither is provided.
//
// Built-in scanners:
//   - arp.Scanner: Reads the ARP cache for MAC/IP mappings
//   - mdns.Scanner: Discovers devices via multicast DNS
//   - ssdp.Scanner: Discovers devices via SSDP
func WithScanners(scanners ...Scanner) Option {
	return func(e *Engine) error {
		if len(scanners) == 0 {
			return errors.New("at least one scanner required")
		}
		e.scanners = scanners
		return nil
	}
}

// WithOUIRegistry enables manufacturer name lookups based on MAC address OUI prefixes.
// The registry maps the first 3 bytes of MAC addresses to vendor names.
// When set, the engine automatically populates the Manufacturer field of discovered devices.
//
// The OUI registry auto-updates from IEEE data when stale (>30 days).
func WithOUIRegistry(registry *oui.Registry) Option {
	return func(e *Engine) error {
		e.ouiRegistry = registry
		return nil
	}
}
