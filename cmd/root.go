package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "whosthere",
	Short: "Local network discovery tool with a modern TUI interface.",
	Long: `About
Local network discovery tool with a modern TUI interface written in Go. Discover, explore, and understand your Local Area Network in an intuitive way. Nok nok, who's there? 🚪`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func init() {
	fmt.Print("Entrypoint of whosthere")
}

// Execute is the entrypoint for the CLI application
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
