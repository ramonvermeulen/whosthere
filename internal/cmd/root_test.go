package cmd

import (
	"bytes"
	"testing"

	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/stretchr/testify/assert"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewRootCommand()
	cmd.Version = "1.0.0"

	assert.Equal(t, "whosthere", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Version)
	assert.True(t, cmd.SilenceUsage)
	assert.NotNil(t, cmd.RunE)
}

func TestNewRootCommand_HasTimeoutFlag(t *testing.T) {
	cmd := NewRootCommand()

	flag := cmd.PersistentFlags().Lookup("timeout")
	assert.NotNil(t, flag)
	assert.Equal(t, "", flag.DefValue)
	assert.Contains(t, flag.Usage, "Scan timeout duration")
}

func TestNewRootCommand_HasInterfaceFlag(t *testing.T) {
	cmd := NewRootCommand()

	flag := cmd.PersistentFlags().Lookup("interface")
	assert.NotNil(t, flag)
	assert.Equal(t, "", flag.DefValue)
	assert.Contains(t, flag.Usage, "Network interface")
}

func TestNewRootCommand_VersionFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.Version = "1.0.0"
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "1.0.0")
}

func TestAddCommands(t *testing.T) {
	root := NewRootCommand()
	AddCommands(root)

	expectedCommands := []string{"version", "daemon", "scan"}
	for _, name := range expectedCommands {
		cmd, _, err := root.Find([]string{name})
		assert.NoError(t, err, "command %s should exist", name)
		assert.Equal(t, name, cmd.Name())
	}
}

func TestAddCommands_Count(t *testing.T) {
	root := NewRootCommand()
	AddCommands(root)

	assert.True(t, root.HasSubCommands())
	assert.Len(t, root.Commands(), 3)
}

func TestNewRootCommand_HasAllPersistentFlags(t *testing.T) {
	cmd := NewRootCommand()

	settings := config.GlobalSettings()
	for _, s := range settings {
		if s.Sources[config.SourceFlag] {
			t.Run(s.FlagName, func(t *testing.T) {
				flag := cmd.PersistentFlags().Lookup(s.FlagName)
				assert.NotNil(t, flag, "persistent flag %s should be present on root command", s.FlagName)
			})
		}
	}
}
