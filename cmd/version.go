package cmd

import (
	"github.com/ramonvermeulen/whosthere/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version and exit",
	Run: func(cmd *cobra.Command, args []string) {
		version.Fprint(cmd.OutOrStdout())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
