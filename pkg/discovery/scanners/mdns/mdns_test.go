package mdns

import (
	"net"
	"testing"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/dns/dnsmessage"
)

func TestNewScanner(t *testing.T) {
	iface := &discovery.InterfaceInfo{}
	scanner, err := New(iface)
	require.NoError(t, err)
	require.Equal(t, iface, scanner.iface)
}

func TestNewScanner_WithLogger(t *testing.T) {
	iface := &discovery.InterfaceInfo{}
	scanner, err := New(iface, WithLogger(discovery.NoOpLogger{}))
	require.NoError(t, err)
	require.NotNil(t, scanner.logger)
}

func TestName(t *testing.T) {
	scanner, err := New(nil)
	require.NoError(t, err)
	require.Equal(t, "mdns", scanner.Name())
}

func TestCleanDisplayName(t *testing.T) {
	require.Equal(t, "My Device", cleanDisplayName("My Device.local."))
	require.Equal(t, "My Device", cleanDisplayName("My Device."))
	require.Equal(t, "My Device", cleanDisplayName("My Device"))
}

func TestExtractServiceName(t *testing.T) {
	require.Equal(t, "http", extractServiceName("_http._tcp.local."))
	require.Equal(t, "", extractServiceName(""))
}

func TestExtractServiceNameFromTarget(t *testing.T) {
	require.Equal(t, "http", extractServiceNameFromTarget("_http._tcp.local."))
	require.Equal(t, "", extractServiceNameFromTarget("invalid"))
}

func TestParseTXTRecords(t *testing.T) {
	ss := &scanSession{}
	dev := discovery.NewDevice(net.IPv4(10, 0, 0, 2).To4())

	txt := &dnsmessage.TXTResource{TXT: []string{
		"manufacturer=Acme",
		"mac=aa:bb:cc:dd:ee:ff",
		"md=Kitchen Speaker",
		"foo=bar",
		"flag",
	}}

	ss.parseTXTRecords(txt, dev)
	require.Equal(t, "Acme", dev.Manufacturer())
	require.Equal(t, "aa:bb:cc:dd:ee:ff", dev.MAC())
	require.Equal(t, "Kitchen Speaker", dev.DisplayName())
	extra := dev.ExtraData()
	require.Equal(t, "bar", extra["foo"])
	require.Equal(t, "true", extra["flag"])
}
