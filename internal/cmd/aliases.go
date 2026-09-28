package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/ramonvermeulen/whosthere/internal/core/aliases"
	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/ramonvermeulen/whosthere/internal/core/devicemeta"
	"github.com/spf13/cobra"
)

func NewAliasesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aliases",
		Short: "Import and export device aliases from a YAML file",
		Long: `Manage device aliases through a user-editable YAML file.

The aliases file is a seed/patch for the local device database, not a mirror:
importing upserts the MAC addresses present in the file and never deletes
aliases for MAC addresses that are absent. Blank aliases clear the record.` + magenta + `

Examples:` + reset + `
  whosthere aliases import
  whosthere aliases import ./aliases.yaml
  whosthere aliases export
  whosthere aliases export --stdout
`,
	}

	cmd.AddCommand(
		newAliasesImportCommand(),
		newAliasesExportCommand(),
	)

	return cmd
}

func newAliasesImportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "import [file]",
		Short: "Import aliases from a file into the local database",
		Long: `Import aliases from a YAML file into the local device database.

When no file is given, the default location is used, which can be overridden
with the aliases.file configuration key (or WHOSTHERE__ALIASES__FILE).`,
		Args: cobra.MaximumNArgs(1),
		RunE: runAliasesImport,
	}
}

func newAliasesExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [file]",
		Short: "Export aliases from the local database to a file",
		Long: `Export all aliases from the local device database into the aliases file
format. The output can be re-imported and is a good starting point for manual
editing.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runAliasesExport,
	}

	cmd.Flags().Bool("stdout", false, "Write the aliases to stdout instead of a file")

	return cmd
}

func resolveAliasesPath(args []string) (string, error) {
	if len(args) > 0 && args[0] != "" {
		return args[0], nil
	}

	// ModeCLI applies environment overrides (WHOSTHERE__ALIASES__FILE) and flag
	// overrides without requiring or creating a config file.
	cfg, err := config.LoadForMode(config.ModeCLI, whosthereFlags)
	if err != nil {
		return "", err
	}

	return aliases.ResolvePath(cfg.Aliases.File)
}

func openAliasesStore() (devicemeta.Store, error) {
	store, err := devicemeta.OpenDefault()
	if err != nil {
		return nil, fmt.Errorf("open device metadata store: %w", err)
	}

	return store, nil
}

func runAliasesImport(cmd *cobra.Command, args []string) error {
	path, err := resolveAliasesPath(args)
	if err != nil {
		return err
	}

	store, err := openAliasesStore()
	if err != nil {
		return err
	}
	defer func() {
		_ = store.Close()
	}()

	report, err := aliases.NewImporter(store).Import(path)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	summary := fmt.Sprintf("Imported aliases from %s\n  imported: %d\n  cleared:  %d\n",
		report.Path, report.Imported, report.Cleared)
	if report.Skipped > 0 {
		summary += fmt.Sprintf("  skipped:  %d\n", report.Skipped)
	}
	if len(report.Invalid) > 0 {
		summary += fmt.Sprintf("  invalid MAC addresses (skipped): %s\n", strings.Join(report.Invalid, ", "))
	}
	if len(report.Duplicates) > 0 {
		summary += fmt.Sprintf("  duplicate MAC addresses (last one wins): %s\n", strings.Join(report.Duplicates, ", "))
	}

	_, err = io.WriteString(out, summary)
	return err
}
func runAliasesExport(cmd *cobra.Command, args []string) error {
	path, err := resolveAliasesPath(args)
	if err != nil {
		return err
	}

	store, err := openAliasesStore()
	if err != nil {
		return err
	}
	defer func() {
		_ = store.Close()
	}()

	asStdout, _ := cmd.Flags().GetBool("stdout")
	if asStdout {
		file, err := aliases.Export(store)
		if err != nil {
			return err
		}

		raw, err := aliases.Marshal(file)
		if err != nil {
			return err
		}

		_, err = cmd.OutOrStdout().Write(raw)
		return err
	}

	if err := aliases.ExportToFile(store, path); err != nil {
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Exported aliases to %s\n", path)
	return err
}
