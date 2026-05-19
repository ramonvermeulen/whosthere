package devicemeta

import (
	"path/filepath"
	"testing"
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

	record, found, err := store.Get(mac)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if found {
		t.Fatalf("expected alias record to be absent, got %+v", record)
	}

	if err := store.SetAlias(mac, " Living Room Speaker "); err != nil {
		t.Fatalf("SetAlias() error = %v", err)
	}

	record, found, err = store.Get(mac)
	if err != nil {
		t.Fatalf("Get() after SetAlias error = %v", err)
	}
	if !found {
		t.Fatal("expected alias record to exist")
	}
	if record.Alias != "Living Room Speaker" {
		t.Fatalf("record.Alias = %q, want %q", record.Alias, "Living Room Speaker")
	}

	if err := store.ClearAlias(mac); err != nil {
		t.Fatalf("ClearAlias() error = %v", err)
	}

	record, found, err = store.Get(mac)
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

	if err := store.SetAlias("aa:bb:cc:dd:ee:ff", "Router"); err != nil {
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

	record, found, err := reopened.Get("AA-BB-CC-DD-EE-FF")
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

	if err := store.SetAlias("aa:bb:cc:dd:ee:ff", "Router"); err != nil {
		t.Fatalf("SetAlias(first) error = %v", err)
	}
	if err := store.SetAlias("aa:bb:cc:dd:ee:11", "Printer"); err != nil {
		t.Fatalf("SetAlias(second) error = %v", err)
	}

	if err := store.ResetAliases(); err != nil {
		t.Fatalf("ResetAliases() error = %v", err)
	}

	tests := []string{"aa:bb:cc:dd:ee:ff", "aa:bb:cc:dd:ee:11"}
	for _, mac := range tests {
		record, found, err := store.Get(mac)
		if err != nil {
			t.Fatalf("Get(%q) after reset error = %v", mac, err)
		}
		if found {
			t.Fatalf("expected alias for %s to be removed, got %+v", mac, record)
		}
	}
}
