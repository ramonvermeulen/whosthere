package discovery

import "context"

// Scanner defines a discovery strategy (SSDP, mDNS, ARP, etc.).
type Scanner interface {
	Name() string
	Scan(ctx context.Context, out chan<- Device) error
}
