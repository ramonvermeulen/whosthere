package mdns

import (
	"errors"
	"net"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/internal/subnet"
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

// WithTargetSubnets constrains the mDNS scanner to only emit devices
// whose IPs fall within the provided IPv4 CIDR subnets.
func WithTargetSubnets(subnets []*net.IPNet) Option {
	return func(s *Scanner) error {
		cloned := subnet.CloneIPNets(subnets)
		for _, s := range cloned {
			if s == nil || s.IP.To4() == nil {
				return errors.New("target subnets must be IPv4 CIDRs")
			}
		}
		s.targetSubnets = cloned
		return nil
	}
}
