package ui

import (
	"testing"

	"github.com/tiagokriok/kanji/internal/application"
	"github.com/tiagokriok/kanji/internal/domain"
)

func newModelWithContextService(repo domain.SetupRepository) Model {
	cs := application.NewContextService(repo)
	return Model{
		contextService: cs,
		state:          persistedUIState{LastBoardByWorkspace: map[string]string{}},
	}
}

// --- containsWorkspace ---

func TestContainsWorkspace_Found(t *testing.T) {
	items := []domain.Workspace{{ID: "ws1", Name: "A"}, {ID: "ws2", Name: "B"}}
	if !containsWorkspace(items, "ws2") {
		t.Error("expected true")
	}
}

func TestContainsWorkspace_NotFound(t *testing.T) {
	items := []domain.Workspace{{ID: "ws1", Name: "A"}}
	if containsWorkspace(items, "ws2") {
		t.Error("expected false")
	}
}

func TestContainsWorkspace_Empty(t *testing.T) {
	if containsWorkspace(nil, "ws1") {
		t.Error("expected false")
	}
}

// --- containsBoard ---

func TestContainsBoard_Found(t *testing.T) {
	items := []domain.Board{{ID: "b1", Name: "A"}, {ID: "b2", Name: "B"}}
	if !containsBoard(items, "b2") {
		t.Error("expected true")
	}
}

func TestContainsBoard_NotFound(t *testing.T) {
	items := []domain.Board{{ID: "b1", Name: "A"}}
	if containsBoard(items, "b2") {
		t.Error("expected false")
	}
}

func TestContainsBoard_Empty(t *testing.T) {
	if containsBoard(nil, "b1") {
		t.Error("expected false")
	}
}

// --- workspaceName ---

func TestWorkspaceName_Found(t *testing.T) {
	items := []domain.Workspace{{ID: "ws1", Name: "Alpha"}}
	if got := workspaceName(items, "ws1"); got != "Alpha" {
		t.Errorf("got %q, want Alpha", got)
	}
}

