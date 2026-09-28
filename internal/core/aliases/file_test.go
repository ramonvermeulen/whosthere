package aliases

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse_ValidFile(t *testing.T) {
	t.Parallel()

	raw := []byte(`version: 1
aliases:
  aa:bb:cc:dd:ee:ff: "Living Room TV"
  de:ad:be:ef:00:01: "Office Printer"
`)

	file, err := Parse(raw)
	require.NoError(t, err, "Parse()")
	require.Equal(t, 1, file.Version, "version")
	require.Equal(t, "Living Room TV", file.Aliases["aa:bb:cc:dd:ee:ff"])
	require.Equal(t, "Office Printer", file.Aliases["de:ad:be:ef:00:01"])
}

func TestParse_MissingVersionDefaults(t *testing.T) {
	t.Parallel()

	file, err := Parse([]byte("aliases:\n  aa:bb:cc:dd:ee:ff: TV\n"))
	require.NoError(t, err, "Parse()")
	require.Equal(t, CurrentVersion, file.Version, "defaulted version")
}

func TestParse_RejectsFutureVersion(t *testing.T) {
	t.Parallel()

	_, err := Parse([]byte("version: 999\naliases: {}\n"))
	require.Error(t, err, "Parse() should reject unknown version")
	require.True(t, errors.Is(err, ErrUnsupportedVersion), "expected ErrUnsupportedVersion, got %v", err)
}

func TestParse_EmptyFileYieldsEmptyMap(t *testing.T) {
	t.Parallel()

	file, err := Parse(nil)
	require.NoError(t, err, "Parse()")
	require.Equal(t, CurrentVersion, file.Version)
	require.NotNil(t, file.Aliases, "aliases map should be initialized")
	require.Empty(t, file.Aliases)
}

func TestParse_InvalidYAML(t *testing.T) {
	t.Parallel()

	_, err := Parse([]byte("aliases: [this, is, not, a, map\n"))
	require.Error(t, err, "Parse() should fail on malformed YAML")
}

func TestMarshal_SortsKeysAndRoundTrips(t *testing.T) {
	t.Parallel()

	file := &File{
		Version: CurrentVersion,
		Aliases: map[string]string{
			"ff:ff:ff:ff:ff:ff": "Zebra",
			"aa:bb:cc:dd:ee:ff": "Apple",
		},
	}

	raw, err := Marshal(file)
	require.NoError(t, err, "Marshal()")

	parsed, err := Parse(raw)
	require.NoError(t, err, "Parse(Marshal())")
	require.Equal(t, file.Aliases, parsed.Aliases, "round-trip aliases")

	require.Less(t, indexOf(raw, "aa:bb:cc:dd:ee:ff"), indexOf(raw, "ff:ff:ff:ff:ff:ff"),
		"keys should be sorted for stable diffs")
}

func indexOf(haystack []byte, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if string(haystack[i:i+len(needle)]) == needle {
			return i
		}
	}
	return -1
}
