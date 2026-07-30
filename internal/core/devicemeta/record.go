package devicemeta

import (
	"strings"
	"time"
)

// Record stores local metadata for a device keyed by MAC address.
type Record struct {
	ScopeID       string            `json:"scope_id,omitempty"`
	MAC           string            `json:"mac,omitempty"`
	Alias         string            `json:"alias,omitempty"`
	LastIP        string            `json:"last_ip,omitempty"`
	DisplayName   string            `json:"display_name,omitempty"`
	Manufacturer  string            `json:"manufacturer,omitempty"`
	InterfaceName string            `json:"interface_name,omitempty"`
	NetworkCIDR   string            `json:"network_cidr,omitempty"`
	FirstSeen     time.Time         `json:"first_seen,omitempty"`
	LastSeen      time.Time         `json:"last_seen,omitempty"`
	ExtraData     map[string]string `json:"extra_data,omitempty"`
	UpdatedAt     time.Time         `json:"updated_at,omitempty"`
}

func (r *Record) normalize() {
	if r == nil {
		return
	}

	r.Alias = strings.TrimSpace(r.Alias)
	r.ScopeID = strings.TrimSpace(r.ScopeID)
	r.LastIP = strings.TrimSpace(r.LastIP)
	r.DisplayName = strings.TrimSpace(r.DisplayName)
	r.Manufacturer = strings.TrimSpace(r.Manufacturer)
	r.InterfaceName = strings.TrimSpace(r.InterfaceName)
	r.NetworkCIDR = strings.TrimSpace(r.NetworkCIDR)
	if len(r.ExtraData) == 0 {
		r.ExtraData = nil
	}
}

func (r Record) empty() bool {
	return r.ScopeID == "" &&
		r.Alias == "" &&
		r.LastIP == "" &&
		r.DisplayName == "" &&
		r.Manufacturer == "" &&
		r.InterfaceName == "" &&
		r.NetworkCIDR == "" &&
		r.FirstSeen.IsZero() &&
		r.LastSeen.IsZero() &&
		len(r.ExtraData) == 0
}
