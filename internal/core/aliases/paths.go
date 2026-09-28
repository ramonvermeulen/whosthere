package aliases

import (
	"fmt"
	"path/filepath"

	"github.com/ramonvermeulen/whosthere/internal/core/paths"
)

const defaultFileName = "aliases.yaml"

// DefaultPath returns the default aliases file location in the app config
// directory. It does not create the file or its parent directory.
func DefaultPath() (string, error) {
	dir, err := paths.ConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}

	return filepath.Join(dir, defaultFileName), nil
}

// ResolvePath returns the effective aliases file path. An explicit override
// (CLI argument or the aliases.file config key, which is also settable via the
// standard WHOSTHERE__ALIASES__FILE environment variable) takes precedence,
// otherwise the default location is used.
func ResolvePath(override string) (string, error) {
	if override != "" {
		return override, nil
	}

	return DefaultPath()
}
