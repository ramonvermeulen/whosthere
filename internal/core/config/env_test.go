package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func SnapshotEnv() map[string]string {
	out := map[string]string{}
	for _, kv := range os.Environ() {
		k, v, ok := strings.Cut(kv, "=")
		if ok {
			out[k] = v
		}
	}
	return out
}

func RestoreEnv(snapshot map[string]string) {
	os.Clearenv()
	for k, v := range snapshot {
		_ = os.Setenv(k, v)
	}
}

func TestApplyEnvSetsNestedValues(t *testing.T) {
	snap := SnapshotEnv()
	RestoreEnv(map[string]string{})
	t.Cleanup(func() { RestoreEnv(snap) })

	_ = os.Setenv("WHOSTHERE__SWEEPER__ENABLED", "false")
	_ = os.Setenv("WHOSTHERE__SCANNERS__MDNS__ENABLED", "false")
	_ = os.Setenv("WHOSTHERE__SCAN_TIMEOUT", "7s")
	_ = os.Setenv("WHOSTHERE__PORT_SCANNER__TCP", "80,443")
	_ = os.Setenv("WHOSTHERE__THEME__NAME", "custom")

	cfg := DefaultConfig()
	if err := ApplyEnv(cfg); err != nil {
		t.Fatalf("ApplyEnv: %v", err)
	}

	if cfg.Sweeper.Enabled {
		t.Fatalf("expected sweeper disabled")
	}
	if cfg.Scanners.MDNS.Enabled {
		t.Fatalf("expected mdns disabled")
	}
	if cfg.ScanTimeout != 7*time.Second {
		t.Fatalf("expected scan_timeout 7s, got %v", cfg.ScanTimeout)
	}
	if len(cfg.PortScanner.TCP) != 2 || cfg.PortScanner.TCP[0] != 80 || cfg.PortScanner.TCP[1] != 443 {
		t.Fatalf("expected tcp ports [80 443], got %v", cfg.PortScanner.TCP)
	}
	if cfg.Theme.Name != "custom" {
		t.Fatalf("expected theme name custom, got %q", cfg.Theme.Name)
	}
}

func TestApplyEnvUnknownKeysAreIgnored(t *testing.T) {
	snap := SnapshotEnv()
	RestoreEnv(map[string]string{})
	t.Cleanup(func() { RestoreEnv(snap) })

	_ = os.Setenv("WHOSTHERE__DOES_NOT_EXIST", "wat")

	cfg := DefaultConfig()
	if err := ApplyEnv(cfg); err != nil {
		t.Fatalf("ApplyEnv: %v", err)
	}
}
