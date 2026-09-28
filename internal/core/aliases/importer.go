package aliases

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ramonvermeulen/whosthere/internal/core/devicemeta"
)

// ErrFileNotFound is returned by Import when the aliases file does not exist.
var ErrFileNotFound = errors.New("aliases file not found")

// stateKeyPrefix namespaces the stored content hash per absolute file path.
const stateKeyPrefix = "aliases_import_hash:"

// Importer applies an aliases file to a metadata store.
type Importer struct {
	store devicemeta.Store
}

// NewImporter returns an Importer backed by store.
func NewImporter(store devicemeta.Store) *Importer {
	return &Importer{store: store}
}

// Report summarizes the outcome of an import.
type Report struct {
	Path     string
	Hash     string
	Imported int
	Cleared  int
	Skipped  int
	// Invalid holds MAC addresses that could not be normalized.
	Invalid []string
	// Duplicates holds normalized MACs that appeared more than once; the last
	// occurrence wins.
	Duplicates []string
}

// Total returns the number of aliases present in the file after de-duplication.
func (r *Report) Total() int {
	return r.Imported + r.Cleared
}

// Import reads the aliases file at path and upserts its entries into the store.
// Entries with invalid MAC addresses are skipped and reported rather than
// failing the import. MAC addresses absent from the file are left untouched.
func (i *Importer) Import(path string) (Report, error) {
	report := Report{Path: path}

	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return report, fmt.Errorf("%w: %s", ErrFileNotFound, path)
		}
		return report, fmt.Errorf("read aliases file: %w", err)
	}

	sum := sha256.Sum256(raw)
	report.Hash = hex.EncodeToString(sum[:])

	file, err := Parse(raw)
	if err != nil {
		return report, err
	}

	normalized, invalid, duplicates := normalizeEntries(file.Aliases)
	report.Invalid = invalid
	report.Duplicates = duplicates

	for _, entry := range normalized {
		if err := i.store.SetAlias(entry.MAC, entry.Alias); err != nil {
			return report, fmt.Errorf("import alias for %s: %w", entry.MAC, err)
		}

		if strings.TrimSpace(entry.Alias) == "" {
			report.Cleared++
			continue
		}
		report.Imported++
	}

	return report, nil
}

// ImportIfChanged imports the file only when its content differs from the last
// successful import recorded in the store. An unchanged file (or a missing
// file) is a no-op. The stored hash is only updated after a successful import.
//
// The imported result reports whether an import actually ran.
func (i *Importer) ImportIfChanged(path string) (report Report, imported bool, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Report{Path: path}, false, nil
		}
		return Report{Path: path}, false, fmt.Errorf("read aliases file: %w", err)
	}

	sum := sha256.Sum256(raw)
	currentHash := hex.EncodeToString(sum[:])

	previousHash, _, err := i.store.GetState(stateKey(path))
	if err != nil {
		return Report{Path: path}, false, fmt.Errorf("read import state: %w", err)
	}
	if previousHash == currentHash {
		return Report{Path: path, Hash: currentHash}, false, nil
	}

	report, err = i.Import(path)
	if err != nil {
		return report, false, err
	}

	if err := i.store.SetState(stateKey(path), currentHash); err != nil {
		return report, false, fmt.Errorf("record import state: %w", err)
	}

	return report, true, nil
}

// stateKey builds the per-path state key used to remember the last imported
// content hash. The path is embedded so that moving or renaming the file
// triggers a re-import.
func stateKey(path string) string {
	return stateKeyPrefix + absPath(path)
}

func absPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}

// normalizeEntries converts raw file entries into normalized MAC/alias pairs.
// Duplicate normalized MACs keep the last occurrence. Invalid MACs are returned
// in invalid, duplicates in duplicates (sorted for deterministic output).
func normalizeEntries(entries map[string]string) (normalizedEntries []Entry, invalid, duplicates []string) {
	normalized := make(map[string]string, len(entries))
	seen := make(map[string]int, len(entries))

	for mac, alias := range entries {
		normalizedMAC, ok := devicemeta.NormalizeMAC(mac)
		if !ok {
			invalid = append(invalid, mac)
			continue
		}

		seen[normalizedMAC]++
		// Later assignments win; map iteration order is non-deterministic but
		// YAML decoding preserves a single value per literal key. Duplicates
		// here only arise from distinct spellings of the same MAC.
		normalized[normalizedMAC] = strings.TrimSpace(alias)
	}

	for mac, count := range seen {
		if count > 1 {
			duplicates = append(duplicates, mac)
		}
	}

	sort.Strings(invalid)
	sort.Strings(duplicates)

	keys := make([]string, 0, len(normalized))
	for mac := range normalized {
		keys = append(keys, mac)
	}
	sort.Strings(keys)

	normalizedEntries = make([]Entry, 0, len(normalized))
	for _, mac := range keys {
		normalizedEntries = append(normalizedEntries, Entry{MAC: mac, Alias: normalized[mac]})
	}

	return normalizedEntries, invalid, duplicates
}
