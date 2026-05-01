package ui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tiagokriok/kanji/internal/application"
	"github.com/tiagokriok/kanji/internal/domain"
)

// --- mock setup repository ---

type mockSetupRepo struct {
	workspaces       []domain.Workspace
	boards           []domain.Board
	columns          []domain.Column
	createdBoard     domain.Board
	createdWorkspace domain.Workspace
	err              error
}

func (r *mockSetupRepo) ListProviders(ctx context.Context) ([]domain.Provider, error) {
	return nil, r.err
}
func (r *mockSetupRepo) CreateProvider(ctx context.Context, p domain.Provider) error { return r.err }
func (r *mockSetupRepo) ListWorkspaces(ctx context.Context) ([]domain.Workspace, error) {
	return r.workspaces, r.err
}
func (r *mockSetupRepo) CreateWorkspace(ctx context.Context, w domain.Workspace) error {
	r.createdWorkspace = w
	return r.err
}
func (r *mockSetupRepo) RenameWorkspace(ctx context.Context, id, name string) error { return r.err }
func (r *mockSetupRepo) ListBoards(ctx context.Context, wsID string) ([]domain.Board, error) {
	return r.boards, r.err
}
func (r *mockSetupRepo) CreateBoard(ctx context.Context, b domain.Board) error {
	r.createdBoard = b
	return r.err
}
func (r *mockSetupRepo) RenameBoard(ctx context.Context, id, name string) error { return r.err }
func (r *mockSetupRepo) ListColumns(ctx context.Context, boardID string) ([]domain.Column, error) {
	return r.columns, r.err
}
func (r *mockSetupRepo) CreateColumn(ctx context.Context, c domain.Column) error { return r.err }
func (r *mockSetupRepo) UpdateColumn(ctx context.Context, columnID string, name, color *string, wipLimit *int, clearWIP bool) error {
	return r.err
}
func (r *mockSetupRepo) ReorderColumns(ctx context.Context, boardID string, ids []string) error {
	return r.err
}

func newMockModelWithContextService(repo domain.SetupRepository) Model {
	cs := application.NewContextService(repo)
	return Model{
		contextService:   cs,
		contextFilter:    newTaskFormInput("Filter", "", 128),
		contextEditInput: newTaskFormInput("Name", "", 256),
		keys:             newKeyMap(),
		width:            80,
		height:           24,
	}
}

// --- contextItems tests ---

func TestContextItems_EmptyWorkspaces(t *testing.T) {
	m := Model{overlayState: overlayState{contextMode: contextWorkspace}, workspaces: []domain.Workspace{}}
	items := m.contextItems()
	if len(items) != 0 {
		t.Errorf("len(items) = %d, want 0", len(items))
	}
}

func TestContextItems_WorkspaceFilterMatch(t *testing.T) {
	m := Model{
		overlayState: overlayState{contextMode: contextWorkspace},
		workspaces:   []domain.Workspace{{ID: "ws1", Name: "Alpha"}, {ID: "ws2", Name: "Beta"}},
	}
	m.contextFilter.SetValue("alp")
	items := m.contextItems()
	if len(items) != 1 || items[0] != "ws1" {
		t.Errorf("items = %v, want [ws1]", items)
	}
}

func TestContextItems_BoardFilterNoMatch(t *testing.T) {
	m := Model{
		overlayState: overlayState{contextMode: contextBoard},
		boards:       []domain.Board{{ID: "b1", Name: "Main"}},
	}
	m.contextFilter.SetValue("zzz")
	items := m.contextItems()
	if len(items) != 0 {
		t.Errorf("len(items) = %d, want 0", len(items))
	}
}

// --- clampContextSelection tests ---

func TestClampContextSelection_Empty(t *testing.T) {
	m := Model{overlayState: overlayState{contextMode: contextWorkspace}, workspaces: []domain.Workspace{}}
	m.contextSelected = 5
	m.clampContextSelection()
	if m.contextSelected != 0 {
		t.Errorf("contextSelected = %d, want 0", m.contextSelected)
	}
}

