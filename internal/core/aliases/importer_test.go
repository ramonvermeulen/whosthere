package aliases

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ramonvermeulen/whosthere/internal/core/devicemeta"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) devicemeta.Store {
	t.Helper()

	store, err := devicemeta.Open(filepath.Join(t.TempDir(), "devices.db"))
	require.NoError(t, err, "Open()")
	t.Cleanup(func() {
		_ = store.Close()
	})

	return store
}

func writeAliasesFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "aliases.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644), "write aliases file")

	return path
}

func TestImport_HappyPath(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	path := writeAliasesFile(t, `version: 1
aliases:
  AA-BB-CC-DD-EE-FF: "Living Room TV"
  de:ad:be:ef:00:01: "Office Printer"
`)

	report, err := NewImporter(store).Import(path)
	require.NoError(t, err, "Import()")
	require.Equal(t, 2, report.Imported, "imported count")
	require.Equal(t, 0, report.Cleared, "cleared count")
	require.Empty(t, report.Invalid, "invalid MACs")
	require.NotEmpty(t, report.Hash, "hash")

	record, found, err := store.Get("aa:bb:cc:dd:ee:ff")
	require.NoError(t, err, "Get()")
	require.True(t, found, "record should exist")
	require.Equal(t, "Living Room TV", record.Alias)
}

func TestImport_RawMACSpellingsNormalizeToSameKey(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	path := writeAliasesFile(t, `aliases:
  "AA-BB-CC-DD-EE-FF": "Dashed"
  "aabbccddeeff": "Compact"
`)

	report, err := NewImporter(store).Import(path)
	require.NoError(t, err, "Import()")
	require.Equal(t, 1, report.Imported, "both spellings collapse to one MAC")
	require.Equal(t, []string{"aa:bb:cc:dd:ee:ff"}, report.Duplicates, "duplicate detected")

	record, _, err := store.Get("aa:bb:cc:dd:ee:ff")
	require.NoError(t, err, "Get()")
	require.NotEmpty(t, record.Alias, "alias should be set")
}

func TestImport_BlankAliasClearsExisting(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	require.NoError(t, store.SetAlias("aa:bb:cc:dd:ee:ff", "Old Name"), "seed alias")

	path := writeAliasesFile(t, `aliases:
  aa:bb:cc:dd:ee:ff: ""
`)

	report, err := NewImporter(store).Import(path)
	require.NoError(t, err, "Import()")
	require.Equal(t, 1, report.Cleared, "cleared count")
	require.Equal(t, 0, report.Imported, "imported count")

	_, found, err := store.Get("aa:bb:cc:dd:ee:ff")
	require.NoError(t, err, "Get()")
	require.False(t, found, "record should be cleared")
}

func TestImport_InvalidMACIsSkipped(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	path := writeAliasesFile(t, `aliases:
  not-a-mac: "Nope"
  aa:bb:cc:dd:ee:ff: "Valid"
`)

	report, err := NewImporter(store).Import(path)
	require.NoError(t, err, "Import() should not fail on invalid MAC")
	require.Equal(t, 1, report.Imported, "imported count")
	require.Equal(t, []string{"not-a-mac"}, report.Invalid, "invalid list")
}

func TestImport_AbsentKeysAreNotDeleted(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	require.NoError(t, store.SetAlias("11:22:33:44:55:66", "Untouched"), "seed alias")

	path := writeAliasesFile(t, `aliases:
  aa:bb:cc:dd:ee:ff: "Imported"
`)

	_, err := NewImporter(store).Import(path)
	require.NoError(t, err, "Import()")

	record, found, err := store.Get("11:22:33:44:55:66")
	require.NoError(t, err, "Get()")
	require.True(t, found, "record outside the file must survive (patch, not mirror)")
	require.Equal(t, "Untouched", record.Alias)
}

func TestImport_MissingFile(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	_, err := NewImporter(store).Import(filepath.Join(t.TempDir(), "nope.yaml"))
	require.Error(t, err, "Import() should fail on missing file")
	require.ErrorIs(t, err, ErrFileNotFound)
}

