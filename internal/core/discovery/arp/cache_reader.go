package arp

import (
	"context"
	"net"
	"runtime"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
)

func (s *Scanner) readARPCache(ctx context.Context, out chan<- discovery.Device) error {
	switch runtime.GOOS {
	case "linux":
		return s.readLinuxARPCache(ctx, out)
	case "darwin", "freebsd", "netbsd", "openbsd":
		return s.readDarwinARPCache(ctx, out)
	case "windows":
		return s.readWindowsARPCache(ctx, out)
	default:
		return nil
	}
}

// Entry represents a single ARP cache entry.
type Entry struct {
	IP            net.IP
	MAC           net.HardwareAddr
	Age           time.Duration
	InterfaceName string
}

// emitARPEntries sends discovered ARP entries to the output channel.
func (s *Scanner) emitARPEntries(ctx context.Context, out chan<- discovery.Device, entries []Entry) error {
	now := time.Now()

	subnet := s.iface.IPv4Net

	for _, entry := range entries {
		if entry.IP == nil || entry.MAC == nil {
			continue
		}

		if entry.InterfaceName != s.iface.Interface.Name {
			continue
		}

		// Filter non-device addresses:
		// - skip multicast MACs (I/G bit set)
		// - skip broadcast MAC (FF:FF:FF:FF:FF:FF)
		// - skip IPv4 broadcast address for our subnet
		// - skip IPv4 multicast ranges (224.0.0.0/4)
		if isMulticastMAC(entry.MAC) || isBroadcastMAC(entry.MAC) || isMulticastIPv4(entry.IP) || isBroadcastIPv4(entry.IP, subnet) {
			continue
		}

		dd := discovery.NewDevice(entry.IP)
		dd.MAC = entry.MAC.String()
		dd.Sources["arp"] = struct{}{}

		if entry.Age > 0 {
			dd.LastSeen = now.Add(-entry.Age)
		} else {
			dd.LastSeen = now
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- dd:
		}
	}

	return nil
}

// isMulticastMAC checks if a MAC address is a multicast address.
// if the LSB of the first byte is set, it's a multicast address.
func isMulticastMAC(mac net.HardwareAddr) bool {
	// Multicast MACs have the least significant bit of the first byte set
	return len(mac) > 0 && (mac[0]&0x01) != 0
}

// isBroadcastMAC checks if a MAC address is a broadcast address.
func isBroadcastMAC(mac net.HardwareAddr) bool {
	// Broadcast MAC is FF:FF:FF:FF:FF:FF
	return len(mac) == 6 && mac[0] == 0xFF && mac[1] == 0xFF && mac[2] == 0xFF && mac[3] == 0xFF && mac[4] == 0xFF && mac[5] == 0xFF
}

// isBroadcastIPv4 checks if an IPv4 address is a broadcast address for the given subnet.
func isBroadcastIPv4(ip net.IP, subnet *net.IPNet) bool {
	if ip == nil || subnet == nil {
		return false
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}

	mask := subnet.Mask
	if len(mask) != net.IPv4len {
		return false
	}

	// Normalize subnet.IP to the actual network address by zeroing host bits.
	network := subnet.IP.Mask(mask).To4()
	if network == nil {
		return false
	}

	var broadcast [net.IPv4len]byte
	for i := 0; i < net.IPv4len; i++ {
		// Compute the broadcast address by setting all host bits to 1:
		//
		//   broadcast = network | ^mask
		//
		// Example:
		//   input CIDR:   192.168.1.42/24
		//   normalized network (IP & mask):
		//                 192.168.1.0
		//   subnet mask:  255.255.255.0
		//   inverted mask (^mask):
		//                 0.0.0.255
		//   broadcast:    192.168.1.255
		broadcast[i] = network[i] | ^mask[i]
	}
	return ip4.Equal(broadcast[:])
}

// isMulticastIPv4 checks if an IPv4 address is in the multicast range (224.0.0.0/4).
func isMulticastIPv4(ip net.IP) bool {
	ip4 := ip.To4()
	// the &0xF0 masks the first 4 bits of the first byte
	// if these bits equal 224 (1110 0000), the IP is in the multicast range
	// it takes the first 4 bits and checks if they match 1110 (224)
	return ip4 != nil && ip4[0]&0xF0 == 224
}