func TestClampContextSelection_InBounds(t *testing.T) {
	m := Model{
		overlayState: overlayState{contextMode: contextWorkspace},
		workspaces:   []domain.Workspace{{ID: "ws1"}, {ID: "ws2"}},
	}
	m.contextSelected = 1
	m.clampContextSelection()
	if m.contextSelected != 1 {
		t.Errorf("contextSelected = %d, want 1", m.contextSelected)
	}
}

func TestClampContextSelection_AboveRange(t *testing.T) {
	m := Model{
		overlayState: overlayState{contextMode: contextWorkspace},
		workspaces:   []domain.Workspace{{ID: "ws1"}},
	}
	m.contextSelected = 5
	m.clampContextSelection()
	if m.contextSelected != 0 {
		t.Errorf("contextSelected = %d, want 0", m.contextSelected)
	}
}

func TestClampContextSelection_BelowRange(t *testing.T) {
	m := Model{
		overlayState: overlayState{contextMode: contextWorkspace},
		workspaces:   []domain.Workspace{{ID: "ws1"}, {ID: "ws2"}},
	}
	m.contextSelected = -3
	m.clampContextSelection()
	if m.contextSelected != 0 {
		t.Errorf("contextSelected = %d, want 0", m.contextSelected)
	}
}

// --- selectedContextID tests ---

func TestSelectedContextID_Empty(t *testing.T) {
	m := Model{overlayState: overlayState{contextMode: contextWorkspace}, workspaces: []domain.Workspace{}}
	if got := m.selectedContextID(); got != "" {
		t.Errorf("selectedContextID() = %q, want empty", got)
	}
}

func TestSelectedContextID_Valid(t *testing.T) {
	m := Model{
		overlayState: overlayState{contextMode: contextWorkspace, contextSelected: 0},
		workspaces:   []domain.Workspace{{ID: "ws1", Name: "A"}},
	}
	if got := m.selectedContextID(); got != "ws1" {
		t.Errorf("selectedContextID() = %q, want ws1", got)
	}
}

// --- contextTitle tests ---

func TestContextTitle_Workspace(t *testing.T) {
	m := Model{overlayState: overlayState{contextMode: contextWorkspace}}
	if got := m.contextTitle(); got != "Workspaces" {
		t.Errorf("contextTitle() = %q, want Workspaces", got)
	}
}

func TestContextTitle_Board(t *testing.T) {
	m := Model{overlayState: overlayState{contextMode: contextBoard}}
	if got := m.contextTitle(); got != "Boards" {
		t.Errorf("contextTitle() = %q, want Boards", got)
	}
}

// --- contextNameByID tests ---

func TestContextNameByID_WorkspaceFound(t *testing.T) {
	m := Model{
		overlayState: overlayState{contextMode: contextWorkspace},
		workspaces:   []domain.Workspace{{ID: "ws1", Name: "Alpha"}},
	}
	if got := m.contextNameByID("ws1"); got != "Alpha" {
		t.Errorf("contextNameByID() = %q, want Alpha", got)
	}
}

func TestContextNameByID_WorkspaceNotFound(t *testing.T) {
	m := Model{overlayState: overlayState{contextMode: contextWorkspace}, workspaces: []domain.Workspace{}}
	if got := m.contextNameByID("ws1"); got != "" {
		t.Errorf("contextNameByID() = %q, want empty", got)
	}
}

func TestContextNameByID_BoardFound(t *testing.T) {
	m := Model{
		overlayState: overlayState{contextMode: contextBoard},
		boards:       []domain.Board{{ID: "b1", Name: "Main"}},
	}
	if got := m.contextNameByID("b1"); got != "Main" {
		t.Errorf("contextNameByID() = %q, want Main", got)
	}
}

// --- openContextPanel / closeContextPanel tests ---

func TestOpenContextPanel_SetsState(t *testing.T) {
	m := Model{
		contextFilter:    newTaskFormInput("Filter", "", 128),
		contextEditInput: newTaskFormInput("Name", "", 256),
	}
	m.openContextPanel(contextBoard)
	if !m.showContexts {
		t.Error("expected showContexts to be true")
	}
	if m.contextMode != contextBoard {
		t.Errorf("contextMode = %v, want contextBoard", m.contextMode)
	}
	if m.contextSelected != 0 {
		t.Errorf("contextSelected = %d, want 0", m.contextSelected)
	}
}

