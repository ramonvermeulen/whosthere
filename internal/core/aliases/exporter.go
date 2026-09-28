package aliases

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ramonvermeulen/whosthere/internal/core/devicemeta"
)

// Export reads every alias from the store and returns them in the versioned
// file format. Ordering is stable so exports are diff-friendly.
func Export(store devicemeta.Store) (*File, error) {
	records, err := store.All()
	if err != nil {
		return nil, fmt.Errorf("read aliases from store: %w", err)
	}

	file := &File{
		Version: CurrentVersion,
		Aliases: make(map[string]string, len(records)),
	}

	for mac, record := range records {
		if record.Alias == "" {
			continue
		}
		file.Aliases[mac] = record.Alias
	}

	return file, nil
}

// ExportToFile writes the store's aliases to path, creating parent directories
// as needed.
func ExportToFile(store devicemeta.Store, path string) error {
	file, err := Export(store)
	if err != nil {
		return err
	}

	raw, err := Marshal(file)
	if err != nil {
		return err
	}

	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create aliases dir: %w", err)
		}
	}

	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return fmt.Errorf("write aliases file: %w", err)
	}

	return nil
}
