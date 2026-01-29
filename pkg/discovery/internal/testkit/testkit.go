package testkit

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

type FakeScanner struct {
	NameStr     string
	Devices     []*discovery.Device
	DeviceBurst int
	Delay       time.Duration
	Err         error
	Scanned     atomic.Int64
}

func (s *FakeScanner) Name() string {
	if s.NameStr == "" {
		return "fake"
	}
	return s.NameStr
}

func (s *FakeScanner) Scan(ctx context.Context, out chan<- *discovery.Device) error {
	s.Scanned.Add(1)

	burst := s.DeviceBurst
	if burst <= 0 {
		burst = 1
	}

	for i := 0; i < len(s.Devices); i += burst {
		if s.Delay > 0 {
			t := time.NewTimer(s.Delay)
			select {
			case <-ctx.Done():
				if !t.Stop() {
					<-t.C
				}
				return ctx.Err()
			case <-t.C:
			}
		}

		end := i + burst
		if end > len(s.Devices) {
			end = len(s.Devices)
		}
		for _, d := range s.Devices[i:end] {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case out <- d:
			}
		}
	}

	return s.Err
}

type FakeSweeper struct{ Started atomic.Int64 }

func (s *FakeSweeper) Start(ctx context.Context) { _ = ctx; s.Started.Add(1) }

func MustIP(t testing.TB, s string) net.IP {
	t.Helper()
	ip := net.ParseIP(s)
	if ip == nil {
		t.Fatalf("invalid ip: %s", s)
	}
	return ip
}

func MustInterfaceInfo(t testing.TB) *discovery.InterfaceInfo {
	t.Helper()
	ip := MustIP(t, "192.168.0.10").To4()
	if ip == nil {
		t.Fatal("expected ipv4")
	}
	_, n, err := net.ParseCIDR("192.168.0.10/24")
	if err != nil {
		t.Fatalf("parse cidr: %v", err)
	}
	return &discovery.InterfaceInfo{Interface: &net.Interface{Name: "test0"}, IPv4Addr: &ip, IPv4Net: n}
}
