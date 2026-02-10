package config

import (
	"strings"

	"github.com/spf13/cobra"
)

type RunMode int

const (
	ModeApp RunMode = iota
	ModeCLI
)

type FlagType int

const (
	FlagTypeString FlagType = iota
	FlagTypeBool
)

type SettingSource int

const (
	SourceYAML SettingSource = iota
	SourceEnv
	SourceFlag
)

type Setter func(cfg *Config, value string) error
type Getter func(cfg *Config) any

type YAMLDoc struct {
	Comment         string
	ExampleValue    string
	CommentedOut    bool
	BlankLineBefore bool
}

type GlobalSetting struct {
	YAMLKey  string
	EnvVar   string
	FlagName string
	Short    string
	Usage    string
	Type     FlagType
	Hidden   bool
	Sources  map[SettingSource]bool
	Set      Setter
	Get      Getter
	Doc      YAMLDoc
}

func (s *GlobalSetting) hasSource(src SettingSource) bool {
	if s == nil || s.Sources == nil {
		return true
	}
	return s.Sources[src]
}

func GlobalSettings() []GlobalSetting {
	all := map[SettingSource]bool{SourceYAML: true, SourceEnv: true, SourceFlag: true}
	yamlEnvOnly := map[SettingSource]bool{SourceYAML: true, SourceEnv: true}
	flagEnvOnly := map[SettingSource]bool{SourceEnv: true, SourceFlag: true}

	return []GlobalSetting{
		// special cases
		// the "config" flag is a special case that only exists as a flag/env var and does not have a corresponding YAML key
		// the reason it doesn't have a YAML key is that it specifies the path to the YAML config file, so it can't be set via YAML itself
		{
			EnvVar:   "WHOSTHERE_CONFIG",
			FlagName: "config",
			Short:    "c",
			Usage:    "Path to config file.",
			Type:     FlagTypeString,
			Sources:  flagEnvOnly,
		},
		// the "pprof-port" flag is a special case that only exists as a flag/env var and does not have a corresponding YAML key
		// the reason it doesn't have a YAML key is that it's meant for debugging and profiling purposes, and we don't want it to be set via YAML in production environments
		{
			FlagName: "pprof-port",
			Short:    "D",
			Usage:    "Pprof HTTP server port for debugging and profiling purposes (e.g., 6060)",
			Type:     FlagTypeString,
			Sources:  flagEnvOnly,
		},

		// general global settings
		{
			YAMLKey:  "network_interface",
			FlagName: "interface",
			Short:    "i",
			Usage:    "Network interface to use for scanning (overrides env/config).",
			Type:     FlagTypeString,
			Sources:  all,
			Set:      func(c *Config, v string) error { c.NetworkInterface = v; return nil },
			Get:      func(c *Config) any { return c.NetworkInterface },
			Doc: YAMLDoc{
				Comment:      "Uncomment the next line to configure a specific network interface - uses OS default if not set",
				ExampleValue: "eth0",
				CommentedOut: true,
			},
		},
		{
			YAMLKey:  "scan_interval",
			FlagName: "interval",
			Short:    "n",
			Usage:    "Scan interval duration (e.g., 30s).",
			Type:     FlagTypeString,
			Sources:  all,
			Set: func(c *Config, v string) error {
				d, err := parseDuration(v)
				if err != nil {
					return err
				}
				c.ScanInterval = d
				return nil
			},
			Get: func(c *Config) any { return c.ScanInterval },
			Doc: YAMLDoc{
				Comment: "How often to run discovery scans",
			},
		},
		{
			YAMLKey:  "scan_timeout",
			FlagName: "timeout",
			Short:    "t",
			Usage:    "Scan timeout duration (e.g., 10s).",
			Type:     FlagTypeString,
			Sources:  all,
			Set: func(c *Config, v string) error {
				d, err := parseDuration(v)
				if err != nil {
					return err
				}
				c.ScanTimeout = d
				return nil
			},
			Get: func(c *Config) any { return c.ScanTimeout },
			Doc: YAMLDoc{
				Comment: "Maximum timeout for each scan",
			},
		},
		{
			YAMLKey:  "scanners.mdns.enabled",
			FlagName: "mdns",
			Short:    "m",
			Usage:    "Enable/disable the mDNS scanner.",
			Type:     FlagTypeBool,
			Sources:  all,
			Set: func(c *Config, v string) error {
				b, err := parseBool(v)
				if err != nil {
					return err
				}
				c.Scanners.MDNS.Enabled = b
				return nil
			},
			Get: func(c *Config) any { return c.Scanners.MDNS.Enabled },
			Doc: YAMLDoc{},
		},
		{
			YAMLKey:  "scanners.ssdp.enabled",
			FlagName: "ssdp",
			Short:    "s",
			Usage:    "Enable/disable the SSDP scanner.",
			Type:     FlagTypeBool,
			Sources:  all,
			Set: func(c *Config, v string) error {
				b, err := parseBool(v)
				if err != nil {
					return err
				}
				c.Scanners.SSDP.Enabled = b
				return nil
			},
			Get: func(c *Config) any { return c.Scanners.SSDP.Enabled },
			Doc: YAMLDoc{},
		},
		{
			YAMLKey:  "scanners.arp.enabled",
			FlagName: "arp",
			Short:    "a",
			Usage:    "Enable/disable the ARP scanner.",
			Type:     FlagTypeBool,
			Sources:  all,
			Set: func(c *Config, v string) error {
				b, err := parseBool(v)
				if err != nil {
					return err
				}
				c.Scanners.ARP.Enabled = b
				return nil
			},
			Get: func(c *Config) any { return c.Scanners.ARP.Enabled },
			Doc: YAMLDoc{},
		},
		{
			YAMLKey:  "sweeper.enabled",
			FlagName: "sweeper",
			Short:    "S",
			Usage:    "Enable/disable the sweeper.",
			Type:     FlagTypeBool,
			Sources:  all,
			Set: func(c *Config, v string) error {
				b, err := parseBool(v)
				if err != nil {
					return err
				}
				c.Sweeper.Enabled = b
				return nil
			},
			Get: func(c *Config) any { return c.Sweeper.Enabled },
			Doc: YAMLDoc{},
		},
		{
			YAMLKey:  "sweeper.interval",
			FlagName: "sweeper-interval",
			Short:    "W",
			Usage:    "Sweeper interval duration (e.g., 5m).",
			Type:     FlagTypeString,
			Sources:  all,
			Set: func(c *Config, v string) error {
				d, err := parseDuration(v)
				if err != nil {
					return err
				}
				c.Sweeper.Interval = d
				return nil
			},
			Get: func(c *Config) any { return c.Sweeper.Interval },
			Doc: YAMLDoc{},
		},
		{
			YAMLKey:  "sweeper.timeout",
			FlagName: "sweeper-timeout",
			Short:    "T",
			Usage:    "Sweeper timeout duration (e.g., 2s).",
			Type:     FlagTypeString,
			Sources:  all,
			Set: func(c *Config, v string) error {
				d, err := parseDuration(v)
				if err != nil {
					return err
				}
				c.Sweeper.Timeout = d
				return nil
			},
			Get: func(c *Config) any { return c.Sweeper.Timeout },
			Doc: YAMLDoc{},
		},
		{
			YAMLKey:  "port_scanner.timeout",
			FlagName: "portscan-timeout",
			Usage:    "Port scan timeout duration per port (e.g., 5s).",
			Type:     FlagTypeString,
			Sources:  yamlEnvOnly,
			Set: func(c *Config, v string) error {
				d, err := parseDuration(v)
				if err != nil {
					return err
				}
				c.PortScanner.Timeout = d
				return nil
			},
			Get: func(c *Config) any { return c.PortScanner.Timeout },
			Doc: YAMLDoc{},
		},
		{
			YAMLKey:  "port_scanner.tcp",
			FlagName: "portscan-tcp",
			Usage:    "Comma-separated TCP ports to scan (e.g., 22,80,443).",
			Type:     FlagTypeString,
			Sources:  yamlEnvOnly,
			Set: func(c *Config, v string) error {
				ports, err := parseIntSlice(v)
				if err != nil {
					return err
				}
				c.PortScanner.TCP = ports
				return nil
			},
			Get: func(c *Config) any { return c.PortScanner.TCP },
			Doc: YAMLDoc{
				Comment: "List of TCP ports to scan on discovered devices",
			},
		},
		{
			YAMLKey:  "splash.enabled",
			FlagName: "splash-enabled",
			Usage:    "Enable/disable the splash screen.",
			Type:     FlagTypeBool,
			Sources:  yamlEnvOnly,
			Set: func(c *Config, v string) error {
				b, err := parseBool(v)
				if err != nil {
					return err
				}
				c.Splash.Enabled = b
				return nil
			},
			Get: func(c *Config) any { return c.Splash.Enabled },
			Doc: YAMLDoc{},
		},
		{
			YAMLKey:  "splash.delay",
			FlagName: "splash-delay",
			Usage:    "Splash screen delay duration (e.g., 1s).",
			Type:     FlagTypeString,
			Sources:  yamlEnvOnly,
			Set: func(c *Config, v string) error {
				d, err := parseDuration(v)
				if err != nil {
					return err
				}
				c.Splash.Delay = d
				return nil
			},
			Get: func(c *Config) any { return c.Splash.Delay },
			Doc: YAMLDoc{},
		},
		{
			YAMLKey:  "theme.enabled",
			FlagName: "theme-enabled",
			Usage:    "Enable/disable theming (ANSI colors).",
			Type:     FlagTypeBool,
			Sources:  yamlEnvOnly,
			Set: func(c *Config, v string) error {
				b, err := parseBool(v)
				if err != nil {
					return err
				}
				c.Theme.Enabled = b
				return nil
			},
			Get: func(c *Config) any { return c.Theme.Enabled },
			Doc: YAMLDoc{
				Comment: "When disabled, the TUI will use the terminal it's default ANSI colors\nAlso see the NO_COLOR environment variable to completely disable ANSI colors",
			},
		},
		{
			YAMLKey:  "theme.name",
			FlagName: "theme",
			Usage:    "Theme name (e.g., default, custom).",
			Type:     FlagTypeString,
			Sources:  yamlEnvOnly,
			Set:      func(c *Config, v string) error { c.Theme.Name = v; return nil },
			Get:      func(c *Config) any { return c.Theme.Name },
			Doc: YAMLDoc{
				Comment: "See the complete list of available themes at https://github.com/ramonvermeulen/whosthere/tree/main/internal/ui/theme/theme.go\nSet name to \"custom\" to use the custom colors below\nFor any color that is not configured it will take the default theme value as fallback",
			},
		},
		{
			YAMLKey: "theme.primitive_background_color",
			Usage:   "Theme primitive background color (hex).",
			Type:    FlagTypeString,
			Sources: yamlEnvOnly,
			Set:     func(c *Config, v string) error { c.Theme.PrimitiveBackgroundColor = v; return nil },
			Get:     func(c *Config) any { return c.Theme.PrimitiveBackgroundColor },
			Doc: YAMLDoc{
				Comment:         "Custom theme colors (uncomment and set name: custom to use)",
				ExampleValue:    "\"#000a1a\"",
				CommentedOut:    true,
				BlankLineBefore: true,
			},
		},
		{
			YAMLKey: "theme.contrast_background_color",
			Usage:   "Theme contrast background color (hex).",
			Type:    FlagTypeString,
			Sources: yamlEnvOnly,
			Set:     func(c *Config, v string) error { c.Theme.ContrastBackgroundColor = v; return nil },
			Get:     func(c *Config) any { return c.Theme.ContrastBackgroundColor },
			Doc: YAMLDoc{
				ExampleValue: "\"#001a33\"",
				CommentedOut: true,
			},
		},
		{
			YAMLKey: "theme.more_contrast_background_color",
			Usage:   "Theme more-contrast background color (hex).",
			Type:    FlagTypeString,
			Sources: yamlEnvOnly,
			Set:     func(c *Config, v string) error { c.Theme.MoreContrastBackgroundColor = v; return nil },
			Get:     func(c *Config) any { return c.Theme.MoreContrastBackgroundColor },
			Doc: YAMLDoc{
				ExampleValue: "\"#003366\"",
				CommentedOut: true,
			},
		},
		{
			YAMLKey: "theme.border_color",
			Usage:   "Theme border color (hex).",
			Type:    FlagTypeString,
			Sources: yamlEnvOnly,
			Set:     func(c *Config, v string) error { c.Theme.BorderColor = v; return nil },
			Get:     func(c *Config) any { return c.Theme.BorderColor },
			Doc: YAMLDoc{
				ExampleValue: "\"#0088ff\"",
				CommentedOut: true,
			},
		},
		{
			YAMLKey: "theme.title_color",
			Usage:   "Theme title color (hex).",
			Type:    FlagTypeString,
			Sources: yamlEnvOnly,
			Set:     func(c *Config, v string) error { c.Theme.TitleColor = v; return nil },
			Get:     func(c *Config) any { return c.Theme.TitleColor },
			Doc: YAMLDoc{
				ExampleValue: "\"#00ffff\"",
				CommentedOut: true,
			},
		},
		{
			YAMLKey: "theme.graphics_color",
			Usage:   "Theme graphics color (hex).",
			Type:    FlagTypeString,
			Sources: yamlEnvOnly,
			Set:     func(c *Config, v string) error { c.Theme.GraphicsColor = v; return nil },
			Get:     func(c *Config) any { return c.Theme.GraphicsColor },
			Doc: YAMLDoc{
				ExampleValue: "\"#00ffaa\"",
				CommentedOut: true,
			},
		},
		{
			YAMLKey: "theme.primary_text_color",
			Usage:   "Theme primary text color (hex).",
			Type:    FlagTypeString,
			Sources: yamlEnvOnly,
			Set:     func(c *Config, v string) error { c.Theme.PrimaryTextColor = v; return nil },
			Get:     func(c *Config) any { return c.Theme.PrimaryTextColor },
			Doc: YAMLDoc{
				ExampleValue: "\"#cceeff\"",
				CommentedOut: true,
			},
		},
		{
			YAMLKey: "theme.secondary_text_color",
			Usage:   "Theme secondary text color (hex).",
			Type:    FlagTypeString,
			Sources: yamlEnvOnly,
			Set:     func(c *Config, v string) error { c.Theme.SecondaryTextColor = v; return nil },
			Get:     func(c *Config) any { return c.Theme.SecondaryTextColor },
			Doc: YAMLDoc{
				ExampleValue: "\"#6699ff\"",
				CommentedOut: true,
			},
		},
		{
			YAMLKey: "theme.tertiary_text_color",
			Usage:   "Theme tertiary text color (hex).",
			Type:    FlagTypeString,
			Sources: yamlEnvOnly,
			Set:     func(c *Config, v string) error { c.Theme.TertiaryTextColor = v; return nil },
			Get:     func(c *Config) any { return c.Theme.TertiaryTextColor },
			Doc: YAMLDoc{
				ExampleValue: "\"#ffaa00\"",
				CommentedOut: true,
			},
		},
		{
			YAMLKey: "theme.inverse_text_color",
			Usage:   "Theme inverse text color (hex).",
			Type:    FlagTypeString,
			Sources: yamlEnvOnly,
			Set:     func(c *Config, v string) error { c.Theme.InverseTextColor = v; return nil },
			Get:     func(c *Config) any { return c.Theme.InverseTextColor },
			Doc: YAMLDoc{
				ExampleValue: "\"#000a1a\"",
				CommentedOut: true,
			},
		},
		{
			YAMLKey: "theme.contrast_secondary_text_color",
			Usage:   "Theme contrast secondary text color (hex).",
			Type:    FlagTypeString,
			Sources: yamlEnvOnly,
			Set:     func(c *Config, v string) error { c.Theme.ContrastSecondaryTextColor = v; return nil },
			Get:     func(c *Config) any { return c.Theme.ContrastSecondaryTextColor },
			Doc: YAMLDoc{
				ExampleValue: "\"#88ddff\"",
				CommentedOut: true,
			},
		},
	}
}

