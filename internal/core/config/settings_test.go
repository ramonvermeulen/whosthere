package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
)

type settingTestCase struct {
	yamlKey      string
	envVar       string
	envValue     string
	expectedEnv  any
	flagValue    string
	expectedFlag any
	yamlValue    string
	expectedYAML any
}

func getSettingTestCases() []settingTestCase {
	return []settingTestCase{
		{
			yamlKey:      "network_interface",
			envVar:       "WHOSTHERE__NETWORK_INTERFACE",
			envValue:     "eth0",
			expectedEnv:  "eth0",
			flagValue:    "wlan0",
			expectedFlag: "wlan0",
			yamlValue:    "en0",
			expectedYAML: "en0",
		},
		{
			yamlKey:      "scan_timeout",
			envVar:       "WHOSTHERE__SCAN_TIMEOUT",
			envValue:     "15s",
			expectedEnv:  15 * time.Second,
			flagValue:    "20s",
			expectedFlag: 20 * time.Second,
			yamlValue:    "10s",
			expectedYAML: 10 * time.Second,
		},
		{
			yamlKey:      "scan_interval",
			envVar:       "WHOSTHERE__SCAN_INTERVAL",
			envValue:     "45s",
			expectedEnv:  45 * time.Second,
			flagValue:    "60s",
			expectedFlag: 60 * time.Second,
			yamlValue:    "30s",
			expectedYAML: 30 * time.Second,
		},
		{
			yamlKey:      "scanners.mdns.enabled",
			envVar:       "WHOSTHERE__SCANNERS__MDNS__ENABLED",
			envValue:     "false",
			expectedEnv:  false,
			flagValue:    "true",
			expectedFlag: true,
			yamlValue:    "false",
			expectedYAML: false,
		},
		{
			yamlKey:      "scanners.ssdp.enabled",
			envVar:       "WHOSTHERE__SCANNERS__SSDP__ENABLED",
			envValue:     "false",
			expectedEnv:  false,
			flagValue:    "true",
			expectedFlag: true,
			yamlValue:    "false",
			expectedYAML: false,
		},
		{
			yamlKey:      "scanners.arp.enabled",
			envVar:       "WHOSTHERE__SCANNERS__ARP__ENABLED",
			envValue:     "false",
			expectedEnv:  false,
			flagValue:    "true",
			expectedFlag: true,
			yamlValue:    "false",
			expectedYAML: false,
		},
		{
			yamlKey:      "sweeper.enabled",
			envVar:       "WHOSTHERE__SWEEPER__ENABLED",
			envValue:     "false",
			expectedEnv:  false,
			flagValue:    "true",
			expectedFlag: true,
			yamlValue:    "false",
			expectedYAML: false,
		},
		{
			yamlKey:      "sweeper.interval",
			envVar:       "WHOSTHERE__SWEEPER__INTERVAL",
			envValue:     "10m",
			expectedEnv:  10 * time.Minute,
			flagValue:    "15m",
			expectedFlag: 15 * time.Minute,
			yamlValue:    "5m",
			expectedYAML: 5 * time.Minute,
		},
		{
			yamlKey:      "sweeper.timeout",
			envVar:       "WHOSTHERE__SWEEPER__TIMEOUT",
			envValue:     "3s",
			expectedEnv:  3 * time.Second,
			flagValue:    "5s",
			expectedFlag: 5 * time.Second,
			yamlValue:    "2s",
			expectedYAML: 2 * time.Second,
		},
		{
			yamlKey:      "port_scanner.timeout",
			envVar:       "WHOSTHERE__PORT_SCANNER__TIMEOUT",
			envValue:     "8s",
			expectedEnv:  8 * time.Second,
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "6s",
			expectedYAML: 6 * time.Second,
		},
		{
			yamlKey:      "port_scanner.tcp",
			envVar:       "WHOSTHERE__PORT_SCANNER__TCP",
			envValue:     "22,80,443",
			expectedEnv:  []int{22, 80, 443},
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "[22, 80]",
			expectedYAML: []int{22, 80},
		},
		{
			yamlKey:      "splash.enabled",
			envVar:       "WHOSTHERE__SPLASH__ENABLED",
			envValue:     "false",
			expectedEnv:  false,
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "false",
			expectedYAML: false,
		},
		{
			yamlKey:      "splash.delay",
			envVar:       "WHOSTHERE__SPLASH__DELAY",
			envValue:     "2s",
			expectedEnv:  2 * time.Second,
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "500ms",
			expectedYAML: 500 * time.Millisecond,
		},
		{
			yamlKey:      "theme.enabled",
			envVar:       "WHOSTHERE__THEME__ENABLED",
			envValue:     "false",
			expectedEnv:  false,
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "false",
			expectedYAML: false,
		},
		{
			yamlKey:      "theme.name",
			envVar:       "WHOSTHERE__THEME__NAME",
			envValue:     "dark",
			expectedEnv:  "dark",
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "light",
			expectedYAML: "light",
		},
		{
			yamlKey:      "theme.no_color",
			envVar:       "WHOSTHERE__THEME__NO_COLOR",
			envValue:     "true",
			expectedEnv:  true,
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "true",
			expectedYAML: true,
		},
		{
			yamlKey:      "theme.primitive_background_color",
			envVar:       "WHOSTHERE__THEME__PRIMITIVE_BACKGROUND_COLOR",
			envValue:     "#111111",
			expectedEnv:  "#111111",
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "#000000",
			expectedYAML: "#000000",
		},
		{
			yamlKey:      "theme.contrast_background_color",
			envVar:       "WHOSTHERE__THEME__CONTRAST_BACKGROUND_COLOR",
			envValue:     "#222222",
			expectedEnv:  "#222222",
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "#111111",
			expectedYAML: "#111111",
		},
		{
			yamlKey:      "theme.more_contrast_background_color",
			envVar:       "WHOSTHERE__THEME__MORE_CONTRAST_BACKGROUND_COLOR",
			envValue:     "#333333",
			expectedEnv:  "#333333",
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "#222222",
			expectedYAML: "#222222",
		},
		{
			yamlKey:      "theme.border_color",
			envVar:       "WHOSTHERE__THEME__BORDER_COLOR",
			envValue:     "#444444",
			expectedEnv:  "#444444",
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "#333333",
			expectedYAML: "#333333",
		},
		{
			yamlKey:      "theme.title_color",
			envVar:       "WHOSTHERE__THEME__TITLE_COLOR",
			envValue:     "#555555",
			expectedEnv:  "#555555",
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "#444444",
			expectedYAML: "#444444",
		},
		{
			yamlKey:      "theme.graphics_color",
			envVar:       "WHOSTHERE__THEME__GRAPHICS_COLOR",
			envValue:     "#666666",
			expectedEnv:  "#666666",
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "#555555",
			expectedYAML: "#555555",
		},
		{
			yamlKey:      "theme.primary_text_color",
			envVar:       "WHOSTHERE__THEME__PRIMARY_TEXT_COLOR",
			envValue:     "#777777",
			expectedEnv:  "#777777",
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "#666666",
			expectedYAML: "#666666",
		},
		{
			yamlKey:      "theme.secondary_text_color",
			envVar:       "WHOSTHERE__THEME__SECONDARY_TEXT_COLOR",
			envValue:     "#888888",
			expectedEnv:  "#888888",
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "#777777",
			expectedYAML: "#777777",
		},
		{
			yamlKey:      "theme.tertiary_text_color",
			envVar:       "WHOSTHERE__THEME__TERTIARY_TEXT_COLOR",
			envValue:     "#999999",
			expectedEnv:  "#999999",
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "#888888",
			expectedYAML: "#888888",
		},
		{
			yamlKey:      "theme.inverse_text_color",
			envVar:       "WHOSTHERE__THEME__INVERSE_TEXT_COLOR",
			envValue:     "#aaaaaa",
			expectedEnv:  "#aaaaaa",
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "#999999",
			expectedYAML: "#999999",
		},
		{
			yamlKey:      "theme.contrast_secondary_text_color",
			envVar:       "WHOSTHERE__THEME__CONTRAST_SECONDARY_TEXT_COLOR",
			envValue:     "#bbbbbb",
			expectedEnv:  "#bbbbbb",
			flagValue:    "",
			expectedFlag: nil,
			yamlValue:    "#aaaaaa",
			expectedYAML: "#aaaaaa",
		},
	}
}

