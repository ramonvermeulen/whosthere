package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadForModeCLISkipsConfigFile(t *testing.T) {
	snapshot := SnapshotEnv()
	RestoreEnv(map[string]string{})
	t.Cleanup(func() { RestoreEnv(snapshot) })

	_ = os.Setenv("WHOSTHERE_CONFIG", "/definitely/does/not/exist.yaml")
	cfg, err := LoadForMode(ModeCLI, &Flags{})
	require.NoError(t, err, "expected no error")
	require.NotNil(t, cfg, "expected cfg")
}
