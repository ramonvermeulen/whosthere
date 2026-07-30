package devicemeta

import (
	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

type Syncer struct {
	storageMode config.Mode
	store       Store
}

func NewSyncer(storageMode config.Mode, store Store) *Syncer {
	return &Syncer{
		storageMode: storageMode,
		store:       store,
	}
}

func (s *Syncer) Enabled() bool {
	return s != nil && s.storageMode == config.ModePersistent && s.store != nil
}

func (s *Syncer) SyncDevice(device *discovery.Device, scope Scope) error {
	if !s.Enabled() || device == nil {
		return nil
	}

	effectiveScope := scope.ScopeForDevice(device)
	normalizedMAC, ok := NormalizedDeviceMAC(device)
	if !ok {
		return nil
	}

	record, found, err := s.store.Get(effectiveScope.ScopeID(), normalizedMAC)
	if err != nil {
		return err
	}
	if found {
		ApplyRecordToDevice(record, device)
	}

	merged, ok := MergeObservedDeviceIntoRecord(record, device, effectiveScope)
	if !ok {
		return nil
	}

	if err := s.store.UpsertScope(effectiveScope.Record); err != nil {
		return err
	}

	return s.store.Upsert(merged)
}

func (s *Syncer) SyncResults(results *discovery.ScanResults, scope Scope) error {
	if !s.Enabled() || results == nil {
		return nil
	}

	seen := make(map[string]struct{}, len(results.Devices))
	for _, device := range results.Devices {
		if normalizedMAC, ok := NormalizedDeviceMAC(device); ok {
			seen[normalizedMAC] = struct{}{}
		}
		if err := s.SyncDevice(device, scope); err != nil {
			return err
		}
	}

	storedDevices, err := s.StoredDevicesForScope(scope)
	if err != nil {
		return err
	}
	for _, device := range storedDevices {
		normalizedMAC, ok := NormalizedDeviceMAC(device)
		if !ok {
			continue
		}
		if _, exists := seen[normalizedMAC]; exists {
			continue
		}
		results.Devices = append(results.Devices, device)
		seen[normalizedMAC] = struct{}{}
	}
	if results.Stats != nil {
		results.Stats.Count = len(results.Devices)
	}

	return nil
}

func (s *Syncer) StoredDevicesForScope(scope Scope) ([]*discovery.Device, error) {
	if !s.Enabled() {
		return nil, nil
	}

	devices := make([]*discovery.Device, 0)
	err := s.store.ForEach(func(record Record) error {
		if !scope.Matches(record) {
			return nil
		}
		device, ok := DeviceFromRecord(record)
		if !ok {
			return nil
		}
		devices = append(devices, device)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return devices, nil
}
