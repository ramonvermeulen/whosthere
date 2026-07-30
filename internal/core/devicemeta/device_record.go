package devicemeta

import (
	"net"
	"strings"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

func RecordFromDevice(device *discovery.Device, scope Scope) (Record, bool) {
	normalizedMAC, ok := NormalizedDeviceMAC(device)
	if !ok {
		return Record{}, false
	}

	record := Record{
		ScopeID:       scope.ScopeID(),
		MAC:           normalizedMAC,
		LastIP:        strings.TrimSpace(device.IP().String()),
		DisplayName:   strings.TrimSpace(device.DisplayName()),
		Manufacturer:  strings.TrimSpace(device.Manufacturer()),
		InterfaceName: strings.TrimSpace(device.InterfaceName()),
		FirstSeen:     device.FirstSeen(),
		LastSeen:      device.LastSeen(),
		ExtraData:     cloneExtraData(device.ExtraData()),
	}
	if scope.InterfaceName() != "" && record.InterfaceName == scope.InterfaceName() {
		record.NetworkCIDR = scope.NetworkCIDR()
	}
	record.normalize()

	return record, true
}

func DeviceFromRecord(record Record) (*discovery.Device, bool) {
	ip := net.ParseIP(strings.TrimSpace(record.LastIP))
	if ip == nil {
		return nil, false
	}

	device := discovery.NewDevice(ip)
	device.SetMAC(record.MAC)
	if record.DisplayName != "" {
		device.SetDisplayName(record.DisplayName)
	}
	if record.Manufacturer != "" {
		device.SetManufacturer(record.Manufacturer)
	}
	if record.InterfaceName != "" {
		device.SetInterfaceName(record.InterfaceName)
	}
	if !record.FirstSeen.IsZero() {
		device.SetFirstSeen(record.FirstSeen)
	}
	if !record.LastSeen.IsZero() {
		device.SetLastSeen(record.LastSeen)
	}
	if len(record.ExtraData) > 0 {
		device.SetExtraData(cloneExtraData(record.ExtraData))
	}

	return device, true
}

func ApplyRecordToDevice(record Record, device *discovery.Device) {
	if device == nil {
		return
	}

	if device.DisplayName() == "" && record.DisplayName != "" {
		device.SetDisplayName(record.DisplayName)
	}
	if device.Manufacturer() == "" && record.Manufacturer != "" {
		device.SetManufacturer(record.Manufacturer)
	}
	if device.InterfaceName() == "" && record.InterfaceName != "" {
		device.SetInterfaceName(record.InterfaceName)
	}

	if firstSeen := device.FirstSeen(); firstSeen.IsZero() || (!record.FirstSeen.IsZero() && record.FirstSeen.Before(firstSeen)) {
		device.SetFirstSeen(record.FirstSeen)
	}
	if lastSeen := device.LastSeen(); lastSeen.IsZero() || (!record.LastSeen.IsZero() && record.LastSeen.After(lastSeen)) {
		device.SetLastSeen(record.LastSeen)
	}

	if len(record.ExtraData) > 0 {
		extraData := device.ExtraData()
		if extraData == nil {
			extraData = map[string]string{}
		}
		for key, value := range record.ExtraData {
			if _, exists := extraData[key]; !exists {
				extraData[key] = value
			}
		}
		device.SetExtraData(extraData)
	}
}

func MergeObservedDeviceIntoRecord(record Record, device *discovery.Device, scope Scope) (Record, bool) {
	observed, ok := RecordFromDevice(device, scope)
	if !ok {
		return Record{}, false
	}

	merged := record
	merged.ScopeID = observed.ScopeID
	merged.MAC = observed.MAC

	if observed.LastIP != "" {
		merged.LastIP = observed.LastIP
	}
	if observed.DisplayName != "" {
		merged.DisplayName = observed.DisplayName
	}
	if observed.Manufacturer != "" {
		merged.Manufacturer = observed.Manufacturer
	}
	if observed.InterfaceName != "" {
		merged.InterfaceName = observed.InterfaceName
	}
	if observed.NetworkCIDR != "" {
		merged.NetworkCIDR = observed.NetworkCIDR
	}

	if merged.FirstSeen.IsZero() || (!observed.FirstSeen.IsZero() && observed.FirstSeen.Before(merged.FirstSeen)) {
		merged.FirstSeen = observed.FirstSeen
	}
	if observed.LastSeen.After(merged.LastSeen) {
		merged.LastSeen = observed.LastSeen
	}

	if len(merged.ExtraData) == 0 {
		merged.ExtraData = map[string]string{}
	} else {
		merged.ExtraData = cloneExtraData(merged.ExtraData)
	}
	for key, value := range observed.ExtraData {
		merged.ExtraData[key] = value
	}
	merged.normalize()

	return merged, true
}

func NormalizedDeviceMAC(device *discovery.Device) (string, bool) {
	if device == nil {
		return "", false
	}

	return NormalizeMAC(device.MAC())
}

func cloneExtraData(data map[string]string) map[string]string {
	if len(data) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(data))
	for key, value := range data {
		cloned[key] = value
	}

	return cloned
}
