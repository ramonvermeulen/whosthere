package discovery

import (
	"encoding/json"
	"net"
	"sync"
	"time"
)

// Device represents a discovered network device with information aggregated
// from multiple discovery protocols (ARP, mDNS, SSDP, etc.).
//
// All fields are private and accessed through thread-safe getters/setters.
// The Device must always be used as a pointer (*Device) to ensure thread-safety.
//
// Fields populated as more information becomes available during scans:
//   - ip: The device's IPv4 address (primary identifier, never nil for valid devices)
//   - mac: Hardware address in colon-separated format (e.g., "aa:bb:cc:dd:ee:ff")
//   - displayName: Human-readable name from mDNS, SSDP, or other protocols
//   - manufacturer: Vendor name derived from the MAC address OUI prefix
//   - sources: Set of scanner names that contributed data (e.g., {"arp-cache", "mdns"})
//   - firstSeen: When this device was first discovered
//   - lastSeen: Most recent discovery time
//   - extraData: Protocol-specific metadata (e.g., SSDP device type, mDNS TXT records)
//   - openPorts: Results from port scans, organized by protocol (not serialized to JSON)
//   - lastPortScan: Timestamp of the most recent port scan (not serialized to JSON)
//
// Devices are uniquely identified by their IP address. When the same IP is seen
// by multiple scanners, their data is merged using the Merge method.
type Device struct {
	mu           sync.RWMutex
	ip           net.IP
	mac          string
	displayName  string
	manufacturer string
	sources      map[string]struct{}
	firstSeen    time.Time
	lastSeen     time.Time
	extraData    map[string]string
	openPorts    map[string][]int
	lastPortScan time.Time
}

// NewDevice creates a Device with the given IP address and initializes all maps.
// FirstSeen and LastSeen are set to the current time. Use this when creating
// devices from scanner implementations.
//
// Example:
//
//	device := discovery.NewDevice(net.ParseIP("192.168.1.100"))
//	device.SetMAC("aa:bb:cc:dd:ee:ff")
//	device.SetDisplayName("Living Room Speaker")
func NewDevice(ip net.IP) *Device {
	now := time.Now()
	return &Device{
		ip:        ip,
		sources:   make(map[string]struct{}),
		firstSeen: now,
		lastSeen:  now,
		extraData: make(map[string]string),
		openPorts: make(map[string][]int),
	}
}

// Merge combines information from another Device into this one.
// Fields are merged as follows:
//   - ip: copied if missing
//   - mac: copied if missing
//   - displayName: copied if missing
//   - manufacturer: copied if missing
//   - sources: union of all sources
//   - extraData: merged, new keys added
//   - firstSeen: earliest time
//   - lastSeen: latest time
//
// Thread-safe: both devices are locked during the operation.
//
// Example:
//
//	base := NewDevice(ip)
//	base.SetMAC("aa:bb:cc:dd:ee:ff")
//	base.AddSource("arp")
//
//	other := NewDevice(ip)
//	other.SetDisplayName("My Device")
//	other.AddSource("mdns")
//
//	// Merge combines both
//	base.Merge(other)
//	// Result: base has both MAC and DisplayName, sources = {"arp", "mdns"}
func (d *Device) Merge(other *Device) {
	if other == nil {
		return
	}
	if d == other {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	other.mu.RLock()
	defer other.mu.RUnlock()

	if d.ip == nil && other.ip != nil {
		d.ip = other.ip
	}
	if d.mac == "" && other.mac != "" {
		d.mac = other.mac
	}
	if d.displayName == "" && other.displayName != "" {
		d.displayName = other.displayName
	}
	if d.manufacturer == "" && other.manufacturer != "" {
		d.manufacturer = other.manufacturer
	}
	if d.sources == nil {
		d.sources = make(map[string]struct{})
	}
	for src := range other.sources {
		d.sources[src] = struct{}{}
	}
	if d.extraData == nil {
		d.extraData = make(map[string]string)
	}
	for k, v := range other.extraData {
		if _, ok := d.extraData[k]; !ok {
			d.extraData[k] = v
		}
	}
	if d.firstSeen.IsZero() || (!other.firstSeen.IsZero() && other.firstSeen.Before(d.firstSeen)) {
		d.firstSeen = other.firstSeen
	}
	if other.lastSeen.After(d.lastSeen) {
		d.lastSeen = other.lastSeen
	}
	if d.openPorts == nil {
		d.openPorts = make(map[string][]int)
	}
	for protocol, ports := range other.openPorts {
		if _, ok := d.openPorts[protocol]; !ok {
			d.openPorts[protocol] = make([]int, len(ports))
			copy(d.openPorts[protocol], ports)
		} else {
			portSet := make(map[int]bool)
			for _, p := range d.openPorts[protocol] {
				portSet[p] = true
			}
			for _, p := range ports {
				if !portSet[p] {
					d.openPorts[protocol] = append(d.openPorts[protocol], p)
					portSet[p] = true
				}
			}
		}
	}
	if other.lastPortScan.After(d.lastPortScan) {
		d.lastPortScan = other.lastPortScan
	}
}

// IP returns a copy of the device's IP address.
func (d *Device) IP() net.IP {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.ip == nil {
		return nil
	}
	return append(net.IP(nil), d.ip...)
}

// MAC returns the device's MAC address.
func (d *Device) MAC() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.mac
}