func TestCloseContextPanel_ClearsState(t *testing.T) {
	m := Model{
		overlayState:     overlayState{showContexts: true, contextEditMode: contextEditCreate},
		contextFilter:    newTaskFormInput("Filter", "", 128),
		contextEditInput: newTaskFormInput("Name", "", 256),
	}
	m.closeContextPanel()
	if m.showContexts {
		t.Error("expected showContexts to be false")
	}
	if m.contextEditMode != contextEditNone {
		t.Errorf("contextEditMode = %v, want contextEditNone", m.contextEditMode)
	}
}

// --- boardColumnsOrderForm tests ---

func TestBoardColumnsOrderForm_ClampSelection_Empty(t *testing.T) {
	f := &boardColumnsOrderForm{columns: []domain.Column{}}
	f.clampSelection()
	if f.selected != 0 {
		t.Errorf("selected = %d, want 0", f.selected)
	}
}

func TestBoardColumnsOrderForm_MoveSelected(t *testing.T) {
	f := &boardColumnsOrderForm{
		columns:  []domain.Column{{ID: "c1", Position: 1}, {ID: "c2", Position: 2}},
		selected: 0,
	}
	if !f.moveSelected(1) {
		t.Error("expected moveSelected to return true")
	}
	if f.columns[0].ID != "c2" || f.columns[1].ID != "c1" {
		t.Errorf("columns = %v, want reversed", f.columns)
	}
	if f.columns[0].Position != 1 || f.columns[1].Position != 2 {
		t.Errorf("positions not updated: %v", f.columns)
	}
}

func TestBoardColumnsOrderForm_MoveSelected_OutOfBounds(t *testing.T) {
	f := &boardColumnsOrderForm{
		columns:  []domain.Column{{ID: "c1"}},
		selected: 0,
	}
	if f.moveSelected(1) {
		t.Error("expected moveSelected to return false for single column")
	}
}

func TestBoardColumnsOrderForm_OrderedColumnIDs(t *testing.T) {
	f := &boardColumnsOrderForm{
		columns: []domain.Column{{ID: "c1"}, {ID: "c2"}, {ID: "c3"}},
	}
	ids := f.orderedColumnIDs()
	if len(ids) != 3 || ids[0] != "c1" || ids[1] != "c2" || ids[2] != "c3" {
		t.Errorf("orderedColumnIDs() = %v", ids)
	}
}

// --- boardCreateForm tests ---

func TestBoardCreateForm_Fields(t *testing.T) {
	f := &boardCreateForm{
		boardName: newTaskFormInput("Board name", "", 128),
		columns: []boardColumnFormRow{
			newBoardColumnRow(1, "Todo", "#60A5FA"),
		},
	}
	fields := f.fields()
	if len(fields) != 3 { // boardName + 2 per column
		t.Errorf("len(fields) = %d, want 3", len(fields))
	}
}

func TestBoardCreateForm_MoveFocus(t *testing.T) {
	f := &boardCreateForm{
		boardName: newTaskFormInput("Board name", "", 128),
		columns: []boardColumnFormRow{
			newBoardColumnRow(1, "Todo", "#60A5FA"),
		},
	}
	f.setFocus(0)
	if f.focus != 0 {
		t.Errorf("focus = %d, want 0", f.focus)
	}
	f.moveFocus(1)
	if f.focus != 1 {
		t.Errorf("focus = %d, want 1", f.focus)
	}
}

func TestBoardCreateForm_AddColumn(t *testing.T) {
	f := &boardCreateForm{
		boardName: newTaskFormInput("Board name", "", 128),
		columns: []boardColumnFormRow{
			newBoardColumnRow(1, "Todo", "#60A5FA"),
		},
	}
	f.setFocus(0)
	f.addColumn()
	if len(f.columns) != 2 {
		t.Errorf("len(columns) = %d, want 2", len(f.columns))
	}
}