func TestImportIfChanged_FirstRunImports(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	path := writeAliasesFile(t, `aliases:
  aa:bb:cc:dd:ee:ff: "TV"
`)

	report, imported, err := NewImporter(store).ImportIfChanged(path)
	require.NoError(t, err, "ImportIfChanged()")
	require.True(t, imported, "first run should import")
	require.Equal(t, 1, report.Imported)
}

func TestImportIfChanged_UnchangedFileIsNoOp(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	path := writeAliasesFile(t, `aliases:
  aa:bb:cc:dd:ee:ff: "TV"
`)

	importer := NewImporter(store)
	_, imported, err := importer.ImportIfChanged(path)
	require.NoError(t, err, "first ImportIfChanged()")
	require.True(t, imported)

	// A TUI-side edit must survive an unchanged file across restarts.
	require.NoError(t, store.SetAlias("aa:bb:cc:dd:ee:ff", "Edited In TUI"), "tui edit")

	_, imported, err = importer.ImportIfChanged(path)
	require.NoError(t, err, "second ImportIfChanged()")
	require.False(t, imported, "unchanged file must not re-import")

	record, _, err := store.Get("aa:bb:cc:dd:ee:ff")
	require.NoError(t, err, "Get()")
	require.Equal(t, "Edited In TUI", record.Alias, "tui edit must survive")
}

func TestImportIfChanged_ChangedFileReimports(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	path := writeAliasesFile(t, `aliases:
  aa:bb:cc:dd:ee:ff: "TV"
`)

	importer := NewImporter(store)
	_, imported, err := importer.ImportIfChanged(path)
	require.NoError(t, err)
	require.True(t, imported)

	require.NoError(t, os.WriteFile(path, []byte(`aliases:
  aa:bb:cc:dd:ee:ff: "TV"
  de:ad:be:ef:00:01: "Printer"
`), 0o644), "rewrite aliases file")

	report, imported, err := importer.ImportIfChanged(path)
	require.NoError(t, err, "ImportIfChanged() after change")
	require.True(t, imported, "changed file should re-import")
	require.Equal(t, 2, report.Imported, "both aliases imported")
}

func TestImportIfChanged_MissingFileIsNoOp(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	_, imported, err := NewImporter(store).ImportIfChanged(filepath.Join(t.TempDir(), "nope.yaml"))
	require.NoError(t, err, "ImportIfChanged() should not error on missing file")
	require.False(t, imported, "missing file is a no-op")
}

func TestExportToFile_RoundTrips(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	require.NoError(t, store.SetAlias("aa:bb:cc:dd:ee:ff", "TV"), "seed")
	require.NoError(t, store.SetAlias("de:ad:be:ef:00:01", "Printer"), "seed")

	path := filepath.Join(t.TempDir(), "out", "aliases.yaml")
	require.NoError(t, ExportToFile(store, path), "ExportToFile()")

	raw, err := os.ReadFile(path)
	require.NoError(t, err, "ReadFile()")
	file, err := Parse(raw)
	require.NoError(t, err, "Parse(export)")
	require.Equal(t, map[string]string{
		"aa:bb:cc:dd:ee:ff": "TV",
		"de:ad:be:ef:00:01": "Printer",
	}, file.Aliases)

	// Importing the export into a fresh store reproduces the same state.
	fresh := newTestStore(t)
	_, err = NewImporter(fresh).Import(path)
	require.NoError(t, err, "Import(export)")

	freshRecords, err := fresh.All()
	require.NoError(t, err, "All()")
	require.Len(t, freshRecords, 2, "round-trip record count")
}

func TestExport_SkipsEmptyAliases(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	require.NoError(t, store.SetAlias("aa:bb:cc:dd:ee:ff", "   "), "blank alias clears")

	file, err := Export(store)
	require.NoError(t, err, "Export()")
	require.Empty(t, file.Aliases, "no empty aliases exported")
}