func TestSettings_EnvOverride(t *testing.T) {
	for _, tc := range getSettingTestCases() {
		tc := tc
		t.Run(tc.yamlKey+"/env", func(t *testing.T) {
			snap := SnapshotEnv()
			RestoreEnv(map[string]string{})
			t.Cleanup(func() { RestoreEnv(snap) })

			_ = os.Setenv(tc.envVar, tc.envValue)

			cfg := DefaultConfig()
			if err := ApplyEnv(cfg); err != nil {
				t.Fatalf("ApplyEnv: %v", err)
			}

			got := getConfigValue(cfg, tc.yamlKey)
			if !equalValues(got, tc.expectedEnv) {
				t.Errorf("got %v, want %v", got, tc.expectedEnv)
			}
		})
	}
}

func TestSettings_FlagOverride(t *testing.T) {
	settings := settingsByYAMLKey()

	for _, tc := range getSettingTestCases() {
		tc := tc
		setting := settings[tc.yamlKey]
		if setting == nil || !setting.hasSource(SourceFlag) || tc.flagValue == "" {
			continue
		}

		t.Run(tc.yamlKey+"/flag", func(t *testing.T) {
			cfg := DefaultConfig()

			if err := SetByYAMLKey(cfg, tc.yamlKey, tc.flagValue); err != nil {
				t.Fatalf("SetByYAMLKey: %v", err)
			}

			got := getConfigValue(cfg, tc.yamlKey)
			if !equalValues(got, tc.expectedFlag) {
				t.Errorf("got %v, want %v", got, tc.expectedFlag)
			}
		})
	}
}

