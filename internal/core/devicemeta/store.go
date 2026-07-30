package devicemeta

import (
	"fmt"
	"path/filepath"

	"github.com/ramonvermeulen/whosthere/internal/core/paths"
)

const defaultDBFileName = "devices.db"

// Store provides persistent local device metadata keyed by normalized MAC address.
type Store interface {
	Get(scopeID, mac string) (Record, bool, error)
	Upsert(record Record) error
	Delete(scopeID, mac string) error
	ForEach(func(Record) error) error
	UpsertScope(scope ScopeRecord) error
	GetScope(scopeID string) (ScopeRecord, bool, error)
	DeleteScopeAndDevices(scopeID string) error
	DeleteAllScopesAndDevices() error
	SetAlias(scopeID, mac, alias string) error
	ClearAlias(scopeID, mac string) error
	ResetAliases() error
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
