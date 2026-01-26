package discovery

import (
	"net"
	"reflect"
	"testing"
)

func TestGenerateSubnetIPs(t *testing.T) {
	tests := []struct {
		name     string
		subnet   *net.IPNet
		skipIP   net.IP
		expected []net.IP
	}{
		{
			name:     "simple /24 subnet, skip local IP",
			subnet:   &net.IPNet{IP: net.IP{192, 168, 0, 0}, Mask: net.IPMask{255, 255, 255, 0}},
			skipIP:   net.IP{192, 168, 0, 104},
			expected: generateExpectedIPs(net.IP{192, 168, 0, 0}, net.IP{192, 168, 0, 255}, net.IP{192, 168, 0, 104}),
		},
		{
			name:     "/24 subnet, skip network IP",
			subnet:   &net.IPNet{IP: net.IP{192, 168, 0, 0}, Mask: net.IPMask{255, 255, 255, 0}},
			skipIP:   net.IP{192, 168, 0, 0},
			expected: generateExpectedIPs(net.IP{192, 168, 0, 0}, net.IP{192, 168, 0, 255}, net.IP{192, 168, 0, 0}),
		},
		{
			name:     "/24 subnet, skip broadcast IP",
			subnet:   &net.IPNet{IP: net.IP{192, 168, 0, 0}, Mask: net.IPMask{255, 255, 255, 0}},
			skipIP:   net.IP{192, 168, 0, 255},
			expected: generateExpectedIPs(net.IP{192, 168, 0, 0}, net.IP{192, 168, 0, 255}, net.IP{192, 168, 0, 255}),
		},
		{
			name:   "/24 subnet, no skip",
			subnet: &net.IPNet{IP: net.IP{192, 168, 0, 0}, Mask: net.IPMask{255, 255, 255, 0}},
			// outside range, should have no effect
			skipIP:   net.IP{192, 168, 1, 1},
			expected: generateExpectedIPs(net.IP{192, 168, 0, 0}, net.IP{192, 168, 0, 255}, nil),
		},
		{
			name:     "/8 subnet, limited to /16",
			subnet:   &net.IPNet{IP: net.IP{10, 0, 0, 0}, Mask: net.IPMask{255, 0, 0, 0}},
			skipIP:   net.IP{10, 0, 0, 1},
			expected: generateExpectedIPs(net.IP{10, 0, 0, 0}, net.IP{10, 0, 255, 255}, net.IP{10, 0, 0, 1}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateSubnetIPs(tt.subnet, tt.skipIP)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("generateSubnetIPs() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func generateExpectedIPs(start, end, skip net.IP) []net.IP {
	var ips []net.IP
	current := make(net.IP, len(start))
	copy(current, start)
	for {
		if skip == nil || !current.Equal(skip) {
			ipCopy := make(net.IP, len(current))
			copy(ipCopy, current)
			ips = append(ips, ipCopy)
		}
		if current.Equal(end) {
			break
		}
		current = incrementIP(current)
	}
	return ips
}
