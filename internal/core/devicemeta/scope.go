package devicemeta

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

type ScopeRecord struct {
	ScopeID       string    `json:"scope_id,omitempty"`
	InterfaceName string    `json:"interface_name,omitempty"`
	NetworkCIDR   string    `json:"network_cidr,omitempty"`
	GatewayIP     string    `json:"gateway_ip,omitempty"`
	GatewayMAC    string    `json:"gateway_mac,omitempty"`
	FirstSeen     time.Time `json:"first_seen,omitempty"`
	LastSeen      time.Time `json:"last_seen,omitempty"`
}

func (r *ScopeRecord) normalize() {
	if r == nil {
		return
	}

	r.ScopeID = strings.TrimSpace(r.ScopeID)
	r.InterfaceName = strings.TrimSpace(r.InterfaceName)
	r.NetworkCIDR = strings.TrimSpace(r.NetworkCIDR)
	r.GatewayIP = strings.TrimSpace(r.GatewayIP)
	r.GatewayMAC = strings.TrimSpace(r.GatewayMAC)
}

type Scope struct {
	AllInterfaces bool
	Record        ScopeRecord
}

func ScopeFromInterfaceInfo(interfaceInfo *discovery.InterfaceInfo, allInterfaces bool) Scope {
	scope := Scope{AllInterfaces: allInterfaces}
	if interfaceInfo == nil {
		return scope
	}

	if interfaceInfo.Interface != nil {
		scope.Record.InterfaceName = strings.TrimSpace(interfaceInfo.Interface.Name)
	}
	if interfaceInfo.IPv4Net != nil {
		scope.Record.NetworkCIDR = strings.TrimSpace(interfaceInfo.IPv4Net.String())
	}
	scope.Record.ScopeID = ScopeIDFromRecord(scope.Record)
	scope.Record.normalize()

	return scope
}

func ScopeIDFromRecord(record ScopeRecord) string {
	record.normalize()

	var parts []string
	if record.GatewayMAC != "" && record.NetworkCIDR != "" {
		parts = append(parts, "gw_mac="+record.GatewayMAC, "cidr="+record.NetworkCIDR)
	} else if record.GatewayIP != "" && record.NetworkCIDR != "" {
		parts = append(parts, "gw_ip="+record.GatewayIP, "cidr="+record.NetworkCIDR)
	} else if record.InterfaceName != "" && record.NetworkCIDR != "" {
		parts = append(parts, "iface="+record.InterfaceName, "cidr="+record.NetworkCIDR)
	} else if record.InterfaceName != "" {
		parts = append(parts, "iface="+record.InterfaceName)
	} else if record.NetworkCIDR != "" {
		parts = append(parts, "cidr="+record.NetworkCIDR)
	}

	if len(parts) == 0 {
		return ""
	}

	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:8])
}

func (s Scope) ScopeID() string {
	return s.Record.ScopeID
}

func (s Scope) InterfaceName() string {
	return s.Record.InterfaceName
}

func (s Scope) NetworkCIDR() string {
	return s.Record.NetworkCIDR
}

func (s Scope) Matches(record Record) bool {
	if s.AllInterfaces {
		return true
	}

	if s.ScopeID() != "" && record.ScopeID != "" {
		return s.ScopeID() == record.ScopeID
	}

	hasLegacyInterfaceMetadata := record.InterfaceName != "" || record.NetworkCIDR != ""
	if !hasLegacyInterfaceMetadata {
		return false
	}

	if s.InterfaceName() != "" && record.InterfaceName != "" && s.InterfaceName() != record.InterfaceName {
		return false
	}
	if s.NetworkCIDR() != "" && record.NetworkCIDR != "" && s.NetworkCIDR() != record.NetworkCIDR {
		return false
	}

	if s.InterfaceName() != "" && record.InterfaceName != "" {
		return true
	}
	if s.NetworkCIDR() != "" && record.NetworkCIDR != "" {
		return true
	}

	return false
}

func (s Scope) ScopeForDevice(device *discovery.Device) Scope {
	if device == nil {
		return s
	}

	deviceInterfaceName := strings.TrimSpace(device.InterfaceName())
	if deviceInterfaceName == "" || deviceInterfaceName == s.InterfaceName() {
		return s
	}

	fallback := Scope{
		AllInterfaces: s.AllInterfaces,
		Record: ScopeRecord{
			InterfaceName: deviceInterfaceName,
		},
	}
	fallback.Record.ScopeID = ScopeIDFromRecord(fallback.Record)
	return fallback
}
