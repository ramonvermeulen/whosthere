package config

import (
	"os"
	"testing"
)

func TestLoadForModeCLISkipsConfigFile(t *testing.T) {
	snapshot := SnapshotEnv()
	RestoreEnv(map[string]string{})
	t.Cleanup(func() { RestoreEnv(snapshot) })

	_ = os.Setenv("WHOSTHERE_CONFIG", "/definitely/does/not/exist.yaml")
	cfg, err := LoadForMode(ModeCLI, &Flags{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg == nil {
		t.Fatalf("expected cfg")
	}
}