func TestSettings_YAMLOverride(t *testing.T) {
	for _, tc := range getSettingTestCases() {
		tc := tc
		t.Run(tc.yamlKey+"/yaml", func(t *testing.T) {
			yamlContent := buildYAMLForKey(tc.yamlKey, tc.yamlValue)

			cfg := DefaultConfig()
			if err := unmarshalYAML([]byte(yamlContent), cfg); err != nil {
				t.Fatalf("unmarshalYAML: %v", err)
			}

			got := getConfigValue(cfg, tc.yamlKey)
			if !equalValues(got, tc.expectedYAML) {
				t.Errorf("got %v, want %v", got, tc.expectedYAML)
			}
		})
	}
}

func TestSettings_Precedence_FlagOverEnv(t *testing.T) {
	settings := settingsByYAMLKey()

	for _, tc := range getSettingTestCases() {
		tc := tc
		setting := settings[tc.yamlKey]
		if setting == nil || !setting.hasSource(SourceFlag) || tc.flagValue == "" {
			continue
		}

		t.Run(tc.yamlKey, func(t *testing.T) {
			snap := SnapshotEnv()
			RestoreEnv(map[string]string{})
			t.Cleanup(func() { RestoreEnv(snap) })

			_ = os.Setenv(tc.envVar, tc.envValue)

			cfg := DefaultConfig()

			if err := ApplyEnv(cfg); err != nil {
				t.Fatalf("ApplyEnv: %v", err)
			}

			if err := SetByYAMLKey(cfg, tc.yamlKey, tc.flagValue); err != nil {
				t.Fatalf("SetByYAMLKey: %v", err)
			}

			got := getConfigValue(cfg, tc.yamlKey)
			if !equalValues(got, tc.expectedFlag) {
				t.Errorf("flag should win over env: got %v, want %v", got, tc.expectedFlag)
			}
		})
	}
}

func TestSettings_Precedence_EnvOverYAML(t *testing.T) {
	for _, tc := range getSettingTestCases() {
		tc := tc
		t.Run(tc.yamlKey, func(t *testing.T) {
			snap := SnapshotEnv()
			RestoreEnv(map[string]string{})
			t.Cleanup(func() { RestoreEnv(snap) })

			cfg := DefaultConfig()

			yamlContent := buildYAMLForKey(tc.yamlKey, tc.yamlValue)
			if err := unmarshalYAML([]byte(yamlContent), cfg); err != nil {
				t.Fatalf("unmarshalYAML: %v", err)
			}

			_ = os.Setenv(tc.envVar, tc.envValue)
			if err := ApplyEnv(cfg); err != nil {
				t.Fatalf("ApplyEnv: %v", err)
			}

			got := getConfigValue(cfg, tc.yamlKey)
			if !equalValues(got, tc.expectedEnv) {
				t.Errorf("env should win over yaml: got %v, want %v", got, tc.expectedEnv)
			}
		})
	}
}

