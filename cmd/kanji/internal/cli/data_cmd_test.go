package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataBootstrapCmd_FirstRun(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/test.db"

	root := NewRootCommand()
	root.SetArgs([]string{"data", "bootstrap", "--db-path", dbPath})

	buf := new(strings.Builder)
	root.SetOut(buf)
	root.SetErr(buf)

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Default Workspace")
	assert.Contains(t, output, "Default Board")
}

func TestDataBootstrapCmd_Idempotent(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/test.db"

	root := NewRootCommand()
	root.SetArgs([]string{"data", "bootstrap", "--db-path", dbPath})

	// First run.
	err := root.Execute()
	require.NoError(t, err)

	// Second run should also succeed.
	buf := new(strings.Builder)
	root.SetOut(buf)
	root.SetErr(buf)
	err = root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Default Workspace")
}

func TestDataBootstrapCmd_JSONOutput(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/test.db"

	root := NewRootCommand()
	root.SetArgs([]string{"data", "bootstrap", "--db-path", dbPath, "--json"})

	buf := new(strings.Builder)
	root.SetOut(buf)
	root.SetErr(buf)

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "workspace")
	assert.Contains(t, output, "board")
	assert.Contains(t, output, "columns")
}