func settingsByYAMLKey() map[string]*GlobalSetting {
	settings := GlobalSettings()
	m := make(map[string]*GlobalSetting, len(settings))
	for i := range settings {
		if settings[i].YAMLKey != "" {
			m[settings[i].YAMLKey] = &settings[i]
		}
	}
	return m
}

func RegisterGlobalConfigFlags(cmd *cobra.Command, flags *Flags) {
	if flags == nil {
		return
	}
	if flags.Overrides == nil {
		flags.Overrides = map[string]string{}
	}

	for _, s := range GlobalSettings() {
		s := s
		if !s.hasSource(SourceFlag) {
			continue
		}

		switch s.FlagName {
		case "config":
			cmd.PersistentFlags().StringVarP(&flags.ConfigFile, s.FlagName, s.Short, "", s.Usage)
			continue
		case "pprof-port":
			cmd.PersistentFlags().StringVar(&flags.PprofPort, s.FlagName, "", s.Usage)
			continue
		}

		switch s.Type {
		case FlagTypeString:
			registerStringSetting(cmd, flags, &s, s.Usage)
		case FlagTypeBool:
			registerBoolSetting(cmd, flags, &s, s.Usage)
		}

		if s.Hidden {
			_ = cmd.PersistentFlags().MarkHidden(s.FlagName)
		}
	}
}

