package cmd

import (
	"testing"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/stretchr/testify/assert"
)

func TestNewDaemonCommand(t *testing.T) {
	cmd := NewDaemonCommand()

	assert.Equal(t, "daemon", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}

func TestNewDaemonCommand_HasAllInheritedPersistentFlags(t *testing.T) {
	rootCmd := NewRootCommand()
	daemonCmd, _, err := rootCmd.Find([]string{"daemon"})
	assert.NoError(t, err)

	settings := config.GlobalSettings()
	for _, s := range settings {
		if s.Sources[config.SourceFlag] {
			t.Run(s.FlagName, func(t *testing.T) {
				flag := daemonCmd.PersistentFlags().Lookup(s.FlagName)
				assert.NotNil(t, flag, "persistent flag %s should be inherited by daemon command", s.FlagName)
			})
		}
	}
}
