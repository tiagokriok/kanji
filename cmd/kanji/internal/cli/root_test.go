package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCommand_Exists(t *testing.T) {
	cmd := NewRootCommand()
	require.NotNil(t, cmd)
	assert.Equal(t, "kanji", cmd.Name())
	assert.Contains(t, cmd.Short, "kanji")
	assert.Contains(t, cmd.Long, "kanji")
}

func TestNewRootCommand_NoArgsShowsHelp(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{})

	buf := new(strings.Builder)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "kanji")
	assert.Contains(t, output, "Usage")
	// With no subcommands yet, Cobra shows "kanji [flags]". Once subcommands
	// are added, this becomes "kanji [command]".
	assert.Contains(t, output, "kanji [flags]")
}

func TestNewRootCommand_PlaceholderForSubcommands(t *testing.T) {
	cmd := NewRootCommand()
	require.NotNil(t, cmd)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	// Root should be runnable (showing help) even before any subcommands exist.
	assert.NotNil(t, cmd.HelpFunc())
}
