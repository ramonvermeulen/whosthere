// Package aliases provides a user-editable YAML file format for importing and
// exporting device aliases into the persistent device metadata store.
//
// The file is a seed/patch, not a mirror: importing upserts the MAC addresses
// present in the file and never deletes records for MAC addresses that are
// absent. See docs and the README for the full semantics.
package aliases

import (
	"errors"
	"fmt"
	"sort"

	"github.com/goccy/go-yaml"
)

// CurrentVersion is the version written to and accepted by the aliases file.
const CurrentVersion = 1

// ErrUnsupportedVersion is returned when a file declares a version this build
// cannot parse.
var ErrUnsupportedVersion = errors.New("unsupported aliases file version")

// Entry is a single alias definition as parsed from the file. MAC is the raw
// value from the file; normalization happens during import.
type Entry struct {
	MAC   string
	Alias string
}

// File is the on-disk representation of an aliases file.
type File struct {
	Version int               `yaml:"version"`
	Aliases map[string]string `yaml:"aliases"`
}

// Marshal serializes a File into the versioned YAML format with stable key
// ordering so exports are diff-friendly.
func Marshal(f *File) ([]byte, error) {
	if f == nil {
		return nil, errors.New("aliases file is nil")
	}

	version := f.Version
	if version == 0 {
		version = CurrentVersion
	}

	ordered := orderedAliases{
		Version: version,
		Aliases: yaml.MapSlice{},
	}

	keys := make([]string, 0, len(f.Aliases))
	for mac := range f.Aliases {
		keys = append(keys, mac)
	}
	sort.Strings(keys)

	for _, mac := range keys {
		ordered.Aliases = append(ordered.Aliases, yaml.MapItem{Key: mac, Value: f.Aliases[mac]})
	}

	raw, err := yaml.Marshal(ordered)
	if err != nil {
		return nil, fmt.Errorf("marshal aliases file: %w", err)
	}

	return raw, nil
}

// Parse decodes the raw bytes of an aliases file. A missing or zero version is
// treated as the current version; unknown future versions are rejected.
func Parse(raw []byte) (*File, error) {
	var parsed File
	if err := yaml.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse aliases file: %w", err)
	}

	if parsed.Version != 0 && parsed.Version != CurrentVersion {
		return nil, fmt.Errorf("%w: got %d, want %d", ErrUnsupportedVersion, parsed.Version, CurrentVersion)
	}

	if parsed.Version == 0 {
		parsed.Version = CurrentVersion
	}

	if parsed.Aliases == nil {
		parsed.Aliases = map[string]string{}
	}

	return &parsed, nil
}

// orderedAliases controls the YAML serialization order of the aliases map.
type orderedAliases struct {
	Version int
	Aliases yaml.MapSlice
}
