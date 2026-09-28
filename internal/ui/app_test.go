package ui

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
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

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644), "write file")
}

func TestMaybeAutoImportAliases_DisabledIsNoOp(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	path := filepath.Join(t.TempDir(), "aliases.yaml")
	writeFile(t, path, "aliases:\n  aa:bb:cc:dd:ee:ff: \"TV\"\n")

	cfg := &config.Config{}
	cfg.Aliases.AutoImport = false
	cfg.Aliases.File = path

	require.NoError(t, maybeAutoImportAliases(cfg, store, testLogger()), "disabled should be a no-op")

	_, found, err := store.Get("aa:bb:cc:dd:ee:ff")
	require.NoError(t, err, "Get()")
	require.False(t, found, "alias must not be imported when auto_import is disabled")
}

func TestMaybeAutoImportAliases_EnabledImports(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	path := filepath.Join(t.TempDir(), "aliases.yaml")
	writeFile(t, path, "aliases:\n  aa:bb:cc:dd:ee:ff: \"TV\"\n")

	cfg := &config.Config{}
	cfg.Aliases.AutoImport = true
	cfg.Aliases.File = path

	require.NoError(t, maybeAutoImportAliases(cfg, store, testLogger()), "auto-import")

	record, found, err := store.Get("aa:bb:cc:dd:ee:ff")
	require.NoError(t, err, "Get()")
	require.True(t, found, "alias should be imported")
	require.Equal(t, "TV", record.Alias)
}

func TestMaybeAutoImportAliases_MissingFileIsNotAnError(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	cfg := &config.Config{}
	cfg.Aliases.AutoImport = true
	cfg.Aliases.File = filepath.Join(t.TempDir(), "does-not-exist.yaml")

	require.NoError(t, maybeAutoImportAliases(cfg, store, testLogger()), "missing file should be tolerated")
}

func TestMaybeAutoImportAliases_UnchangedFileDoesNotClobberTUIEdit(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	path := filepath.Join(t.TempDir(), "aliases.yaml")
	writeFile(t, path, "aliases:\n  aa:bb:cc:dd:ee:ff: \"From File\"\n")

	cfg := &config.Config{}
	cfg.Aliases.AutoImport = true
	cfg.Aliases.File = path

	require.NoError(t, maybeAutoImportAliases(cfg, store, testLogger()), "first import")

	require.NoError(t, store.SetAlias("aa:bb:cc:dd:ee:ff", "Edited In TUI"), "tui edit")

	require.NoError(t, maybeAutoImportAliases(cfg, store, testLogger()), "second import")

	record, _, err := store.Get("aa:bb:cc:dd:ee:ff")
	require.NoError(t, err, "Get()")
	require.Equal(t, "Edited In TUI", record.Alias, "unchanged file must not clobber a TUI edit")
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}
