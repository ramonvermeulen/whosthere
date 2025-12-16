package discovery

import (
	"net"
	"time"
)

// TODO(ramon): Maybe it could be nice to have a merge strategy? E.g. when multiple scanners return the same device.
// Per column we could for example choose to have a specific Merge strategy, so that the "best" data from different
// scanners could be combined into a single device representation.
// maybe something like?:
// const (
//    MergeFirstWins     MergeStrategy = iota   // First scanner's value wins
//    MergeMostSpecific                         // More specific value wins
//    MergeLongest                              // Longer value (usually more complete)
//    MergeUnion                                // Combine arrays/maps
//    MergeMostRecent                           // Newer timestamp wins
// )
// END TODO

// Device represents a discovered network device aggregated from multiple scanners.
type Device struct {
	IP           net.IP              // Primary IP address (identity key)
	MAC          string              // MAC if known
	DisplayName  string              // Reverse DNS or reported name
	Manufacturer string              // Vendor from OUI or protocol metadata
	Model        string              // Reported model
	Services     map[string]int      // service name -> port (or 0 if unknown)
	Sources      map[string]struct{} // scanners that contributed info
	LastSeen     time.Time           // last time any scanner saw the device
	Extras       map[string]string   // additional key/value metadata discovered from protocols (TXT, SSDP, etc.)
}

// NewDevice builds a Device with initialized maps and current timestamp as last seen.
func NewDevice(ip net.IP) Device {
	return Device{IP: ip, Services: map[string]int{}, Sources: map[string]struct{}{}, LastSeen: time.Now(), Extras: map[string]string{}}
}

// Merge merges fields from 'other' into d, preferring non-empty/newer data and unioning maps.
func (d *Device) Merge(other Device) {
	if d.IP == nil && other.IP != nil {
		d.IP = other.IP
	}
	if d.MAC == "" && other.MAC != "" {
		d.MAC = other.MAC
	}
	if d.DisplayName == "" && other.DisplayName != "" {
		d.DisplayName = other.DisplayName
	}
	if d.Manufacturer == "" && other.Manufacturer != "" {
		d.Manufacturer = other.Manufacturer
	}
	if d.Model == "" && other.Model != "" {
		d.Model = other.Model
	}
	if d.Services == nil {
		d.Services = map[string]int{}
	}
	for name, port := range other.Services {
		if _, ok := d.Services[name]; !ok || d.Services[name] == 0 {
			if d.Services[name] == 0 || port != 0 {
				d.Services[name] = port
			}
		}
	}
	if d.Sources == nil {
		d.Sources = map[string]struct{}{}
	}
	for src := range other.Sources {
		d.Sources[src] = struct{}{}
	}
	if d.Extras == nil {
		d.Extras = map[string]string{}
	}
	for k, v := range other.Extras {
		// prefer existing value; only set if missing
		if _, ok := d.Extras[k]; !ok {
			d.Extras[k] = v
		}
	}
	if other.LastSeen.After(d.LastSeen) {
		d.LastSeen = other.LastSeen
	}
}
