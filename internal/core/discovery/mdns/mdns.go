package mdns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
	"go.uber.org/zap"
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

type Scanner struct {
	iface *discovery.InterfaceInfo
}

func NewScanner(iface *discovery.InterfaceInfo) *Scanner {
	return &Scanner{iface: iface}
}

func (s *Scanner) Name() string {
	return "mdns"
}

func (s *Scanner) Scan(ctx context.Context, out chan<- discovery.Device) error {
	session := &scanSession{
		log:   zap.L().Named("mdns"),
		iface: s.iface,
	}
	return session.run(ctx, out)
}

// scanSession manages state for one mDNS scan
type scanSession struct {
	log                 *zap.Logger
	conn                *net.UDPConn
	multicastAddr       *net.UDPAddr
	iface               *discovery.InterfaceInfo
	queriedServiceTypes map[string]bool
	reportedDevices     map[string]bool
}

func (ss *scanSession) setupConnection() error {
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

func (ss *scanSession) run(ctx context.Context, out chan<- discovery.Device) error {
	if err := ss.setupConnection(); err != nil {
		return fmt.Errorf("setup connection: %w", err)
	}
	defer func() {
		_ = ss.conn.Close()
	}()

	ss.queriedServiceTypes = make(map[string]bool)
	ss.reportedDevices = make(map[string]bool)

	// Sends the initial service discovery query (multicast)
	// See https://datatracker.ietf.org/doc/html/rfc6763#section-9
	if err := ss.queryService(serviceDiscoveryQuery); err != nil {
		return fmt.Errorf("initial service discovery: %w", err)
	}

	// Listens for the mDNS responses until the timeout has reached
	return ss.listenForResponses(ctx, out)
}

func (ss *scanSession) listenForResponses(ctx context.Context, out chan<- discovery.Device) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		return fmt.Errorf("context has no deadline")
	}
	timeout := time.Until(deadline)
	if err := ss.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return fmt.Errorf("set read deadline: %w", err)
	}

	buffer := make([]byte, maxBufferSize)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			shouldExit, err := ss.readAndProcessPacket(buffer, out)
			if err != nil || shouldExit {
				return err
			}
		}
	}
}

// readAndProcessPacket handles ONE complete mDNS response packet
func (ss *scanSession) readAndProcessPacket(buffer []byte, out chan<- discovery.Device) (bool, error) {
	packetSize, sender, err := ss.conn.ReadFromUDP(buffer)
	if err != nil {
		if isTimeout(err) {
			return true, nil
		}
		return false, fmt.Errorf("read UDP packet: %w", err)
	}

	dnsMsg, err := parseDNSMessage(buffer[:packetSize])
	if err != nil {
		ss.log.Debug("Failed to parse DNS", zap.Error(err))
		return false, nil
	}

	if !dnsMsg.Response {
		return false, nil
	}

	ss.processDNSResponse(dnsMsg, sender, out)
	return false, nil
}

// processDNSResponse handles all records in one DNS message
func (ss *scanSession) processDNSResponse(msg *dnsmessage.Message, sender *net.UDPAddr, out chan<- discovery.Device) {
	for _, answer := range msg.Answers {
		if ptr, ok := answer.Body.(*dnsmessage.PTRResource); ok {
			serviceName := answer.Header.Name.String()
			ptrValue := ptr.PTR.String()

			ss.log.Debug("Got PTR record",
				zap.String("question", serviceName),
				zap.String("points_to", ptrValue))

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
	if ss.queriedServiceTypes[serviceType] {
		return
	}

	ss.queriedServiceTypes[serviceType] = true
	ss.log.Debug("Discovered new service type", zap.String("type", serviceType))

	if err := ss.queryService(serviceType); err != nil {
		ss.log.Warn("Failed to query service",
			zap.String("service", serviceType),
			zap.Error(err))
	}
}

func (ss *scanSession) handleDiscoveredDevice(
	answer *dnsmessage.Resource,
	ptrValue string,
	sender *net.UDPAddr,
	out chan<- discovery.Device,
) {
	deviceID := fmt.Sprintf("%s-%s", sender.IP.String(), ptrValue)

	if ss.reportedDevices[deviceID] {
		return
	}

	now := time.Now()
	device := &discovery.Device{
		IP:          sender.IP,
		DisplayName: cleanDisplayName(ptrValue),
		Services:    make(map[string]int),
		Sources:     map[string]struct{}{"mdns": {}},
		ExtraData:   make(map[string]string),
		FirstSeen:   now,
		LastSeen:    now,
	}

	if service := extractServiceName(answer.Header.Name.String()); service != "" {
		device.Services[service] = 0
	}

	out <- *device
	ss.reportedDevices[deviceID] = true
}

func (ss *scanSession) extractDeviceDetails(
	records []dnsmessage.Resource,
	sender *net.UDPAddr,
	out chan<- discovery.Device,
) {
	if len(records) == 0 {
		return
	}

	device := discovery.NewDevice(sender.IP)
	device.Sources["mdns"] = struct{}{}

	for _, record := range records {
		switch r := record.Body.(type) {
		case *dnsmessage.SRVResource:
			device.DisplayName = cleanDisplayName(r.Target.String())
			if service := extractServiceNameFromTarget(r.Target.String()); service != "" {
				device.Services[service] = int(r.Port)
			}
		case *dnsmessage.TXTResource:
			ss.parseTXTRecords(r, &device)
		}
	}

	if len(device.Services) > 0 || device.DisplayName != "" {
		out <- device
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
				device.Manufacturer = value
			case "mac":
				device.MAC = value
			// todo(ramon): think about device merge strategy, often `md` is a better display name, however at this point often other scanners have already set a name
			case "md":
				device.DisplayName = value
			default:
				if device.ExtraData == nil {
					device.ExtraData = make(map[string]string)
				}
				device.ExtraData[key] = value
			}
		} else {
			if device.ExtraData == nil {
				device.ExtraData = make(map[string]string)
			}
			device.ExtraData[text] = "true"
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

func extractServiceName(dnsName string) string {
	parts := strings.Split(dnsName, ".")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimPrefix(parts[0], "_")
}

func extractServiceNameFromTarget(target string) string {
	parts := strings.Split(target, ".")
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimPrefix(parts[0], "_")
}
