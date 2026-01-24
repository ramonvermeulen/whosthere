package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	appName          = "whosthere"
	xdgConfigDirEnv  = "XDG_CONFIG_HOME"
	xdgStateDirEnv   = "XDG_STATE_HOME"
	defaultConfigDir = ".config"
	defaultStateDir  = ".local/state"
)

// ConfigDir returns the XDG config directory for this app without creating it.
// It follows XDG_CONFIG_HOME when set, otherwise falls back to:
// - ~/.config/whosthere (Linux, MacOS)
// - %APPDATA%/whosthere (Windows)
func ConfigDir() (string, error) {
	if env := os.Getenv(xdgConfigDirEnv); env != "" {
		return filepath.Join(env, appName), nil
	}

	// os.UserConfigDir returns:
	// - ~/.config on Linux
	// - ~/Library/Application Support on macOS
	// - %APPDATA% on Windows
	dir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home dir construction if UserConfigDir fails
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, defaultConfigDir)
	}

	return filepath.Join(dir, appName), nil
}

// StateDir returns the XDG state directory for this app, creating it if
// necessary. It follows XDG_STATE_HOME when set, otherwise falls back to:
// - ~/.local/state/whosthere (Linux, MacOS)
// - %LOCALAPPDATA%/whosthere (Windows)
func StateDir() (string, error) {
	if env := os.Getenv(xdgStateDirEnv); env != "" {
		dir := filepath.Join(env, appName)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
		return dir, nil
	}

	// State directory: Use %LOCALAPPDATA% on Windows (via UserCacheDir).
	// On Linux/Unix, use ~/.local/state to comply with XDG spec,
	// as UserCacheDir returns ~/.cache which is volatile/temporary.
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var base string
	// Check if on Windows
	if runtime.GOOS == "windows" {
		// os.UserCacheDir() usually points to LocalAppData
		// On Windows, use LocalAppData
		ucd, err := os.UserCacheDir()
		if err == nil {
			base = ucd
		} else {
			base = filepath.Join(home, "AppData", "Local")
		}
	} else {
		// On Linux/macOS, default to ~/.local/state
		base = filepath.Join(home, defaultStateDir)
	}

	dir := filepath.Join(base, appName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}
