package ssdp

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"net/url"
	"strings"

	"github.com/ramonvermeulen/whosthere/internal/discovery"
	"go.uber.org/zap"
)

const (
	MulticastAddr = "239.255.255.250:1900"
	HeaderMan     = `"ssdp:discover"`
	HeaderST      = "ssdp:all"
	HeaderMX      = 2
)

var _ discovery.Scanner = (*Scanner)(nil)

// Scanner implements SSDP discovery (UPnP) via manual M-SEARCH over UDP.
// Implemented as described in the RFC: https://datatracker.ietf.org/doc/html/draft-cai-ssdp-v1-03#section-4.1
type Scanner struct{}

func (s *Scanner) Name() string { return "ssdp" }

// Scan sends an SSDP M-SEARCH and streams responses incrementally until the ctx deadline.
func (s *Scanner) Scan(ctx context.Context, out chan<- discovery.Device) error {
	log := zap.L()

	mAddr, err := net.ResolveUDPAddr("udp4", MulticastAddr)
	if err != nil {
		return fmt.Errorf("resolve ssdp addr: %w", err)
	}
	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return fmt.Errorf("listen udp: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if err := sendSearch(conn, mAddr, log); err != nil {
		return err
	}
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
		handlePacket(out, src, buf[:n], log)
	}
}

// sendSearch builds and sends the SSDP M-SEARCH request.
func sendSearch(conn *net.UDPConn, addr *net.UDPAddr, log *zap.Logger) error {
	req := fmt.Sprintf(
		"M-SEARCH * HTTP/1.1\r\n"+
			"HOST: %s\r\n"+
			"MAN: %s\r\n"+
			"MX: %d\r\n"+
			"ST: %s\r\n"+
			"USER-AGENT: whosthere/0.1\r\n\r\n",
		MulticastAddr, HeaderMan, HeaderMX, HeaderST,
	)
	log.Debug("ssdp m-search send", zap.String("host", MulticastAddr), zap.Int("mx", HeaderMX), zap.String("st", HeaderST))
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
func handlePacket(out chan<- discovery.Device, src *net.UDPAddr, payload []byte, log *zap.Logger) {
	loc, server := parseHeaders(payload)
	ip := ipFromAddr(src)
	if ip == nil && loc != "" {
		ip = ipFromLocation(loc)
	}
	if ip == nil {
		log.Debug("ssdp response skipped; no ip", zap.String("src", src.String()), zap.String("location", loc))
		return
	}
	dd := discovery.NewDevice(ip)
	dd.DisplayName = server
	dd.Services["upnp"] = 0
	dd.Sources["ssdp"] = struct{}{}
	out <- dd
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
