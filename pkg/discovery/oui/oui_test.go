package oui

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCSVBytesHeaderAndLookup(t *testing.T) {
	csvData := []byte("Registry,Assignment,Organization Name,Organization Address\n" +
		"MA-L,286FB9,Test Org,Somewhere\n")

	m, err := parseCSVBytes(csvData)
	require.NoError(t, err, "parseCSVBytes error")
	require.NotEmpty(t, m, "expected at least one entry")
	require.Equal(t, "Test Org", m["286FB9"], "expected prefix 286FB9 -> 'Test Org'")
}

func TestLookup(t *testing.T) {
	csvData := []byte("Registry,Assignment,Organization Name,Organization Address\n" +
		"MA-L,286FB9,Test Org,Somewhere\n")

	m, err := parseCSVBytes(csvData)
	require.NoError(t, err, "parseCSVBytes error")
	reg := &Registry{prefixMap: m, logger: slog.Default()}

	tests := []struct {
		name    string
		mac     string
		wantOrg string
		wantOK  bool
	}{
		{"valid MAC", "28:6f:b9:00:11:22", "Test Org", true},
		{"invalid MAC", "ff:ff:ff:00:00:00", "", false},
		{"empty MAC", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := reg.Lookup(tt.mac)
			require.Equal(t, tt.wantOK, ok, "Lookup(%q) ok", tt.mac)
			require.Equal(t, tt.wantOrg, got, "Lookup(%q) org", tt.mac)
		})
	}
}
