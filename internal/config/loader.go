package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

const (
	defaultConfigDirName  = "whosthere"
	defaultConfigFileName = "config.yaml"
	// Environment variable to override config file path
	configEnvVar = "WHOSTHERE_CONFIG"
	xdgConfigEnv = "XDG_CONFIG_HOME"
)

// Load resolves the config path, reads/creates YAML, and returns the merged config.
func Load(pathOverride string) (*Config, string, error) {
	resolvedPath, err := resolveConfigPath(pathOverride)
	if err != nil {
		return nil, "", err
	}

	cfg := DefaultConfig()

	if err := ensureConfigFile(resolvedPath, cfg); err != nil {
		return nil, "", err
	}

	raw, err := os.ReadFile(resolvedPath)
	if err != nil {
		return cfg, "", fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(raw, cfg); err != nil {
		return cfg, "", fmt.Errorf("parse config: %w", err)
	}

	if cfg.Splash.Delay < 0 {
		cfg.Splash.Delay = DefaultSplashDelay
	}

	return cfg, resolvedPath, nil
}

func ensureConfigFile(path string, defaults *Config) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(defaults)
	if err != nil {
		return fmt.Errorf("serialize defaults: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write default config: %w", err)
	}

	return nil
}

// resolveConfigPath returns the path using precedence: flag override > env var > XDG default.
func resolveConfigPath(pathOverride string) (string, error) {
	if pathOverride != "" {
		return pathOverride, nil
	}

	if env := os.Getenv(configEnvVar); env != "" {
		return env, nil
	}

	base := os.Getenv(xdgConfigEnv)
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}
		base = filepath.Join(home, ".config")
	}

	return filepath.Join(base, defaultConfigDirName, defaultConfigFileName), nil
}
