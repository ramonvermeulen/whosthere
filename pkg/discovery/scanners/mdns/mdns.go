package mdns

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"golang.org/x/net/dns/dnsmessage"
	"golang.org/x/net/ipv4"
)

var _ discovery.Scanner = (*Scanner)(nil)

const (
	serviceDiscoveryQuery = "_services._dns-sd._udp.local."
	mdnsMulticastAddress  = "224.0.0.251"
	mdnsPort              = 5353
	maxBufferSize         = 16384
)

// Scanner discovers devices using multicast DNS (mDNS), also known as Bonjour or Avahi.
// mDNS is commonly used by printers, smart home devices, Apple devices, and Linux systems
// to advertise services on the local network without requiring a DNS server.
//
// The scanner sends DNS-SD queries and listens for responses containing device names,
// services, IP addresses, and additional metadata (TXT records).
//
// Provides richer information than ARP (device names, service types, metadata) but
// only discovers devices that advertise via mDNS.
type Scanner struct {
	iface  *discovery.InterfaceInfo
	logger discovery.Logger
}

// New creates an mDNS scanner for the specified network interface.
//
// Example:
//
//	import "github.com/ramonvermeulen/whosthere/pkg/discovery/scanners/mdns"
//
//	iface, _ := discovery.NewInterfaceInfo("en0")
//	scanner, err := mdns.New(iface)
//	if err != nil {
//	    log.Fatal(err)
//	}
func New(iface *discovery.InterfaceInfo, opts ...Option) (*Scanner, error) {
	s := &Scanner{iface: iface, logger: discovery.NoOpLogger{}}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Scanner) Name() string {
	return "mdns"
}

// Scan sends mDNS queries and listens for responses until the context is canceled.
// Discovered devices are sent to the out channel as they're found.
//
// The scanner queries for all services (_services._dns-sd._udp.local) and parses responses.
// Multicast responses from other devices on the network are also captured.
//
// Returns when ctx is canceled or on unrecoverable network errors.
func (s *Scanner) Scan(ctx context.Context, out chan<- *discovery.Device) error {
	session := &scanSession{
		logger: s.logger,
		iface:  s.iface,
	}
	return session.run(ctx, out)
}

// scanSession manages state for one mDNS scan
type scanSession struct {
	logger              discovery.Logger
	conn                *net.UDPConn
	multicastAddr       *net.UDPAddr
	iface               *discovery.InterfaceInfo
	queriedServiceTypes map[string]bool
	reportedDevices     map[string]bool
	mu                  sync.RWMutex
}

func (ss *scanSession) setupConnection() (err error) {
	addr, err := net.ResolveUDPAddr("udp4",
		fmt.Sprintf("%s:%d", mdnsMulticastAddress, mdnsPort))
	if err != nil {
		return fmt.Errorf("resolve multicast address: %w", err)
	}

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: *ss.iface.IPv4Addr, Port: 0})
	if err != nil {
		return fmt.Errorf("create UDP socket: %w", err)
	}

	p := ipv4.NewPacketConn(conn)
	if err := p.JoinGroup(ss.iface.Interface, addr); err != nil {
		_ = conn.Close()
		return fmt.Errorf("join multicast group: %w", err)
	}

	ss.conn = conn
	ss.multicastAddr = addr
	return nil
}

func (ss *scanSession) queryService(serviceName string) error {
	msg := dnsmessage.Message{
		Header: dnsmessage.Header{ID: 0, RecursionDesired: false},
		Questions: []dnsmessage.Question{{
			Name:  dnsmessage.MustNewName(serviceName),
			Type:  dnsmessage.TypePTR,
			Class: dnsmessage.ClassINET,
		}},
	}

	packet, err := msg.Pack()
	if err != nil {
		return fmt.Errorf("pack DNS query: %w", err)
	}

	_, err = ss.conn.WriteToUDP(packet, ss.multicastAddr)
	return err
}

func (ss *scanSession) run(ctx context.Context, out chan<- *discovery.Device) error {
	if err := ss.setupConnection(); err != nil {
		return fmt.Errorf("setup connection: %w", err)
	}
	defer func() {
		_ = ss.conn.Close()
	}()

	ss.queriedServiceTypes = make(map[string]bool)
	ss.reportedDevices = make(map[string]bool)

	// Send multiple initial service discovery queries to improve reliability
	// mDNS packets can be dropped, so we send 3 queries with slight delays
	for i := 0; i < 3; i++ {
		if err := ss.queryService(serviceDiscoveryQuery); err != nil {
			return fmt.Errorf("initial service discovery: %w", err)
		}
		if i < 2 {
			time.Sleep(50 * time.Millisecond)
		}
	}

	go ss.periodicQuerySender(ctx)

	return ss.listenForResponses(ctx, out)
}

func (ss *scanSession) periodicQuerySender(ctx context.Context) {
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Send service discovery query twice per tick for better reliability
			_ = ss.queryService(serviceDiscoveryQuery)
			time.Sleep(10 * time.Millisecond)
			_ = ss.queryService(serviceDiscoveryQuery)

			ss.mu.RLock()
			services := make([]string, 0, len(ss.queriedServiceTypes))
			for serviceType := range ss.queriedServiceTypes {
				services = append(services, serviceType)
			}
			ss.mu.RUnlock()

			// Query each discovered service type twice
			for _, service := range services {
				_ = ss.queryService(service)
				time.Sleep(5 * time.Millisecond)
				_ = ss.queryService(service)
			}
		}
	}
}

