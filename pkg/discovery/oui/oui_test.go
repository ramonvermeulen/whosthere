package oui

import (
	"log/slog"
	"testing"
)

func TestParseCSVBytesHeaderAndLookup(t *testing.T) {
	csvData := []byte("Registry,Assignment,Organization Name,Organization Address\n" +
		"MA-L,286FB9,Test Org,Somewhere\n")

	m, err := parseCSVBytes(csvData)
	if err != nil {
		t.Fatalf("parseCSVBytes error: %v", err)
	}
	if len(m) == 0 {
		t.Fatalf("expected at least one entry, got 0")
	}
	if got, ok := m["286FB9"]; !ok || got != "Test Org" {
		t.Fatalf("expected prefix 286FB9 -> 'Test Org', got %q, ok=%v", got, ok)
	}
}

func TestLookup(t *testing.T) {
	csvData := []byte("Registry,Assignment,Organization Name,Organization Address\n" +
		"MA-L,286FB9,Test Org,Somewhere\n")

	m, err := parseCSVBytes(csvData)
	if err != nil {
		t.Fatalf("parseCSVBytes error: %v", err)
	}
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
			if ok != tt.wantOK || got != tt.wantOrg {
				t.Errorf("Lookup(%q) = %q, %v; want %q, %v",
					tt.mac, got, ok, tt.wantOrg, tt.wantOK)
			}
		})
	}
}
