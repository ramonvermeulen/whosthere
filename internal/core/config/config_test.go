package config

import (
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/ramonvermeulen/whosthere/pkg/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAndNormalizeDurations(t *testing.T) {
	cfg := &Config{
		ScanInterval: -1,
		ScanDuration: 0,
		Splash:       SplashConfig{Enabled: true, Delay: -1},
		Scanners:     ScannerConfig{MDNS: ScannerToggle{Enabled: true}},
	}

	err := cfg.validateAndNormalize()
	require.Error(t, err, "expected validation errors")

	assert.Contains(t, err.Error(), "scan_interval must be > 0", "expected scan_interval error")
	assert.Equal(t, discovery.DefaultScanInterval, cfg.ScanInterval, "expected scan interval default")

	assert.Contains(t, err.Error(), "scan_duration must be > 0", "expected scan_duration error")
	assert.Equal(t, discovery.DefaultScanTimeout, cfg.ScanDuration, "expected scan duration default")

	assert.Contains(t, err.Error(), "splash.delay must be >= 0", "expected splash delay error")
	assert.Equal(t, DefaultSplashDelay, cfg.Splash.Delay, "expected splash delay default")
}

func TestValidateAndNormalizeHappyPath(t *testing.T) {
	cfg := &Config{
		ScanInterval: 15 * time.Second,
		ScanDuration: 5 * time.Second,
		ScanTimeout:  5 * time.Second,
		Splash:       SplashConfig{Enabled: false, Delay: 2 * time.Second},
		Scanners: ScannerConfig{
			MDNS: ScannerToggle{Enabled: true},
			SSDP: ScannerToggle{Enabled: false},
			ARP:  ScannerToggle{Enabled: true},
		},
	}

	require.NoError(t, cfg.validateAndNormalize())
}

func TestDefaultConfigProducesValidConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.NoError(t, cfg.validateAndNormalize(), "expected default config to be valid")

	require.Equal(t, DefaultThemeName, cfg.Theme.Name, "expected default theme")
	require.False(t, cfg.ScanLargeSubnets, "expected scan_large_subnets to default to false")
}

func TestYAMLUnmarshalAndValidateHappyPath(t *testing.T) {
	raw := `
target_subnets:
  - 10.0.0.42/24
  - 10.0.1.0/24
  - 10.0.1.0/24
scan_large_subnets: true
scan_interval: 15s
scan_duration: 5s
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
splash:
  enabled: false
  delay: 750ms
`

	cfg := DefaultConfig()
	require.NoError(t, yaml.Unmarshal([]byte(raw), cfg), "unmarshal yaml")
	require.NoError(t, cfg.validateAndNormalize(), "validate")

	assert.Equal(t, 15*time.Second, cfg.ScanInterval, "scan interval")
	assert.Equal(t, []string{"10.0.0.0/24", "10.0.1.0/24"}, cfg.TargetSubnets, "target subnets")
	assert.Equal(t, 5*time.Second, cfg.ScanDuration, "scan duration")
	assert.False(t, cfg.Splash.Enabled, "expected splash disabled")
	assert.Equal(t, 750*time.Millisecond, cfg.Splash.Delay, "splash delay")
	assert.True(t, cfg.Scanners.MDNS.Enabled, "mdns should be enabled")
	assert.False(t, cfg.Scanners.SSDP.Enabled, "ssdp should be disabled")
	assert.True(t, cfg.Scanners.ARP.Enabled, "arp should be enabled")
	assert.Equal(t, []int{80, 443}, cfg.PortScanner.TCP, "tcp ports")
	assert.Equal(t, DefaultPortScanTimeout, cfg.PortScanner.Timeout, "port scanner timeout")
	assert.Equal(t, DefaultThemeEnabled, cfg.Theme.Enabled, "theme enabled")
	assert.True(t, cfg.ScanLargeSubnets, "expected scan_large_subnets to be true from YAML")
}

func TestValidateAndNormalizeTargetSubnetsRejectsInvalidCIDRs(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TargetSubnets = []string{
		"10.0.0.0/24",
		"not-a-cidr",
		"2001:db8::/64",
	}

	err := cfg.validateAndNormalize()
	require.Error(t, err, "expected validation error")

	msg := err.Error()
	for _, expected := range []string{
		"target_subnets contains invalid CIDR: not-a-cidr",
		"target_subnets only supports IPv4 CIDRs: 2001:db8::/64",
	} {
		assert.Contains(t, msg, expected, "expected error in message")
	}
}

func TestYAMLUnmarshalAndValidateFixesInvalids(t *testing.T) {
	raw := `
scan_interval: -5s
scan_duration: 0s
scanners:
  mdns:
    enabled: false
  ssdp:
    enabled: false
  arp:
    enabled: false
splash:
  enabled: true
  delay: -2s
`

	cfg := DefaultConfig()
	require.NoError(t, yaml.Unmarshal([]byte(raw), cfg), "unmarshal yaml")

	err := cfg.validateAndNormalize()
	require.Error(t, err, "expected validation error")

	msg := err.Error()
	for _, expected := range []string{
		"scan_interval must be > 0",
		"scan_duration must be > 0",
		"splash.delay must be >= 0",
	} {
		assert.Contains(t, msg, expected, "expected error in message")
	}

	assert.Equal(t, discovery.DefaultScanInterval, cfg.ScanInterval, "expected default scan interval")
	assert.Equal(t, discovery.DefaultScanTimeout, cfg.ScanDuration, "expected default scan duration")
	assert.Equal(t, DefaultSplashDelay, cfg.Splash.Delay, "expected default splash delay")
}