// DisplayName returns the device's display name.
func (d *Device) DisplayName() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.displayName
}

// Manufacturer returns the device's manufacturer.
func (d *Device) Manufacturer() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.manufacturer
}

// Sources returns a copy of the device's sources map.
func (d *Device) Sources() map[string]struct{} {
	d.mu.RLock()
	defer d.mu.RUnlock()
	m := make(map[string]struct{}, len(d.sources))
	for k, v := range d.sources {
		m[k] = v
	}
	return m
}

// FirstSeen returns the first seen time.
func (d *Device) FirstSeen() time.Time {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.firstSeen
}

// LastSeen returns the last seen time.
func (d *Device) LastSeen() time.Time {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.lastSeen
}

// ExtraData returns a copy of the extra data map.
func (d *Device) ExtraData() map[string]string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	m := make(map[string]string, len(d.extraData))
	for k, v := range d.extraData {
		m[k] = v
	}
	return m
}

// OpenPorts returns a deep copy of the open ports map.
func (d *Device) OpenPorts() map[string][]int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	m := make(map[string][]int, len(d.openPorts))
	for k, v := range d.openPorts {
		m[k] = append([]int(nil), v...)
	}
	return m
}

// LastPortScan returns the last port scan time.
func (d *Device) LastPortScan() time.Time {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.lastPortScan
}

// SetIP sets the device's IP address.
func (d *Device) SetIP(ip net.IP) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if ip == nil {
		d.ip = nil
	} else {
		d.ip = append(net.IP(nil), ip...)
	}
}

// SetMAC sets the device's MAC address.
func (d *Device) SetMAC(mac string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.mac = mac
}

// SetDisplayName sets the device's display name.
func (d *Device) SetDisplayName(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.displayName = name
}

// SetManufacturer sets the device's manufacturer.
func (d *Device) SetManufacturer(manufacturer string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.manufacturer = manufacturer
}

// SetSources sets the device's sources map.
func (d *Device) SetSources(sources map[string]struct{}) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.sources = make(map[string]struct{}, len(sources))
	for k, v := range sources {
		d.sources[k] = v
	}
}

// SetFirstSeen sets the first seen time.
func (d *Device) SetFirstSeen(t time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.firstSeen = t
}

// SetLastSeen sets the last seen time.
func (d *Device) SetLastSeen(t time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.lastSeen = t
}

// SetExtraData sets the extra data map.
func (d *Device) SetExtraData(data map[string]string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.extraData = make(map[string]string, len(data))
	for k, v := range data {
		d.extraData[k] = v
	}
}

// SetOpenPorts sets the open ports map.
func (d *Device) SetOpenPorts(ports map[string][]int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.openPorts = make(map[string][]int, len(ports))
	for k, v := range ports {
		d.openPorts[k] = append([]int(nil), v...)
	}
}

// SetLastPortScan sets the last port scan time.
func (d *Device) SetLastPortScan(t time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.lastPortScan = t
}

// AddSource adds a scanner source to the device.
func (d *Device) AddSource(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.sources == nil {
		d.sources = make(map[string]struct{})
	}
	d.sources[name] = struct{}{}
}

// AddExtraData adds a key-value pair to extra data.
func (d *Device) AddExtraData(key, value string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.extraData == nil {
		d.extraData = make(map[string]string)
	}
	d.extraData[key] = value
}

// Copy creates a deep copy of the device.
func (d *Device) Copy() *Device {
	d.mu.RLock()
	defer d.mu.RUnlock()

	newD := &Device{
		ip:           append(net.IP(nil), d.ip...),
		mac:          d.mac,
		displayName:  d.displayName,
		manufacturer: d.manufacturer,
		sources:      make(map[string]struct{}),
		firstSeen:    d.firstSeen,
		lastSeen:     d.lastSeen,
		extraData:    make(map[string]string),
		openPorts:    make(map[string][]int),
		lastPortScan: d.lastPortScan,
	}

	for k := range d.sources {
		newD.sources[k] = struct{}{}
	}
	for k, v := range d.extraData {
		newD.extraData[k] = v
	}
	for k, v := range d.openPorts {
		newD.openPorts[k] = append([]int(nil), v...)
	}

	return newD
}

// MarshalJSON customizes the JSON encoding of the Device struct.
// It ensures thread-safe access to the fields.
func (d *Device) MarshalJSON() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	type temp struct {
		IP           string            `json:"ip"`
		MAC          string            `json:"mac"`
		DisplayName  string            `json:"displayName"`
		Manufacturer string            `json:"manufacturer"`
		Sources      []string          `json:"sources"`
		FirstSeen    time.Time         `json:"firstSeen"`
		LastSeen     time.Time         `json:"lastSeen"`
		ExtraData    map[string]string `json:"extraData"`
	}

	ipStr := ""
	if d.ip != nil {
		ipStr = d.ip.String()
	}

	t := temp{
		IP:           ipStr,
		MAC:          d.mac,
		DisplayName:  d.displayName,
		Manufacturer: d.manufacturer,
		Sources:      make([]string, 0, len(d.sources)),
		FirstSeen:    d.firstSeen,
		LastSeen:     d.lastSeen,
		ExtraData:    make(map[string]string, len(d.extraData)),
	}

	for source := range d.sources {
		t.Sources = append(t.Sources, source)
	}
	for k, v := range d.extraData {
		t.ExtraData[k] = v
	}

	return json.Marshal(t)
}
