package mdns

import (
	"errors"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

// Option is a functional option for configuring the mDNS Scanner.
type Option func(*Scanner) error

// WithLogger sets a custom logger for the MDNS scanner.
func WithLogger(logger discovery.Logger) Option {
	return func(s *Scanner) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		s.logger = logger
		return nil
	}
}
