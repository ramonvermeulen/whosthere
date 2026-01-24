package utils

import (
	"fmt"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
)

// ColorToHexTag converts a tcell.Color to a tview dynamic color hex tag.
func ColorToHexTag(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// SortedKeys is a helper to return asc sorted map keys.
func SortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func FmtDuration(d time.Duration) string {
	if d < time.Second {
		return "<1s"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d/time.Second))
	}
	return fmt.Sprintf("%dm", int(d/time.Minute))
}

func Truncate(s string, maxLen int) string {
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return s[:maxLen]
	}
	return s[:maxLen-1] + "…"
}

// SanitizeString returns the string if it contains only printable characters.
// Otherwise, it returns a hex representation.
func SanitizeString(s string) string {
	for _, r := range s {
		// Check for non-printable ASCII characters.
		if r < 32 || r > 126 {
			return fmt.Sprintf("0x%x", []byte(s))
		}
	}
	return s
}