func TestBoardCreateForm_RemoveFocusedColumn(t *testing.T) {
	f := &boardCreateForm{
		boardName: newTaskFormInput("Board name", "", 128),
		columns: []boardColumnFormRow{
			newBoardColumnRow(1, "Todo", "#60A5FA"),
			newBoardColumnRow(2, "Doing", "#F59E0B"),
		},
	}
	f.setFocus(1) // focus on first column name
	if !f.removeFocusedColumn() {
		t.Error("expected removeFocusedColumn to return true")
	}
	if len(f.columns) != 1 {
		t.Errorf("len(columns) = %d, want 1", len(f.columns))
	}
}

func TestBoardCreateForm_RemoveFocusedColumn_LastRemaining(t *testing.T) {
	f := &boardCreateForm{
		boardName: newTaskFormInput("Board name", "", 128),
		columns: []boardColumnFormRow{
			newBoardColumnRow(1, "Todo", "#60A5FA"),
		},
	}
	f.setFocus(1)
	if f.removeFocusedColumn() {
		t.Error("expected removeFocusedColumn to return false for last column")
	}
}

func TestBoardCreateForm_CycleFocusedColor(t *testing.T) {
	f := &boardCreateForm{
		boardName: newTaskFormInput("Board name", "", 128),
		columns: []boardColumnFormRow{
			newBoardColumnRow(1, "Todo", "#60A5FA"),
		},
	}
	f.setFocus(2) // focus on color field
	if !f.cycleFocusedColor(1) {
		t.Error("expected cycleFocusedColor to return true")
	}
	val := f.columns[0].color.Value()
	if val == "#60A5FA" {
		t.Error("expected color to change after cycle")
	}
}

// --- beginContextCreate tests ---

func TestBeginContextCreate_Workspace(t *testing.T) {
	m := Model{
		overlayState:     overlayState{contextMode: contextWorkspace},
		contextEditInput: newTaskFormInput("Name", "", 256),
	}
	m.beginContextCreate()
	if m.contextEditMode != contextEditCreate {
		t.Errorf("contextEditMode = %v, want contextEditCreate", m.contextEditMode)
	}
}

func TestBeginContextCreate_Board(t *testing.T) {
	m := Model{overlayState: overlayState{contextMode: contextBoard}}
	m.beginContextCreate()
	if m.boardForm == nil {
		t.Fatal("expected boardForm to be set")
	}
	if len(m.boardForm.columns) != 3 {
		t.Errorf("len(columns) = %d, want 3", len(m.boardForm.columns))
	}
}

// --- startBoardCreateForm / closeBoardCreateForm tests ---

func TestStartBoardCreateForm(t *testing.T) {
	m := Model{}
	m.startBoardCreateForm()
	if m.boardForm == nil {
		t.Fatal("expected boardForm to be set")
	}
	if m.boardForm.focus != 0 {
		t.Errorf("focus = %d, want 0", m.boardForm.focus)
	}
}

func TestCloseBoardCreateForm(t *testing.T) {
	m := Model{
		overlayState:  overlayState{boardForm: &boardCreateForm{}},
		contextFilter: newTaskFormInput("Filter", "", 128),
	}
	m.closeBoardCreateForm()
	if m.boardForm != nil {
		t.Error("expected boardForm to be nil")
	}
}

// --- beginBoardColumnsReorder tests ---

