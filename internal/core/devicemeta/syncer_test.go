package devicemeta

import (
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

func TestSyncerSessionModeIsNoOp(t *testing.T) {
	t.Parallel()

	store, err := Open(filepath.Join(t.TempDir(), "devices.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	syncer := NewSyncer(config.ModeSession, store)
	device := discovery.NewDevice(net.ParseIP("192.168.1.99"))
	device.SetMAC("aa:bb:cc:dd:ee:ff")

	if err := syncer.SyncDevice(device, Scope{}); err != nil {
		t.Fatalf("SyncDevice() error = %v", err)
	}

	_, found, err := store.Get("", "aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if found {
		t.Fatal("session mode should not persist device records")
	}
}

func TestSyncerPersistentModeHydratesAndPersists(t *testing.T) {
	t.Parallel()

	store, err := Open(filepath.Join(t.TempDir(), "devices.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	firstSeen := time.Unix(100, 0).UTC()
	lastSeen := time.Unix(200, 0).UTC()
	if err := store.Upsert(Record{
		MAC:           "aa:bb:cc:dd:ee:ff",
		Alias:         "Trusted Device",
		DisplayName:   "Stored Name",
		Manufacturer:  "Stored Inc",
		InterfaceName: "en0",
		FirstSeen:     firstSeen,
		LastSeen:      lastSeen,
		ExtraData:     map[string]string{"persisted": "true"},
	}); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	syncer := NewSyncer(config.ModePersistent, store)
	device := discovery.NewDevice(net.ParseIP("192.168.1.77"))
	device.SetMAC("aa:bb:cc:dd:ee:ff")
	device.SetLastSeen(time.Unix(300, 0).UTC())

	scope := Scope{Record: ScopeRecord{InterfaceName: "en0", NetworkCIDR: "192.168.1.0/24", ScopeID: ScopeIDFromRecord(ScopeRecord{InterfaceName: "en0", NetworkCIDR: "192.168.1.0/24"})}}
	if err := syncer.SyncDevice(device, scope); err != nil {
		t.Fatalf("SyncDevice() error = %v", err)
	}

	if got := device.DisplayName(); got != "Stored Name" {
		t.Fatalf("DisplayName() = %q, want hydrated stored value", got)
	}
	if got := device.Manufacturer(); got != "Stored Inc" {
		t.Fatalf("Manufacturer() = %q, want hydrated stored value", got)
	}
	if got := device.InterfaceName(); got != "en0" {
		t.Fatalf("InterfaceName() = %q, want hydrated stored value", got)
	}
	if got := device.FirstSeen(); !got.Equal(firstSeen) {
		t.Fatalf("FirstSeen() = %v, want %v", got, firstSeen)
	}

	record, found, err := store.Get(scope.ScopeID(), "aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !found {
		t.Fatal("expected persisted record after sync")
	}
	if record.Alias != "Trusted Device" {
		t.Fatalf("Alias = %q, want preserved alias", record.Alias)
	}
	if record.LastIP != "192.168.1.77" {
		t.Fatalf("LastIP = %q, want observed IP", record.LastIP)
	}
	if !record.LastSeen.Equal(time.Unix(300, 0).UTC()) {
		t.Fatalf("LastSeen = %v, want latest observed timestamp", record.LastSeen)
	}
	if record.ExtraData["persisted"] != "true" {
		t.Fatalf("persisted extra data missing: %+v", record.ExtraData)
	}
	if record.NetworkCIDR != "192.168.1.0/24" {
		t.Fatalf("NetworkCIDR = %q, want %q", record.NetworkCIDR, "192.168.1.0/24")
	}
}

func TestSyncerSyncResultsIncludesStoredOnlyDevices(t *testing.T) {
	t.Parallel()

	store, err := Open(filepath.Join(t.TempDir(), "devices.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	if err := store.Upsert(Record{
		MAC:         "aa:bb:cc:dd:ee:ff",
		LastIP:      "192.168.1.10",
		DisplayName: "Stored Device",
		FirstSeen:   time.Unix(100, 0).UTC(),
		LastSeen:    time.Unix(200, 0).UTC(),
	}); err != nil {
		t.Fatalf("Upsert(stored only) error = %v", err)
	}

	results := &discovery.ScanResults{
		Devices: []*discovery.Device{
			discovery.NewDevice(net.ParseIP("192.168.1.20")),
		},
		Stats: &discovery.ScanStats{Count: 1},
	}
	results.Devices[0].SetMAC("11:22:33:44:55:66")

	syncer := NewSyncer(config.ModePersistent, store)
	if err := syncer.SyncResults(results, Scope{AllInterfaces: true}); err != nil {
		t.Fatalf("SyncResults() error = %v", err)
	}

	if len(results.Devices) != 2 {
		t.Fatalf("len(results.Devices) = %d, want 2", len(results.Devices))
	}
	if results.Stats.Count != 2 {
		t.Fatalf("results.Stats.Count = %d, want 2", results.Stats.Count)
	}
}

func TestStoredDevicesForScopeFiltersByInterface(t *testing.T) {
	t.Parallel()

	store, err := Open(filepath.Join(t.TempDir(), "devices.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	en0Scope := ScopeFromInterfaceInfo(&discovery.InterfaceInfo{
		Interface: mustInterface("en0"),
		IPv4Net:   mustIPv4Net("192.168.1.0/24"),
	}, false)
	en1Scope := ScopeFromInterfaceInfo(&discovery.InterfaceInfo{
		Interface: mustInterface("en1"),
		IPv4Net:   mustIPv4Net("10.0.0.0/24"),
	}, false)

	if err := store.Upsert(Record{
		ScopeID:       en0Scope.ScopeID(),
		MAC:           "aa:bb:cc:dd:ee:01",
		LastIP:        "192.168.1.10",
		DisplayName:   "en0-device",
		InterfaceName: "en0",
		NetworkCIDR:   "192.168.1.0/24",
	}); err != nil {
		t.Fatalf("Upsert(en0) error = %v", err)
	}
	if err := store.Upsert(Record{
		ScopeID:       en1Scope.ScopeID(),
		MAC:           "aa:bb:cc:dd:ee:02",
		LastIP:        "10.0.0.10",
		DisplayName:   "en1-device",
		InterfaceName: "en1",
		NetworkCIDR:   "10.0.0.0/24",
	}); err != nil {
		t.Fatalf("Upsert(en1) error = %v", err)
	}
	if err := store.Upsert(Record{
		MAC:         "aa:bb:cc:dd:ee:03",
		LastIP:      "172.16.0.10",
		DisplayName: "legacy",
	}); err != nil {
		t.Fatalf("Upsert(legacy) error = %v", err)
	}

	syncer := NewSyncer(config.ModePersistent, store)
	devices, err := syncer.StoredDevicesForScope(en0Scope)
	if err != nil {
		t.Fatalf("StoredDevicesForScope() error = %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("len(devices) = %d, want 1", len(devices))
	}
	if got := devices[0].DisplayName(); got != "en0-device" {
		t.Fatalf("DisplayName() = %q, want %q", got, "en0-device")
	}

	devices, err = syncer.StoredDevicesForScope(Scope{AllInterfaces: true})
	if err != nil {
		t.Fatalf("StoredDevicesForScope(all) error = %v", err)
	}
	if len(devices) != 3 {
		t.Fatalf("len(devices) with all interfaces = %d, want 3", len(devices))
	}
}

func mustInterface(name string) *net.Interface {
	return &net.Interface{Name: name}
}

func mustIPv4Net(cidr string) *net.IPNet {
	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}
	return subnet
}
