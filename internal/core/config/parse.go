package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseBool(s string) (bool, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "1", "true", "t", "yes", "y", "on":
		return true, nil
	case "0", "false", "f", "no", "n", "off":
		return false, nil
	default:
		return strconv.ParseBool(s)
	}
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	d, err := time.ParseDuration(s)
	if err == nil {
		return d, nil
	}

	if seconds, numErr := strconv.ParseInt(s, 10, 64); numErr == nil {
		return time.Duration(seconds) * time.Second, nil
	}

	return 0, err
}

func parseInt(s string) (int, error) {
	i, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return int(i), err
}

func parseIntSlice(s string) ([]int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return []int{}, nil
	}

	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		i, err := parseInt(p)
		if err != nil {
			return nil, fmt.Errorf("invalid integer %q: %w", p, err)
		}
		result = append(result, i)
	}
	return result, nil
}
