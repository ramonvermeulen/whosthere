package config

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
)

func TestValidateAndNormalizeDurations(t *testing.T) {
	cfg := &Config{
		ScanInterval: -1,
		ScanDuration: 0,
		Splash:       SplashConfig{Enabled: true, Delay: -1},
		Scanners:     ScannerConfig{MDNS: ScannerToggle{Enabled: true}},
	}

	err := cfg.validateAndNormalize()
	if err == nil {
		t.Fatalf("expected validation errors")
	}

	if !strings.Contains(err.Error(), "scan_interval must be > 0") {
		t.Errorf("expected scan_interval error, got %v", err)
	}
	if cfg.ScanInterval != DefaultScanInterval {
		t.Errorf("expected scan interval default %v, got %v", DefaultScanInterval, cfg.ScanInterval)
	}

	if !strings.Contains(err.Error(), "scan_duration must be > 0") {
		t.Errorf("expected scan_duration error, got %v", err)
	}
	if cfg.ScanDuration != DefaultScanDuration {
		t.Errorf("expected scan duration default %v, got %v", DefaultScanDuration, cfg.ScanDuration)
	}

	if !strings.Contains(err.Error(), "splash.delay must be >= 0") {
		t.Errorf("expected splash delay error, got %v", err)
	}
	if cfg.Splash.Delay != DefaultSplashDelay {
		t.Errorf("expected splash delay default %v, got %v", DefaultSplashDelay, cfg.Splash.Delay)
	}
}

func TestValidateAndNormalizeDurationRelationship(t *testing.T) {
	cfg := &Config{
		ScanInterval: 5 * time.Second,
		ScanDuration: 10 * time.Second,
		Splash:       SplashConfig{Enabled: true, Delay: DefaultSplashDelay},
		Scanners:     ScannerConfig{MDNS: ScannerToggle{Enabled: true}},
	}

	err := cfg.validateAndNormalize()
	if err == nil {
		t.Fatalf("expected validation error")
	}

	if !strings.Contains(err.Error(), "scan_duration must be <= scan_interval") {
		t.Errorf("expected relationship error, got %v", err)
	}
	if cfg.ScanDuration != cfg.ScanInterval {
		t.Errorf("expected scan duration coerced to interval %v, got %v", cfg.ScanInterval, cfg.ScanDuration)
	}
}

func TestValidateAndNormalizeScanners(t *testing.T) {
	cfg := &Config{
		ScanInterval: DefaultScanInterval,
		ScanDuration: DefaultScanDuration,
		Splash:       SplashConfig{Enabled: true, Delay: DefaultSplashDelay},
		Scanners:     ScannerConfig{},
	}

	err := cfg.validateAndNormalize()
	if err == nil {
		t.Fatalf("expected validation error")
	}

	if !strings.Contains(err.Error(), "at least one scanner must be enabled") {
		t.Errorf("expected scanner error, got %v", err)
	}

	if !cfg.Scanners.MDNS.Enabled || !cfg.Scanners.SSDP.Enabled || !cfg.Scanners.ARP.Enabled {
		t.Errorf("expected all scanners default-enabled when none specified, got %+v", cfg.Scanners)
	}
}

