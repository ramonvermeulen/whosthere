package ssdp

import (
	"errors"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

// Option configures an SSDP Scanner during construction.
type Option func(*Scanner) error

// WithLogger sets a custom logger for the SSDP scanner.
func WithLogger(logger discovery.Logger) Option {
	return func(s *Scanner) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		s.logger = logger
		return nil
	}
}
