package ssdp

import (
	"net"
	"testing"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/stretchr/testify/require"
)

func TestNewScanner(t *testing.T) {
	iface := &discovery.InterfaceInfo{}
	scanner, err := New(iface)
	require.NoError(t, err)
	if scanner.iface != iface {
		t.Errorf("expected iface to be set")
	}
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
	if scanner.Name() != "ssdp" {
		t.Errorf("expected name ssdp, got %s", scanner.Name())
	}
}

func TestParseHeaders_ExtractsLocationAndServer(t *testing.T) {
	payload := []byte("HTTP/1.1 200 OK\r\nLOCATION: http://10.0.0.2:80/device.xml\r\nServer: test/1.0\r\n\r\n")
	loc, server := parseHeaders(payload)
	require.Equal(t, "http://10.0.0.2:80/device.xml", loc)
	require.Equal(t, "test/1.0", server)
}

func TestParseHeaders_AppendsTerminatorIfMissing(t *testing.T) {
	payload := []byte("HTTP/1.1 200 OK\r\nLocation: http://10.0.0.2/device.xml\r\nServer: test\r\n")
	loc, server := parseHeaders(payload)
	require.Equal(t, "http://10.0.0.2/device.xml", loc)
	require.Equal(t, "test", server)
}

func TestHandlePacket_UsesSrcIP(t *testing.T) {
	out := make(chan *discovery.Device, 1)
	src := &net.UDPAddr{IP: net.IPv4(10, 0, 0, 2).To4(), Port: 1900}
	payload := []byte("HTTP/1.1 200 OK\r\nServer: unit-test\r\n\r\n")

	handlePacket(out, src, payload)

	require.Len(t, out, 1)
	d := <-out
	require.Equal(t, "10.0.0.2", d.IP().String())
	require.Equal(t, "unit-test", d.DisplayName())
}

func TestHandlePacket_UsesLocationWhenSrcIPMissing(t *testing.T) {
	out := make(chan *discovery.Device, 1)
	src := &net.UDPAddr{IP: nil, Port: 1900}
	payload := []byte("HTTP/1.1 200 OK\r\nLocation: http://10.0.0.3:80/device.xml\r\nServer: unit-test\r\n\r\n")

	handlePacket(out, src, payload)

	require.Len(t, out, 1)
	d := <-out
	require.Equal(t, "10.0.0.3", d.IP().String())
	require.Equal(t, "http://10.0.0.3:80/device.xml", d.ExtraData()["location"])
}

func TestHandlePacket_DoesNotEmitWithoutResolvableIP(t *testing.T) {
	out := make(chan *discovery.Device, 1)
	src := &net.UDPAddr{IP: nil, Port: 1900}
	payload := []byte("HTTP/1.1 200 OK\r\nServer: unit-test\r\n\r\n")

	handlePacket(out, src, payload)

	require.Len(t, out, 0)
}
