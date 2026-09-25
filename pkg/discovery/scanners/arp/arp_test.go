package arp

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/ramonvermeulen/whosthere/pkg/discovery/internal/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsMulticastMAC(t *testing.T) {
	tests := []struct {
		name string
		mac  net.HardwareAddr
		want bool
	}{
		{"unicast MAC", net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}, false},
		{"multicast MAC", net.HardwareAddr{0x01, 0x00, 0x5e, 0x00, 0x00, 0xfb}, true},
		{"broadcast MAC", net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, true},
		{"empty MAC", net.HardwareAddr{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isMulticastMAC(tt.mac), "isMulticastMAC(%v)", tt.mac)
		})
	}
}

func TestIsBroadcastMAC(t *testing.T) {
	tests := []struct {
		name string
		mac  net.HardwareAddr
		want bool
	}{
		{"broadcast MAC", net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, true},
		{"multicast MAC", net.HardwareAddr{0x01, 0x00, 0x5e, 0x00, 0x00, 0xfb}, false},
		{"unicast MAC", net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}, false},
		{"short MAC", net.HardwareAddr{0xff, 0xff, 0xff}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isBroadcastMAC(tt.mac), "isBroadcastMAC(%v)", tt.mac)
		})
	}
}

func TestIsBroadcastIPv4(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		cidr string
		want bool
	}{
		{"broadcast /24", "192.168.1.255", "192.168.1.42/24", true},
		{"broadcast /25", "192.168.1.127", "192.168.1.42/25", true},
		{"broadcast /26", "192.168.1.63", "192.168.1.42/26", true},
		{"not broadcast", "192.168.1.42", "192.168.1.42/24", false},
		{"different subnet", "192.168.2.255", "192.168.1.1/24", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			_, subnet, _ := net.ParseCIDR(tt.cidr)

			assert.Equal(t, tt.want, isBroadcastIPv4(ip, subnet), "isBroadcastIPv4(%s, %s)", tt.ip, tt.cidr)
		})
	}
}

func TestEmitARPEntries_FiltersOutsideTargetSubnets(t *testing.T) {
	iface := testkit.MustInterfaceInfo(t)
	_, targetSubnet, err := net.ParseCIDR("10.0.1.0/24")
	require.NoError(t, err, "parse target subnet")

	s, err := New(iface, WithTargetSubnets([]*net.IPNet{targetSubnet}))
	require.NoError(t, err, "new scanner")

	out := make(chan *discovery.Device, 2)
	entries := []Entry{
		{
			IP:            net.ParseIP("10.0.1.42"),
			MAC:           net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			InterfaceName: iface.Interface.Name,
		},
		{
			IP:            net.ParseIP("10.0.1.255"),
			MAC:           net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x66},
			InterfaceName: iface.Interface.Name,
		},
		{
			IP:            net.ParseIP("192.168.1.42"),
			MAC:           net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x77},
			InterfaceName: iface.Interface.Name,
		},
	}

	require.NoError(t, s.emitARPEntries(context.Background(), out, entries), "emit entries")
	close(out)

	var got []string
	for dev := range out {
		got = append(got, dev.IP().String())
	}

	require.Equal(t, []string{"10.0.1.42"}, got, "filtered ARP entries")
}
func TestIsMulticastIPv4(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{"multicast lower bound", "224.0.0.1", true},
		{"multicast upper bound", "239.255.255.255", true},
		{"unicast IPv4", "192.168.1.1", false},
		{"broadcast IPv4", "255.255.255.255", false},
		{"IPv6 address", "ff02::1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)

			assert.Equal(t, tt.want, isMulticastIPv4(ip), "isMulticastIPv4(%s)", tt.ip)
		})
	}
}

func TestWithPollInterval_RejectsNonPositive(t *testing.T) {
	s, err := New(testkit.MustInterfaceInfo(t), WithPollInterval(0))
	require.Error(t, err, "expected error for zero poll interval")
	require.Nil(t, s, "expected nil scanner on invalid option")
}

func TestWithPollInterval_SetsInterval(t *testing.T) {
	interval := 2 * time.Millisecond
	s, err := New(testkit.MustInterfaceInfo(t), WithPollInterval(interval))
	require.NoError(t, err, "unexpected error")
	require.NotNil(t, s, "expected scanner")
	require.Equal(t, interval, s.pollInterval, "pollInterval")
}
