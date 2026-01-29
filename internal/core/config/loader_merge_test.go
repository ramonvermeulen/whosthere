package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoadMergedPrecedenceFlagsOverrideEnv(t *testing.T) {
	old := os.Environ()
	os.Clearenv()
	t.Cleanup(func() {
		os.Clearenv()
		for _, kv := range old {
			k, v, ok := strings.Cut(kv, "=")
			if ok {
				_ = os.Setenv(k, v)
			}
		}
	})

	_ = os.Setenv("WHOSTHERE__SCAN_TIMEOUT", "3s")
	_ = os.Setenv("HOME", "/tmp")

	flags := &Flags{Overrides: map[string]string{"scan_timeout": "9s"}}
	cfg, err := LoadMerged(flags)
	if err != nil {
		t.Fatalf("LoadMerged: %v", err)
	}

	if cfg.ScanTimeout != 9*time.Second {
		t.Fatalf("expected scan_timeout 9s, got %v", cfg.ScanTimeout)
	}
}
