//go:build darwin || freebsd || netbsd || openbsd

package arp

import (
	"context"
	"fmt"
	"net"
	"runtime"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"golang.org/x/net/route"
	"golang.org/x/sys/unix"
)

// readDarwinARPCache reads ARP table on Darwin/BSD systems using x/net/route.
// We fetch the IPv4 routing RIB and filter RouteMessages with RTF_LLINFO (neighbor cache).
// see https://man.freebsd.org/cgi/man.cgi?query=rtentry&sektion=9&manpath=FreeBSD+6.1-RELEASE
func (s *Scanner) readDarwinARPCache(ctx context.Context, out chan<- *discovery.Device) error {
	entries, err := s.readDarwinARPCacheRaw()
	if err != nil {
		return fmt.Errorf("read darwin arp cache: %w", err)
	}
	return s.emitARPEntries(ctx, out, entries)
}

func (s *Scanner) readDarwinARPCacheRaw() ([]Entry, error) {
	var ribType route.RIBType
	if runtime.GOOS == "freebsd" {
		ribType = unix.NET_RT_FLAGS
	} else {
		ribType = route.RIBTypeRoute
	}

	b, err := route.FetchRIB(unix.AF_INET, ribType, 0)
	if err != nil {
		return nil, fmt.Errorf("route.FetchRIB: %w", err)
	}

	msgs, err := route.ParseRIB(route.RIBTypeRoute, b)
	if err != nil {
		return nil, fmt.Errorf("route.ParseRIB: %w", err)
	}

	var entries []Entry
	for _, m := range msgs {
		rm, ok := m.(*route.RouteMessage)
		if !ok {
			continue
		}
		if rm.Flags&unix.RTF_LLINFO == 0 {
			continue
		}

		var ip net.IP
		var mac net.HardwareAddr

		for _, a := range rm.Addrs {
			switch v := a.(type) {
			case *route.Inet4Addr:
				if ip == nil {
					ip = net.IPv4(v.IP[0], v.IP[1], v.IP[2], v.IP[3])
				}
			case *route.LinkAddr:
				if mac == nil && len(v.Addr) >= 6 {
					mac = v.Addr[:6]
				}
			}
		}

		var interfaceName string
		if rm.Index > 0 {
			iface, err := net.InterfaceByIndex(rm.Index)
			if err == nil {
				interfaceName = iface.Name
			}
		}

		if ip == nil || mac == nil {
			continue
		}

		entries = append(entries, Entry{IP: ip, MAC: mac, InterfaceName: interfaceName})
	}
	return entries, nil
}
