package discovery

import (
	"context"
	"net"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
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

// Sweeper triggers ARP resolution to populate the OS ARP cache.
// Because whosthere is designed to not run with elevated privileges,
// it cannot send ARP requests directly. Instead, it triggers ARP resolution
// by sending UDP/TCP packets to IPs in the target subnet. This causes the OS
// to send ARP requests for those IPs, populating the ARP cache which can
// then be read by the ARP scanner.
type Sweeper struct {
	iface    *InterfaceInfo
	interval time.Duration

	logger *zap.Logger
}

func NewSweeper(iface *InterfaceInfo, interval time.Duration) *Sweeper {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	logger := zap.L().With(
		zap.String("scanner", "arp"),
		zap.String("component", "Sweeper"),
	)
	return &Sweeper{
		iface:    iface,
		interval: interval,
		logger:   logger,
	}
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

	// run the initial sweep so that we don't have to wait for the first tick
	go s.runSweep(ctx, subnet, localIP)
}

func (s *Sweeper) runSweep(ctx context.Context, subnet *net.IPNet, localIP net.IP) {
	ips := generateSubnetIPs(subnet, localIP)
	if len(ips) == 0 {
		return
	}

	s.logger.Debug("Triggering ARP requests for subnet", zap.String("subnet", subnet.Mask.String()))
	triggerSubnetSweep(ctx, ips)
	s.logger.Debug("ARP triggering completed", zap.String("subnet", subnet.String()))
}

func triggerSubnetSweep(ctx context.Context, ips []net.IP) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrentTriggers)
	total := len(ips)
	triggered := 0

	for _, ip := range ips {
		zap.L().Debug("Triggering ARP for IP", zap.String("ip", ip.String()))
		select {
		case <-ctx.Done():
			zap.L().Warn("ARP sweep interrupted by context cancellation, this can indicate you have a short scan duration configured",
				zap.Int("triggered", triggered),
				zap.Int("total", total),
				zap.Int("remaining", total-triggered))
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
func generateSubnetIPs(subnet *net.IPNet, skipIP net.IP) []net.IP {
	// If users request it, we could potentially add an option to override the /16 limit via configuration?
	var ips []net.IP
	network := subnet.IP.To4()
	if network == nil {
		return ips
	}

	ones, _ := subnet.Mask.Size()
	if ones < 16 {
		zap.L().Warn("large subnet detected, limiting ARP scan to /16 equivalent", zap.Int("prefix", ones), zap.String("subnet", subnet.String()))
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
