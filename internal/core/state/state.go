package state

import (
	"sort"
	"sync"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

// ReadOnly provides read-only access to application state.
// This interface is intended for "dumb" components that only need to read state.
type ReadOnly interface {
	DevicesSnapshot() []*discovery.Device
	Selected() (*discovery.Device, bool)
	SelectedIP() string
	CurrentTheme() string
	PreviousTheme() string
	Version() string
	FilterPattern() string
	IsDiscovering() bool
	IsPortscanning() bool
	Config() config.Config
	GetDevice(ip string) (*discovery.Device, bool)
	SearchActive() bool
	SearchText() string
	SearchError() bool
	NoColor() bool
}

// AppState holds application-level state shared across views and
// orchestrated by the App. Scanners do not write here directly.
type AppState struct {
	mu sync.RWMutex

	devices        map[string]*discovery.Device
	selectedIP     string
	previousTheme  string
	version        string
	filterPattern  string
	isDiscovering  bool
	isPortscanning bool
	cfg            *config.Config
	searchError    bool
	searchActive   bool
	noColor        bool
}

func NewAppState(cfg *config.Config, version string) *AppState {
	s := &AppState{
		devices: make(map[string]*discovery.Device),
		version: version,
		cfg:     cfg,
		noColor: theme.IsNoColor() || (cfg != nil && cfg.Theme.NoColor),
	}

	themeName := config.DefaultThemeName
	if cfg != nil && cfg.Theme.Name != "" {
		themeName = cfg.Theme.Name
	}
	s.previousTheme = themeName

	return s
}

// UpsertDevice merges a device into the canonical device map.
func (s *AppState) UpsertDevice(d *discovery.Device) {
	if d.IP() == nil {
		return
	}
	key := d.IP().String()
	if key == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.devices[key]; ok {
		existing.Merge(d)
		s.devices[key] = existing
	} else {
		// Stores a copy to prevent race conditions between discovery engine and UI rendering
		s.devices[key] = d.Copy()
	}
}

// DevicesSnapshot returns a copy of all devices for rendering.
func (s *AppState) DevicesSnapshot() []*discovery.Device {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*discovery.Device, 0, len(s.devices))
	for _, d := range s.devices {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool {
		return discovery.CompareIPs(out[i].IP(), out[j].IP())
	})
	return out
}

// SetSelectedIP stores the currently selected device IP.
func (s *AppState) SetSelectedIP(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selectedIP = ip
}

// Selected returns the currently selected device, if any.
func (s *AppState) Selected() (*discovery.Device, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.selectedIP == "" {
		return nil, false
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
func (s *AppState) SetCurrentTheme(t string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.Theme.Name = t
}

// CurrentTheme returns the current theme.
func (s *AppState) CurrentTheme() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.Theme.Name
}

// PreviousTheme returns the previous theme.
func (s *AppState) PreviousTheme() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.previousTheme
}

// SetPreviousTheme sets the previous theme.
func (s *AppState) SetPreviousTheme(t string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.previousTheme = t
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

// SetIsPortscanning sets the portscanning state.
func (s *AppState) SetIsPortscanning(portscanning bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isPortscanning = portscanning
}

// IsPortscanning returns the portscanning state.
func (s *AppState) IsPortscanning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isPortscanning
}

// Config returns the port scanner configuration.
func (s *AppState) Config() config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.cfg
}

// ReadOnly returns a read-only interface to the state.
func (s *AppState) ReadOnly() ReadOnly {
	return s
}

// GetDevice retrieves a device by IP address.
func (s *AppState) GetDevice(ip string) (*discovery.Device, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	device, ok := s.devices[ip]
	return device, ok
}

// SearchActive returns the search active state.
func (s *AppState) SearchActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.searchActive
}

// SearchText returns the current search text.
func (s *AppState) SearchText() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.filterPattern
}

// SearchError returns if there is an error in the search.
func (s *AppState) SearchError() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.searchError
}

// SetSearchError sets the search error state.
func (s *AppState) SetSearchError(err bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.searchError = err
}

// SetSearchActive sets the search active state.
func (s *AppState) SetSearchActive(active bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.searchActive = active
}

// NoColor returns the no color state.
func (s *AppState) NoColor() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.noColor
}
