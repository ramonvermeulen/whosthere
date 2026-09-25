package devicemeta

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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
			require.Equal(t, tt.ok, ok, "NormalizeMAC(%q) ok", tt.input)
			require.Equal(t, tt.want, got, "NormalizeMAC(%q)", tt.input)
		})
	}
}

func TestBoltStore_SetGetClearAlias(t *testing.T) {
	t.Parallel()

	store, err := Open(filepath.Join(t.TempDir(), "devices.db"))
	require.NoError(t, err, "Open()")
	t.Cleanup(func() {
		_ = store.Close()
	})

	const mac = "AA:BB:CC:DD:EE:FF"

	record, found, err := store.Get(mac)
	require.NoError(t, err, "Get()")
	require.False(t, found, "expected alias record to be absent, got %+v", record)

	require.NoError(t, store.SetAlias(mac, " Living Room Speaker "), "SetAlias()")

	record, found, err = store.Get(mac)
	require.NoError(t, err, "Get() after SetAlias")
	require.True(t, found, "expected alias record to exist")
	require.Equal(t, "Living Room Speaker", record.Alias)

	require.NoError(t, store.ClearAlias(mac), "ClearAlias()")

	record, found, err = store.Get(mac)
	require.NoError(t, err, "Get() after ClearAlias")
	require.False(t, found, "expected alias record to be removed, got %+v", record)
}

func TestBoltStore_PersistsAcrossReopen(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "devices.db")

	store, err := Open(path)
	require.NoError(t, err, "Open()")

	require.NoError(t, store.SetAlias("aa:bb:cc:dd:ee:ff", "Router"), "SetAlias()")
	require.NoError(t, store.Close(), "Close()")

	reopened, err := Open(path)
	require.NoError(t, err, "Open() reopen")
	t.Cleanup(func() {
		_ = reopened.Close()
	})

	record, found, err := reopened.Get("AA-BB-CC-DD-EE-FF")
	require.NoError(t, err, "Get() after reopen")
	require.True(t, found, "expected alias record after reopen")
	require.Equal(t, "Router", record.Alias)
}

func TestBoltStore_ResetAliases(t *testing.T) {
	t.Parallel()

	store, err := Open(filepath.Join(t.TempDir(), "devices.db"))
	require.NoError(t, err, "Open()")
	t.Cleanup(func() {
		_ = store.Close()
	})

	require.NoError(t, store.SetAlias("aa:bb:cc:dd:ee:ff", "Router"), "SetAlias(first)")
	require.NoError(t, store.SetAlias("aa:bb:cc:dd:ee:11", "Printer"), "SetAlias(second)")

	require.NoError(t, store.ResetAliases(), "ResetAliases()")

	tests := []string{"aa:bb:cc:dd:ee:ff", "aa:bb:cc:dd:ee:11"}
	for _, mac := range tests {
		record, found, err := store.Get(mac)
		require.NoError(t, err, "Get(%s) after reset", mac)
		require.False(t, found, "expected alias for %s to be removed, got %+v", mac, record)
	}
}
