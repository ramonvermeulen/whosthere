package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestConfigDir(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	tmpDir := t.TempDir()
	t.Setenv(xdgConfigDirEnv, tmpDir)
	dir, err := ConfigDir()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected := filepath.Join(tmpDir, appName)
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
	tmpDir := t.TempDir()
	t.Setenv(xdgStateDirEnv, tmpDir)
	dir, err := StateDir()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected := filepath.Join(tmpDir, appName)
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}

	// Test without XDG_STATE_HOME - use temp HOME to avoid writing to real home
	t.Setenv(xdgStateDirEnv, "")
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
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
		expected = filepath.Join(tmpHome, defaultStateDir, appName)
	}

	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}
