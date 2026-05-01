package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelpTopics_Exist(t *testing.T) {
	root := NewRootCommand()

	topics := []string{"concepts", "context", "selectors", "output"}
	for _, topic := range topics {
		cmd, _, err := root.Find([]string{"help", topic})
		require.NoError(t, err, "topic %s should exist", topic)
		assert.NotNil(t, cmd, "topic %s command should not be nil", topic)
	}
}

func TestHelpTopic_Concepts(t *testing.T) {
	root := NewRootCommand()
	root.SetArgs([]string{"help", "concepts"})

	buf := new(strings.Builder)
	root.SetOut(buf)
	root.SetErr(buf)

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "workspace")
	assert.Contains(t, output, "board")
	assert.Contains(t, output, "column")
	assert.Contains(t, output, "task")
}

func TestHelpTopic_Context(t *testing.T) {
	root := NewRootCommand()
	root.SetArgs([]string{"help", "context"})

	buf := new(strings.Builder)
	root.SetOut(buf)
	root.SetErr(buf)

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "cli_context")
	assert.Contains(t, output, "tui_state")
	assert.Contains(t, output, "namespace")
}

func TestHelpTopic_Selectors(t *testing.T) {
	root := NewRootCommand()
	root.SetArgs([]string{"help", "selectors"})

	buf := new(strings.Builder)
	root.SetOut(buf)
	root.SetErr(buf)

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "name")
	assert.Contains(t, output, "exact")
}

func TestHelpTopic_Output(t *testing.T) {
	root := NewRootCommand()
	root.SetArgs([]string{"help", "output"})

	buf := new(strings.Builder)
	root.SetOut(buf)
	root.SetErr(buf)

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "JSON")
	assert.Contains(t, output, "Human")
	assert.Contains(t, output, "verbose")
}
