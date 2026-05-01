package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tiagokriok/kanji/internal/state"
)

func TestContextClear_RemovesCLIContext(t *testing.T) {
	dir := t.TempDir()
	store := state.NewStore(dir + "/state.json")
	_ = store.SetCLIContext("test-ns", state.CLIContext{WorkspaceID: "ws-1", BoardID: "board-1"})

	cmd := &cobra.Command{}
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runContextClearWithNamespace(cmd, store, ns)
	require.NoError(t, err)

	ctx, err := store.GetCLIContext("test-ns")
	require.NoError(t, err)
	assert.Empty(t, ctx.WorkspaceID)
	assert.Empty(t, ctx.BoardID)

	output := buf.String()
	assert.Contains(t, output, "Cleared")
}

func TestContextClear_PreservesTUIState(t *testing.T) {
	dir := t.TempDir()
	store := state.NewStore(dir + "/state.json")
	_ = store.SetCLIContext("test-ns", state.CLIContext{WorkspaceID: "ws-1"})
	_ = store.SetTUIState("test-ns", state.TUIState{LastWorkspaceID: "ws-tui"})

	cmd := &cobra.Command{}
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runContextClearWithNamespace(cmd, store, ns)
	require.NoError(t, err)

	ts, err := store.GetTUIState("test-ns")
	require.NoError(t, err)
	assert.Equal(t, "ws-tui", ts.LastWorkspaceID)
}

func TestContextClear_JSON(t *testing.T) {
	dir := t.TempDir()
	store := state.NewStore(dir + "/state.json")
	_ = store.SetCLIContext("test-ns", state.CLIContext{WorkspaceID: "ws-1"})

	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", true, "")
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runContextClearWithNamespace(cmd, store, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "namespace")
	assert.Contains(t, output, "cleared")
}

func runContextClearWithNamespace(cmd *cobra.Command, store *state.Store, ns Namespace) error {
	cfg, err := ResolveConfig(cmd)
	if err != nil {
		return err
	}

	if err := store.ClearCLIContext(ns.Key); err != nil {
		return err
	}

	if cfg.JSON {
		payload := map[string]string{
			"namespace": ns.Key,
			"status":    "cleared",
		}
		return RenderWrappedJSON(cmd.OutOrStdout(), "context", payload)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Cleared CLI context for namespace: %s\n", ns.Key)
	return nil
}
