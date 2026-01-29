package oui

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/paths"
	"go.uber.org/zap"
)

// see https://pkg.go.dev/embed
// this will embed the oui.csv file while compiling the binary
//
//go:embed oui.csv
var embeddedOUIDB []byte

const (
	updateURL       = "https://standards-oui.ieee.org/oui/oui.csv"
	maxAge          = 30 * 24 * time.Hour
	clientTimeout   = 30 * time.Second
	userAgentHeader = "whosthere/1.0 (+https://github.com/ramonvermeulen/whosthere)"
	acceptHeader    = "text/csv,application/vnd.ms-excel;q=0.9,*/*;q=0.8"
)

type Registry struct {
	mu        sync.RWMutex
	prefixMap map[string]string
	loadedAt  time.Time
	path      string
}

// Init initializes the OUI database.
func Init(ctx context.Context) (*Registry, error) {
	logger := zap.L()

	stateDir, err := paths.StateDir()
	if err != nil {
		return nil, fmt.Errorf("creating state dir: %w", err)
	}

	path := filepath.Join(stateDir, "oui.csv")
	reg := &Registry{
		prefixMap: make(map[string]string),
		path:      path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		logger.Debug("OUI: state file not found, using embedded CSV", zap.String("path", path), zap.Error(err))
		// fall back to embedded data
		data = embeddedOUIDB
		if writeErr := os.WriteFile(path, data, 0o644); writeErr != nil {
			logger.Debug("OUI: failed to write embedded oui.csv to state dir", zap.Error(writeErr))
		}
		reg.loadedAt = time.Now()
	} else {
		logger.Debug("OUI: loaded CSV from state dir", zap.String("path", path), zap.Int("bytes", len(data)))
		if info, err := os.Stat(path); err == nil {
			reg.loadedAt = info.ModTime()
		} else {
			return nil, fmt.Errorf("stat oui.csv: %w", err)
		}
	}

	if err := reg.loadFromBytes(data); err != nil {
		logger.Error("OUI: failed to parse CSV", zap.Error(err))
		return nil, err
	}

	reg.mu.RLock()
	entryCount := len(reg.prefixMap)
	loadedAt := reg.loadedAt
	reg.mu.RUnlock()
	logger.Debug("OUI: registry initialized", zap.Int("entries", entryCount), zap.String("path", path), zap.Time("loaded_at", loadedAt))

	age := time.Since(loadedAt)
	if age > maxAge {
		logger.Info("OUI: data older than maxAge, triggering one-time refresh", zap.Duration("age", age), zap.Duration("max_age", maxAge))
		go func() {
			if err := reg.Refresh(ctx); err != nil {
				logger.Debug("OUI: initial one-time refresh failed", zap.Error(err))
			}
		}()
	} else {
		logger.Debug("OUI: data is fresh enough, skipping initial refresh", zap.Duration("age", age), zap.Duration("max_age", maxAge))
	}

	return reg, nil
}

// Refresh downloads the latest CSV and replaces in-memory and on-disk data.
func (reg *Registry) Refresh(ctx context.Context) error {
	logger := zap.L()

	ctx, cancel := context.WithTimeout(ctx, clientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, updateURL, http.NoBody)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgentHeader)
	req.Header.Set("Accept", acceptHeader)

	client := &http.Client{Timeout: clientTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("non-2xx response fetching OUI Registry: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	m, err := parseCSVBytes(data)
	if err != nil {
		return err
	}

	reg.mu.Lock()
	defer reg.mu.Unlock()
	reg.prefixMap = m
	reg.loadedAt = time.Now()

	if err := os.WriteFile(reg.path, data, 0o644); err != nil {
		logger.Debug("failed to persist refreshed oui.csv", zap.Error(err))
	}

	return nil
}

func (reg *Registry) loadFromBytes(b []byte) error {
	m, err := parseCSVBytes(b)
	if err != nil {
		return err
	}
	reg.mu.Lock()
	defer reg.mu.Unlock()
	reg.prefixMap = m
	if reg.loadedAt.IsZero() {
		reg.loadedAt = time.Now()
	}
	return nil
}

func parseCSVBytes(b []byte) (map[string]string, error) {
	r := csv.NewReader(bufio.NewReader(strings.NewReader(string(b))))
	r.FieldsPerRecord = -1

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}
	if len(header) < 3 {
		return nil, fmt.Errorf("unexpected header format in OUI CSV")
	}

	const (
		macCol = 1
		orgCol = 2
	)

	m := make(map[string]string)

	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if macCol >= len(rec) || orgCol >= len(rec) {
			continue
		}
		macField := strings.TrimSpace(rec[macCol])
		org := strings.TrimSpace(rec[orgCol])
		if macField == "" || org == "" {
			continue
		}

		prefix := normalizeMACPrefix(macField)
		if prefix == "" {
			continue
		}
		if _, exists := m[prefix]; !exists {
			m[prefix] = org
		}
	}

	return m, nil
}

func normalizeMACPrefix(s string) string {
	s = strings.ToUpper(s)
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, ".", "")
	// Some IEEE CSVs contain only the prefix already, others may contain full MAC
	if len(s) < 6 {
		return ""
	}
	return s[:6]
}

// Lookup returns the organization for a given MAC address string.
// It uses a 24-bit prefix (first 3 bytes).
func (reg *Registry) Lookup(mac string) (string, bool) {
	logger := zap.L()

	prefix := normalizeMACPrefix(mac)
	if prefix == "" {
		logger.Debug("OUI: lookup skipped, empty/invalid MAC", zap.String("mac", mac))
		return "", false
	}
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	org, ok := reg.prefixMap[prefix]
	if !ok {
		logger.Debug("OUI: no entry for prefix", zap.String("mac", mac), zap.String("prefix", prefix))
		return "", false
	}
	logger.Debug("OUI: lookup hit", zap.String("mac", mac), zap.String("prefix", prefix), zap.String("org", org))
	return org, true
}
