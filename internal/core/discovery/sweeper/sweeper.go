package sweeper

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
)

const (
	maxConcurrentTriggers = 200
	triggerDeadline       = 300 * time.Millisecond
	tcpDialTimeout        = 300 * time.Millisecond
)

var (
	udpTriggerPorts = []int{9, 33434}
	tcpTriggerPorts = []int{80, 443}
)

var _ discovery.Sweeper = (*Sweeper)(nil)

// Sweeper triggers ARP resolution to populate the OS ARP cache.
// Because whosthere is designed to not run with elevated privileges,
// it cannot send ARP requests directly. Instead, it triggers ARP resolution
// by sending UDP/TCP packets to IPs in the target subnet. This causes the OS
// to send ARP requests for those IPs, populating the ARP cache which can
// then be read by the ARP scanner.
type Sweeper struct {
	iface    *discovery.InterfaceInfo
	interval time.Duration
	timeout  time.Duration
	logger   discovery.Logger
}

func NewSweeper(opts ...Option) (*Sweeper, error) {
	s := &Sweeper{
		interval: discovery.DefaultSweepInterval,
		timeout:  discovery.DefaultSweepTimeout,
		logger:   &discovery.NoOpLogger{},
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	if s.iface == nil {
		return nil, errors.New("interface is required for sweeper")
	}

	return s, nil
}

func (s *Sweeper) Start(ctx context.Context) {
	subnet := s.iface.IPv4Net
	localIP := *s.iface.IPv4Addr

	ticker := time.NewTicker(s.interval)
	go func() {
		defer ticker.Stop()
		defer ctx.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.runSweep(ctx, subnet, localIP)
			}
		}
	}()

	go s.runSweep(ctx, subnet, localIP)
}

func (s *Sweeper) runSweep(ctx context.Context, subnet *net.IPNet, localIP net.IP) {
	ips := s.generateSubnetIPs(subnet, localIP)
	if len(ips) == 0 {
		return
	}

	s.logger.Log(ctx, slog.LevelDebug, "Triggering ARP requests for subnet", "subnet", subnet.Mask.String())
	s.triggerSubnetSweep(ctx, ips)
	s.logger.Log(ctx, slog.LevelDebug, "ARP triggering completed", "subnet", subnet.String())
}

func (s *Sweeper) triggerSubnetSweep(ctx context.Context, ips []net.IP) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrentTriggers)
	total := len(ips)
	triggered := 0

	for _, ip := range ips {
		s.logger.Log(ctx, slog.LevelDebug, "Triggering ARP for IP", "ip", ip.String())
		select {
		case <-ctx.Done():
			s.logger.Log(ctx, slog.LevelWarn, "ARP sweep interrupted by context cancellation, this can indicate you have a short scan duration configured", "triggered", triggered, "total", total, "remaining", total-triggered)
			return
		default:
		}

		wg.Add(1)
		sem <- struct{}{}
		triggered++

		go func(targetIP net.IP) {
			defer wg.Done()
			defer func() { <-sem }()
			sendARPTarget(targetIP)
		}(ip)
	}

	wg.Wait()
}

func sendARPTarget(ip net.IP) {
	deadline := time.Now().Add(triggerDeadline)

	for _, p := range udpTriggerPorts {
		addr := &net.UDPAddr{IP: ip, Port: p}
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			continue
		}
		_ = conn.SetWriteDeadline(deadline)
		_, _ = conn.Write([]byte{0})
		_ = conn.Close()
	}

	for _, p := range tcpTriggerPorts {
		addr := net.JoinHostPort(ip.String(), strconv.Itoa(p))
		c, err := net.DialTimeout("tcp", addr, tcpDialTimeout)
		if err == nil {
			_ = c.Close()
		}
	}
}

// generateSubnetIPs generates a list of IPs in the given subnet,
// Skipping the specified IP (usually the interface's own IP).
// It includes the network address and broadcast address.
// It limits the scan to a /16 equivalent if the subnet is larger.
// In that case it will only scan the first 65534 IPs of that subnet.
func (s *Sweeper) generateSubnetIPs(subnet *net.IPNet, skipIP net.IP) []net.IP {
	// If users request it, we could potentially add an option to override the /16 limit via configuration?
	var ips []net.IP
	network := subnet.IP.To4()
	if network == nil {
		return ips
	}

	ones, _ := subnet.Mask.Size()
	if ones < 16 {
		s.logger.Log(context.Background(), slog.LevelWarn, "large subnet detected, limiting ARP scan to /16 equivalent", "prefix", ones, "subnet", subnet.String())
	}

	networkIP := subnet.IP.Mask(subnet.Mask)
	broadcastIP := make(net.IP, len(networkIP))
	copy(broadcastIP, networkIP)

	effectiveMask := subnet.Mask
	if ones < 16 {
		effectiveMask = net.CIDRMask(16, 32)
	}
	for i := range network {
		// sets broadcast IP to a /16 equivalent if subnet is larger
		broadcastIP[i] |= ^effectiveMask[i]
	}

	currentIP := make(net.IP, len(networkIP))
	copy(currentIP, networkIP)

	for {
		if !currentIP.Equal(skipIP) {
			ipCopy := make(net.IP, len(currentIP))
			copy(ipCopy, currentIP)
			ips = append(ips, ipCopy)
		}
		if currentIP.Equal(broadcastIP) {
			break
		}
		currentIP = incrementIP(currentIP)
	}

	return ips
}

// incrementIP increments the IP address by 1
func incrementIP(ip net.IP) net.IP {
	newIP := make(net.IP, len(ip))
	copy(newIP, ip)
	for i := len(newIP) - 1; i >= 0; i-- {
		newIP[i]++
		if newIP[i] != 0 {
			break
		}
	}
	return newIP
}
