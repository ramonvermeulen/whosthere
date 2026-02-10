package arp

import (
	"errors"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

// Option configures an ARP Scanner during construction.
type Option func(*Scanner) error

// WithLogger sets a custom logger for the ARP scanner.
func WithLogger(logger discovery.Logger) Option {
	return func(s *Scanner) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		s.logger = logger
		return nil
	}
}

// WithPollInterval sets how often the ARP cache is read during scanning.
// Faster polling detects new devices sooner but uses more CPU.
// Must be positive.
//
// Default: 250ms
func WithPollInterval(interval time.Duration) Option {
	return func(s *Scanner) error {
		if interval <= 0 {
			return errors.New("poll interval must be positive")
		}
		s.pollInterval = interval
		return nil
	}
}
