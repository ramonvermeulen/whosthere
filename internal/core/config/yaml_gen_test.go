package config

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestGenerateDefaultYAML(t *testing.T) {
	yaml := GenerateDefaultYAML()

	if yaml == "" {
		t.Fatal("generated YAML is empty")
	}

	mustContain := []string{
		"# Uncomment the next line to configure a specific network interface",
		"# network_interface: eth0",
		"scan_interval: 20s",
		"scan_timeout: 10s",
		"scanners:",
		"mdns:",
		"enabled: true",
		"ssdp:",
		"arp:",
		"sweeper:",
		"interval: 5m",
		"port_scanner:",
		"timeout: 5s",
		"tcp: [",
		"splash:",
		"delay: 1s",
		"theme:",
		"name: default",
		"# primitive_background_color:",
	}

	for _, s := range mustContain {
		if !strings.Contains(yaml, s) {
			t.Errorf("generated YAML missing expected content: %q", s)
		}
	}

	lines := strings.Split(yaml, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.Contains(line, ":") && !strings.HasSuffix(strings.TrimSpace(line), ":") {
			if strings.Count(line, "  ") > 0 && !strings.HasPrefix(line, strings.Repeat("  ", strings.Count(line, "  ")/2)) {
				continue
			}
		}
		_ = i
	}
}

func TestGenerateDefaultYAMLMatchesREADME(t *testing.T) {
	readmePath := "../../../README.md"
	readmeContent, err := os.ReadFile(readmePath)
	if err != nil {
		t.Skipf("README.md not found at %s: %v", readmePath, err)
	}

	re := regexp.MustCompile("(?s)Example of the default configuration file:\n\n```yaml\n(.*?)```")
	matches := re.FindSubmatch(readmeContent)
	if matches == nil {
		t.Fatal("could not find YAML config block in README.md")
	}

	readmeYAML := strings.TrimSpace(string(matches[1]))
	generatedYAML := strings.TrimSpace(GenerateDefaultYAML())

	if readmeYAML != generatedYAML {
		t.Errorf("generated YAML does not match README.md\n\n--- README.md ---\n%s\n\n--- Generated ---\n%s", readmeYAML, generatedYAML)
	}
}

func TestYAMLSettingsHaveDefaults(t *testing.T) {
	settings := getYAMLSettings()
	defaults := DefaultConfig()

	for _, s := range settings {
		if s.Doc.CommentedOut {
			if s.Doc.ExampleValue == "" {
				t.Errorf("commented out setting %q should have ExampleValue", s.YAMLKey)
			}
			continue
		}
		if s.Get == nil {
			t.Errorf("setting %q should have a Get function", s.YAMLKey)
			continue
		}
		val := s.Get(defaults)
		if val == nil || val == "" {
			t.Errorf("setting %q should have a default value in DefaultConfig()", s.YAMLKey)
		}
	}
}
