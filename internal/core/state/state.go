package state

import (
	"sort"
	"strings"
	"sync"
	"time"

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
	AliasFor(d *discovery.Device) string
	DetectedNameFor(d *discovery.Device) string
	AliasOrDetectedNameFor(d *discovery.Device) string
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
	CurrentInterface() string
	StatusMessage() string
	StatusSeverity() StatusSeverity
	AliasEditorDraft() string
}

type StatusSeverity string

const (
	StatusSeverityInfo    StatusSeverity = "info"
	StatusSeveritySuccess StatusSeverity = "success"
	StatusSeverityError   StatusSeverity = "error"
)

// AppState holds application-level state shared across views and
// orchestrated by the App. Scanners do not write here directly.
type AppState struct {
	mu sync.RWMutex

	devices          map[string]*deviceEntry
	selectedIP       string
	previousTheme    string
	version          string
	filterPattern    string
	isDiscovering    bool
	isPortscanning   bool
	cfg              *config.Config
	searchError      bool
	searchActive     bool
	noColor          bool
	currentInterface string
	aliasEditorDraft string
	statusMessage    string
	statusSeverity   StatusSeverity
	statusUntil      time.Time
}

func NewAppState(cfg *config.Config, version string) *AppState {
	s := &AppState{
		devices: make(map[string]*deviceEntry),
		version: version,
		cfg:     cfg,
		noColor: theme.IsNoColor() || (cfg != nil && cfg.Theme.NoColor),
	}
	theme.UpdateNoColor(s.noColor)

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

	if existingEntry, exists := s.devices[key]; exists {
		existingEntry.Merge(d)
		return
	}

	s.devices[key] = newDeviceEntry(d)
}

// DevicesSnapshot returns a copy of all devices for rendering.
func (s *AppState) DevicesSnapshot() []*discovery.Device {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*discovery.Device, 0, len(s.devices))
	for _, entry := range s.devices {
		if entry == nil || entry.Device() == nil {
			continue
		}
		out = append(out, entry.Device())
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
	entry, ok := s.devices[s.selectedIP]
	if !ok || entry == nil || entry.Device() == nil {
		return nil, false
	}

	return entry.Device(), true
}

// SelectedIP returns the currently selected device IP, if any.
func (s *AppState) SelectedIP() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selectedIP
}

// AliasFor returns the cached alias for a device, if present.
func (s *AppState) AliasFor(d *discovery.Device) string {
	if d == nil {
		return ""
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.findDeviceEntryLocked(d)
	if !ok {
		return ""
	}

	return entry.Alias()
}

// AliasOrDetectedNameFor returns the detected name with alias override.
func (s *AppState) AliasOrDetectedNameFor(d *discovery.Device) string {
	if d == nil {
		return ""
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if entry, ok := s.findDeviceEntryLocked(d); ok {
		if alias := entry.Alias(); alias != "" {
			return alias
		}
		return entry.DetectedName()
	}

	return (&deviceEntry{device: d}).DetectedName()
}

// DetectedNameFor returns the device name without alias precedence.
func (s *AppState) DetectedNameFor(d *discovery.Device) string {
	if d == nil {
		return ""
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if entry, ok := s.findDeviceEntryLocked(d); ok {
		return entry.DetectedName()
	}

	return (&deviceEntry{device: d}).DetectedName()
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

	entry, ok := s.devices[ip]
	if !ok || entry == nil || entry.Device() == nil {
		return nil, false
	}

	return entry.Device(), true
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

// CurrentInterface returns the active network interface name.
func (s *AppState) CurrentInterface() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentInterface
}

// SetCurrentInterface sets the active network interface name.
func (s *AppState) SetCurrentInterface(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentInterface = name
}

// SetAliasEditorDraft updates the current alias editor draft.
func (s *AppState) SetAliasEditorDraft(alias string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.aliasEditorDraft = alias
}

// ClearAliasEditorDraft removes the current alias editor draft.
func (s *AppState) ClearAliasEditorDraft() {
	s.SetAliasEditorDraft("")
}

// AliasEditorDraft returns the current alias editor draft.
func (s *AppState) AliasEditorDraft() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.aliasEditorDraft
}

// SetStatusMessage sets a transient status message until the given duration expires.
func (s *AppState) SetStatusMessage(message string, severity StatusSeverity, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.statusMessage = strings.TrimSpace(message)
	s.statusSeverity = severity
	if s.statusMessage == "" || duration <= 0 {
		s.statusSeverity = StatusSeverityInfo
		s.statusUntil = time.Time{}
		return
	}

	s.statusUntil = time.Now().Add(duration)
}

// ClearStatusMessage removes any active transient status message.
func (s *AppState) ClearStatusMessage() {
	s.SetStatusMessage("", StatusSeverityInfo, 0)
}

// StatusMessage returns the active transient status message, if any.
func (s *AppState) StatusMessage() string {
	s.mu.RLock()
	message := s.statusMessage
	until := s.statusUntil
	s.mu.RUnlock()

	if message == "" {
		return ""
	}
	if !until.IsZero() && time.Now().After(until) {
		s.ClearStatusMessage()
		return ""
	}

	return message
}

// StatusSeverity returns the severity for the current status message.
func (s *AppState) StatusSeverity() StatusSeverity {
	s.mu.RLock()
	severity := s.statusSeverity
	message := s.statusMessage
	until := s.statusUntil
	s.mu.RUnlock()

	if message == "" {
		return StatusSeverityInfo
	}
	if !until.IsZero() && time.Now().After(until) {
		s.ClearStatusMessage()
		return StatusSeverityInfo
	}

	return severity
}

// HasAliasMetadataForMAC reports whether alias metadata for a normalized MAC is cached.
func (s *AppState) HasAliasMetadataForMAC(normalizedMAC string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, entry := range s.devices {
		if entry != nil && entry.HasMAC(normalizedMAC) {
			return entry.AliasMetadataLoaded()
		}
	}

	return false
}

// SetAliasForMAC caches an alias for all devices with the given normalized MAC address.
func (s *AppState) SetAliasForMAC(normalizedMAC, alias string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, entry := range s.devices {
		if entry != nil && entry.HasMAC(normalizedMAC) {
			entry.SetAlias(alias)
		}
	}
}

// ClearAliasForMAC removes the cached alias for all devices with the given normalized MAC address.
func (s *AppState) ClearAliasForMAC(normalizedMAC string) {
	s.SetAliasForMAC(normalizedMAC, "")
}

// ResetAliases clears all cached aliases and loaded markers.
func (s *AppState) ResetAliases() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, entry := range s.devices {
		if entry != nil {
			entry.ResetAlias()
		}
	}
}

// ClearDevices removes all discovered devices and resets selection.
func (s *AppState) ClearDevices() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices = make(map[string]*deviceEntry)
	s.selectedIP = ""
}

func (s *AppState) findDeviceEntryLocked(device *discovery.Device) (*deviceEntry, bool) {
	if device == nil {
		return nil, false
	}

	if ip := device.IP(); ip != nil {
		if entry, ok := s.devices[ip.String()]; ok {
			return entry, true
		}
	}

	normalizedMAC, hasValidMAC := normalizedDeviceMAC(device)
	if !hasValidMAC {
		return nil, false
	}

	for _, entry := range s.devices {
		if entry != nil && entry.HasMAC(normalizedMAC) {
			return entry, true
		}
	}

	return nil, false
}
