package cli

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tiagokriok/kanji/internal/state"
)

func TestResolveWorkspaceScope_ByWorkspaceIDFlag(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")
	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	require.Len(t, ws, 1)

	cmd := &cobra.Command{}
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--workspace-id", ws[0].ID}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	id, from, err := ResolveWorkspaceScope(cmd, rt, store, ns)
	require.NoError(t, err)
	assert.Equal(t, ws[0].ID, id)
	assert.Equal(t, "flag", from)
}

func TestResolveWorkspaceScope_ByWorkspaceNameFlag(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--workspace", "Default Workspace"}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	id, from, err := ResolveWorkspaceScope(cmd, rt, store, ns)
	require.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.Equal(t, "flag", from)
}

func TestResolveWorkspaceScope_FromCLIContext(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")
	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	require.Len(t, ws, 1)

	require.NoError(t, store.SetCLIContext("test-ns", state.CLIContext{WorkspaceID: ws[0].ID}))

	cmd := &cobra.Command{}
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	id, from, err := ResolveWorkspaceScope(cmd, rt, store, ns)
	require.NoError(t, err)
	assert.Equal(t, ws[0].ID, id)
	assert.Equal(t, "context", from)
}

func TestResolveWorkspaceScope_FailsWhenNoSource(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	_, _, err := ResolveWorkspaceScope(cmd, rt, store, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace")
}

func TestResolveWorkspaceScope_FlagTakesPrecedenceOverContext(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")
	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	require.Len(t, ws, 1)

	require.NoError(t, store.SetCLIContext("test-ns", state.CLIContext{WorkspaceID: ws[0].ID}))

	cmd := &cobra.Command{}
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--workspace-id", ws[0].ID}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	_, from, err := ResolveWorkspaceScope(cmd, rt, store, ns)
	require.NoError(t, err)
	assert.Equal(t, "flag", from)
}

func TestResolveWorkspaceScope_NotFoundByID(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--workspace-id", "invalid-id"}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	_, _, err := ResolveWorkspaceScope(cmd, rt, store, ns)
	require.Error(t, err)
	assert.True(t, isNotFound(err))
}

func TestResolveWorkspaceScope_NotFoundByName(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--workspace", "Nonexistent"}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	_, _, err := ResolveWorkspaceScope(cmd, rt, store, ns)
	require.Error(t, err)
	assert.True(t, isNotFound(err))
}

func TestResolveBoardScope_ByBoardIDFlag(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")
	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	boards, err := rt.ContextService.ListBoards(context.Background(), ws[0].ID)
	require.NoError(t, err)
	require.Len(t, boards, 1)

	cmd := &cobra.Command{}
	cmd.Flags().String("board-id", "", "")
	cmd.Flags().String("board", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--board-id", boards[0].ID}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	id, from, err := ResolveBoardScope(cmd, rt, store, ns, ws[0].ID)
	require.NoError(t, err)
	assert.Equal(t, boards[0].ID, id)
	assert.Equal(t, "flag", from)
}

func TestResolveBoardScope_ByBoardNameFlag(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")
	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	boards, err := rt.ContextService.ListBoards(context.Background(), ws[0].ID)
	require.NoError(t, err)
	require.Len(t, boards, 1)

	cmd := &cobra.Command{}
	cmd.Flags().String("board-id", "", "")
	cmd.Flags().String("board", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--board", boards[0].Name}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	id, from, err := ResolveBoardScope(cmd, rt, store, ns, ws[0].ID)
	require.NoError(t, err)
	assert.Equal(t, boards[0].ID, id)
	assert.Equal(t, "flag", from)
}

func TestResolveBoardScope_FromCLIContext(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")
	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	boards, err := rt.ContextService.ListBoards(context.Background(), ws[0].ID)
	require.NoError(t, err)
	require.Len(t, boards, 1)

	require.NoError(t, store.SetCLIContext("test-ns", state.CLIContext{
		WorkspaceID: ws[0].ID,
		BoardID:     boards[0].ID,
	}))

	cmd := &cobra.Command{}
	cmd.Flags().String("board-id", "", "")
	cmd.Flags().String("board", "", "")

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	id, from, err := ResolveBoardScope(cmd, rt, store, ns, ws[0].ID)
	require.NoError(t, err)
	assert.Equal(t, boards[0].ID, id)
	assert.Equal(t, "context", from)
}

func TestResolveBoardScope_ContextSkippedWhenWorkspaceMismatch(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")
	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	boards, err := rt.ContextService.ListBoards(context.Background(), ws[0].ID)
	require.NoError(t, err)
	require.Len(t, boards, 1)

	require.NoError(t, store.SetCLIContext("test-ns", state.CLIContext{
		WorkspaceID: ws[0].ID,
		BoardID:     boards[0].ID,
	}))

	cmd := &cobra.Command{}
	cmd.Flags().String("board-id", "", "")
	cmd.Flags().String("board", "", "")

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	_, _, err = ResolveBoardScope(cmd, rt, store, ns, "other-workspace-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "board")
}

func TestResolveBoardScope_FailsWhenNoSource(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")
	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().String("board-id", "", "")
	cmd.Flags().String("board", "", "")

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	_, _, err = ResolveBoardScope(cmd, rt, store, ns, ws[0].ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "board")
}

func TestResolveBoardScope_FlagTakesPrecedenceOverContext(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")
	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	boards, err := rt.ContextService.ListBoards(context.Background(), ws[0].ID)
	require.NoError(t, err)
	require.Len(t, boards, 1)

	require.NoError(t, store.SetCLIContext("test-ns", state.CLIContext{
		WorkspaceID: ws[0].ID,
		BoardID:     boards[0].ID,
	}))

	cmd := &cobra.Command{}
	cmd.Flags().String("board-id", "", "")
	cmd.Flags().String("board", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--board-id", boards[0].ID}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	_, from, err := ResolveBoardScope(cmd, rt, store, ns, ws[0].ID)
	require.NoError(t, err)
	assert.Equal(t, "flag", from)
}

func TestResolveBoardScope_NotFoundByID(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")
	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().String("board-id", "", "")
	cmd.Flags().String("board", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--board-id", "invalid-id"}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	_, _, err = ResolveBoardScope(cmd, rt, store, ns, ws[0].ID)
	require.Error(t, err)
	assert.True(t, isNotFound(err))
}

func TestResolveBoardScope_NotFoundByName(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")
	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().String("board-id", "", "")
	cmd.Flags().String("board", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--board", "Nonexistent"}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	_, _, err = ResolveBoardScope(cmd, rt, store, ns, ws[0].ID)
	require.Error(t, err)
	assert.True(t, isNotFound(err))
}

func TestResolveBoardScope_DoesNotUseTUIState(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")
	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	boards, err := rt.ContextService.ListBoards(context.Background(), ws[0].ID)
	require.NoError(t, err)
	require.Len(t, boards, 1)

	require.NoError(t, store.SetTUIState("test-ns", state.TUIState{
		LastWorkspaceID:      ws[0].ID,
		LastBoardByWorkspace: map[string]string{ws[0].ID: boards[0].ID},
	}))

	cmd := &cobra.Command{}
	cmd.Flags().String("board-id", "", "")
	cmd.Flags().String("board", "", "")

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	_, _, err = ResolveBoardScope(cmd, rt, store, ns, ws[0].ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "board")
}

func TestResolveWorkspaceScope_DoesNotUseTUIState(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")
	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	require.Len(t, ws, 1)

	require.NoError(t, store.SetTUIState("test-ns", state.TUIState{
		LastWorkspaceID: ws[0].ID,
	}))

	cmd := &cobra.Command{}
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	_, _, err = ResolveWorkspaceScope(cmd, rt, store, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace")
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	se, ok := err.(*SelectorError)
	return ok && se.Code == "not_found"
}
