package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ramonvermeulen/whosthere/internal/core/paths"
)

const (
	defaultConfigFileName = "config.yaml"
	// Environment variable to override config file path
	configEnvVar = "WHOSTHERE_CONFIG"
)

var ErrConfigNil = errors.New("config is nil")

// ensureConfigFile ensures that the config file exists at the given path.
// If the file does not exist, it is created with default values.
func ensureConfigFile(path string, defaults *Config) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	return writeConfigFile(path, defaults)
}

// resolveConfigPath returns the path using precedence: flag override > env var > XDG default.
func resolveConfigPath(pathOverride string) (string, error) {
	if pathOverride != "" {
		return pathOverride, nil
	}

	if env := os.Getenv(configEnvVar); env != "" {
		return env, nil
	}

	dir, err := paths.ConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}

	return filepath.Join(dir, defaultConfigFileName), nil
}

// writeConfigFile marshals the config to YAML and writes it to the specified path,
// ensuring the parent directory exists.
func writeConfigFile(path string, _ *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data := GenerateDefaultYAML()

	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// Save writes the config to the specified path (or resolves the default path if empty).
func Save(cfg *Config, pathOverride string) error {
	if cfg == nil {
		return ErrConfigNil
	}

	resolvedPath, err := resolveConfigPath(pathOverride)
	if err != nil {
		return fmt.Errorf("resolve config path: %w", err)
	}

	return writeConfigFile(resolvedPath, cfg)
}
