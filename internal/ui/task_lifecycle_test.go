package ui

import (
	"testing"

	"github.com/tiagokriok/kanji/internal/domain"
)

// --- handleOpResult tests ---

func TestHandleOpResult_ErrorClearsReturn(t *testing.T) {
	m := Model{
		overlayState: overlayState{returnTaskView: true, returnTaskID: "t1"},
	}
	updated, cmd := m.handleOpResult(opResultMsg{err: errTest("fail")})
	if updated.err == nil {
		t.Error("expected err to be set")
	}
	if updated.statusLine != "fail" {
		t.Errorf("statusLine = %q, want %q", updated.statusLine, "fail")
	}
	if updated.returnTaskView {
		t.Error("expected returnTaskView to be cleared")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestHandleOpResult_KanbanSetsPending(t *testing.T) {
	m := Model{
		viewMode: viewKanban,
		columns:  []domain.Column{{ID: "c1", Name: "Todo"}, {ID: "c2", Name: "Doing"}},
	}
	updated, cmd := m.handleOpResult(opResultMsg{taskID: "t1", columnID: "c2"})
	if updated.pendingKanbanTaskID != "t1" {
		t.Errorf("pendingKanbanTaskID = %q, want %q", updated.pendingKanbanTaskID, "t1")
	}
	if updated.pendingKanbanColumnID != "c2" {
		t.Errorf("pendingKanbanColumnID = %q, want %q", updated.pendingKanbanColumnID, "c2")
	}
	if updated.activeColumn != 1 {
		t.Errorf("activeColumn = %d, want 1", updated.activeColumn)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestHandleOpResult_KanbanSkipsWhitespaceTaskID(t *testing.T) {
	m := Model{
		viewMode: viewKanban,
		columns:  []domain.Column{{ID: "c1", Name: "Todo"}},
	}
	updated, _ := m.handleOpResult(opResultMsg{taskID: "  ", columnID: "c1"})
	if updated.pendingKanbanTaskID != "" {
		t.Errorf("pendingKanbanTaskID = %q, want empty", updated.pendingKanbanTaskID)
	}
}

func TestHandleOpResult_KanbanSkipsWhitespaceColumnID(t *testing.T) {
	m := Model{
		viewMode: viewKanban,
		columns:  []domain.Column{{ID: "c1", Name: "Todo"}},
	}
	updated, _ := m.handleOpResult(opResultMsg{taskID: "t1", columnID: "  "})
	if updated.pendingKanbanColumnID != "" {
		t.Errorf("pendingKanbanColumnID = %q, want empty", updated.pendingKanbanColumnID)
	}
}

func TestHandleOpResult_ReturnToTaskViewer(t *testing.T) {
	m := Model{
		overlayState: overlayState{returnTaskView: true, returnTaskID: "t1"},
	}
	updated, cmd := m.handleOpResult(opResultMsg{})
	if updated.returnTaskView {
		t.Error("expected returnTaskView to be cleared")
	}
	if updated.returnTaskID != "" {
		t.Errorf("returnTaskID = %q, want empty", updated.returnTaskID)
	}
	if !updated.showTaskView {
		t.Error("expected showTaskView to be true")
	}
	if updated.viewTaskID != "t1" {
		t.Errorf("viewTaskID = %q, want %q", updated.viewTaskID, "t1")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestHandleOpResult_ReturnSkipsWhitespaceTaskID(t *testing.T) {
	m := Model{
		overlayState: overlayState{returnTaskView: true, returnTaskID: "  "},
	}
	updated, _ := m.handleOpResult(opResultMsg{})
	if !updated.returnTaskView {
		t.Error("expected returnTaskView to remain true")
	}
	if updated.showTaskView {
		t.Error("expected showTaskView to remain false")
	}
}

func TestHandleOpResult_PlainSuccessReloads(t *testing.T) {
	m := Model{statusLine: "working"}
	updated, cmd := m.handleOpResult(opResultMsg{status: "ok"})
	if updated.statusLine != "" {
		t.Errorf("statusLine = %q, want empty", updated.statusLine)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}
