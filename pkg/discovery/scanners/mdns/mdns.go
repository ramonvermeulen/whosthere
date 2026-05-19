package mdns

import (
	"context"
	"io"
	"log"
	"log/slog"
	"net"
	"strings"
	"time"

	hashimdns "github.com/hashicorp/mdns"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

const (
	serviceDiscoveryQuery = "_services._dns-sd._udp"
)

var _ discovery.Scanner = (*Scanner)(nil)

// Scanner implements the discovery.Scanner interface using hashicorp/mdns
// Only a minimal, idiomatic implementation is provided for maintainability.
type Scanner struct {
	iface  *discovery.InterfaceInfo
	logger discovery.Logger
	// queryFunc allows injection of a mock mDNS query function for testing.
	// This is only settable via a test-only Option and should not be changed in production code.
	// It exists solely to enable safe, race-free unit testing without global state.
	queryFunc func(params *hashimdns.QueryParam) error
}

// New creates an mDNS scanner for the specified network interface.
func New(iface *discovery.InterfaceInfo, opts ...Option) (*Scanner, error) {
	s := &Scanner{iface: iface, logger: discovery.NoOpLogger{}, queryFunc: hashimdns.Query}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Name returns the scanner name for engine compatibility.
func (s *Scanner) Name() string {
	return "mdns"
}

func (s *Scanner) Scan(ctx context.Context, results chan<- *discovery.Device) error {
	entriesCh := make(chan *hashimdns.ServiceEntry, 256)
	errCh := make(chan error, 1)

	go func() {
		params := hashimdns.DefaultParams(serviceDiscoveryQuery)
		params.Entries = entriesCh
		params.Interface = s.iface.Interface
		params.DisableIPv6 = true
		params.Logger = log.Default()
		params.Logger.SetOutput(io.Discard)

		if deadline, ok := ctx.Deadline(); ok {
			params.Timeout = time.Until(deadline)
		}

		if err := s.queryFunc(params); err != nil {
			errCh <- err
		}
		close(entriesCh)
		close(errCh)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case entry, ok := <-entriesCh:
			if !ok {
				select {
				case err := <-errCh:
					return err
				default:
					return nil
				}
			}

			if !s.acceptsEntry(entry) {
				s.logger.Log(ctx, slog.LevelDebug, "ignoring mDNS device outside selected interface network", "name", entry.Name, "ip", entry.AddrV4.String())
				continue
			}

			dev := discovery.NewDevice(entry.AddrV4)
			dev.SetDisplayName(entry.Name)
			if s.iface != nil && s.iface.Interface != nil {
				dev.SetInterfaceName(s.iface.Interface.Name)
			}
			dev.AddSource("mdns")
			if entry.Info != "" {
				fields := entry.InfoFields
				for _, f := range fields {
					if kv := splitKeyValue(f); kv != nil {
						dev.AddExtraData(kv[0], kv[1])
					} else {
						dev.AddExtraData(f, "true")
					}
				}
			}

			s.logger.Log(ctx, slog.LevelDebug, "discovered device via mDNS", "name", entry.Name, "ip", entry.AddrV4.String())

			select {
			case results <- dev:
			case <-ctx.Done():
				return ctx.Err()
			}

		case err := <-errCh:
			return err
		}
	}
}

func (s *Scanner) acceptsEntry(entry *hashimdns.ServiceEntry) bool {
	if entry == nil {
		return false
	}

	return s.acceptsIPv4(entry.AddrV4)
}

func (s *Scanner) acceptsIPv4(ip net.IP) bool {
	if ip == nil {
		return false
	}

	if s.iface == nil || s.iface.IPv4Net == nil {
		return true
	}

	ipv4 := ip.To4()
	if ipv4 == nil {
		return false
	}

	return s.iface.IPv4Net.Contains(ipv4)
}

// splitKeyValue splits a string like "key=value" and returns [key, value], or nil if not present.
func splitKeyValue(s string) []string {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) == 2 {
		return parts
	}
	return nil
}
