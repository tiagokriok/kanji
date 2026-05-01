package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tiagokriok/kanji/internal/state"
)

func setupBootstrappedRuntime(t *testing.T) (*Runtime, string) {
	dir := t.TempDir()
	cfg := RuntimeConfig{DBPath: filepath.Join(dir, "test.db")}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	_, err = rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	return rt, dir
}

func TestContextSet_ByWorkspaceID(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")
	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	require.Len(t, ws, 1)

	cmd := &cobra.Command{}
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")
	cmd.Flags().String("board-id", "", "")
	cmd.Flags().String("board", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--workspace-id", ws[0].ID}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runContextSetWithNamespace(cmd, rt, store, ns)
	require.NoError(t, err)

	ctx, err := store.GetCLIContext("test-ns")
	require.NoError(t, err)
	assert.Equal(t, ws[0].ID, ctx.WorkspaceID)
	assert.Empty(t, ctx.BoardID)
}

func TestContextSet_ByWorkspaceName(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")
	cmd.Flags().String("board-id", "", "")
	cmd.Flags().String("board", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--workspace", "Default Workspace"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runContextSetWithNamespace(cmd, rt, store, ns)
	require.NoError(t, err)

	ctx, err := store.GetCLIContext("test-ns")
	require.NoError(t, err)
	assert.NotEmpty(t, ctx.WorkspaceID)
	assert.Empty(t, ctx.BoardID)
}

func TestContextSet_WithBoard(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")
	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	boards, err := rt.ContextService.ListBoards(context.Background(), ws[0].ID)
	require.NoError(t, err)
	require.Len(t, boards, 1)

	cmd := &cobra.Command{}
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")
	cmd.Flags().String("board-id", "", "")
	cmd.Flags().String("board", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--workspace-id", ws[0].ID, "--board-id", boards[0].ID}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runContextSetWithNamespace(cmd, rt, store, ns)
	require.NoError(t, err)

	ctx, err := store.GetCLIContext("test-ns")
	require.NoError(t, err)
	assert.Equal(t, ws[0].ID, ctx.WorkspaceID)
	assert.Equal(t, boards[0].ID, ctx.BoardID)
}

func TestContextSet_WorkspaceOnlyClearsBoard(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")
	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	boards, err := rt.ContextService.ListBoards(context.Background(), ws[0].ID)
	require.NoError(t, err)

	// Pre-set both workspace and board.
	_ = store.SetCLIContext("test-ns", state.CLIContext{WorkspaceID: ws[0].ID, BoardID: boards[0].ID})

	// Now set workspace only.
	cmd := &cobra.Command{}
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")
	cmd.Flags().String("board-id", "", "")
	cmd.Flags().String("board", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--workspace-id", ws[0].ID}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runContextSetWithNamespace(cmd, rt, store, ns)
	require.NoError(t, err)

	ctx, err := store.GetCLIContext("test-ns")
	require.NoError(t, err)
	assert.Equal(t, ws[0].ID, ctx.WorkspaceID)
	assert.Empty(t, ctx.BoardID)
}

func TestContextSet_InvalidWorkspace(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")
	cmd.Flags().String("board-id", "", "")
	cmd.Flags().String("board", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--workspace-id", "invalid-id"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runContextSetWithNamespace(cmd, rt, store, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestContextSet_InvalidBoard(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")
	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")
	cmd.Flags().String("board-id", "", "")
	cmd.Flags().String("board", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--workspace-id", ws[0].ID, "--board-id", "invalid-id"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runContextSetWithNamespace(cmd, rt, store, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func runContextSetWithNamespace(cmd *cobra.Command, rt *Runtime, store *state.Store, ns Namespace) error {
	cfg, err := ResolveConfig(cmd)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Resolve workspace.
	var workspaceID string
	if cmd.Flags().Changed("workspace-id") {
		id, _ := cmd.Flags().GetString("workspace-id")
		workspaceID = id
		workspaces, err := rt.ContextService.ListWorkspaces(ctx)
		if err != nil {
			return err
		}
		found := false
		for _, ws := range workspaces {
			if ws.ID == workspaceID {
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("workspace", workspaceID)
		}
	} else if cmd.Flags().Changed("workspace") {
		name, _ := cmd.Flags().GetString("workspace")
		workspaces, err := rt.ContextService.ListWorkspaces(ctx)
		if err != nil {
			return err
		}
		found := false
		for _, ws := range workspaces {
			if ExactMatch(ws.Name, name) {
				workspaceID = ws.ID
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("workspace", name)
		}
	} else {
		return NewValidation("workspace-id or workspace is required")
	}

	// Resolve board.
	var boardID string
	if cmd.Flags().Changed("board-id") {
		id, _ := cmd.Flags().GetString("board-id")
		boardID = id
		boards, err := rt.ContextService.ListBoards(ctx, workspaceID)
		if err != nil {
			return err
		}
		found := false
		for _, b := range boards {
			if b.ID == boardID {
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("board", boardID)
		}
	} else if cmd.Flags().Changed("board") {
		name, _ := cmd.Flags().GetString("board")
		boards, err := rt.ContextService.ListBoards(ctx, workspaceID)
		if err != nil {
			return err
		}
		found := false
		for _, b := range boards {
			if ExactMatch(b.Name, name) {
				boardID = b.ID
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("board", name)
		}
	}

	if err := store.SetCLIContext(ns.Key, state.CLIContext{
		WorkspaceID: workspaceID,
		BoardID:     boardID,
	}); err != nil {
		return err
	}

	if cfg.JSON {
		payload := map[string]string{
			"namespace":    ns.Key,
			"workspace_id": workspaceID,
		}
		if boardID != "" {
			payload["board_id"] = boardID
		}
		return RenderWrappedJSON(cmd.OutOrStdout(), "context", payload)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Context set for namespace: %s\n", ns.Key)
	fmt.Fprintf(cmd.OutOrStdout(), "Workspace: %s\n", workspaceID)
	if boardID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Board:     %s\n", boardID)
	}
	return nil
}
