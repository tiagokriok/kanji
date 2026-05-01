package state

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) *Store {
	dir := t.TempDir()
	return NewStore(filepath.Join(dir, "state.json"))
}

func TestStore_Load_MissingFile(t *testing.T) {
	s := newTestStore(t)

	st, err := s.Load()
	require.NoError(t, err)

	assert.NotNil(t, st.Namespaces)
	assert.Empty(t, st.Namespaces)
}

func TestStore_SaveAndLoad_RoundTrip(t *testing.T) {
	s := newTestStore(t)

	st := SharedState{
		Namespaces: map[string]NamespaceState{
			"/home/user/project1": {
				CLIContext: CLIContext{WorkspaceID: "ws-1", BoardID: "board-1"},
				TUIState:   TUIState{LastWorkspaceID: "ws-1", LastBoardByWorkspace: map[string]string{"ws-1": "board-1"}},
			},
			"/home/user/project2": {
				CLIContext: CLIContext{WorkspaceID: "ws-2"},
				TUIState:   TUIState{},
			},
		},
	}

	err := s.Save(st)
	require.NoError(t, err)

	loaded, err := s.Load()
	require.NoError(t, err)

	require.Len(t, loaded.Namespaces, 2)
	assert.Equal(t, "ws-1", loaded.Namespaces["/home/user/project1"].CLIContext.WorkspaceID)
	assert.Equal(t, "board-1", loaded.Namespaces["/home/user/project1"].CLIContext.BoardID)
	assert.Equal(t, "ws-1", loaded.Namespaces["/home/user/project1"].TUIState.LastWorkspaceID)
	assert.Equal(t, "board-1", loaded.Namespaces["/home/user/project1"].TUIState.LastBoardByWorkspace["ws-1"])
	assert.Equal(t, "ws-2", loaded.Namespaces["/home/user/project2"].CLIContext.WorkspaceID)
	assert.Empty(t, loaded.Namespaces["/home/user/project2"].CLIContext.BoardID)
}

func TestStore_NamespaceIsolation(t *testing.T) {
	s := newTestStore(t)

	st := SharedState{
		Namespaces: map[string]NamespaceState{
			"ns-a": {CLIContext: CLIContext{WorkspaceID: "ws-a"}},
			"ns-b": {CLIContext: CLIContext{WorkspaceID: "ws-b"}},
		},
	}

	err := s.Save(st)
	require.NoError(t, err)

	loaded, err := s.Load()
	require.NoError(t, err)

	assert.Equal(t, "ws-a", loaded.Namespaces["ns-a"].CLIContext.WorkspaceID)
	assert.Equal(t, "ws-b", loaded.Namespaces["ns-b"].CLIContext.WorkspaceID)
}

func TestStore_GetCLIContext_MissingNamespace(t *testing.T) {
	s := newTestStore(t)

	ctx, err := s.GetCLIContext("nonexistent")
	require.NoError(t, err)

	assert.Empty(t, ctx.WorkspaceID)
	assert.Empty(t, ctx.BoardID)
}

func TestStore_SetAndGetCLIContext(t *testing.T) {
	s := newTestStore(t)

	err := s.SetCLIContext("ns-1", CLIContext{WorkspaceID: "ws-1", BoardID: "board-1"})
	require.NoError(t, err)

	ctx, err := s.GetCLIContext("ns-1")
	require.NoError(t, err)

	assert.Equal(t, "ws-1", ctx.WorkspaceID)
	assert.Equal(t, "board-1", ctx.BoardID)
}

func TestStore_ClearCLIContext(t *testing.T) {
	s := newTestStore(t)

	err := s.SetCLIContext("ns-1", CLIContext{WorkspaceID: "ws-1", BoardID: "board-1"})
	require.NoError(t, err)

	err = s.ClearCLIContext("ns-1")
	require.NoError(t, err)

	ctx, err := s.GetCLIContext("ns-1")
	require.NoError(t, err)

	assert.Empty(t, ctx.WorkspaceID)
	assert.Empty(t, ctx.BoardID)
}

func TestStore_ClearCLIContext_PreservesTUIState(t *testing.T) {
	s := newTestStore(t)

	st := SharedState{
		Namespaces: map[string]NamespaceState{
			"ns-1": {
				CLIContext: CLIContext{WorkspaceID: "ws-1"},
				TUIState:   TUIState{LastWorkspaceID: "ws-tui"},
			},
		},
	}
	err := s.Save(st)
	require.NoError(t, err)

	err = s.ClearCLIContext("ns-1")
	require.NoError(t, err)

	loaded, err := s.Load()
	require.NoError(t, err)

	assert.Empty(t, loaded.Namespaces["ns-1"].CLIContext.WorkspaceID)
	assert.Equal(t, "ws-tui", loaded.Namespaces["ns-1"].TUIState.LastWorkspaceID)
}

func TestStore_GetTUIState_MissingNamespace(t *testing.T) {
	s := newTestStore(t)

	ts, err := s.GetTUIState("nonexistent")
	require.NoError(t, err)

	assert.Empty(t, ts.LastWorkspaceID)
	assert.Nil(t, ts.LastBoardByWorkspace)
}

