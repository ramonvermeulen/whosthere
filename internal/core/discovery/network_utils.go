package discovery

import (
	"fmt"
	"net"

	"go.uber.org/zap"
)

// InterfaceInfo holds the essential network interface information for scanning
type InterfaceInfo struct {
	Interface *net.Interface // Network interface
	IPv4Addr  *net.IP        // IPv4 host address used by the interface
	IPv4Net   *net.IPNet     // Subnet (e.g., 192.168.1.0/24)
}

// NewInterfaceInfo creates InterfaceInfo from a net.Interface
// It returns an error if the interface has no IPv4 address.
// This makes sure that every scanner has the necessary information to perform network scans.
// And makes interface handling consistent and swappable.
func NewInterfaceInfo(interfaceName string) (*InterfaceInfo, error) {
	iface, err := getNetworkInterface(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("get network interface %s: %w", interfaceName, err)
	}
	info := &InterfaceInfo{Interface: iface}

	addresses, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("get addresses for %s: %w", iface.Name, err)
	}

	for _, addr := range addresses {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			info.IPv4Addr = &ipnet.IP
			info.IPv4Net = ipnet
			break
		}
	}

	if info.IPv4Addr == nil {
		return nil, fmt.Errorf("interface %s has no IPv4 address", iface.Name)
	}

	return info, nil
}

// getNetworkInterface returns the network interface by name.
// If interfaceName is empty, it attempts to return the OS default network interface.
func getNetworkInterface(interfaceName string) (*net.Interface, error) {
	var iface *net.Interface
	var err error
	if interfaceName != "" {
		if iface, err = net.InterfaceByName(interfaceName); err != nil {
			return nil, err
		}
		zap.L().Info("using specified network interface", zap.String("interface", interfaceName))
		return iface, nil
	}

	if iface, err = getDefaultInterface(); err != nil {
		zap.L().Info("failed to get default network interface", zap.Error(err))
		return nil, err
	}

	zap.L().Info("using default network interface", zap.String("interface", iface.Name))
	return iface, nil
}

// getDefaultInterface attempts to return the OS default network interface
func getDefaultInterface() (*net.Interface, error) {
	// try to get the default interface by sending UDP packet
	if name, err := getInterfaceNameByUDP(); err == nil {
		return name, nil
	}

	// if that fails, return the first non-loopback interface that is up
	// this is often the default interface, but in special cases it might not be
	// todo: find better solution in the future, maybe by parsing routing table?
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp != 0 {
			return &iface, nil
		}
	}

	return nil, fmt.Errorf("no network interface found")
}

// getInterfaceNameByUDP tries to determine the default network interface
// by creating a UDP connection to a public IP and checking the local address used.
func getInterfaceNameByUDP() (*net.Interface, error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return nil, err
	}
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip != nil && ip.Equal(localAddr.IP) {
				return &iface, nil
			}
		}
	}

	return nil, fmt.Errorf("interface not found for IP %s", localAddr.IP)
}

// CompareIPs compares two IP addresses for sorting.
// IPv4 addresses are compared numerically (byte-by-byte), ensuring
// correct ordering like 192.168.1.2 < 192.168.1.100.
// IPv6 addresses or mixed IPv4/IPv6 comparisons fall back to
// string comparison (lexicographic).
// Returns true if a should be sorted before b.
func CompareIPs(a, b net.IP) bool {
	aBytes := a.To4()
	bBytes := b.To4()
	if aBytes == nil || bBytes == nil {
		return a.String() < b.String()
	}
	for i := 0; i < 4; i++ {
		if aBytes[i] != bBytes[i] {
			return aBytes[i] < bBytes[i]
		}
	}
	return false
}
