package cmd

import (
	"github.com/ramonvermeulen/whosthere/internal/core/version"
	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version and exit",
		Run: func(cmd *cobra.Command, args []string) {
			version.Fprint(cmd.OutOrStdout())
		},
	}
}