func TestValidateAndNormalizeHappyPath(t *testing.T) {
	cfg := &Config{
		ScanInterval: 15 * time.Second,
		ScanDuration: 5 * time.Second,
		Splash:       SplashConfig{Enabled: false, Delay: 2 * time.Second},
		Scanners: ScannerConfig{
			MDNS: ScannerToggle{Enabled: true},
			SSDP: ScannerToggle{Enabled: false},
			ARP:  ScannerToggle{Enabled: true},
		},
	}

	if err := cfg.validateAndNormalize(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDefaultConfigProducesValidConfig(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.validateAndNormalize(); err != nil {
		t.Fatalf("expected default config to be valid, got %v", err)
	}

	if cfg.Theme.Name != DefaultThemeName {
		t.Fatalf("expected default theme %q, got %q", DefaultThemeName, cfg.Theme.Name)
	}
}

func TestLoaderValidateNilConfig(t *testing.T) {
	if err := validateAndNormalize(nil); !errors.Is(err, ErrConfigNil) {
		t.Fatalf("expected ErrConfigNil, got %v", err)
	}
}

func TestYAMLUnmarshalAndValidateHappyPath(t *testing.T) {
	raw := `
scan_interval: 15s
scan_duration: 5s
splash:
  enabled: false
  delay: 750ms
scanners:
  mdns:
    enabled: true
  ssdp:
    enabled: false
  arp:
    enabled: true
port_scanner:
  timeout: 5s
  tcp: [80, 443]
`

	cfg := DefaultConfig()
	if err := yaml.Unmarshal([]byte(raw), cfg); err != nil {
		t.Fatalf("unmarshal yaml: %v", err)
	}

	if err := cfg.validateAndNormalize(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	if got, want := cfg.ScanInterval, 15*time.Second; got != want {
		t.Errorf("scan interval: got %v, want %v", got, want)
	}
	if got, want := cfg.ScanDuration, 5*time.Second; got != want {
		t.Errorf("scan duration: got %v, want %v", got, want)
	}
	if cfg.Splash.Enabled {
		t.Errorf("expected splash disabled")
	}
	if got, want := cfg.Splash.Delay, 750*time.Millisecond; got != want {
		t.Errorf("splash delay: got %v, want %v", got, want)
	}
	if !cfg.Scanners.MDNS.Enabled || cfg.Scanners.SSDP.Enabled || !cfg.Scanners.ARP.Enabled {
		t.Errorf("scanner flags unexpected: %+v", cfg.Scanners)
	}
	if len(cfg.PortScanner.TCP) != 2 || cfg.PortScanner.TCP[0] != 80 || cfg.PortScanner.TCP[1] != 443 {
		t.Errorf("tcp ports unexpected: %v", cfg.PortScanner.TCP)
	}
	if cfg.PortScanner.Timeout != DefaultPortScanTimeout {
		t.Errorf("timeout unexpected: got %v, want %v", cfg.PortScanner.Timeout, DefaultPortScanTimeout)
	}
	if cfg.Theme.Enabled != DefaultThemeEnabled {
		t.Errorf("theme enabled unexpected: got %v, want %v", cfg.Theme.Enabled, DefaultThemeEnabled)
	}
}

func TestYAMLUnmarshalAndValidateFixesInvalids(t *testing.T) {
	raw := `
scan_interval: -5s
scan_duration: 0s
splash:
  enabled: true
  delay: -2s
scanners:
  mdns:
    enabled: false
  ssdp:
    enabled: false
  arp:
    enabled: false
`

	cfg := DefaultConfig()
	if err := yaml.Unmarshal([]byte(raw), cfg); err != nil {
		t.Fatalf("unmarshal yaml: %v", err)
	}

	err := cfg.validateAndNormalize()
	if err == nil {
		t.Fatalf("expected validation error")
	}

	msg := err.Error()
	for _, expected := range []string{
		"scan_interval must be > 0",
		"scan_duration must be > 0",
		"splash.delay must be >= 0",
		"at least one scanner must be enabled",
	} {
		if !strings.Contains(msg, expected) {
			t.Errorf("expected error %q in %q", expected, msg)
		}
	}

	if cfg.ScanInterval != DefaultScanInterval {
		t.Errorf("expected default scan interval %v, got %v", DefaultScanInterval, cfg.ScanInterval)
	}
	if cfg.ScanDuration != DefaultScanDuration {
		t.Errorf("expected default scan duration %v, got %v", DefaultScanDuration, cfg.ScanDuration)
	}
	if cfg.Splash.Delay != DefaultSplashDelay {
		t.Errorf("expected default splash delay %v, got %v", DefaultSplashDelay, cfg.Splash.Delay)
	}
	if !cfg.Scanners.MDNS.Enabled || !cfg.Scanners.SSDP.Enabled || !cfg.Scanners.ARP.Enabled {
		t.Errorf("expected scanners re-enabled to defaults, got %+v", cfg.Scanners)
	}
}