func TestSettings_Precedence_FlagOverEnvOverYAML(t *testing.T) {
	settings := settingsByYAMLKey()

	for _, tc := range getSettingTestCases() {
		tc := tc
		setting := settings[tc.yamlKey]
		if setting == nil || !setting.hasSource(SourceFlag) || tc.flagValue == "" {
			continue
		}

		t.Run(tc.yamlKey, func(t *testing.T) {
			snap := SnapshotEnv()
			RestoreEnv(map[string]string{})
			t.Cleanup(func() { RestoreEnv(snap) })

			cfg := DefaultConfig()

			yamlContent := buildYAMLForKey(tc.yamlKey, tc.yamlValue)
			if err := unmarshalYAML([]byte(yamlContent), cfg); err != nil {
				t.Fatalf("unmarshalYAML: %v", err)
			}

			_ = os.Setenv(tc.envVar, tc.envValue)
			if err := ApplyEnv(cfg); err != nil {
				t.Fatalf("ApplyEnv: %v", err)
			}

			if err := SetByYAMLKey(cfg, tc.yamlKey, tc.flagValue); err != nil {
				t.Fatalf("SetByYAMLKey: %v", err)
			}

			got := getConfigValue(cfg, tc.yamlKey)
			if !equalValues(got, tc.expectedFlag) {
				t.Errorf("flag should win over env and yaml: got %v, want %v", got, tc.expectedFlag)
			}
		})
	}
}

