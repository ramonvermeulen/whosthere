package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAliasesImportAndExportCommand(t *testing.T) {
	tmpDir := t.TempDir()

	// Point both the store and the aliases file at isolated temp locations.
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmpDir, "state"))

	aliasesFile := filepath.Join(tmpDir, "aliases.yaml")
	aliasesContent := "version: 1\n" +
		"aliases:\n" +
		"  AA-BB-CC-DD-EE-FF: \"Living Room TV\"\n" +
		"  not-a-mac: \"Skipped\"\n"
	require.NoError(t, os.WriteFile(aliasesFile, []byte(aliasesContent), 0o644), "write aliases file")

	importCmd := newAliasesImportCommand()
	var importOut bytes.Buffer
	importCmd.SetOut(&importOut)

	require.NoError(t, importCmd.RunE(importCmd, []string{aliasesFile}), "import")
	require.Contains(t, importOut.String(), "imported: 1", "import output")
	require.Contains(t, importOut.String(), "not-a-mac", "invalid MAC reported")

	exportFile := filepath.Join(tmpDir, "exported.yaml")
	exportCmd := newAliasesExportCommand()
	var exportOut bytes.Buffer
	exportCmd.SetOut(&exportOut)

	require.NoError(t, exportCmd.RunE(exportCmd, []string{exportFile}), "export")
	require.Contains(t, exportOut.String(), "Exported aliases to", "export output")

	content, err := os.ReadFile(exportFile)
	require.NoError(t, err, "read exported file")
	require.Contains(t, string(content), "aa:bb:cc:dd:ee:ff", "normalized MAC exported")
	require.Contains(t, string(content), "Living Room TV", "alias exported")
}

func TestAliasesExportCommand_Stdout(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmpDir, "state"))

	aliasesFile := filepath.Join(tmpDir, "aliases.yaml")
	require.NoError(t, os.WriteFile(aliasesFile, []byte("aliases:\n  aa:bb:cc:dd:ee:ff: \"TV\"\n"), 0o644), "write aliases file")

	importCmd := newAliasesImportCommand()
	importCmd.SetOut(&bytes.Buffer{})
	require.NoError(t, importCmd.RunE(importCmd, []string{aliasesFile}), "import")

	exportCmd := newAliasesExportCommand()
	var out bytes.Buffer
	exportCmd.SetOut(&out)
	require.NoError(t, exportCmd.Flags().Set("stdout", "true"), "set stdout flag")
	require.NoError(t, exportCmd.RunE(exportCmd, nil), "export --stdout")

	require.Contains(t, out.String(), "aa:bb:cc:dd:ee:ff", "alias on stdout")
	require.Contains(t, out.String(), "version: 1", "versioned output")
}
