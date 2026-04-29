package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tiagokriok/kanji/internal/domain"
)

func newTestModelWithWindowSize(width, height int) Model {
	m := newTestModelWithServices(&fakeTaskRepoForCommands{}, &fakeCommentRepoForCommands{})
	ta := textarea.New()
	ta.SetHeight(8)
	m.textArea = ta
	m.width = width
	m.height = height
	return m
}

// --- dispatchOverlayUpdate tests ---

func TestDispatchOverlayUpdate_NoOverlayReturnsFalse(t *testing.T) {
	m := Model{overlayState: overlayState{}}
	_, _, ok := m.dispatchOverlayUpdate(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if ok {
		t.Error("expected false when no overlay is active")
	}
}

func TestDispatchOverlayUpdate_TaskViewReturnsTrue(t *testing.T) {
	m := Model{
		overlayState: overlayState{showTaskView: true, viewTaskID: "t1"},
		tasks:        []domain.Task{{ID: "t1", Title: "Task"}},
		columns:      []domain.Column{{ID: "c1", Name: "Todo"}},
		width:        80,
		height:       24,
		keys:         newKeyMap(),
	}
	model, cmd, ok := m.dispatchOverlayUpdate(tea.KeyMsg{Type: tea.KeyEscape})
	if !ok {
		t.Fatal("expected true when task view overlay is active")
	}
	updated := model.(Model)
	if updated.showTaskView {
		t.Error("expected task view to be closed on escape")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestDispatchOverlayUpdate_KeybindsReturnsTrue(t *testing.T) {
	m := Model{
		overlayState: overlayState{showKeybinds: true},
		keys:         newKeyMap(),
		width:        80,
		height:       24,
	}
	model, _, ok := m.dispatchOverlayUpdate(tea.KeyMsg{Type: tea.KeyEscape})
	if !ok {
		t.Fatal("expected true when keybinds overlay is active")
	}
	updated := model.(Model)
	if updated.showKeybinds {
		t.Error("expected keybinds to be closed on escape")
	}
}

// --- dispatchGlobalMessage tests ---

func TestDispatchGlobalMessage_WindowSizeUpdatesDimensions(t *testing.T) {
	m := newTestModelWithWindowSize(10, 10)
	model, cmd, ok := m.dispatchGlobalMessage(tea.WindowSizeMsg{Width: 100, Height: 50})
	if !ok {
		t.Fatal("expected true for WindowSizeMsg")
	}
	updated := model.(Model)
	if updated.width != 100 {
		t.Errorf("width = %d, want 100", updated.width)
	}
	if updated.height != 50 {
		t.Errorf("height = %d, want 50", updated.height)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestDispatchGlobalMessage_ExecuteActionRoutes(t *testing.T) {
	m := newTestModelWithWindowSize(80, 24)
	model, cmd, ok := m.dispatchGlobalMessage(executeActionMsg{action: "quit"})
	if !ok {
		t.Fatal("expected true for executeActionMsg")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd for quit action")
	}
	msg := cmd()
	if _, isQuit := msg.(tea.QuitMsg); !isQuit {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
	_ = model
}

func TestDispatchGlobalMessage_TasksLoadedRoutes(t *testing.T) {
	m := newTestModelWithWindowSize(80, 24)
	m.columns = []domain.Column{{ID: "c1", Name: "Todo"}}
	m.showDetails = true
	model, cmd, ok := m.dispatchGlobalMessage(tasksLoadedMsg{tasks: []domain.Task{{ID: "t1", Title: "Task"}}})
	if !ok {
		t.Fatal("expected true for tasksLoadedMsg")
	}
	updated := model.(Model)
	if len(updated.tasks) != 1 {
		t.Errorf("tasks len = %d, want 1", len(updated.tasks))
	}
	if cmd == nil {
		t.Error("expected non-nil cmd for details refresh")
	}
}

func TestDispatchGlobalMessage_CommentsLoadedRoutes(t *testing.T) {
	m := newTestModelWithWindowSize(80, 24)
	model, cmd, ok := m.dispatchGlobalMessage(commentsLoadedMsg{comments: []domain.Comment{{ID: "c1", BodyMD: "hi"}}})
	if !ok {
		t.Fatal("expected true for commentsLoadedMsg")
	}
	updated := model.(Model)
	if len(updated.comments) != 1 {
		t.Errorf("comments len = %d, want 1", len(updated.comments))
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestDispatchGlobalMessage_OpResultRoutes(t *testing.T) {
	m := newTestModelWithWindowSize(80, 24)
	model, cmd, ok := m.dispatchGlobalMessage(opResultMsg{status: "ok"})
	if !ok {
		t.Fatal("expected true for opResultMsg")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd for reload")
	}
	_ = model
}

func TestDispatchGlobalMessage_UnknownReturnsFalse(t *testing.T) {
	m := newTestModelWithWindowSize(80, 24)
	_, _, ok := m.dispatchGlobalMessage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	if ok {
		t.Error("expected false for unhandled message type")
	}
}

// --- handleDeleteConfirmKey tests ---

func TestHandleDeleteConfirmKey_NotConfirmingReturnsFalse(t *testing.T) {
	m := newTestModelWithWindowSize(80, 24)
	m.confirmingDelete = false
	_, _, ok := m.handleDeleteConfirmKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if ok {
		t.Error("expected false when not confirming delete")
	}
}

func TestHandleDeleteConfirmKey_YesWithTaskDeletes(t *testing.T) {
	m := newTestModelWithWindowSize(80, 24)
	m.confirmingDelete = true
	m.statusLine = "delete task? y/n"
	colID := "c1"
	m.tasks = []domain.Task{{ID: "t1", Title: "Task", ColumnID: &colID}}
	m.selected = 0
	m.columns = []domain.Column{{ID: "c1", Name: "Todo"}}
	m.viewMode = viewKanban
	updated, cmd, ok := m.handleDeleteConfirmKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if !ok {
		t.Fatal("expected true when confirming delete")
	}
	if updated.confirmingDelete {
		t.Error("expected confirmingDelete to be cleared")
	}
	if updated.statusLine != "" {
		t.Errorf("statusLine = %q, want empty", updated.statusLine)
	}
	if cmd == nil {
		t.Error("expected non-nil delete cmd")
	}
}

func TestHandleDeleteConfirmKey_YesWithoutTaskCancels(t *testing.T) {
	m := newTestModelWithWindowSize(80, 24)
	m.confirmingDelete = true
	m.statusLine = "delete task? y/n"
	m.tasks = []domain.Task{}
	updated, cmd, ok := m.handleDeleteConfirmKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if !ok {
		t.Fatal("expected true when confirming delete")
	}
	if updated.confirmingDelete {
		t.Error("expected confirmingDelete to be cleared")
	}
	if cmd != nil {
		t.Error("expected nil cmd when no task")
	}
}

func TestHandleDeleteConfirmKey_OtherKeyCancels(t *testing.T) {
	m := newTestModelWithWindowSize(80, 24)
	m.confirmingDelete = true
	m.statusLine = "delete task? y/n"
	updated, cmd, ok := m.handleDeleteConfirmKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if !ok {
		t.Fatal("expected true when confirming delete")
	}
	if updated.confirmingDelete {
		t.Error("expected confirmingDelete to be cleared")
	}
	if updated.statusLine != "" {
		t.Errorf("statusLine = %q, want empty", updated.statusLine)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

// --- renderBaseView tests ---

func TestRenderBaseView_ListMode(t *testing.T) {
	m := newTestModelWithWindowSize(80, 24)
	m.workspaceName = "WS"
	m.boardName = "Board"
	m.columns = []domain.Column{{ID: "c1", Name: "Todo"}}
	m.tasks = []domain.Task{{ID: "t1", Title: "Task"}}
	m.viewMode = viewList
	rendered := m.renderBaseView()
	if rendered == "" {
		t.Error("expected non-empty rendered view")
	}
}

func TestRenderBaseView_KanbanMode(t *testing.T) {
	m := newTestModelWithWindowSize(80, 24)
	m.workspaceName = "WS"
	m.boardName = "Board"
	m.columns = []domain.Column{{ID: "c1", Name: "Todo"}}
	m.tasks = []domain.Task{{ID: "t1", Title: "Task"}}
	m.viewMode = viewKanban
	rendered := m.renderBaseView()
	if rendered == "" {
		t.Error("expected non-empty rendered view")
	}
}

// --- wrapOverlays tests ---

func TestWrapOverlays_NoOverlayReturnsBase(t *testing.T) {
	m := newTestModelWithWindowSize(80, 24)
	m.overlayState = overlayState{}
	base := "base-view"
	result := m.wrapOverlays(base)
	if result != base {
		t.Errorf("result = %q, want %q", result, base)
	}
}

func TestWrapOverlays_KeybindsOverlay(t *testing.T) {
	m := newTestModelWithWindowSize(80, 24)
	m.overlayState = overlayState{showKeybinds: true}
	m.keys = newKeyMap()
	base := "base-view"
	result := m.wrapOverlays(base)
	if result == base {
		t.Error("expected result to differ from base when keybinds overlay is active")
	}
}

func TestWrapOverlays_TaskViewOverlay(t *testing.T) {
	m := newTestModelWithWindowSize(80, 24)
	m.overlayState = overlayState{showTaskView: true, viewTaskID: "t1"}
	m.tasks = []domain.Task{{ID: "t1", Title: "Task"}}
	m.columns = []domain.Column{{ID: "c1", Name: "Todo"}}
	base := "base-view"
	result := m.wrapOverlays(base)
	if result == base {
		t.Error("expected result to differ from base when task view overlay is active")
	}
}
