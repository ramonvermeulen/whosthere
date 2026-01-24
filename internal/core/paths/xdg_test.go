package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestConfigDir(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	t.Setenv(xdgConfigDirEnv, filepath.FromSlash("/tmp/xdg_config"))
	dir, err := ConfigDir()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected := filepath.FromSlash("/tmp/xdg_config/whosthere")
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}

	// Test without XDG_CONFIG_HOME
	t.Setenv(xdgConfigDirEnv, "")
	dir, err = ConfigDir()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Expectation depends on OS now
	ucd, err := os.UserConfigDir()
	if err != nil {
		// Fallback expectation
		home, _ := os.UserHomeDir()
		expected = filepath.Join(home, defaultConfigDir, appName)
	} else {
		expected = filepath.Join(ucd, appName)
	}

	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestStateDir(t *testing.T) {
	// Test with XDG_STATE_HOME set
	t.Setenv(xdgStateDirEnv, filepath.FromSlash("/tmp/xdg_state"))
	dir, err := StateDir()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected := filepath.FromSlash("/tmp/xdg_state/whosthere")
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}

	// Test without XDG_STATE_HOME
	t.Setenv(xdgStateDirEnv, "")
	dir, err = StateDir()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if runtime.GOOS == "windows" {
		// Windows expectation
		ucd, err := os.UserCacheDir()
		if err == nil {
			expected = filepath.Join(ucd, appName)
		} else {
			home, _ := os.UserHomeDir()
			expected = filepath.Join(home, "AppData", "Local", appName)
		}
	} else {
		home, _ := os.UserHomeDir()
		expected = filepath.Join(home, defaultStateDir, appName)
	}

	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}