func TestStore_SetAndGetTUIState(t *testing.T) {
	s := newTestStore(t)

	err := s.SetTUIState("ns-1", TUIState{LastWorkspaceID: "ws-tui", LastBoardByWorkspace: map[string]string{"ws-tui": "board-tui"}})
	require.NoError(t, err)

	ts, err := s.GetTUIState("ns-1")
	require.NoError(t, err)

	assert.Equal(t, "ws-tui", ts.LastWorkspaceID)
	assert.Equal(t, "board-tui", ts.LastBoardByWorkspace["ws-tui"])
}

func TestStore_DefaultStorePath(t *testing.T) {
	path, err := DefaultStorePath()
	require.NoError(t, err)

	assert.Contains(t, path, "kanji")
	assert.Contains(t, path, "namespaced_state.json")
}

func TestStore_SanitizeNamespace_ClearsCLIContext(t *testing.T) {
	s := newTestStore(t)

	err := s.SetCLIContext("ns-1", CLIContext{WorkspaceID: "ws-1", BoardID: "board-1"})
	require.NoError(t, err)

	err = s.SanitizeNamespace("ns-1", "ws-1")
	require.NoError(t, err)

	ctx, err := s.GetCLIContext("ns-1")
	require.NoError(t, err)
	assert.Empty(t, ctx.WorkspaceID)
	assert.Empty(t, ctx.BoardID)
}

func TestStore_SanitizeNamespace_PreservesOtherWorkspace(t *testing.T) {
	s := newTestStore(t)

	err := s.SetCLIContext("ns-1", CLIContext{WorkspaceID: "ws-2", BoardID: "board-2"})
	require.NoError(t, err)

	err = s.SanitizeNamespace("ns-1", "ws-1")
	require.NoError(t, err)

	ctx, err := s.GetCLIContext("ns-1")
	require.NoError(t, err)
	assert.Equal(t, "ws-2", ctx.WorkspaceID)
	assert.Equal(t, "board-2", ctx.BoardID)
}

func TestStore_SanitizeNamespace_ClearsTUIState(t *testing.T) {
	s := newTestStore(t)

	err := s.SetTUIState("ns-1", TUIState{
		LastWorkspaceID:      "ws-1",
		LastBoardByWorkspace: map[string]string{"ws-1": "board-1", "ws-2": "board-2"},
	})
	require.NoError(t, err)

	err = s.SanitizeNamespace("ns-1", "ws-1")
	require.NoError(t, err)

	ts, err := s.GetTUIState("ns-1")
	require.NoError(t, err)
	assert.Empty(t, ts.LastWorkspaceID)
	assert.NotContains(t, ts.LastBoardByWorkspace, "ws-1")
	assert.Equal(t, "board-2", ts.LastBoardByWorkspace["ws-2"])
}

func TestStore_SanitizeBoard_RemovesCLIContext(t *testing.T) {
	s := newTestStore(t)

	err := s.SetCLIContext("ns-1", CLIContext{WorkspaceID: "ws-1", BoardID: "board-1"})
	require.NoError(t, err)

	err = s.SanitizeBoard("ns-1", "board-1")
	require.NoError(t, err)

	ctx, err := s.GetCLIContext("ns-1")
	require.NoError(t, err)
	assert.Equal(t, "ws-1", ctx.WorkspaceID)
	assert.Empty(t, ctx.BoardID)
}

func TestStore_SanitizeBoard_RemovesTUIState(t *testing.T) {
	s := newTestStore(t)

	err := s.SetTUIState("ns-1", TUIState{
		LastBoardByWorkspace: map[string]string{"ws-1": "board-1", "ws-2": "board-2"},
	})
	require.NoError(t, err)

	err = s.SanitizeBoard("ns-1", "board-1")
	require.NoError(t, err)

	ts, err := s.GetTUIState("ns-1")
	require.NoError(t, err)
	assert.NotContains(t, ts.LastBoardByWorkspace, "ws-1")
	assert.Equal(t, "board-2", ts.LastBoardByWorkspace["ws-2"])
}

func TestStore_SanitizeBoard_PreservesUnrelatedEntries(t *testing.T) {
	s := newTestStore(t)

	err := s.SetCLIContext("ns-1", CLIContext{WorkspaceID: "ws-1", BoardID: "board-1"})
	require.NoError(t, err)
	err = s.SetTUIState("ns-1", TUIState{
		LastBoardByWorkspace: map[string]string{"ws-1": "board-1", "ws-2": "board-2"},
	})
	require.NoError(t, err)

	err = s.SanitizeBoard("ns-1", "board-3")
	require.NoError(t, err)

	ctx, err := s.GetCLIContext("ns-1")
	require.NoError(t, err)
	assert.Equal(t, "board-1", ctx.BoardID)

	ts, err := s.GetTUIState("ns-1")
	require.NoError(t, err)
	assert.Equal(t, "board-1", ts.LastBoardByWorkspace["ws-1"])
	assert.Equal(t, "board-2", ts.LastBoardByWorkspace["ws-2"])
}

func TestStore_SanitizeNamespace_MissingNamespace(t *testing.T) {
	s := newTestStore(t)

	err := s.SanitizeNamespace("nonexistent", "ws-1")
	require.NoError(t, err)
}
