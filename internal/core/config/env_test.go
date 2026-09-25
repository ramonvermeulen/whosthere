package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, ApplyEnv(cfg), "ApplyEnv")

	require.False(t, cfg.Sweeper.Enabled, "expected sweeper disabled")
	require.False(t, cfg.Scanners.MDNS.Enabled, "expected mdns disabled")
	require.Equal(t, 7*time.Second, cfg.ScanTimeout, "expected scan_timeout 7s")
	require.Equal(t, []int{80, 443}, cfg.PortScanner.TCP, "expected tcp ports [80 443]")
	require.Equal(t, "custom", cfg.Theme.Name, "expected theme name custom")
}

func TestApplyEnvUnknownKeysAreIgnored(t *testing.T) {
	snap := SnapshotEnv()
	RestoreEnv(map[string]string{})
	t.Cleanup(func() { RestoreEnv(snap) })

	_ = os.Setenv("WHOSTHERE__DOES_NOT_EXIST", "wat")

	cfg := DefaultConfig()
	require.NoError(t, ApplyEnv(cfg), "ApplyEnv")
}
