package config

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateDefaultYAML(t *testing.T) {
	yaml := GenerateDefaultYAML()

	require.NotEmpty(t, yaml, "generated YAML is empty")

	mustContain := []string{
		"# Uncomment the next line to configure a specific network interface",
		"# network_interface: eth0",
		"# target_subnets: [\"10.0.0.0/24\", \"10.0.1.0/24\"]",
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
		assert.Contains(t, yaml, s, "generated YAML missing expected content")
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

	re := regexp.MustCompile(`(?s)\*\*Example configuration:\*\*[\r\n]+` + "```yaml[\r\n]+" + `(.*?)` + "```" + ``)
	matches := re.FindSubmatch(readmeContent)
	require.NotNil(t, matches, "could not find YAML config block in README.md")

	readmeYAML := strings.TrimSpace(string(matches[1]))
	readmeYAML = strings.ReplaceAll(readmeYAML, "\r\n", "\n")
	generatedYAML := strings.TrimSpace(GenerateDefaultYAML())

	assert.Equal(t, readmeYAML, generatedYAML, "generated YAML does not match README.md")
}

func TestYAMLSettingsHaveDefaults(t *testing.T) {
	settings := getYAMLSettings()
	defaults := DefaultConfig()

	for _, s := range settings {
		if s.Doc.CommentedOut {
			assert.NotEmpty(t, s.Doc.ExampleValue, "commented out setting %q should have ExampleValue", s.YAMLKey)
			continue
		}
		assert.NotNil(t, s.Get, "setting %q should have a Get function", s.YAMLKey)
		if s.Get == nil {
			continue
		}
		val := s.Get(defaults)
		assert.NotNil(t, val, "setting %q should have a default value in DefaultConfig()", s.YAMLKey)
	}
}
