package events

// Event represents a UI event emitted by components or views.
type Event interface{}

// DeviceSelected is emitted when a device is selected in the table.
type DeviceSelected struct {
	IP string
}

// FilterChanged is emitted when the filter pattern changes.
type FilterChanged struct {
	Pattern string
}

// NavigateTo is emitted to navigate to a route.
type NavigateTo struct {
	Route   string
	Overlay bool
}

// ThemeSelected is emitted when a theme is selected.
type ThemeSelected struct {
	Name string
}

// ThemeConfirmed is emitted when a theme is confirmed.
type ThemeConfirmed struct{}

// ThemeSaved is emitted when a theme is saved to config.
type ThemeSaved struct {
	Name string
}

// HideView is emitted to hide the current modal.
type HideView struct{}

// DiscoveryStarted is emitted when discovery starts.
type DiscoveryStarted struct{}

// DiscoveryStopped is emitted when discovery stops.
type DiscoveryStopped struct{}

// PortScanStarted is emitted when port scan starts.
type PortScanStarted struct{}

// PortScanStopped is emitted when port scan stops.
type PortScanStopped struct{}

// SearchStarted is emitted when search mode starts.
type SearchStarted struct{}

// SearchError is emitted when search input is invalid.
type SearchError struct {
	Error bool
}

// SearchFinished is emitted when search mode ends.
type SearchFinished struct{}

// CopyIP is emitted to copy the IP to clipboard.
type CopyIP struct {
	IP string
}

// CopyMac is emitted to copy the MAC to clipboard.
type CopyMac struct {
	MAC string
}
