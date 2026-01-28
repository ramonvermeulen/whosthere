package config

import (
	"errors"
	"net"
	"strings"
	"time"
)

const (
	DefaultSplashEnabled  = true
	DefaultThemeEnabled   = true
	DefaultSweeperEnabled = true
	DefaultSplashDelay    = 1 * time.Second

	DefaultScanInterval    = 20 * time.Second
	DefaultScanDuration    = 10 * time.Second
	DefaultPortScanTimeout = 5 * time.Second
	DefaultSweeperInterval = 5 * time.Minute

	DefaultThemeName = "default"
	CustomThemeName  = "custom"
)

var DefaultTCPPorts = []int{21, 22, 23, 25, 80, 110, 135, 139, 143, 389, 443, 445, 993, 995, 1433, 1521, 3306, 3389, 5432, 5900, 8080, 8443, 9000, 9090, 9200, 9300, 10000, 27017}

// Config captures all configurable parameters for the application.
type Config struct {
	NetworkInterface string            `yaml:"network_interface"`
	ScanInterval     time.Duration     `yaml:"scan_interval"`
	ScanDuration     time.Duration     `yaml:"scan_duration"`
	Scanners         ScannerConfig     `yaml:"scanners"`
	Sweeper          SweeperConfig     `yaml:"sweeper"`
	PortScanner      PortScannerConfig `yaml:"port_scanner"`
	Splash           SplashConfig      `yaml:"splash"`
	Theme            ThemeConfig       `yaml:"theme"`
}

// ScannerToggle lets users enable/disable a scanner.
type ScannerToggle struct {
	Enabled bool `yaml:"enabled"`
}

// ScannerConfig groups scanner enablement flags.
type ScannerConfig struct {
	MDNS ScannerToggle `yaml:"mdns"`
	SSDP ScannerToggle `yaml:"ssdp"`
	ARP  ScannerToggle `yaml:"arp"`
}

// SweeperConfig controls the sweeper behavior.
type SweeperConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Interval time.Duration `yaml:"interval"`
}

// PortScannerConfig defines TCP ports to scan.
type PortScannerConfig struct {
	TCP     []int         `yaml:"tcp"`
	Timeout time.Duration `yaml:"timeout"`
}

// SplashConfig controls the splash screen visibility and timing.
type SplashConfig struct {
	Enabled bool          `yaml:"enabled"`
	Delay   time.Duration `yaml:"delay"`
}

// ThemeConfig selects a theme by name and optionally carries custom color overrides.
type ThemeConfig struct {
	Enabled                     bool   `yaml:"enabled"`
	Name                        string `yaml:"name"`
	PrimitiveBackgroundColor    string `yaml:"primitive_background_color"`
	ContrastBackgroundColor     string `yaml:"contrast_background_color"`
	MoreContrastBackgroundColor string `yaml:"more_contrast_background_color"`
	BorderColor                 string `yaml:"border_color"`
	TitleColor                  string `yaml:"title_color"`
	GraphicsColor               string `yaml:"graphics_color"`
	PrimaryTextColor            string `yaml:"primary_text_color"`
	SecondaryTextColor          string `yaml:"secondary_text_color"`
	TertiaryTextColor           string `yaml:"tertiary_text_color"`
	InverseTextColor            string `yaml:"inverse_text_color"`
	ContrastSecondaryTextColor  string `yaml:"contrast_secondary_text_color"`
}

// DefaultConfig builds a Config pre-populated with baked-in defaults.
func DefaultConfig() *Config {
	return &Config{
		ScanInterval: DefaultScanInterval,
		ScanDuration: DefaultScanDuration,
		Scanners: ScannerConfig{
			MDNS: ScannerToggle{Enabled: true},
			SSDP: ScannerToggle{Enabled: true},
			ARP:  ScannerToggle{Enabled: true},
		},
		Sweeper: SweeperConfig{
			Enabled:  DefaultSweeperEnabled,
			Interval: DefaultSweeperInterval,
		},
		PortScanner: PortScannerConfig{
			TCP:     DefaultTCPPorts,
			Timeout: DefaultPortScanTimeout,
		},
		Splash: SplashConfig{
			Enabled: DefaultSplashEnabled,
			Delay:   DefaultSplashDelay,
		},
		Theme: ThemeConfig{
			Name:    DefaultThemeName,
			Enabled: DefaultThemeEnabled,
		},
	}
}

// validateAndNormalize validates the config and fixes up out-of-range values.
func (c *Config) validateAndNormalize() error {
	var errs []string

	if c.Splash.Delay < 0 {
		errs = append(errs, "splash.delay must be >= 0")
		c.Splash.Delay = DefaultSplashDelay
	}

	if c.ScanInterval <= 0 {
		errs = append(errs, "scan_interval must be > 0")
		c.ScanInterval = DefaultScanInterval
	}

	if c.ScanDuration <= 0 {
		errs = append(errs, "scan_duration must be > 0")
		c.ScanDuration = DefaultScanDuration
	}

	if c.ScanDuration > c.ScanInterval {
		errs = append(errs, "scan_duration must be <= scan_interval")
		c.ScanDuration = c.ScanInterval
	}

	if !c.Scanners.MDNS.Enabled && !c.Scanners.SSDP.Enabled && !c.Scanners.ARP.Enabled {
		errs = append(errs, "at least one scanner must be enabled")
		c.Scanners.MDNS.Enabled = true
		c.Scanners.SSDP.Enabled = true
		c.Scanners.ARP.Enabled = true
	}

	if len(c.PortScanner.TCP) == 0 {
		c.PortScanner.TCP = DefaultTCPPorts
	}

	if c.PortScanner.Timeout <= 0 {
		c.PortScanner.Timeout = DefaultPortScanTimeout
	}

	if c.Sweeper.Interval <= 0 {
		c.Sweeper.Interval = DefaultSweeperInterval
	}

	if strings.TrimSpace(c.Theme.Name) == "" {
		c.Theme.Name = DefaultThemeName
	}

	if c.NetworkInterface != "" {
		if _, err := net.InterfaceByName(c.NetworkInterface); err != nil {
			errs = append(errs, "network_interface does not exist: "+c.NetworkInterface)
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}
