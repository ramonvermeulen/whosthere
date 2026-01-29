package arp

import (
	"net"
	"testing"
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
			if got := isMulticastMAC(tt.mac); got != tt.want {
				t.Errorf("isMulticastMAC(%v) = %v, want %v", tt.mac, got, tt.want)
			}
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
			if got := isBroadcastMAC(tt.mac); got != tt.want {
				t.Errorf("isBroadcastMAC(%v) = %v, want %v", tt.mac, got, tt.want)
			}
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

			if got := isBroadcastIPv4(ip, subnet); got != tt.want {
				t.Errorf("isBroadcastIPv4(%s, %s) = %v, want %v",
					tt.ip, tt.cidr, got, tt.want)
			}
		})
	}
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

			if got := isMulticastIPv4(ip); got != tt.want {
				t.Errorf("isMulticastIPv4(%s) = %v, want %v",
					tt.ip, got, tt.want)
			}
		})
	}
}
