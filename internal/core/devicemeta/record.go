package devicemeta

import "time"

// Record stores local metadata for a device keyed by MAC address.
type Record struct {
	Alias     string    `json:"alias,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func (r Record) empty() bool {
	return r.Alias == ""
}
