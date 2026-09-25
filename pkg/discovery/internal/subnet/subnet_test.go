package subnet

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCloneIPNet_Nil(t *testing.T) {
	require.Nil(t, CloneIPNet(nil))
}

func TestCloneIPNet_DeepCopy(t *testing.T) {
	_, original, err := net.ParseCIDR("10.0.0.0/24")
	require.NoError(t, err)

	clone := CloneIPNet(original)

	require.Equal(t, original.String(), clone.String())
	require.NotSame(t, original, clone, "should be a different pointer")

	clone.IP[3] = 99
	require.NotEqual(t, original.IP[3], clone.IP[3], "modifying clone IP should not affect original")

	clone.Mask[2] = 0
	require.NotEqual(t, original.Mask[2], clone.Mask[2], "modifying clone Mask should not affect original")
}

func TestCloneIPNets_Empty(t *testing.T) {
	result := CloneIPNets(nil)
	require.Empty(t, result)

	result = CloneIPNets([]*net.IPNet{})
	require.Empty(t, result)
}

func TestCloneIPNets_PreservesNil(t *testing.T) {
	_, subnet, err := net.ParseCIDR("10.0.0.0/24")
	require.NoError(t, err)

	result := CloneIPNets([]*net.IPNet{nil, subnet, nil})
	require.Len(t, result, 3)
	require.Nil(t, result[0])
	require.NotNil(t, result[1])
	require.Nil(t, result[2])
}

func TestCloneIPNets_DeepCopy(t *testing.T) {
	_, subnetA, err := net.ParseCIDR("10.0.0.0/24")
	require.NoError(t, err)
	_, subnetB, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)

	result := CloneIPNets([]*net.IPNet{subnetA, subnetB})
	require.Len(t, result, 2)
	require.NotSame(t, subnetA, result[0])
	require.NotSame(t, subnetB, result[1])
	require.Equal(t, "10.0.0.0/24", result[0].String())
	require.Equal(t, "192.168.1.0/24", result[1].String())
}

func TestIPInAnySubnet_NilOrEmptySubnets(t *testing.T) {
	ip := net.ParseIP("10.0.0.5")
	require.False(t, IPInAnySubnet(ip, nil))
	require.False(t, IPInAnySubnet(ip, []*net.IPNet{}))
}

func TestIPInAnySubnet_IPv6(t *testing.T) {
	_, subnet, err := net.ParseCIDR("::1/128")
	require.NoError(t, err)
	ip := net.ParseIP("::1")
	require.False(t, IPInAnySubnet(ip, []*net.IPNet{subnet}))
}

func TestIPInAnySubnet_Inside(t *testing.T) {
	_, subnet, err := net.ParseCIDR("10.0.0.0/24")
	require.NoError(t, err)
	require.True(t, IPInAnySubnet(net.ParseIP("10.0.0.5"), []*net.IPNet{subnet}))
	require.True(t, IPInAnySubnet(net.ParseIP("10.0.0.0"), []*net.IPNet{subnet}))
	require.True(t, IPInAnySubnet(net.ParseIP("10.0.0.255"), []*net.IPNet{subnet}))
}

func TestIPInAnySubnet_Outside(t *testing.T) {
	_, subnet, err := net.ParseCIDR("10.0.0.0/24")
	require.NoError(t, err)
	require.False(t, IPInAnySubnet(net.ParseIP("10.0.1.1"), []*net.IPNet{subnet}))
	require.False(t, IPInAnySubnet(net.ParseIP("192.168.1.1"), []*net.IPNet{subnet}))
}

func TestIPInAnySubnet_MultipleSubnets(t *testing.T) {
	_, subnetA, err := net.ParseCIDR("10.0.0.0/24")
	require.NoError(t, err)
	_, subnetB, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)

	subnets := []*net.IPNet{subnetA, subnetB}
	require.True(t, IPInAnySubnet(net.ParseIP("10.0.0.1"), subnets))
	require.True(t, IPInAnySubnet(net.ParseIP("192.168.1.1"), subnets))
	require.False(t, IPInAnySubnet(net.ParseIP("172.16.0.1"), subnets))
}

func TestIPInAnySubnet_NilSubnetInList(t *testing.T) {
	_, subnet, err := net.ParseCIDR("10.0.0.0/24")
	require.NoError(t, err)
	subnets := []*net.IPNet{nil, subnet}
	require.True(t, IPInAnySubnet(net.ParseIP("10.0.0.5"), subnets))
}