package discovery

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// Dialer abstracts network connection creation for testability.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// PortScanner performs TCP port scanning on network devices.
type PortScanner struct {
	workers int
	dialer  Dialer
	iface   *InterfaceInfo
}

// NewPortScanner creates a PortScanner with the specified number of concurrent workers.
// More workers scan faster but consume more system resources (file descriptors, memory).
// The scanner binds to the provided interface's IPv4 address.
//
// Example:
//
//	iface, _ := discovery.NewInterfaceInfo("en0")
//	scanner := discovery.NewPortScanner(20, iface)
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

// Stream scans TCP ports on the target IP address and calls the callback for each open port.
// Scanning happens concurrently using the configured number of workers.
// The callback is invoked from multiple goroutines - ensure it's thread-safe.
func (ps *PortScanner) Stream(ctx context.Context, ip string, ports []int, timeout time.Duration, callback func(int)) error {
	// todo(ramon): consider using a channel for results instead of a callback, just like the Engine does for devices.
	// This would allow more flexible handling of results and better integration with the rest of the system.
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
