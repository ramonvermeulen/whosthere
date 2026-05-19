package state

import (
	"strings"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

type deviceEntry struct {
	device              *discovery.Device
	normalizedMAC       string
	alias               string
	aliasMetadataLoaded bool
}

func newDeviceEntry(device *discovery.Device) *deviceEntry {
	deviceCopy := device.Copy()
	normalizedMAC, _ := normalizedDeviceMAC(deviceCopy)

	return &deviceEntry{
		device:        deviceCopy,
		normalizedMAC: normalizedMAC,
	}
}

func (e *deviceEntry) Merge(device *discovery.Device) {
	if e == nil || e.device == nil || device == nil {
		return
	}

	e.device.Merge(device)
	if e.normalizedMAC == "" {
		e.normalizedMAC, _ = normalizedDeviceMAC(e.device)
	}
}

func (e *deviceEntry) Device() *discovery.Device {
	if e == nil {
		return nil
	}

	return e.device
}

func (e *deviceEntry) HasMAC(normalizedMAC string) bool {
	return e != nil && e.normalizedMAC != "" && e.normalizedMAC == normalizedMAC
}

func (e *deviceEntry) Alias() string {
	if e == nil {
		return ""
	}

	return e.alias
}

func (e *deviceEntry) SetAlias(alias string) {
	if e == nil {
		return
	}

	e.aliasMetadataLoaded = true
	e.alias = strings.TrimSpace(alias)
}

func (e *deviceEntry) ClearAlias() {
	e.SetAlias("")
}

func (e *deviceEntry) ResetAlias() {
	if e == nil {
		return
	}

	e.alias = ""
	e.aliasMetadataLoaded = false
}

func (e *deviceEntry) AliasMetadataLoaded() bool {
	return e != nil && e.aliasMetadataLoaded
}

func (e *deviceEntry) PreferredName() string {
	if e == nil || e.device == nil {
		return ""
	}

	if e.alias != "" {
		return e.alias
	}
	if discoveredName := e.device.DisplayName(); discoveredName != "" {
		return discoveredName
	}
	if manufacturer := e.device.Manufacturer(); manufacturer != "" {
		return manufacturer
	}
	if ip := e.device.IP(); ip != nil {
		return ip.String()
	}

	return ""
}
