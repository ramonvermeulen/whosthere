package mdns

import (
	"errors"
	"net"

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

// WithTargetSubnets constrains the mDNS scanner to only emit devices
// whose IPs fall within the provided IPv4 CIDR subnets.
func WithTargetSubnets(subnets []*net.IPNet) Option {
	return func(s *Scanner) error {
		cloned := cloneIPNets(subnets)
		for _, subnet := range cloned {
			if subnet == nil || subnet.IP.To4() == nil {
				return errors.New("target subnets must be IPv4 CIDRs")
			}
		}
		s.targetSubnets = cloned
		return nil
	}
}

func cloneIPNets(subnets []*net.IPNet) []*net.IPNet {
	if len(subnets) == 0 {
		return []*net.IPNet{}
	}

	cloned := make([]*net.IPNet, 0, len(subnets))
	for _, subnet := range subnets {
		if subnet == nil {
			cloned = append(cloned, nil)
			continue
		}
		cloned = append(cloned, cloneIPNet(subnet))
	}
	return cloned
}

func cloneIPNet(subnet *net.IPNet) *net.IPNet {
	if subnet == nil {
		return nil
	}

	ip := make(net.IP, len(subnet.IP))
	copy(ip, subnet.IP)
	mask := make(net.IPMask, len(subnet.Mask))
	copy(mask, subnet.Mask)
	return &net.IPNet{
		IP:   ip,
		Mask: mask,
	}
}
