package config

import (
	"fmt"
	"strings"
	"time"
)

func GenerateDefaultYAML() string {
	var sb strings.Builder
	settings := getYAMLSettings()
	defaults := DefaultConfig()

	var prevPath []string
	for i, s := range settings {
		parts := strings.Split(s.YAMLKey, ".")
		depth := len(parts) - 1
		indent := strings.Repeat("  ", depth)

		if i > 0 && parts[0] != prevPath[0] {
			sb.WriteString("\n")
		}

		if s.Doc.BlankLineBefore {
			sb.WriteString("\n")
		}

		writeNesting(&sb, parts, prevPath)

		if s.Doc.Comment != "" {
			for _, line := range strings.Split(s.Doc.Comment, "\n") {
				sb.WriteString(indent)
				sb.WriteString("# ")
				sb.WriteString(line)
				sb.WriteString("\n")
			}
		}

		writeSettingLine(&sb, &s, parts, indent, defaults)
		if s.Doc.BlankLineAfter {
			sb.WriteString("\n")
		}
		prevPath = parts
	}

	return sb.String()
}

func writeNesting(sb *strings.Builder, parts, prevPath []string) {
	for depth := 0; depth < len(parts)-1; depth++ {
		if depth >= len(prevPath) || prevPath[depth] != parts[depth] {
			indent := strings.Repeat("  ", depth)
			sb.WriteString(indent)
			sb.WriteString(parts[depth])
			sb.WriteString(":\n")
		}
	}
}

func writeSettingLine(sb *strings.Builder, s *GlobalSetting, parts []string, indent string, defaults *Config) {
	key := parts[len(parts)-1]
	value := getDisplayValue(s, defaults)

	if s.Doc.CommentedOut {
		sb.WriteString(indent)
		sb.WriteString("# ")
		sb.WriteString(key)
		sb.WriteString(": ")
		sb.WriteString(value)
		sb.WriteString("\n")
	} else {
		sb.WriteString(indent)
		sb.WriteString(key)
		sb.WriteString(": ")
		sb.WriteString(value)
		sb.WriteString("\n")
	}
}

func getDisplayValue(s *GlobalSetting, defaults *Config) string {
	if s.Doc.ExampleValue != "" {
		return s.Doc.ExampleValue
	}
	if s.Get != nil {
		return formatValue(s.Get(defaults))
	}
	return ""
}

func getYAMLSettings() []GlobalSetting {
	var yamlSettings []GlobalSetting
	for _, s := range GlobalSettings() {
		if s.YAMLKey != "" && s.hasSource(SourceYAML) {
			yamlSettings = append(yamlSettings, s)
		}
	}
	return yamlSettings
}

func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case bool:
		return fmt.Sprintf("%t", val)
	case int:
		return fmt.Sprintf("%d", val)
	case []int:
		parts := make([]string, len(val))
		for i, port := range val {
			parts[i] = fmt.Sprintf("%d", port)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case time.Duration:
		return formatDuration(val)
	case fmt.Stringer:
		return val.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatDuration(d time.Duration) string {
	if d >= time.Hour && d%time.Hour == 0 {
		return fmt.Sprintf("%dh", d/time.Hour)
	}
	if d >= time.Minute && d%time.Minute == 0 {
		return fmt.Sprintf("%dm", d/time.Minute)
	}
	if d >= time.Second && d%time.Second == 0 {
		return fmt.Sprintf("%ds", d/time.Second)
	}
	if d >= time.Millisecond && d%time.Millisecond == 0 {
		return fmt.Sprintf("%dms", d/time.Millisecond)
	}
	return d.String()
}
