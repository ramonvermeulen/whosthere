package config

import "fmt"

func validateAndNormalizeForMode(cfg *Config, mode RunMode) error {
	if cfg == nil {
		return ErrConfigNil
	}

	switch mode {
	case ModeApp:
		return cfg.validateAndNormalize()
	case ModeCLI:
		return cfg.normalizeBasics()
	default:
		return fmt.Errorf("unknown mode %d", mode)
	}
}
