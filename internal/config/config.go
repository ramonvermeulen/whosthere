package config

const (
	DefaultSplashEnabled = true
	DefaultSplashDelay   = float32(1.0)
)

// ThemeConfig mirrors tview.Theme; all fields are optional strings that
// override the tview.Theme defaults when set.
type ThemeConfig struct {
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

// Config captures runtime configuration values loaded from the YAML config file.
type Config struct {
	Splash SplashConfig `yaml:"splash"`
	Theme  ThemeConfig  `yaml:"theme"`
}

// SplashConfig controls the splash screen visibility and timing.
type SplashConfig struct {
	Enabled bool    `yaml:"enabled"`
	Delay   float32 `yaml:"delay"` // seconds, supports fractional values like 0.5
}

// DefaultConfig builds a Config pre-populated with baked-in defaults.
func DefaultConfig() *Config {
	return &Config{
		Splash: SplashConfig{
			Enabled: DefaultSplashEnabled,
			Delay:   DefaultSplashDelay,
		},
		Theme: ThemeConfig{
			PrimitiveBackgroundColor:    "#000a1a",
			ContrastBackgroundColor:     "#001a33",
			MoreContrastBackgroundColor: "#003366",
			BorderColor:                 "#0088ff",
			TitleColor:                  "#00ffff",
			GraphicsColor:               "#00ffaa",
			PrimaryTextColor:            "#cceeff",
			SecondaryTextColor:          "#6699ff",
			TertiaryTextColor:           "#ffaa00",
			InverseTextColor:            "#000a1a",
			ContrastSecondaryTextColor:  "#88ddff",
		},
	}
}
