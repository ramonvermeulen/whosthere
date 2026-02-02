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

	home, _ := os.UserHomeDir()
	expected = filepath.Join(home, defaultConfigDir, appName)
	if runtime.GOOS == "windows" {
		dir, _ := os.UserConfigDir()
		expected = filepath.Join(dir, appName)
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

	if runtime.GOOS == "windows" {
		tmpLocalAppData := filepath.Join(tmpHome, "AppData", "Local")
		t.Setenv("LOCALAPPDATA", tmpLocalAppData)
		expected = filepath.Join(tmpLocalAppData, appName)
	} else {
		expected = filepath.Join(tmpHome, defaultStateDir, appName)
	}

	dir, err = StateDir()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}
