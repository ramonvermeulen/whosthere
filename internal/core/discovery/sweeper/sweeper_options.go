package sweeper

import (
	"errors"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
)

// Option configures a Sweeper
type Option func(*Sweeper) error

// WithSweeperInterface sets the network interface for the sweeper
func WithSweeperInterface(iface *discovery.InterfaceInfo) Option {
	return func(s *Sweeper) error {
		if iface == nil {
			return errors.New("interface cannot be nil")
		}
		s.iface = iface
		return nil
	}
}

// WithSweeperInterval sets the sweep interval
func WithSweeperInterval(interval time.Duration) Option {
	return func(s *Sweeper) error {
		if interval <= 0 {
			return errors.New("sweep interval must be positive")
		}
		s.interval = interval
		return nil
	}
}

// WithSweeperTimeout sets the sweep timeout
func WithSweeperTimeout(timeout time.Duration) Option {
	return func(s *Sweeper) error {
		if timeout <= 0 {
			return errors.New("sweep timeout must be positive")
		}
		s.timeout = timeout
		return nil
	}
}

// WithSweeperLogger sets the logger for the sweeper
func WithSweeperLogger(logger discovery.Logger) Option {
	return func(s *Sweeper) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		s.logger = logger
		return nil
	}
}
