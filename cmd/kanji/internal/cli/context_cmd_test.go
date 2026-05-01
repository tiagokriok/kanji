package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tiagokriok/kanji/internal/state"
)

func TestContextShow_Empty(t *testing.T) {
	dir := t.TempDir()
	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runContextShowWithNamespace(cmd, store, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "test-ns")
	assert.Contains(t, output, "cwd")
	assert.Contains(t, output, "(none)")
}

func TestContextShow_Populated(t *testing.T) {
	dir := t.TempDir()
	store := state.NewStore(dir + "/state.json")
	_ = store.SetCLIContext("test-ns", state.CLIContext{WorkspaceID: "ws-1", BoardID: "board-1"})
	_ = store.SetTUIState("test-ns", state.TUIState{LastWorkspaceID: "ws-tui"})

	cmd := &cobra.Command{}
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "env"}
	err := runContextShowWithNamespace(cmd, store, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "test-ns")
	assert.Contains(t, output, "env")
	assert.Contains(t, output, "ws-1")
	assert.Contains(t, output, "board-1")
	assert.Contains(t, output, "ws-tui")
}

func TestContextShow_JSON(t *testing.T) {
	dir := t.TempDir()
	store := state.NewStore(dir + "/state.json")
	_ = store.SetCLIContext("test-ns", state.CLIContext{WorkspaceID: "ws-1"})

	cmd := &cobra.Command{}
	// Simulate --json flag by setting it on the command.
	cmd.Flags().Bool("json", true, "")
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runContextShowWithNamespace(cmd, store, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "namespace")
	assert.Contains(t, output, "cli_context")
	assert.Contains(t, output, "tui_state")
	assert.Contains(t, output, "ws-1")
}

// runContextShowWithNamespace is a test helper that bypasses ResolveNamespace.
func runContextShowWithNamespace(cmd *cobra.Command, store *state.Store, ns Namespace) error {
	cfg, err := ResolveConfig(cmd)
	if err != nil {
		return err
	}

	cliCtx, _ := store.GetCLIContext(ns.Key)
	tuiState, _ := store.GetTUIState(ns.Key)

	if cfg.JSON {
		return renderContextJSON(cmd.OutOrStdout(), ns, cliCtx, tuiState)
	}
	return renderContextHuman(cmd.OutOrStdout(), ns, cliCtx, tuiState)
}
