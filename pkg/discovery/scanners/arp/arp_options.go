package arp

import (
	"errors"
	"net"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/internal/subnet"
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

// WithAllInterfaces disables interface filtering on the ARP cache.
// When enabled, the scanner reports devices from all network interfaces
// instead of only the configured one.
func WithAllInterfaces(all bool) Option {
	return func(s *Scanner) error {
		s.allInterfaces = all
		return nil
	}
}

// WithTargetSubnets sets the IPv4 CIDR subnets used for target-aware ARP filtering.
// The scanner still reads the OS ARP cache, but only entries inside these subnets
// are emitted as devices.
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
