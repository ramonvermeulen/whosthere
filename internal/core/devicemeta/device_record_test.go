package devicemeta

import (
	"net"
	"testing"
	"time"

	"github.com/ramonvermeulen/whosthere/pkg/discovery"
)

func TestApplyRecordToDeviceHydratesMissingFieldsAndPreservesObservedValues(t *testing.T) {
	t.Parallel()

	device := discovery.NewDevice(net.ParseIP("192.168.1.50"))
	device.SetMAC("aa:bb:cc:dd:ee:ff")
	device.SetDisplayName("Observed Name")
	device.SetLastSeen(time.Unix(250, 0).UTC())
	device.SetExtraData(map[string]string{"observed": "yes"})

	record := Record{
		MAC:           "aa:bb:cc:dd:ee:ff",
		DisplayName:   "Stored Name",
		Manufacturer:  "Acme",
		InterfaceName: "en0",
		FirstSeen:     time.Unix(100, 0).UTC(),
		LastSeen:      time.Unix(200, 0).UTC(),
		ExtraData: map[string]string{
			"persisted": "true",
			"observed":  "no",
		},
	}

	ApplyRecordToDevice(record, device)

	if got := device.DisplayName(); got != "Observed Name" {
		t.Fatalf("DisplayName() = %q, want observed value", got)
	}
	if got := device.Manufacturer(); got != "Acme" {
		t.Fatalf("Manufacturer() = %q, want %q", got, "Acme")
	}
	if got := device.InterfaceName(); got != "en0" {
		t.Fatalf("InterfaceName() = %q, want %q", got, "en0")
	}
	if got := device.FirstSeen(); !got.Equal(record.FirstSeen) {
		t.Fatalf("FirstSeen() = %v, want %v", got, record.FirstSeen)
	}
	if got := device.LastSeen(); !got.Equal(time.Unix(250, 0).UTC()) {
		t.Fatalf("LastSeen() = %v, want latest observed timestamp", got)
	}

	extraData := device.ExtraData()
	if extraData["persisted"] != "true" {
		t.Fatalf("persisted extra data missing: %+v", extraData)
	}
	if extraData["observed"] != "yes" {
		t.Fatalf("observed extra data should win, got %+v", extraData)
	}
}

func TestMergeObservedDeviceIntoRecordPreservesAliasAndHistory(t *testing.T) {
	t.Parallel()

	device := discovery.NewDevice(net.ParseIP("192.168.1.60"))
	device.SetMAC("aa:bb:cc:dd:ee:ff")
	device.SetDisplayName("Observed Name")
	device.SetManufacturer("Observed Inc")
	device.SetInterfaceName("en1")
	device.SetFirstSeen(time.Unix(150, 0).UTC())
	device.SetLastSeen(time.Unix(300, 0).UTC())
	device.SetExtraData(map[string]string{"source": "mdns"})

	record := Record{
		MAC:           "aa:bb:cc:dd:ee:ff",
		Alias:         "Trusted Device",
		LastIP:        "192.168.1.10",
		DisplayName:   "Stored Name",
		Manufacturer:  "Stored Inc",
		InterfaceName: "en0",
		NetworkCIDR:   "192.168.1.0/24",
		FirstSeen:     time.Unix(100, 0).UTC(),
		LastSeen:      time.Unix(200, 0).UTC(),
		ExtraData:     map[string]string{"persisted": "true"},
	}

	scope := Scope{Record: ScopeRecord{InterfaceName: "en1", NetworkCIDR: "192.168.1.0/24", ScopeID: ScopeIDFromRecord(ScopeRecord{InterfaceName: "en1", NetworkCIDR: "192.168.1.0/24"})}}
	merged, ok := MergeObservedDeviceIntoRecord(record, device, scope)
	if !ok {
		t.Fatal("expected merge to succeed")
	}

	if merged.Alias != "Trusted Device" {
		t.Fatalf("Alias = %q, want preserved alias", merged.Alias)
	}
	if merged.LastIP != "192.168.1.60" {
		t.Fatalf("LastIP = %q, want observed IP", merged.LastIP)
	}
	if merged.DisplayName != "Observed Name" || merged.Manufacturer != "Observed Inc" || merged.InterfaceName != "en1" {
		t.Fatalf("observed fields should win, got %+v", merged)
	}
	if merged.NetworkCIDR != "192.168.1.0/24" {
		t.Fatalf("NetworkCIDR = %q, want %q", merged.NetworkCIDR, "192.168.1.0/24")
	}
	if !merged.FirstSeen.Equal(time.Unix(100, 0).UTC()) {
		t.Fatalf("FirstSeen = %v, want earliest stored timestamp", merged.FirstSeen)
	}
	if !merged.LastSeen.Equal(time.Unix(300, 0).UTC()) {
		t.Fatalf("LastSeen = %v, want latest observed timestamp", merged.LastSeen)
	}
	if merged.ExtraData["persisted"] != "true" || merged.ExtraData["source"] != "mdns" {
		t.Fatalf("extra data not merged correctly: %+v", merged.ExtraData)
	}
}

func TestRecordFromDeviceUsesScopeNetworkCIDRForMatchingInterface(t *testing.T) {
	t.Parallel()

	device := discovery.NewDevice(net.ParseIP("192.168.1.60"))
	device.SetMAC("aa:bb:cc:dd:ee:ff")
	device.SetInterfaceName("en1")

	record, ok := RecordFromDevice(device, Scope{Record: ScopeRecord{InterfaceName: "en1", NetworkCIDR: "192.168.1.0/24", ScopeID: ScopeIDFromRecord(ScopeRecord{InterfaceName: "en1", NetworkCIDR: "192.168.1.0/24"})}})
	if !ok {
		t.Fatal("expected record to be created")
	}
	if record.NetworkCIDR != "192.168.1.0/24" {
		t.Fatalf("NetworkCIDR = %q, want %q", record.NetworkCIDR, "192.168.1.0/24")
	}

	record, ok = RecordFromDevice(device, Scope{Record: ScopeRecord{InterfaceName: "en0", NetworkCIDR: "10.0.0.0/24", ScopeID: ScopeIDFromRecord(ScopeRecord{InterfaceName: "en0", NetworkCIDR: "10.0.0.0/24"})}})
	if !ok {
		t.Fatal("expected record to be created")
	}
	if record.NetworkCIDR != "" {
		t.Fatalf("NetworkCIDR = %q, want empty when interface differs", record.NetworkCIDR)
	}
}

func TestDeviceFromRecord(t *testing.T) {
	t.Parallel()

	record := Record{
		MAC:           "aa:bb:cc:dd:ee:ff",
		LastIP:        "192.168.1.70",
		DisplayName:   "Stored Name",
		Manufacturer:  "Stored Inc",
		InterfaceName: "en0",
		FirstSeen:     time.Unix(100, 0).UTC(),
		LastSeen:      time.Unix(200, 0).UTC(),
		ExtraData:     map[string]string{"persisted": "true"},
	}

	device, ok := DeviceFromRecord(record)
	if !ok {
		t.Fatal("expected device to be created from record")
	}
	if got := device.IP().String(); got != "192.168.1.70" {
		t.Fatalf("IP() = %q, want %q", got, "192.168.1.70")
	}
	if got := device.DisplayName(); got != "Stored Name" {
		t.Fatalf("DisplayName() = %q, want %q", got, "Stored Name")
	}
	if got := device.ExtraData()["persisted"]; got != "true" {
		t.Fatalf("ExtraData()[persisted] = %q, want %q", got, "true")
	}
}
