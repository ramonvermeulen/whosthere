// Package discovery provides network device discovery using multiple protocols.
//
// The package enables discovering devices on a local network through various
// methods including ARP cache reading, mDNS, and SSDP.
// It merges results from different sources into unified device records and
// enriches them with manufacturer information via OUI lookups.
//
// # Basic Usage
//
// Create an engine with at least one scanner and a network interface:
//
//	import (
//	    "context"
//	    "fmt"
//	    "time"
//
//	    "github.com/ramonvermeulen/whosthere/pkg/discovery"
//	    "github.com/ramonvermeulen/whosthere/pkg/discovery/scanners/arp"
//	    "github.com/ramonvermeulen/whosthere/pkg/discovery/scanners/mdns"
//	    "github.com/ramonvermeulen/whosthere/pkg/discovery/sweeper"
//	)
//
//	func main() {
//	    // Auto-detect network interface
//	    iface, err := discovery.NewInterfaceInfo("")
//	    if err != nil {
//	        panic(err)
//	    }
//
//	    // Create scanners
//	    arpScanner, _ := arp.New(iface)
//	    mdnsScanner, _ := mdns.New(iface)
//
//	    // Create sweeper to populate ARP cache
//	    sw, _ := sweeper.New(
//	        sweeper.WithSweeperInterface(iface),
//	    )
//
//	    // Create engine
//	    engine, err := discovery.NewEngine(
//	        discovery.WithInterface(iface),
//	        discovery.WithScanners(arpScanner, mdnsScanner),
//	        discovery.WithSweeper(sw),
//	        discovery.WithScanInterval(30 * time.Second),
//	    )
//	    if err != nil {
//	        panic(err)
//	    }
//	    defer engine.Stop()
//
//	    // Start continuous discovery
//	    events := engine.Start(context.Background())
//	    for event := range events {
//	        switch event.Type {
//	        case discovery.EventDeviceDiscovered:
//	            dev := event.Device
//	            fmt.Printf("Found: %s (%s) - %s\n",
//	                dev.IP().String(), dev.MAC(), dev.DisplayName())
//	        case discovery.EventScanCompleted:
//	            fmt.Printf("Scan completed: %d devices in %v\n",
//	                event.Stats.Count, event.Stats.Duration)
//	        case discovery.EventError:
//	            fmt.Printf("Error: %v\n", event.Error)
//	        }
//	    }
//	}
//
// # One-Shot Scanning
//
// For a single scan without continuous monitoring:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
//	defer cancel()
//
//	results, err := engine.Scan(ctx)
//	if err != nil {
//	    panic(err)
//	}
//
//	for _, dev := range results.Devices {
//	    fmt.Printf("%s - %s - %s\n", dev.IP().String(), dev.MAC(), dev.DisplayName())
//	}
//
// # Architecture
//
// The discovery package is built around these core components:
//
//   - Engine: Orchestrates scanners, merges results, emits events
//   - Scanner: Protocol-specific discovery implementation (ARP, mDNS, SSDP)
//   - Sweeper: Populates the ARP cache by triggering network traffic
//   - Device: Unified device record aggregating data from all scanners
//   - Event: Asynchronous notification of discoveries and lifecycle changes
//
// # No Elevated Privileges Required
//
// The package is designed to work without root/admin privileges:
//
//   - ARP reading uses OS-provided cache files/commands
//   - mDNS and SSDP use standard UDP sockets
//   - Sweeper uses UDP/TCP connections, not raw ARP packets
//
// This makes the package suitable for user-space applications and containers.
//
// # API
// As long as the package is in early development (pre-v1.0.0), be aware, the API may change without a major version bump.
package discovery
