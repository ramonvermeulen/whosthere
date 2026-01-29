package discovery

import (
	"errors"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/discovery/oui"
)

// Option configures an Engine
type Option func(*Engine) error

// WithScanTimeout sets scan timeout
func WithScanTimeout(timeout time.Duration) Option {
	return func(e *Engine) error {
		if timeout <= 0 {
			return errors.New("timeout must be positive")
		}
		e.scanTimeout = timeout
		return nil
	}
}

// WithScanInterval sets interval for continuous scans
func WithScanInterval(interval time.Duration) Option {
	return func(e *Engine) error {
		if interval <= 0 {
			return errors.New("interval must be positive")
		}
		e.scanInterval = interval
		return nil
	}
}

// WithSweeper enables ARP cache sweeping
func WithSweeper(sweeper Sweeper) Option {
	return func(e *Engine) error {
		e.sweeper = sweeper
		return nil
	}
}

// WithLogger sets a logger
func WithLogger(logger Logger) Option {
	return func(e *Engine) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		e.logger = logger
		return nil
	}
}

// WithInterface sets network interface
func WithInterface(iface *InterfaceInfo) Option {
	return func(e *Engine) error {
		if iface == nil {
			return errors.New("interface cannot be nil")
		}
		e.Iface = iface
		return nil
	}
}

// WithScanners adds scanners
func WithScanners(scanners ...Scanner) Option {
	return func(e *Engine) error {
		if len(scanners) == 0 {
			return errors.New("at least one scanner required")
		}
		e.scanners = scanners
		return nil
	}
}

// WithOUIRegistry sets the OUI registry
func WithOUIRegistry(registry *oui.Registry) Option {
	return func(e *Engine) error {
		e.ouiRegistry = registry
		return nil
	}
}