func TestWorkspaceName_NotFound(t *testing.T) {
	if got := workspaceName(nil, "ws1"); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

// --- boardName ---

func TestBoardName_Found(t *testing.T) {
	items := []domain.Board{{ID: "b1", Name: "Main"}}
	if got := boardName(items, "b1"); got != "Main" {
		t.Errorf("got %q, want Main", got)
	}
}

func TestBoardName_NotFound(t *testing.T) {
	if got := boardName(nil, "b1"); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

// --- workspaceIndexByID ---

func TestWorkspaceIndexByID_Found(t *testing.T) {
	items := []domain.Workspace{{ID: "ws1"}, {ID: "ws2"}}
	if got := workspaceIndexByID(items, "ws2"); got != 1 {
		t.Errorf("got %d, want 1", got)
	}
}

func TestWorkspaceIndexByID_NotFound(t *testing.T) {
	if got := workspaceIndexByID([]domain.Workspace{{ID: "ws1"}}, "ws2"); got != -1 {
		t.Errorf("got %d, want -1", got)
	}
}

// --- boardIndexByID ---

func TestBoardIndexByID_Found(t *testing.T) {
	items := []domain.Board{{ID: "b1"}, {ID: "b2"}}
	if got := boardIndexByID(items, "b2"); got != 1 {
		t.Errorf("got %d, want 1", got)
	}
}

func TestBoardIndexByID_NotFound(t *testing.T) {
	if got := boardIndexByID([]domain.Board{{ID: "b1"}}, "b2"); got != -1 {
		t.Errorf("got %d, want -1", got)
	}
}

// --- bootstrapContexts ---

func TestBootstrapContexts_NilService(t *testing.T) {
	m := Model{state: persistedUIState{LastBoardByWorkspace: map[string]string{}}}
	m.bootstrapContexts()
	if len(m.workspaces) != 0 {
		t.Errorf("workspaces = %d, want 0", len(m.workspaces))
	}
}

func TestBootstrapContexts_WithService(t *testing.T) {
	repo := &mockSetupRepo{workspaces: []domain.Workspace{{ID: "ws1", Name: "A"}}}
	m := newModelWithContextService(repo)
	m.bootstrapContexts()
	if len(m.workspaces) != 1 || m.workspaces[0].ID != "ws1" {
		t.Errorf("workspaces = %v, want [ws1]", m.workspaces)
	}
}

// --- switchWorkspace ---

func TestSwitchWorkspace_NilService(t *testing.T) {
	m := Model{workspaceID: "ws1"}
	if err := m.switchWorkspace("ws2"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestSwitchWorkspace_EmptyID(t *testing.T) {
	repo := &mockSetupRepo{}
	m := newModelWithContextService(repo)
	if err := m.switchWorkspace(""); err == nil {
		t.Error("expected error for empty workspace id")
	}
}

func TestSwitchWorkspace_NotFound(t *testing.T) {
	repo := &mockSetupRepo{workspaces: []domain.Workspace{{ID: "ws1", Name: "A"}}}
	m := newModelWithContextService(repo)
	m.workspaces = repo.workspaces
	if err := m.switchWorkspace("ws2"); err == nil {
		t.Error("expected error for missing workspace")
	}
}

func TestSwitchWorkspace_Success(t *testing.T) {
	repo := &mockSetupRepo{
		workspaces: []domain.Workspace{{ID: "ws1", Name: "A"}},
		boards:     []domain.Board{{ID: "b1", Name: "Main", WorkspaceID: "ws1"}},
		columns:    []domain.Column{{ID: "c1", Name: "Todo", Position: 1}},
	}
	m := newModelWithContextService(repo)
	m.workspaces = repo.workspaces
	if err := m.switchWorkspace("ws1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.workspaceID != "ws1" {
		t.Errorf("workspaceID = %q, want ws1", m.workspaceID)
	}
	if m.workspaceName != "A" {
		t.Errorf("workspaceName = %q, want A", m.workspaceName)
	}
	if m.boardID != "b1" {
		t.Errorf("boardID = %q, want b1", m.boardID)
	}
	if m.boardName != "Main" {
		t.Errorf("boardName = %q, want Main", m.boardName)
	}
	if len(m.columns) != 1 || m.columns[0].ID != "c1" {
		t.Errorf("columns = %v, want [c1]", m.columns)
	}
}

func TestSwitchWorkspace_CreatesMainBoardWhenEmpty(t *testing.T) {
	repo := &mockSetupRepo{
		workspaces: []domain.Workspace{{ID: "ws1", Name: "A"}},
		columns:    []domain.Column{{ID: "c1", Name: "Todo", Position: 1}},
	}
	m := newModelWithContextService(repo)
	m.workspaces = repo.workspaces
	if err := m.switchWorkspace("ws1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createdBoard.Name != "Main" || repo.createdBoard.WorkspaceID != "ws1" {
		t.Errorf("createdBoard = %+v, want Main board in ws1", repo.createdBoard)
	}
	if m.boardID != repo.createdBoard.ID {
		t.Errorf("boardID = %q, want %q", m.boardID, repo.createdBoard.ID)
	}
}

func TestSwitchWorkspace_RestoresSavedBoard(t *testing.T) {
	repo := &mockSetupRepo{
		workspaces: []domain.Workspace{{ID: "ws1", Name: "A"}},
		boards: []domain.Board{
			{ID: "b1", Name: "First", WorkspaceID: "ws1"},
			{ID: "b2", Name: "Second", WorkspaceID: "ws1"},
		},
		columns: []domain.Column{{ID: "c1", Name: "Todo", Position: 1}},
	}
	m := newModelWithContextService(repo)
	m.workspaces = repo.workspaces
	m.state.LastWorkspaceID = "ws1"
	m.state.LastBoardByWorkspace["ws1"] = "b2"
	if err := m.switchWorkspace("ws1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.boardID != "b2" {
		t.Errorf("boardID = %q, want b2", m.boardID)
	}
}

// --- switchBoard ---

func TestSwitchBoard_NilService(t *testing.T) {
	m := Model{boardID: "b1"}
	if err := m.switchBoard("b2"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestSwitchBoard_EmptyID(t *testing.T) {
	repo := &mockSetupRepo{}
	m := newModelWithContextService(repo)
	if err := m.switchBoard(""); err == nil {
		t.Error("expected error for empty board id")
	}
}

func TestSwitchBoard_NotFound(t *testing.T) {
	repo := &mockSetupRepo{boards: []domain.Board{{ID: "b1", Name: "Main"}}}
	m := newModelWithContextService(repo)
	m.boards = repo.boards
	if err := m.switchBoard("b2"); err == nil {
		t.Error("expected error for missing board")
	}
}

func TestSwitchBoard_Success(t *testing.T) {
	repo := &mockSetupRepo{
		boards:  []domain.Board{{ID: "b1", Name: "Main"}},
		columns: []domain.Column{{ID: "c1", Name: "Todo", Position: 1}, {ID: "c2", Name: "Done", Position: 2}},
	}
	m := newModelWithContextService(repo)
	m.boards = repo.boards
	if err := m.switchBoard("b1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.boardID != "b1" {
		t.Errorf("boardID = %q, want b1", m.boardID)
	}
	if m.boardName != "Main" {
		t.Errorf("boardName = %q, want Main", m.boardName)
	}
	if len(m.columns) != 2 {
		t.Errorf("columns = %d, want 2", len(m.columns))
	}
	if m.columns[0].Position != 1 || m.columns[1].Position != 2 {
		t.Errorf("column order wrong: %v", m.columns)
	}
}

// --- switchBoardByOffset ---

func TestSwitchBoardByOffset_EmptyBoards(t *testing.T) {
	m := Model{boards: nil, boardID: "b1"}
	changed, err := m.switchBoardByOffset(1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if changed {
		t.Error("expected no change")
	}
}

func TestSwitchBoardByOffset_SingleBoard(t *testing.T) {
	m := Model{boards: []domain.Board{{ID: "b1"}}, boardID: "b1"}
	changed, err := m.switchBoardByOffset(1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if changed {
		t.Error("expected no change for single board")
	}
}

func TestSwitchBoardByOffset_Next(t *testing.T) {
	repo := &mockSetupRepo{
		boards:  []domain.Board{{ID: "b1", Name: "A"}, {ID: "b2", Name: "B"}},
		columns: []domain.Column{{ID: "c1", Name: "Todo", Position: 1}},
	}
	m := newModelWithContextService(repo)
	m.boards = repo.boards
	m.boardID = "b1"
	changed, err := m.switchBoardByOffset(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected change")
	}
	if m.boardID != "b2" {
		t.Errorf("boardID = %q, want b2", m.boardID)
	}
}

func TestSwitchBoardByOffset_PreviousWrap(t *testing.T) {
	repo := &mockSetupRepo{
		boards:  []domain.Board{{ID: "b1", Name: "A"}, {ID: "b2", Name: "B"}},
		columns: []domain.Column{{ID: "c1", Name: "Todo", Position: 1}},
	}
	m := newModelWithContextService(repo)
	m.boards = repo.boards
	m.boardID = "b1"
	changed, err := m.switchBoardByOffset(-1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected change")
	}
	if m.boardID != "b2" {
		t.Errorf("boardID = %q, want b2", m.boardID)
	}
}

// --- reloadContextsFromStorage ---

func TestReloadContextsFromStorage_NilService(t *testing.T) {
	m := Model{state: persistedUIState{LastBoardByWorkspace: map[string]string{}}}
	if err := m.reloadContextsFromStorage(); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestReloadContextsFromStorage_UsesLastWorkspace(t *testing.T) {
	repo := &mockSetupRepo{
		workspaces: []domain.Workspace{{ID: "ws1", Name: "A"}, {ID: "ws2", Name: "B"}},
		boards:     []domain.Board{{ID: "b1", Name: "Main", WorkspaceID: "ws2"}},
		columns:    []domain.Column{{ID: "c1", Name: "Todo", Position: 1}},
	}
	m := newModelWithContextService(repo)
	m.state.LastWorkspaceID = "ws2"
	if err := m.reloadContextsFromStorage(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.workspaceID != "ws2" {
		t.Errorf("workspaceID = %q, want ws2", m.workspaceID)
	}
}

func TestReloadContextsFromStorage_FallsBackToFirstWorkspace(t *testing.T) {
	repo := &mockSetupRepo{
		workspaces: []domain.Workspace{{ID: "ws1", Name: "A"}},
		boards:     []domain.Board{{ID: "b1", Name: "Main", WorkspaceID: "ws1"}},
		columns:    []domain.Column{{ID: "c1", Name: "Todo", Position: 1}},
	}
	m := newModelWithContextService(repo)
	if err := m.reloadContextsFromStorage(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.workspaceID != "ws1" {
		t.Errorf("workspaceID = %q, want ws1", m.workspaceID)
	}
}

func TestReloadContextsFromStorage_NoWorkspaces(t *testing.T) {
	repo := &mockSetupRepo{workspaces: []domain.Workspace{}}
	m := newModelWithContextService(repo)
	if err := m.reloadContextsFromStorage(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.workspaces) != 0 {
		t.Errorf("workspaces = %d, want 0", len(m.workspaces))
	}
}

// --- persistContextSelection ---

func TestPersistContextSelection_UpdatesState(t *testing.T) {
	repo := &mockSetupRepo{workspaces: []domain.Workspace{{ID: "ws1", Name: "A"}}}
	m := newModelWithContextService(repo)
	m.workspaceID = "ws1"
	m.boardID = "b1"
	m.workspaces = repo.workspaces
	// avoid disk write by not calling through switchWorkspace; test state directly
	m.state.LastWorkspaceID = ""
	m.state.LastBoardByWorkspace = map[string]string{}
	m.persistContextSelection()
	if m.state.LastWorkspaceID != "ws1" {
		t.Errorf("LastWorkspaceID = %q, want ws1", m.state.LastWorkspaceID)
	}
	if m.state.LastBoardByWorkspace["ws1"] != "b1" {
		t.Errorf("LastBoardByWorkspace[ws1] = %q, want b1", m.state.LastBoardByWorkspace["ws1"])
	}
}
