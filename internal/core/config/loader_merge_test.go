package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
	_ = os.Setenv("USERPROFILE", "/tmp")

	flags := &Flags{Overrides: map[string]string{"scan_timeout": "9s"}}
	cfg, err := LoadMerged(flags)
	require.NoError(t, err, "LoadMerged")
	require.Equal(t, 9*time.Second, cfg.ScanTimeout, "expected scan_timeout 9s")
}
