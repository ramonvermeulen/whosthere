package discovery

import (
	"fmt"
	"net"
)

// InterfaceInfo contains network interface information required for device discovery.
// Scanners need both the interface itself and its IPv4 configuration to operate.
// Use NewInterfaceInfo() to create instances with proper validation.
type InterfaceInfo struct {
	Interface *net.Interface // The network interface device
	IPv4Addr  *net.IP        // Host's IPv4 address on this interface
	IPv4Net   *net.IPNet     // The subnet CIDR (e.g., 192.168.1.0/24)
}

// NewInterfaceInfo creates an InterfaceInfo from a network interface name.
// Pass an empty string to auto-detect the system's default network interface.
//
// Returns an error if:
//   - The interface doesn't exist
//   - The interface has no IPv4 address configured
//   - No suitable default interface is found (when interfaceName is "")
//
// Example:
//
//	// Explicit interface
//	iface, err := discovery.NewInterfaceInfo("en0")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Auto-detect
//	iface, err := discovery.NewInterfaceInfo("")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Using %s with IP %s\n", iface.Interface.Name, iface.IPv4Addr)
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
		return iface, nil
	}

	if iface, err = getDefaultInterface(); err != nil {
		return nil, err
	}

	return iface, nil
}

// getDefaultInterface attempts to return the OS default network interface
// todo(ramon) find more reliable way to get default interface across platforms
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

// CompareIPs compares two IP addresses for sorting purposes.
// IPv4 addresses are compared numerically (192.168.1.2 < 192.168.1.100),
// not lexicographically. IPv6 addresses fall back to string comparison.
//
// Returns true if a should be sorted before b.
//
// Useful when displaying device lists in a user-friendly order.
//
// Example:
//
//	devices := []Device{...}
//	sort.Slice(devices, func(i, j int) bool {
//	    return discovery.CompareIPs(devices[i].IP, devices[j].IP)
//	})
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
