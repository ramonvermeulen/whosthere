package discovery

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// Dialer interface for testing.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// PortScanner scans ports on a given IP address.
type PortScanner struct {
	workers int
	dialer  Dialer
	iface   *InterfaceInfo
}

// NewPortScanner creates a new PortScanner.
func NewPortScanner(workers int, iface *InterfaceInfo) *PortScanner {
	return &PortScanner{
		workers: workers,
		dialer:  &netDialer{iface: iface},
		iface:   iface,
	}
}

// netDialer implements Dialer using net.Dialer.
type netDialer struct {
	iface *InterfaceInfo
}

func (d *netDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	var dialer net.Dialer
	dialer.LocalAddr = &net.TCPAddr{IP: *d.iface.IPv4Addr}
	return dialer.DialContext(ctx, network, address)
}

// Stream scans the TCP ports on the given IP and calls the callback for each open port found.
// It uses the provided context for cancellation.
// TODO: instead of using a callback, maybe expose a channel for results?
func (ps *PortScanner) Stream(ctx context.Context, ip string, ports []int, timeout time.Duration, callback func(int)) error {
	if len(ports) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	portChan := make(chan int, ps.workers)
	errChan := make(chan error, 1)

	for i := 0; i < ps.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ps.streamWorker(ctx, ip, portChan, callback, timeout)
		}()
	}

	go func() {
		defer close(portChan)
		for _, port := range ports {
			select {
			case portChan <- port:
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Return first error if any
	if err := <-errChan; err != nil {
		return err
	}
	return ctx.Err()
}

// streamWorker performs the actual port scanning for streaming.
func (ps *PortScanner) streamWorker(ctx context.Context, ip string, ports <-chan int, callback func(int), timeout time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		case port, ok := <-ports:
			if !ok {
				return
			}
			if ps.isPortOpen(ctx, ip, port, timeout) {
				callback(port)
			}
		}
	}
}

// isPortOpen checks if a TCP port is open using context-aware dialing.
func (ps *PortScanner) isPortOpen(ctx context.Context, ip string, port int, timeout time.Duration) bool {
	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	conn, err := ps.dialer.DialContext(dialCtx, "tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return false
	}
	err = conn.Close()
	return err == nil
}
