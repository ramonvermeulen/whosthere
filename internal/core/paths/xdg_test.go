package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	t.Setenv(xdgConfigDirEnv, "/tmp/xdg_config")
	dir, err := ConfigDir()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected := "/tmp/xdg_config/whosthere"
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}

	// Test without XDG_CONFIG_HOME
	t.Setenv(xdgConfigDirEnv, "")
	dir, err = ConfigDir()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	home, _ := os.UserHomeDir()
	expected = filepath.Join(home, defaultConfigDir, appName)
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestStateDir(t *testing.T) {
	// Test with XDG_STATE_HOME set
	t.Setenv(xdgStateDirEnv, "/tmp/xdg_state")
	dir, err := StateDir()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected := "/tmp/xdg_state/whosthere"
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}

	// Test without XDG_STATE_HOME
	t.Setenv(xdgStateDirEnv, "")
	dir, err = StateDir()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	home, _ := os.UserHomeDir()
	expected = filepath.Join(home, defaultStateDir, appName)
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}
