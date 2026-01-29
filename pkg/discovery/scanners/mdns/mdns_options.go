package mdns

import (
	"errors"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

// Option configures an mDNS Scanner during construction.
type Option func(*Scanner) error

// WithLogger sets a custom logger for the mDNS scanner.
func WithLogger(logger discovery.Logger) Option {
	return func(s *Scanner) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		s.logger = logger
		return nil
	}
}
