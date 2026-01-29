package ssdp

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/textproto"
	"net/url"
	"strings"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

const (
	MulticastAddr = "239.255.255.250:1900"
	HeaderMan     = `"ssdp:discover"`
	HeaderST      = "ssdp:all"
	HeaderMX      = 2
)

var _ discovery.Scanner = (*Scanner)(nil)

// Scanner discovers devices using SSDP (Simple Service Discovery Protocol),
// part of the UPnP standard. SSDP is commonly used by smart TVs, media servers,
// IoT devices, network printers, and home automation devices.
//
// The scanner sends an M-SEARCH multicast query and collects responses from
// devices advertising their services. Each response may include device location
// (XML descriptor URL), server information, and service type.
//
// Implements the discovery protocol as specified in:
// https://datatracker.ietf.org/doc/html/draft-cai-ssdp-v1-03
type Scanner struct {
	iface  *discovery.InterfaceInfo
	logger discovery.Logger
}

// New creates an SSDP scanner for the specified network interface.
func New(iface *discovery.InterfaceInfo, opts ...Option) *Scanner {
	s := &Scanner{iface: iface, logger: discovery.NoOpLogger{}}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil
		}
	}
	return s
}

func (s *Scanner) Name() string { return "ssdp" }

// Scan sends an SSDP M-SEARCH multicast and collects responses until ctx deadline.
// Discovered devices are sent to the out channel as they respond.
//
// The M-SEARCH requests devices to respond within MX seconds (default: 2).
// The scanner listens for the context duration, which should be at least MX + 1 second
// to allow all devices time to respond.
//
// Returns an error on network failures, nil otherwise.
func (s *Scanner) Scan(ctx context.Context, out chan<- *discovery.Device) error {
	mAddr, err := net.ResolveUDPAddr("udp4", MulticastAddr)
	if err != nil {
		return fmt.Errorf("resolve ssdp addr: %w", err)
	}
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: *s.iface.IPv4Addr, Port: 0})
	if err != nil {
		return fmt.Errorf("listen udp: %w", err)
	}
	defer func() { _ = conn.Close() }()

	s.logger.Log(ctx, slog.LevelDebug, "sending SSDP M-SEARCH", "to", mAddr.String(), "from", conn.LocalAddr().String())
	if err := sendSearch(conn, mAddr); err != nil {
		return err
	}

	s.logger.Log(ctx, slog.LevelDebug, "waiting for SSDP responses")
	if err := applyDeadlineFromContext(conn, ctx); err != nil {
		return err
	}

	buf := make([]byte, 8192)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				return nil
			}
			return fmt.Errorf("read ssdp: %w", err)
		}
		handlePacket(out, src, buf[:n])
	}
}

// sendSearch builds and sends the SSDP M-SEARCH request.
func sendSearch(conn *net.UDPConn, addr *net.UDPAddr) error {
	req := fmt.Sprintf(
		"M-SEARCH * HTTP/1.1\r\n"+
			"HOST: %s\r\n"+
			"MAN: %s\r\n"+
			"MX: %d\r\n"+
			"ST: %s\r\n"+
			"USER-AGENT: whosthere/0.1\r\n\r\n",
		MulticastAddr, HeaderMan, HeaderMX, HeaderST,
	)
	if _, err := conn.WriteToUDP([]byte(req), addr); err != nil {
		return fmt.Errorf("send m-search: %w", err)
	}
	return nil
}

// applyDeadlineFromContext sets the UDP read deadline from the context.
func applyDeadlineFromContext(conn *net.UDPConn, ctx context.Context) error {
	if dl, ok := ctx.Deadline(); ok {
		if err := conn.SetReadDeadline(dl); err != nil {
			return fmt.Errorf("set read deadline: %w", err)
		}
		return nil
	}
	return fmt.Errorf("ssdp scan requires context with deadline")
}

// handlePacket parses the packet and emits a Device if an IP can be resolved.
func handlePacket(out chan<- *discovery.Device, src *net.UDPAddr, payload []byte) {
	loc, server := parseHeaders(payload)
	ip := ipFromAddr(src)
	if ip == nil && loc != "" {
		ip = ipFromLocation(loc)
	}
	if ip == nil {
		return
	}
	d := discovery.NewDevice(ip)
	d.SetDisplayName(server)
	d.AddSource("ssdp")
	if loc != "" {
		d.AddExtraData("location", loc)
	}
	if server != "" {
		d.AddExtraData("server", server)
	}
	select {
	case out <- d:
	default:
	}
}

// parseHeaders extracts LOCATION and SERVER using HTTP-like header parsing.
func parseHeaders(b []byte) (location, server string) {
	// Ensures the buffer ends with CRLFCRLF to satisfy textproto header reader
	data := b
	if !bytes.HasSuffix(data, []byte("\r\n\r\n")) {
		data = append(append([]byte{}, data...), []byte("\r\n\r\n")...)
	}
	br := bufio.NewReader(bytes.NewReader(data))
	tr := textproto.NewReader(br)
	// Read the first status line and ignore errors (best-effort)
	_, _ = tr.ReadLine()
	hdr, err := tr.ReadMIMEHeader()
	if err != nil {
		return "", ""
	}
	location = strings.TrimSpace(hdr.Get("Location"))
	server = strings.TrimSpace(hdr.Get("Server"))
	return
}

// Helper: extract IP from net.Addr (UDP address)
func ipFromAddr(a net.Addr) net.IP {
	if a == nil {
		return nil
	}
	if ua, ok := a.(*net.UDPAddr); ok {
		return ua.IP
	}
	host, _, err := net.SplitHostPort(a.String())
	if err == nil {
		return net.ParseIP(host)
	}
	return nil
}

// Helper: extract host/IP from Location URL and return IP literal if present
func ipFromLocation(loc string) net.IP {
	u, err := url.Parse(loc)
	if err != nil {
		return nil
	}
	host := u.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return net.ParseIP(host)
}
