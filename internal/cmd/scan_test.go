package cmd

import (
	"testing"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/stretchr/testify/assert"
)

func TestNewScanCommand(t *testing.T) {
	cmd := NewScanCommand()

	assert.Equal(t, "scan", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}

func TestNewScanCommand_HasJSONFlag(t *testing.T) {
	cmd := NewScanCommand()

	flag := cmd.Flags().Lookup("json")
	assert.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
	assert.Equal(t, "Output results in JSON format", flag.Usage)
}

func TestNewScanCommand_HasPrettyFlag(t *testing.T) {
	cmd := NewScanCommand()

	flag := cmd.Flags().Lookup("pretty")
	assert.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
	assert.Equal(t, "Pretty print output", flag.Usage)
}

func TestNewScanCommand_HasAllInheritedPersistentFlags(t *testing.T) {
	rootCmd := NewRootCommand()
	scanCmd, _, err := rootCmd.Find([]string{"scan"})
	assert.NoError(t, err)

	settings := config.GlobalSettings()
	for _, s := range settings {
		if s.Sources[config.SourceFlag] {
			t.Run(s.FlagName, func(t *testing.T) {
				flag := scanCmd.PersistentFlags().Lookup(s.FlagName)
				assert.NotNil(t, flag, "persistent flag %s should be inherited by scan command", s.FlagName)
			})
		}
	}
}
