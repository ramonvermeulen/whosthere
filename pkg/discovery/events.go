package discovery

import "time"

// Event represents something that happened during device discovery.
// Events are emitted through the Events channel. Each Event has a Type
// indicating what happened. Based on the Type, exactly one of Device,
// Error, or Stats will be non-nil:
//
//   - EventDeviceDiscovered: Device is non-nil
//   - EventScanCompleted: Stats is non-nil
//   - EventError: Error is non-nil
//   - EventScanStarted, EventEngineStarted, EventEngineStopped:
//     all fields are nil
//
// Example usage:
//
//	for event := range engine.Events {
//	    switch event.Type {
//	    case discovery.EventDeviceDiscovered:
//	        fmt.Println(event.Device.IP)
//	    case discovery.EventScanCompleted:
//	        fmt.Printf("Found %d devices\n", event.Stats.DeviceCount)
//	    }
//	}
type Event struct {
	Type   EventType
	Device *Device    // non-nil when Type == EventDeviceDiscovered
	Error  error      // non-nil when Type == EventError
	Stats  *ScanStats // non-nil when Type == EventScanCompleted
}

// EventType indicates what kind of event this is.
type EventType int

const (
	EventDeviceDiscovered EventType = iota
	EventScanStarted
	EventScanCompleted
	EventError
	EventEngineStarted
	EventEngineStopped
)

// ScanStats contains statistics about a completed scan.
type ScanStats struct {
	DeviceCount int
	Duration    time.Duration
}

// NewDeviceEvent creates a device discovery event.
func NewDeviceEvent(device *Device) Event {
	return Event{
		Type:   EventDeviceDiscovered,
		Device: device,
	}
}

// NewScanCompletedEvent creates a scan completion event.
func NewScanCompletedEvent(stats *ScanStats) Event {
	return Event{
		Type:  EventScanCompleted,
		Stats: stats,
	}
}

// NewErrorEvent creates an error event.
func NewErrorEvent(err error) Event {
	return Event{
		Type:  EventError,
		Error: err,
	}
}

// NewScanStartedEvent creates a scan start event.
func NewScanStartedEvent() Event {
	return Event{
		Type: EventScanStarted,
	}
}

// NewEngineStartedEvent creates an engine start event.
func NewEngineStartedEvent() Event {
	return Event{
		Type: EventEngineStarted,
	}
}

// NewEngineStoppedEvent creates an engine stop event.
func NewEngineStoppedEvent() Event {
	return Event{
		Type: EventEngineStopped,
	}
}
