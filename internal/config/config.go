package config

const (
	DefaultSplashEnabled = true
	DefaultSplashDelay   = float32(1.0)
)

// Config captures runtime configuration values loaded from the YAML config file.
type Config struct {
	Splash SplashConfig `yaml:"splash"`
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
	}
}
