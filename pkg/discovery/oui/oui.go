package oui

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
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

// Registry provides MAC address OUI (Organizationally Unique Identifier) lookups
// to resolve manufacturer names from MAC addresses. The first 3 bytes of a MAC
// address identify the manufacturer.
//
// The registry embeds IEEE OUI data and can optionally cache updates to disk.
// It automatically refreshes data older than 30 days from the IEEE website.
//
// Thread-safe for concurrent lookups.
type Registry struct {
	mu        sync.RWMutex
	prefixMap map[string]string
	loadedAt  time.Time
	path      string
	logger    *slog.Logger
}

// Option configures a Registry during construction.
type Option func(*Registry)

// WithCacheDir sets the directory for caching OUI data.
// The registry saves downloaded IEEE data to oui.csv in this directory,
// reducing network requests on subsequent runs. If empty, no caching occurs.
func WithCacheDir(dir string) Option {
	return func(r *Registry) {
		if dir == "" {
			return
		}
		r.path = filepath.Join(dir, "oui.csv")
	}
}

// New creates an OUI registry with embedded IEEE data.
// If WithCacheDir is used and cached data exists, it's loaded instead.
// Automatically triggers a background refresh if data is older than 30 days.
//
// The registry is immediately usable even if the background refresh fails.
func New(ctx context.Context, opts ...Option) (*Registry, error) {
	reg := &Registry{
		prefixMap: make(map[string]string),
		logger:    slog.Default(),
	}
	for _, opt := range opts {
		opt(reg)
	}

	data := embeddedOUIDB
	if reg.path != "" {
		b, err := os.ReadFile(reg.path)
		switch err {
		case nil:
			data = b
			if info, statErr := os.Stat(reg.path); statErr == nil {
				reg.loadedAt = info.ModTime()
			}
			reg.logger.Debug("OUI: loaded CSV from cache", "path", reg.path, "bytes", len(data))
		default:
			reg.logger.Debug("OUI: cache not available, using embedded CSV", "path", reg.path, "err", err)
			if mkErr := os.MkdirAll(filepath.Dir(reg.path), 0o755); mkErr == nil {
				if writeErr := os.WriteFile(reg.path, data, 0o644); writeErr != nil {
					reg.logger.Debug("OUI: failed to write embedded oui.csv to cache", "err", writeErr)
				}
			}
			reg.loadedAt = time.Now()
		}
	} else {
		reg.loadedAt = time.Now()
	}

	if err := reg.loadFromBytes(data); err != nil {
		reg.logger.Error("OUI: failed to parse CSV", "err", err)
		return nil, err
	}

	reg.mu.RLock()
	entryCount := len(reg.prefixMap)
	loadedAt := reg.loadedAt
	path := reg.path
	reg.mu.RUnlock()
	reg.logger.Debug("OUI: registry initialized", "entries", entryCount, "path", path, "loaded_at", loadedAt)

	age := time.Since(loadedAt)
	// todo(ramon): make this configurable, e.g. the time period as well as whether to trigger a background refresh at all
	if reg.path != "" && age > maxAge {
		reg.logger.Info("OUI: data older than maxAge, triggering one-time refresh", "age", age, "max_age", maxAge)
		go func() {
			if err := reg.Refresh(ctx); err != nil {
				reg.logger.Debug("OUI: initial one-time refresh failed", "err", err)
			}
		}()
	} else {
		reg.logger.Debug("OUI: data is fresh enough, skipping initial refresh", "age", age, "max_age", maxAge)
	}

	return reg, nil
}

// Refresh downloads the latest OUI data from the IEEE website and updates the registry.
// If a cache directory was configured, the downloaded data is saved to disk.
//
// Called automatically in the background when data is older than 30 days.
// You can also call it manually to force an update.
//
// Returns an error if the download fails or the data is malformed.
// The registry continues using existing data if the refresh fails.
func (reg *Registry) Refresh(ctx context.Context) error {
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
	defer func() { _ = resp.Body.Close() }()

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
	reg.prefixMap = m
	reg.loadedAt = time.Now()
	path := reg.path
	reg.mu.Unlock()

	if path != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err == nil {
			if err := os.WriteFile(path, data, 0o644); err != nil {
				reg.logger.Debug("OUI: failed to persist refreshed oui.csv", "err", err)
			}
		}
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
	if len(s) < 6 {
		return ""
	}
	return s[:6]
}

// Lookup returns the manufacturer name for a MAC address.
// Accepts various MAC formats: "AA:BB:CC:DD:EE:FF", "AA-BB-CC-DD-EE-FF", "AABBCCDDEEFF".
// Case-insensitive. Only the first 3 bytes (OUI prefix) are used.
//
// Returns the manufacturer name and true if found, empty string and false otherwise.
func (reg *Registry) Lookup(mac string) (string, bool) {
	prefix := normalizeMACPrefix(mac)
	if prefix == "" {
		reg.logger.Debug("OUI: lookup skipped, empty/invalid MAC", "mac", mac)
		return "", false
	}
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	org, ok := reg.prefixMap[prefix]
	if !ok {
		reg.logger.Debug("OUI: no entry for prefix", "mac", mac, "prefix", prefix)
		return "", false
	}
	reg.logger.Debug("OUI: lookup hit", "mac", mac, "prefix", prefix, "org", org)
	return org, true
}
