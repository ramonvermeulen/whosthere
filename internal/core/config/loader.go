package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/ramonvermeulen/whosthere/internal/core/paths"
)

const (
	defaultConfigFileName = "config.yaml"
	// Environment variable to override config file path
	configEnvVar = "WHOSTHERE_CONFIG"
)

var ErrConfigNil = errors.New("config is nil")

// Load resolves the config path, reads/creates YAML, and returns the merged config.
func Load(pathOverride string) (*Config, error) {
	resolvedPath, err := resolveConfigPath(pathOverride)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()

	if err := ensureConfigFile(resolvedPath, cfg); err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(resolvedPath)
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(raw, cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}

	if err := validateAndNormalize(cfg); err != nil {
		return cfg, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

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

func validateAndNormalize(cfg *Config) error {
	if cfg == nil {
		return ErrConfigNil
	}

	return cfg.validateAndNormalize()
}

// writeConfigFile marshals the config to YAML and writes it to the specified path,
// ensuring the parent directory exists.
func writeConfigFile(path string, cfg *Config) error {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := marshalConfigWithComments(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// marshalConfigWithComments creates a YAML representation with helpful comments.
func marshalConfigWithComments(cfg *Config) ([]byte, error) {
	tcpPorts := make([]string, len(cfg.PortScanner.TCP))
	for i, p := range cfg.PortScanner.TCP {
		tcpPorts[i] = fmt.Sprintf("%d", p)
	}

	commented := fmt.Sprintf(`# whosthere configuration file
# For more information, visit: https://github.com/ramonvermeulen/whosthere
# Uncomment the next line to configure a specific network interface - uses OS default if not set
# network_interface: eth0

# How often to run discovery scans
scan_interval: %s

# Maximum duration for each scan
scan_duration: %s

# Scanner configuration
scanners:
  mdns:
    enabled: %t
  ssdp:
    enabled: %t
  arp:
    enabled: %t

sweeper:
  enabled: %t
  sweep_interval: %s

# Port scanner configuration
port_scanner:
  timeout: %s
  # List of TCP ports to scan on discovered devices
  tcp: [%s]

# Splash screen configuration
splash:
  enabled: %t
  delay: %s

# Theme configuration
theme:
  # When disabled, the TUI will use the terminal it's default ANSI colors
  # Also see the NO_COLOR environment variable to completely disable ANSI colors
  enabled: %t
  # See the complete list of available themes at https://github.com/ramonvermeulen/whosthere/tree/main/internal/ui/theme/theme.go
  # Set name to "custom" to use the custom colors below
  # For any color that is not configured it will take the default theme value as fallback
  name: %s
  
  # Custom theme colors (uncomment and set name: custom to use)
  # primitive_background_color: "#000a1a"
  # contrast_background_color: "#001a33"
  # more_contrast_background_color: "#003366"
  # border_color: "#0088ff"
  # title_color: "#00ffff"
  # graphics_color: "#00ffaa"
  # primary_text_color: "#cceeff"
  # secondary_text_color: "#6699ff"
  # tertiary_text_color: "#ffaa00"
  # inverse_text_color: "#000a1a"
  # contrast_secondary_text_color: "#88ddff"
`,
		cfg.ScanInterval,
		cfg.ScanDuration,
		cfg.Scanners.MDNS.Enabled,
		cfg.Scanners.SSDP.Enabled,
		cfg.Scanners.ARP.Enabled,
		cfg.Sweeper.Enabled,
		cfg.Sweeper.Interval,
		cfg.PortScanner.Timeout,
		strings.Join(tcpPorts, ", "),
		cfg.Splash.Enabled,
		cfg.Splash.Delay,
		cfg.Theme.Enabled,
		cfg.Theme.Name,
	)

	return []byte(commented), nil
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