func TestFullYAMLConfig_LoadFromFile(t *testing.T) {
	snap := SnapshotEnv()
	RestoreEnv(map[string]string{})
	t.Cleanup(func() { RestoreEnv(snap) })

	// Note: network_interface is excluded from this test because:
	// 1. It requires a valid interface name which varies by system (lo/lo0/Loopback Pseudo-Interface 1)
	// 2. It's already tested in individual setting tests (env/flag/yaml)
	// 3. This test focuses on the full loading path, not individual field validation
	fullYAML := `
scan_timeout: 12s
scan_interval: 45s

scanners:
  mdns:
    enabled: false
  ssdp:
    enabled: false
  arp:
    enabled: true

sweeper:
  enabled: false
  interval: 8m
  timeout: 4s

port_scanner:
  timeout: 7s
  tcp: [22, 80, 443, 8080]

splash:
  enabled: false
  delay: 750ms

theme:
  enabled: false
  name: "custom"
  primitive_background_color: "#000001"
  contrast_background_color: "#000002"
  more_contrast_background_color: "#000003"
  border_color: "#000004"
  title_color: "#000005"
  graphics_color: "#000006"
  primary_text_color: "#000007"
  secondary_text_color: "#000008"
  tertiary_text_color: "#000009"
  inverse_text_color: "#00000a"
  contrast_secondary_text_color: "#00000b"
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(fullYAML), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	cfg, err := LoadForMode(ModeApp, &Flags{ConfigFile: configPath})
	if err != nil {
		t.Fatalf("LoadForMode: %v", err)
	}

	assertions := []struct {
		yamlKey  string
		got      any
		expected any
	}{
		{"scan_timeout", cfg.ScanTimeout, 12 * time.Second},
		{"scan_interval", cfg.ScanInterval, 45 * time.Second},
		{"scanners.mdns.enabled", cfg.Scanners.MDNS.Enabled, false},
		{"scanners.ssdp.enabled", cfg.Scanners.SSDP.Enabled, false},
		{"scanners.arp.enabled", cfg.Scanners.ARP.Enabled, true},
		{"sweeper.enabled", cfg.Sweeper.Enabled, false},
		{"sweeper.interval", cfg.Sweeper.Interval, 8 * time.Minute},
		{"sweeper.timeout", cfg.Sweeper.Timeout, 4 * time.Second},
		{"port_scanner.timeout", cfg.PortScanner.Timeout, 7 * time.Second},
		{"port_scanner.tcp", cfg.PortScanner.TCP, []int{22, 80, 443, 8080}},
		{"splash.enabled", cfg.Splash.Enabled, false},
		{"splash.delay", cfg.Splash.Delay, 750 * time.Millisecond},
		{"theme.enabled", cfg.Theme.Enabled, false},
		{"theme.no_color", cfg.Theme.NoColor, false},
		{"theme.name", cfg.Theme.Name, "custom"},
		{"theme.primitive_background_color", cfg.Theme.PrimitiveBackgroundColor, "#000001"},
		{"theme.contrast_background_color", cfg.Theme.ContrastBackgroundColor, "#000002"},
		{"theme.more_contrast_background_color", cfg.Theme.MoreContrastBackgroundColor, "#000003"},
		{"theme.border_color", cfg.Theme.BorderColor, "#000004"},
		{"theme.title_color", cfg.Theme.TitleColor, "#000005"},
		{"theme.graphics_color", cfg.Theme.GraphicsColor, "#000006"},
		{"theme.primary_text_color", cfg.Theme.PrimaryTextColor, "#000007"},
		{"theme.secondary_text_color", cfg.Theme.SecondaryTextColor, "#000008"},
		{"theme.tertiary_text_color", cfg.Theme.TertiaryTextColor, "#000009"},
		{"theme.inverse_text_color", cfg.Theme.InverseTextColor, "#00000a"},
		{"theme.contrast_secondary_text_color", cfg.Theme.ContrastSecondaryTextColor, "#00000b"},
	}

	testedKeys := make(map[string]bool)
	// network_interface is tested in individual setting tests but excluded from full YAML test
	// due to system-dependent interface names (lo/lo0/Loopback Pseudo-Interface 1)
	testedKeys["network_interface"] = true
	for _, a := range assertions {
		testedKeys[a.yamlKey] = true
		if !equalValues(a.got, a.expected) {
			t.Errorf("%s: got %v, want %v", a.yamlKey, a.got, a.expected)
		}
	}

	for _, s := range GlobalSettings() {
		if s.YAMLKey == "" {
			continue
		}
		if !testedKeys[s.YAMLKey] {
			t.Errorf("setting %q is not covered in TestFullYAMLConfig_LoadFromFile", s.YAMLKey)
		}
	}
}

func TestMeta_AllSettingsHaveTestCases(t *testing.T) {
	testedKeys := make(map[string]bool)
	for _, tc := range getSettingTestCases() {
		testedKeys[tc.yamlKey] = true
	}

	for _, s := range GlobalSettings() {
		if s.YAMLKey == "" {
			continue
		}

		if !testedKeys[s.YAMLKey] {
			t.Errorf("setting %q has no test case in getSettingTestCases()", s.YAMLKey)
		}
	}
}

func TestMeta_AllSettingsHaveSetterAndGetter(t *testing.T) {
	for _, s := range GlobalSettings() {
		if s.YAMLKey == "" {
			continue
		}

		if s.Set == nil {
			t.Errorf("setting %q is missing Setter", s.YAMLKey)
		}
		if s.Get == nil {
			t.Errorf("setting %q is missing Getter", s.YAMLKey)
		}
	}
}

func getConfigValue(cfg *Config, yamlKey string) any {
	settings := settingsByYAMLKey()
	s := settings[yamlKey]
	if s == nil || s.Get == nil {
		return nil
	}
	return s.Get(cfg)
}

func buildYAMLForKey(yamlKey, value string) string {
	parts := strings.Split(yamlKey, ".")
	indent := ""
	var lines []string

	for i, part := range parts {
		if i == len(parts)-1 {
			switch {
			case strings.HasPrefix(value, "[") || strings.HasPrefix(value, "{"):
				lines = append(lines, indent+part+": "+value)
			case strings.HasPrefix(value, "#") || strings.Contains(value, " "):
				lines = append(lines, indent+part+": \""+value+"\"")
			default:
				lines = append(lines, indent+part+": "+value)
			}
		} else {
			lines = append(lines, indent+part+":")
			indent += "  "
		}
	}

	return strings.Join(lines, "\n")
}

func equalValues(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	aSlice, aIsSlice := a.([]int)
	bSlice, bIsSlice := b.([]int)
	if aIsSlice && bIsSlice {
		if len(aSlice) != len(bSlice) {
			return false
		}
		for i := range aSlice {
			if aSlice[i] != bSlice[i] {
				return false
			}
		}
		return true
	}

	return reflect.DeepEqual(a, b)
}

func unmarshalYAML(data []byte, cfg *Config) error {
	return yaml.Unmarshal(data, cfg)
}
