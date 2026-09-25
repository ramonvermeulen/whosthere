package sweeper

import (
	"errors"
	"net"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/internal/subnet"
)

// Option configures a Sweeper during construction.
type Option func(*Sweeper) error

// WithSweeperInterface sets the network interface for sweeping.
func WithSweeperInterface(iface *discovery.InterfaceInfo) Option {
	return func(s *Sweeper) error {
		if iface == nil {
			return errors.New("interface cannot be nil")
		}
		s.iface = iface
		return nil
	}
}

// WithSweeperInterval sets the time between sweep cycles.
// Each sweep triggers ARP for all IPs in the subnet.
// Must be positive.
//
// Default: 5 minutes (discovery.DefaultSweepInterval)
func WithSweeperInterval(interval time.Duration) Option {
	return func(s *Sweeper) error {
		if interval <= 0 {
			return errors.New("sweep interval must be positive")
		}
		s.interval = interval
		return nil
	}
}

// WithSweeperTimeout sets the maximum duration for each sweep cycle.
// If a sweep takes longer, it's canceled and the next one begins.
// Must be positive.
//
// Default: 20 seconds (discovery.DefaultSweepTimeout)
func WithSweeperTimeout(timeout time.Duration) Option {
	return func(s *Sweeper) error {
		if timeout <= 0 {
			return errors.New("sweep timeout must be positive")
		}
		s.timeout = timeout
		return nil
	}
}

// WithTargetSubnets sets the IPv4 CIDR subnets to sweep.
// When at least one target subnet is provided, the sweeper does not automatically
// include the selected interface's subnet. Add it explicitly if desired.
func WithTargetSubnets(subnets []*net.IPNet) Option {
	return func(s *Sweeper) error {
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

// WithSweeperLogger sets a custom logger for the sweeper.
func WithSweeperLogger(logger discovery.Logger) Option {
	return func(s *Sweeper) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		s.logger = logger
		return nil
	}
}

// WithAllowLargeSubnets disables the /16 sweep limit for large subnets.
// When false (default), subnets larger than /16 are capped to a /16 equivalent
// (65534 IPs). When true, the full subnet is scanned regardless of size.
func WithAllowLargeSubnets(allow bool) Option {
	return func(s *Sweeper) error {
		s.allowLargeSubnets = allow
		return nil
	}
}
