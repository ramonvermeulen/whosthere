package config

import (
	"fmt"
	"os"
	"strings"
)

const envPrefix = "WHOSTHERE__"

func ApplyEnv(cfg *Config) error {
	if cfg == nil {
		return ErrConfigNil
	}

	settings := settingsByYAMLKey()

	for _, kv := range os.Environ() {
		k, v, ok := strings.Cut(kv, "=")
		if !ok || !strings.HasPrefix(k, envPrefix) {
			continue
		}

		rest := strings.TrimPrefix(k, envPrefix)
		if rest == "" {
			continue
		}

		yamlKey := envVarToYAMLKey(rest)
		setting, exists := settings[yamlKey]
		if !exists || setting.Set == nil {
			continue
		}

		if err := setting.Set(cfg, v); err != nil {
			return fmt.Errorf("env %s: %w", k, err)
		}
	}

	return nil
}

func envVarToYAMLKey(s string) string {
	parts := strings.Split(s, "__")
	for i := range parts {
		parts[i] = strings.ToLower(parts[i])
	}
	return strings.Join(parts, ".")
}

func SetByYAMLKey(cfg *Config, yamlKey, value string) error {
	if cfg == nil {
		return ErrConfigNil
	}

	settings := settingsByYAMLKey()
	setting, exists := settings[yamlKey]
	if !exists || setting.Set == nil {
		return nil
	}

	if err := setting.Set(cfg, value); err != nil {
		return fmt.Errorf("setting %s: %w", yamlKey, err)
	}
	return nil
}