func (ss *scanSession) listenForResponses(ctx context.Context, out chan<- *discovery.Device) error {
	buffer := make([]byte, maxBufferSize)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Set a short read deadline so we can check context periodically
			_ = ss.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

			packetSize, sender, err := ss.conn.ReadFromUDP(buffer)
			if err != nil {
				if isTimeout(err) {
					// Timeout is normal, continue listening
					continue
				}
				// Real error, but check if context is done first
				if ctx.Err() != nil {
					return nil
				}
				return fmt.Errorf("read UDP packet: %w", err)
			}

			// Process the received packet
			dnsMsg, err := parseDNSMessage(buffer[:packetSize])
			if err != nil {
				// Skip malformed packets
				continue
			}

			if dnsMsg.Response {
				ss.processDNSResponse(dnsMsg, sender, out)
			}
		}
	}
}

// processDNSResponse handles all records in one DNS message
func (ss *scanSession) processDNSResponse(msg *dnsmessage.Message, sender *net.UDPAddr, out chan<- *discovery.Device) {
	for _, answer := range msg.Answers {
		if ptr, ok := answer.Body.(*dnsmessage.PTRResource); ok {
			serviceName := answer.Header.Name.String()
			ptrValue := ptr.PTR.String()

			if serviceName == serviceDiscoveryQuery {
				// This is a service type announcement (e.g., "_http._tcp.local")
				ss.handleDiscoveredServiceType(ptrValue)
			} else {
				// This is a device announcement (e.g., "My Device._http._tcp.local")
				ss.handleDiscoveredDevice(&answer, ptrValue, sender, out)
			}
		}
	}

	ss.extractDeviceDetails(msg.Additionals, sender, out)
}

func (ss *scanSession) handleDiscoveredServiceType(serviceType string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.queriedServiceTypes[serviceType] {
		return
	}

	ss.queriedServiceTypes[serviceType] = true

	for i := 0; i < 3; i++ {
		if err := ss.queryService(serviceType); err != nil {
			ss.logger.Log(context.Background(), slog.LevelWarn, "query service type", "serviceType", serviceType, "error", err)
			return
		}
		if i < 2 {
			time.Sleep(20 * time.Millisecond)
		}
	}
}

func (ss *scanSession) handleDiscoveredDevice(
	answer *dnsmessage.Resource,
	ptrValue string,
	sender *net.UDPAddr,
	out chan<- *discovery.Device,
) {
	deviceID := fmt.Sprintf("%s-%s", sender.IP.String(), ptrValue)

	if ss.reportedDevices[deviceID] {
		return
	}

	device := discovery.NewDevice(sender.IP)
	device.SetDisplayName(cleanDisplayName(ptrValue))
	device.AddSource("mdns")

	select {
	case out <- device:
		ss.reportedDevices[deviceID] = true
	default:
	}
}

func (ss *scanSession) extractDeviceDetails(
	records []dnsmessage.Resource,
	sender *net.UDPAddr,
	out chan<- *discovery.Device,
) {
	if len(records) == 0 {
		return
	}

	device := discovery.NewDevice(sender.IP)
	device.AddSource("mdns")

	for _, record := range records {
		switch r := record.Body.(type) {
		case *dnsmessage.SRVResource:
			device.SetDisplayName(cleanDisplayName(r.Target.String()))
		case *dnsmessage.TXTResource:
			ss.parseTXTRecords(r, device)
		}
	}

	if device.DisplayName() != "" {
		select {
		case out <- device:
		default:
		}
	}
}

// parseTXTRecords extracts device details from TXT records
// see https://datatracker.ietf.org/doc/html/rfc6763#section-6.3
// it implements common keys used by various devices
func (ss *scanSession) parseTXTRecords(txt *dnsmessage.TXTResource, device *discovery.Device) {
	for _, text := range txt.TXT {
		// Split key=value
		if idx := strings.IndexByte(text, '='); idx > 0 {
			key := strings.ToLower(text[:idx])
			value := text[idx+1:]

			switch key {
			case "manufacturer":
				device.SetManufacturer(value)
			case "mac":
				device.SetMAC(value)
			// todo(ramon): think about device merge strategy, often `md` is a better display name, however at this point often other scanners have already set a name
			case "md":
				device.SetDisplayName(value)
			default:
				device.AddExtraData(key, value)
			}
		} else {
			device.AddExtraData(text, "true")
		}
	}
}

// utils
// todo(ramon): after multiple scanner implementations look for overlap and move to common package
func parseDNSMessage(data []byte) (*dnsmessage.Message, error) {
	var msg dnsmessage.Message
	err := msg.Unpack(data)
	return &msg, err
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func cleanDisplayName(name string) string {
	name = strings.TrimSuffix(name, ".local.")
	return strings.TrimSuffix(name, ".")
}

func extractServiceNameFromTarget(target string) string {
	parts := strings.Split(target, ".")
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimPrefix(parts[0], "_")
}
