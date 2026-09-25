package subnet

import "net"

func CloneIPNets(subnets []*net.IPNet) []*net.IPNet {
	if len(subnets) == 0 {
		return []*net.IPNet{}
	}

	cloned := make([]*net.IPNet, 0, len(subnets))
	for _, subnet := range subnets {
		if subnet == nil {
			cloned = append(cloned, nil)
			continue
		}
		cloned = append(cloned, CloneIPNet(subnet))
	}
	return cloned
}

func CloneIPNet(subnet *net.IPNet) *net.IPNet {
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

func IPInAnySubnet(ip net.IP, subnets []*net.IPNet) bool {
	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}
	for _, subnet := range subnets {
		if subnet != nil && subnet.Contains(ip4) {
			return true
		}
	}
	return false
}