package devicemeta

import (
	"fmt"
	"path/filepath"

	"github.com/ramonvermeulen/whosthere/internal/core/paths"
)

const defaultDBFileName = "devices.db"

// Store provides persistent local device metadata keyed by normalized MAC address.
type Store interface {
	Get(mac string) (Record, bool, error)
	SetAlias(mac, alias string) error
	ClearAlias(mac string) error
	ResetAliases() error
	// All returns every stored record keyed by normalized MAC address. Records
	// without an alias are omitted.
	All() (map[string]Record, error)
	// GetState reads an opaque key/value pair from the metadata bucket. The
	// second return value reports whether the key was present.
	GetState(key string) (string, bool, error)
	// SetState writes an opaque key/value pair to the metadata bucket. An empty
	// value deletes the key.
	SetState(key, value string) error
	Close() error
}

// OpenDefault opens the metadata store in the application's state directory.
func OpenDefault() (Store, error) {
	dir, err := paths.StateDir()
	if err != nil {
		return nil, fmt.Errorf("resolve state dir: %w", err)
	}

	return Open(filepath.Join(dir, defaultDBFileName))
}