func TestBeginBoardColumnsReorder_NilService(t *testing.T) {
	m := Model{overlayState: overlayState{contextMode: contextBoard}, contextService: nil}
	if err := m.beginBoardColumnsReorder(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBeginBoardColumnsReorder_WrongMode(t *testing.T) {
	repo := &mockSetupRepo{columns: []domain.Column{{ID: "c1", Name: "Todo", Position: 1}}}
	m := newMockModelWithContextService(repo)
	m.overlayState.contextMode = contextWorkspace
	if err := m.beginBoardColumnsReorder(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBeginBoardColumnsReorder_NoSelection(t *testing.T) {
	repo := &mockSetupRepo{columns: []domain.Column{{ID: "c1", Name: "Todo", Position: 1}}}
	m := newMockModelWithContextService(repo)
	m.overlayState.contextMode = contextBoard
	m.boards = []domain.Board{}
	if err := m.beginBoardColumnsReorder(); err == nil {
		t.Error("expected error for no selected board")
	}
}

func TestBeginBoardColumnsReorder_Success(t *testing.T) {
	repo := &mockSetupRepo{columns: []domain.Column{{ID: "c1", Name: "Todo", Position: 1}}}
	m := newMockModelWithContextService(repo)
	m.contextMode = contextBoard
	m.boards = []domain.Board{{ID: "b1", Name: "Main"}}
	m.contextSelected = 0
	if err := m.beginBoardColumnsReorder(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.boardOrder == nil {
		t.Fatal("expected boardOrder to be set")
	}
	if m.boardOrder.boardID != "b1" {
		t.Errorf("boardID = %q, want b1", m.boardOrder.boardID)
	}
}

// --- closeBoardColumnsReorder tests ---

func TestCloseBoardColumnsReorder(t *testing.T) {
	m := Model{
		overlayState:  overlayState{boardOrder: &boardColumnsOrderForm{}},
		contextFilter: newTaskFormInput("Filter", "", 128),
	}
	m.closeBoardColumnsReorder()
	if m.boardOrder != nil {
		t.Error("expected boardOrder to be nil")
	}
}

// --- submitBoardColumnsReorder tests ---

func TestSubmitBoardColumnsReorder_NilService(t *testing.T) {
	m := Model{overlayState: overlayState{boardOrder: &boardColumnsOrderForm{boardID: "b1", columns: []domain.Column{{ID: "c1"}}}}}
	if err := m.submitBoardColumnsReorder(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- submitBoardCreateForm tests ---

func TestSubmitBoardCreateForm_NilForm(t *testing.T) {
	repo := &mockSetupRepo{}
	m := newMockModelWithContextService(repo)
	if err := m.submitBoardCreateForm(); err == nil {
		t.Error("expected error for nil form")
	}
}

func TestSubmitBoardCreateForm_EmptyName(t *testing.T) {
	repo := &mockSetupRepo{}
	m := newMockModelWithContextService(repo)
	m.startBoardCreateForm()
	m.boardForm.boardName.SetValue("")
	if err := m.submitBoardCreateForm(); err == nil {
		t.Error("expected error for empty board name")
	}
}

func TestSubmitBoardCreateForm_InvalidColor(t *testing.T) {
	repo := &mockSetupRepo{}
	m := newMockModelWithContextService(repo)
	m.startBoardCreateForm()
	m.boardForm.columns[0].color.SetValue("BAD")
	if err := m.submitBoardCreateForm(); err == nil {
		t.Error("expected error for invalid color")
	}
}

// --- beginContextRename tests ---

func TestBeginContextRename_NoSelection(t *testing.T) {
	m := Model{
		overlayState:     overlayState{contextMode: contextWorkspace},
		workspaces:       []domain.Workspace{},
		contextEditInput: newTaskFormInput("Name", "", 256),
	}
	m.beginContextRename()
	if m.contextEditMode != contextEditNone {
		t.Errorf("contextEditMode = %v, want contextEditNone", m.contextEditMode)
	}
}

func TestBeginContextRename_WithSelection(t *testing.T) {
	m := Model{
		overlayState:     overlayState{contextMode: contextWorkspace, contextSelected: 0},
		workspaces:       []domain.Workspace{{ID: "ws1", Name: "Alpha"}},
		contextEditInput: newTaskFormInput("Name", "", 256),
	}
	m.beginContextRename()
	if m.contextEditMode != contextEditRename {
		t.Errorf("contextEditMode = %v, want contextEditRename", m.contextEditMode)
	}
	if m.contextEditInput.Value() != "Alpha" {
		t.Errorf("value = %q, want Alpha", m.contextEditInput.Value())
	}
}

// --- updateContextPanel tests ---

func TestUpdateContextPanel_CancelClosesPanel(t *testing.T) {
	m := Model{overlayState: overlayState{showContexts: true}, keys: newKeyMap()}
	updated, cmd := m.updateContextPanel(tea.KeyMsg{Type: tea.KeyEscape})
	_ = cmd
	um := updated.(Model)
	if um.showContexts {
		t.Error("expected showContexts to be false after cancel")
	}
}

func TestUpdateContextPanel_UpMovesSelection(t *testing.T) {
	m := Model{
		overlayState: overlayState{showContexts: true, contextMode: contextWorkspace, contextSelected: 1},
		workspaces:   []domain.Workspace{{ID: "ws1"}, {ID: "ws2"}},
		keys:         newKeyMap(),
	}
	updated, _ := m.updateContextPanel(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	um := updated.(Model)
	if um.contextSelected != 0 {
		t.Errorf("contextSelected = %d, want 0", um.contextSelected)
	}
}

func TestUpdateContextPanel_DownMovesSelection(t *testing.T) {
	m := Model{
		overlayState: overlayState{showContexts: true, contextMode: contextWorkspace, contextSelected: 0},
		workspaces:   []domain.Workspace{{ID: "ws1"}, {ID: "ws2"}},
		keys:         newKeyMap(),
	}
	updated, _ := m.updateContextPanel(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	um := updated.(Model)
	if um.contextSelected != 1 {
		t.Errorf("contextSelected = %d, want 1", um.contextSelected)
	}
}

func TestUpdateContextPanel_WindowSize(t *testing.T) {
	m := Model{overlayState: overlayState{showContexts: true}, width: 10, height: 10}
	updated, cmd := m.updateContextPanel(tea.WindowSizeMsg{Width: 100, Height: 50})
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	um := updated.(Model)
	if um.width != 100 || um.height != 50 {
		t.Errorf("width=%d height=%d, want 100,50", um.width, um.height)
	}
}

func TestUpdateContextPanel_nKeyCreates(t *testing.T) {
	m := Model{
		overlayState:     overlayState{showContexts: true, contextMode: contextWorkspace},
		keys:             newKeyMap(),
		contextEditInput: newTaskFormInput("Name", "", 256),
	}
	updated, _ := m.updateContextPanel(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	um := updated.(Model)
	if um.contextEditMode != contextEditCreate {
		t.Errorf("contextEditMode = %v, want contextEditCreate", um.contextEditMode)
	}
}

func TestUpdateContextPanel_rKeyRenames(t *testing.T) {
	m := Model{
		overlayState:     overlayState{showContexts: true, contextMode: contextWorkspace, contextSelected: 0},
		workspaces:       []domain.Workspace{{ID: "ws1", Name: "Alpha"}},
		keys:             newKeyMap(),
		contextEditInput: newTaskFormInput("Name", "", 256),
	}
	updated, _ := m.updateContextPanel(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	um := updated.(Model)
	if um.contextEditMode != contextEditRename {
		t.Errorf("contextEditMode = %v, want contextEditRename", um.contextEditMode)
	}
}

func TestUpdateContextPanel_ConfirmSwitchesWorkspace(t *testing.T) {
	repo := &mockSetupRepo{
		workspaces: []domain.Workspace{{ID: "ws1", Name: "Alpha"}},
		boards:     []domain.Board{{ID: "b1", Name: "Main", WorkspaceID: "ws1"}},
		columns:    []domain.Column{{ID: "c1", Name: "Todo", Position: 1}},
	}
	m := newMockModelWithContextService(repo)
	m.overlayState = overlayState{showContexts: true, contextMode: contextWorkspace, contextSelected: 0}
	m.workspaces = []domain.Workspace{{ID: "ws1", Name: "Alpha"}}
	updated, cmd := m.updateContextPanel(tea.KeyMsg{Type: tea.KeyEnter})
	_ = cmd
	um := updated.(Model)
	if um.showContexts {
		t.Error("expected panel to close after confirm")
	}
}

func TestUpdateContextPanel_BoardOrderCancel(t *testing.T) {
	m := Model{
		overlayState:  overlayState{showContexts: true, contextMode: contextBoard, boardOrder: &boardColumnsOrderForm{}},
		keys:          newKeyMap(),
		contextFilter: newTaskFormInput("Filter", "", 128),
	}
	updated, _ := m.updateContextPanel(tea.KeyMsg{Type: tea.KeyEscape})
	um := updated.(Model)
	if um.boardOrder != nil {
		t.Error("expected boardOrder to be nil after cancel")
	}
}

func TestUpdateContextPanel_BoardOrderMoveSelection(t *testing.T) {
	m := Model{
		overlayState: overlayState{showContexts: true, contextMode: contextBoard, boardOrder: &boardColumnsOrderForm{
			columns:  []domain.Column{{ID: "c1"}, {ID: "c2"}},
			selected: 0,
		}},
		keys: newKeyMap(),
	}
	updated, _ := m.updateContextPanel(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	um := updated.(Model)
	if um.boardOrder.selected != 1 {
		t.Errorf("selected = %d, want 1", um.boardOrder.selected)
	}
}

func TestUpdateContextPanel_BoardFormCancel(t *testing.T) {
	m := Model{
		overlayState:  overlayState{showContexts: true, contextMode: contextBoard, boardForm: &boardCreateForm{boardName: newTaskFormInput("Board name", "", 128)}},
		keys:          newKeyMap(),
		contextFilter: newTaskFormInput("Filter", "", 128),
	}
	updated, _ := m.updateContextPanel(tea.KeyMsg{Type: tea.KeyEscape})
	um := updated.(Model)
	if um.boardForm != nil {
		t.Error("expected boardForm to be nil after cancel")
	}
}

func TestUpdateContextPanel_BoardFormTabMovesFocus(t *testing.T) {
	m := Model{
		overlayState: overlayState{showContexts: true, contextMode: contextBoard, boardForm: &boardCreateForm{
			boardName: newTaskFormInput("Board name", "", 128),
			columns:   []boardColumnFormRow{newBoardColumnRow(1, "Todo", "#60A5FA")},
		}},
		keys: newKeyMap(),
	}
	m.boardForm.setFocus(0)
	updated, _ := m.updateContextPanel(tea.KeyMsg{Type: tea.KeyTab})
	um := updated.(Model)
	if um.boardForm.focus != 1 {
		t.Errorf("focus = %d, want 1", um.boardForm.focus)
	}
}

func TestUpdateContextPanel_EditModeCancel(t *testing.T) {
	m := Model{
		overlayState:  overlayState{showContexts: true, contextMode: contextWorkspace, contextEditMode: contextEditCreate},
		keys:          newKeyMap(),
		contextFilter: newTaskFormInput("Filter", "", 128),
	}
	updated, _ := m.updateContextPanel(tea.KeyMsg{Type: tea.KeyEscape})
	um := updated.(Model)
	if um.contextEditMode != contextEditNone {
		t.Errorf("contextEditMode = %v, want contextEditNone", um.contextEditMode)
	}
}

// --- renderContextPanel tests ---

func TestRenderContextPanel_NotEmpty(t *testing.T) {
	m := Model{
		overlayState: overlayState{showContexts: true, contextMode: contextWorkspace, contextSelected: 0},
		workspaces:   []domain.Workspace{{ID: "ws1", Name: "Alpha"}},
		width:        80,
		height:       24,
	}
	out := m.renderContextPanel("")
	if out == "" {
		t.Error("expected non-empty render output")
	}
}

func TestRenderBoardColumnsOrderContextPanel(t *testing.T) {
	m := Model{
		overlayState: overlayState{boardOrder: &boardColumnsOrderForm{
			boardName: "Main",
			columns:   []domain.Column{{ID: "c1", Name: "Todo", Color: "#60A5FA"}},
			selected:  0,
		}},
		width:  80,
		height: 24,
	}
	out := m.renderBoardColumnsOrderContextPanel(60, 20)
	if out == "" {
		t.Error("expected non-empty render output")
	}
}

func TestRenderBoardCreateContextPanel(t *testing.T) {
	m := Model{
		overlayState: overlayState{boardForm: &boardCreateForm{
			boardName: newTaskFormInput("Board name", "Test", 128),
			columns:   []boardColumnFormRow{newBoardColumnRow(1, "Todo", "#60A5FA")},
		}},
		width:  80,
		height: 24,
	}
	out := m.renderBoardCreateContextPanel(60, 20)
	if out == "" {
		t.Error("expected non-empty render output")
	}
}
