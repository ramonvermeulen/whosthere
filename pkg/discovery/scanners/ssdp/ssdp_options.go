package ssdp

import (
	"errors"
	"net"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/internal/subnet"
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

// WithTargetSubnets constrains the SSDP scanner to only emit devices
// whose IPs fall within the provided IPv4 CIDR subnets.
func WithTargetSubnets(subnets []*net.IPNet) Option {
	return func(s *Scanner) error {
		cloned := subnet.CloneIPNets(subnets)
		for _, sub := range cloned {
			if sub == nil || sub.IP.To4() == nil {
				return errors.New("target subnets must be IPv4 CIDRs")
			}
		}
		s.targetSubnets = cloned
		return nil
	}
}