func registerStringSetting(cmd *cobra.Command, flags *Flags, s *GlobalSetting, usage string) {
	if s == nil {
		return
	}

	if s.Short != "" {
		cmd.PersistentFlags().StringP(s.FlagName, s.Short, "", usage)
	} else {
		cmd.PersistentFlags().String(s.FlagName, "", usage)
	}

	cmd.PersistentPreRunE = chainPersistentPreRun(cmd.PersistentPreRunE, func(cmd *cobra.Command, _ []string) error {
		if !cmd.Flags().Changed(s.FlagName) {
			return nil
		}
		val, err := cmd.Flags().GetString(s.FlagName)
		if err != nil {
			return err
		}
		flags.Overrides[s.YAMLKey] = strings.TrimSpace(val)
		return nil
	})
}

func registerBoolSetting(cmd *cobra.Command, flags *Flags, s *GlobalSetting, usage string) {
	if s == nil {
		return
	}

	if s.Short != "" {
		cmd.PersistentFlags().BoolP(s.FlagName, s.Short, false, usage)
	} else {
		cmd.PersistentFlags().Bool(s.FlagName, false, usage)
	}

	cmd.PersistentPreRunE = chainPersistentPreRun(cmd.PersistentPreRunE, func(cmd *cobra.Command, _ []string) error {
		if !cmd.Flags().Changed(s.FlagName) {
			return nil
		}
		val, err := cmd.Flags().GetBool(s.FlagName)
		if err != nil {
			return err
		}
		if val {
			flags.Overrides[s.YAMLKey] = "true"
		} else {
			flags.Overrides[s.YAMLKey] = "false"
		}
		return nil
	})
}

type persistentPreRunE func(cmd *cobra.Command, args []string) error

func chainPersistentPreRun(existing, next persistentPreRunE) persistentPreRunE {
	if existing == nil {
		return next
	}
	return func(cmd *cobra.Command, args []string) error {
		if err := existing(cmd, args); err != nil {
			return err
		}
		return next(cmd, args)
	}
}
