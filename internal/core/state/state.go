package state

import (
	"sync"

	"github.com/ramonvermeulen/whosthere/internal/core/discovery"
)

// ReadOnly provides read-only access to application state.
// This interface is intended for "dumb" components that only need to read state.
type ReadOnly interface {
	DevicesSnapshot() []discovery.Device
	Selected() (discovery.Device, bool)
	SelectedIP() string
	CurrentTheme() string
	PreviousTheme() string
	Version() string
	FilterPattern() string
	IsDiscovering() bool
}

// AppState holds application-level state shared across views and
// orchestrated by the App. Scanners do not write here directly.
type AppState struct {
	mu sync.RWMutex

	devices       map[string]discovery.Device
	selectedIP    string
	currentTheme  string
	previousTheme string
	version       string
	filterPattern string
	isDiscovering bool
}

func NewAppState() *AppState {
	return &AppState{
		devices: make(map[string]discovery.Device),
	}
}

// UpsertDevice merges a device into the canonical device map.
func (s *AppState) UpsertDevice(d *discovery.Device) {
	if d.IP == nil {
		return
	}
	key := d.IP.String()
	if key == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.devices[key]; ok {
		existing.Merge(d)
		s.devices[key] = existing
	} else {
		s.devices[key] = *d
	}
}

// DevicesSnapshot returns a copy of all devices for rendering.
func (s *AppState) DevicesSnapshot() []discovery.Device {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]discovery.Device, 0, len(s.devices))
	for _, d := range s.devices {
		out = append(out, d)
	}
	return out
}

// SetSelectedIP stores the currently selected device IP.
func (s *AppState) SetSelectedIP(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selectedIP = ip
}

// Selected returns the currently selected device, if any.
func (s *AppState) Selected() (discovery.Device, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.selectedIP == "" {
		return discovery.Device{}, false
	}
	d, ok := s.devices[s.selectedIP]
	return d, ok
}

// SelectedIP returns the currently selected device IP, if any.
func (s *AppState) SelectedIP() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selectedIP
}

// SetCurrentTheme sets the current theme.
func (s *AppState) SetCurrentTheme(theme string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentTheme = theme
}

// CurrentTheme returns the current theme.
func (s *AppState) CurrentTheme() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentTheme
}

// PreviousTheme returns the previous theme.
func (s *AppState) PreviousTheme() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.previousTheme
}

// SetPreviousTheme sets the previous theme.
func (s *AppState) SetPreviousTheme(theme string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.previousTheme = theme
}

// SetVersion sets the current version.
func (s *AppState) SetVersion(version string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.version = version
}

// Version returns the current version.
func (s *AppState) Version() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.version
}

// SetFilterPattern sets the filter pattern.
func (s *AppState) SetFilterPattern(pattern string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filterPattern = pattern
}

// FilterPattern returns the filter pattern.
func (s *AppState) FilterPattern() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.filterPattern
}

// SetIsDiscovering sets the discovering state.
func (s *AppState) SetIsDiscovering(discovering bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isDiscovering = discovering
}

// IsDiscovering returns the discovering state.
func (s *AppState) IsDiscovering() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isDiscovering
}

// ReadOnly returns a read-only interface to the state.
func (s *AppState) ReadOnly() ReadOnly {
	return s
}
