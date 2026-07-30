package devicemeta

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

func TestNormalizeMAC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
		ok    bool
	}{
		{
			name:  "normalizes uppercase with dashes",
			input: "AA-BB-CC-DD-EE-FF",
			want:  "aa:bb:cc:dd:ee:ff",
			ok:    true,
		},
		{
			name:  "trims surrounding whitespace",
			input: "  aa:bb:cc:dd:ee:ff  ",
			want:  "aa:bb:cc:dd:ee:ff",
			ok:    true,
		},
		{
			name:  "rejects invalid MAC",
			input: "not-a-mac",
			want:  "",
			ok:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := NormalizeMAC(tt.input)
			if ok != tt.ok {
				t.Fatalf("NormalizeMAC(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("NormalizeMAC(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBoltStore_SetGetClearAlias(t *testing.T) {
	t.Parallel()

	store, err := Open(filepath.Join(t.TempDir(), "devices.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	const mac = "AA:BB:CC:DD:EE:FF"

	record, found, err := store.Get("", mac)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if found {
		t.Fatalf("expected alias record to be absent, got %+v", record)
	}

	if err := store.SetAlias("", mac, " Living Room Speaker "); err != nil {
		t.Fatalf("SetAlias() error = %v", err)
	}

	record, found, err = store.Get("", mac)
	if err != nil {
		t.Fatalf("Get() after SetAlias error = %v", err)
	}
	if !found {
		t.Fatal("expected alias record to exist")
	}
	if record.Alias != "Living Room Speaker" {
		t.Fatalf("record.Alias = %q, want %q", record.Alias, "Living Room Speaker")
	}

	if err := store.ClearAlias("", mac); err != nil {
		t.Fatalf("ClearAlias() error = %v", err)
	}

	record, found, err = store.Get("", mac)
	if err != nil {
		t.Fatalf("Get() after ClearAlias error = %v", err)
	}
	if found {
		t.Fatalf("expected alias record to be removed, got %+v", record)
	}
}

func TestBoltStore_PersistsAcrossReopen(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "devices.db")

	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if err := store.SetAlias("", "aa:bb:cc:dd:ee:ff", "Router"); err != nil {
		t.Fatalf("SetAlias() error = %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("Open() reopen error = %v", err)
	}
	t.Cleanup(func() {
		_ = reopened.Close()
	})

	record, found, err := reopened.Get("", "AA-BB-CC-DD-EE-FF")
	if err != nil {
		t.Fatalf("Get() after reopen error = %v", err)
	}
	if !found {
		t.Fatal("expected alias record after reopen")
	}
	if record.Alias != "Router" {
		t.Fatalf("record.Alias = %q, want %q", record.Alias, "Router")
	}
}

func TestBoltStore_ResetAliases(t *testing.T) {
	t.Parallel()

	store, err := Open(filepath.Join(t.TempDir(), "devices.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	if err := store.SetAlias("", "aa:bb:cc:dd:ee:ff", "Router"); err != nil {
		t.Fatalf("SetAlias(first) error = %v", err)
	}
	if err := store.SetAlias("", "aa:bb:cc:dd:ee:11", "Printer"); err != nil {
		t.Fatalf("SetAlias(second) error = %v", err)
	}

	if err := store.ResetAliases(); err != nil {
		t.Fatalf("ResetAliases() error = %v", err)
	}

	tests := []string{"aa:bb:cc:dd:ee:ff", "aa:bb:cc:dd:ee:11"}
	for _, mac := range tests {
		record, found, err := store.Get("", mac)
		if err != nil {
			t.Fatalf("Get(%q) after reset error = %v", mac, err)
		}
		if found {
			t.Fatalf("expected alias for %s to be removed, got %+v", mac, record)
		}
	}
}

func TestBoltStore_UpsertAndForEach(t *testing.T) {
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

	record := Record{
		MAC:           "AA:BB:CC:DD:EE:FF",
		Alias:         "Router",
		LastIP:        "192.168.1.1",
		DisplayName:   "Gateway",
		Manufacturer:  "Acme",
		InterfaceName: "en0",
		FirstSeen:     firstSeen,
		LastSeen:      lastSeen,
		ExtraData:     map[string]string{"source": "mdns"},
	}

	if err := store.Upsert(record); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	got, found, err := store.Get("", record.MAC)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !found {
		t.Fatal("expected full record to exist")
	}
	if got.MAC != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("record.MAC = %q, want %q", got.MAC, "aa:bb:cc:dd:ee:ff")
	}
	if got.LastIP != record.LastIP || got.DisplayName != record.DisplayName || got.Manufacturer != record.Manufacturer || got.InterfaceName != record.InterfaceName {
		t.Fatalf("record fields not persisted correctly: %+v", got)
	}
	if !got.FirstSeen.Equal(firstSeen) || !got.LastSeen.Equal(lastSeen) {
		t.Fatalf("timestamps not persisted correctly: %+v", got)
	}
	if got.ExtraData["source"] != "mdns" {
		t.Fatalf("ExtraData not persisted correctly: %+v", got.ExtraData)
	}

	var records []Record
	if err := store.ForEach(func(record Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("ForEach() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("ForEach() count = %d, want 1", len(records))
	}
}

func TestBoltStore_SetAliasPreservesOtherFields(t *testing.T) {
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
		LastIP:        "192.168.1.20",
		DisplayName:   "Existing Device",
		Manufacturer:  "Acme",
		InterfaceName: "en0",
		FirstSeen:     firstSeen,
		LastSeen:      lastSeen,
		ExtraData:     map[string]string{"kind": "speaker"},
	}); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	if err := store.SetAlias("", "aa:bb:cc:dd:ee:ff", "Kitchen"); err != nil {
		t.Fatalf("SetAlias() error = %v", err)
	}

	record, found, err := store.Get("", "aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !found {
		t.Fatal("expected record to exist")
	}
	if record.Alias != "Kitchen" {
		t.Fatalf("record.Alias = %q, want %q", record.Alias, "Kitchen")
	}
	if record.DisplayName != "Existing Device" || record.LastIP != "192.168.1.20" || record.ExtraData["kind"] != "speaker" {
		t.Fatalf("SetAlias() should preserve non-alias fields, got %+v", record)
	}
	if !record.FirstSeen.Equal(firstSeen) || !record.LastSeen.Equal(lastSeen) {
		t.Fatalf("SetAlias() should preserve timestamps, got %+v", record)
	}
}

func TestBoltStore_ClearAliasPreservesNonAliasRecord(t *testing.T) {
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
		Alias:       "Kitchen",
		DisplayName: "Existing Device",
		LastIP:      "192.168.1.20",
	}); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	if err := store.ClearAlias("", "aa:bb:cc:dd:ee:ff"); err != nil {
		t.Fatalf("ClearAlias() error = %v", err)
	}

	record, found, err := store.Get("", "aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !found {
		t.Fatal("expected record to remain after alias clear")
	}
	if record.Alias != "" {
		t.Fatalf("record.Alias = %q, want empty", record.Alias)
	}
	if record.DisplayName != "Existing Device" || record.LastIP != "192.168.1.20" {
		t.Fatalf("ClearAlias() should preserve non-alias fields, got %+v", record)
	}
}

func TestBoltStore_GetLegacyAliasOnlyRecord(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "devices.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	db := store.(*boltStore).db
	if err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(deviceMetadataBucket)
		raw, err := json.Marshal(map[string]any{
			"alias":      "Legacy Router",
			"updated_at": time.Unix(123, 0).UTC(),
		})
		if err != nil {
			return err
		}
		return bucket.Put([]byte("aa:bb:cc:dd:ee:ff"), raw)
	}); err != nil {
		t.Fatalf("seed legacy record: %v", err)
	}

	record, found, err := store.Get("", "aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !found {
		t.Fatal("expected legacy record to be found")
	}
	if record.MAC != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("record.MAC = %q, want normalized key", record.MAC)
	}
	if record.Alias != "Legacy Router" {
		t.Fatalf("record.Alias = %q, want %q", record.Alias, "Legacy Router")
	}
}
