package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

func LoadForMode(mode RunMode, flags *Flags) (*Config, error) {
	cfg := DefaultConfig()

	// if the RunMode is ModeCLI, we skip loading the config file
	// and only use env vars + flag overrides
	if mode == ModeApp {
		pathOverride := ""
		if flags != nil {
			pathOverride = flags.ConfigFile
		}

		resolvedPath, err := resolveConfigPath(pathOverride)
		if err != nil {
			return nil, err
		}

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
	}

	if err := ApplyEnv(cfg); err != nil {
		return nil, err
	}

	if flags != nil {
		for k, v := range flags.Overrides {
			if err := SetByYAMLKey(cfg, k, v); err != nil {
				return nil, err
			}
		}
	}

	if err := validateAndNormalizeForMode(cfg, mode); err != nil {
		return cfg, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func LoadMerged(flags *Flags) (*Config, error) {
	return LoadForMode(ModeApp, flags)
}
