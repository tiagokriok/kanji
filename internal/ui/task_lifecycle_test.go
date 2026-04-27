package ui

import (
	"testing"

	"github.com/tiagokriok/kanji/internal/domain"
)

// --- handleCommentsLoaded tests ---

func TestHandleCommentsLoaded_Success(t *testing.T) {
	m := Model{comments: nil}
	updated, cmd := m.handleCommentsLoaded(commentsLoadedMsg{comments: []domain.Comment{{ID: "c1", BodyMD: "hi"}}})
	if len(updated.comments) != 1 {
		t.Errorf("comments len = %d, want 1", len(updated.comments))
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestHandleCommentsLoaded_Error(t *testing.T) {
	m := Model{}
	updated, cmd := m.handleCommentsLoaded(commentsLoadedMsg{err: errTest("boom")})
	if updated.err == nil {
		t.Error("expected err to be set")
	}
	if updated.statusLine != "boom" {
		t.Errorf("statusLine = %q, want %q", updated.statusLine, "boom")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

// --- handleTasksLoaded tests ---

func TestHandleTasksLoaded_Error(t *testing.T) {
	m := Model{}
	updated, cmd := m.handleTasksLoaded(tasksLoadedMsg{err: errTest("fail")}, true, true)
	if updated.err == nil {
		t.Error("expected err to be set")
	}
	if updated.statusLine != "fail" {
		t.Errorf("statusLine = %q, want %q", updated.statusLine, "fail")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestHandleTasksLoaded_RestoreKanbanAndRefreshDetails(t *testing.T) {
	m := Model{
		tasks:       []domain.Task{{ID: "t1", Title: "Old", ColumnID: strPtr("c1")}},
		selected:    0,
		columns:     []domain.Column{{ID: "c1", Name: "Todo"}},
		width:       80,
		height:      24,
		showDetails: true,
	}
	updated, cmd := m.handleTasksLoaded(tasksLoadedMsg{tasks: []domain.Task{{ID: "t2", Title: "New", ColumnID: strPtr("c1")}}}, true, true)
	if len(updated.tasks) != 1 || updated.tasks[0].ID != "t2" {
		t.Error("expected tasks to be updated")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd for details refresh")
	}
}

func TestHandleTasksLoaded_NoRestoreNoRefresh(t *testing.T) {
	m := Model{
		tasks:    []domain.Task{{ID: "t1", Title: "Old"}},
		selected: 0,
		columns:  []domain.Column{{ID: "c1", Name: "Todo"}},
		width:    80,
		height:   24,
	}
	updated, cmd := m.handleTasksLoaded(tasksLoadedMsg{tasks: []domain.Task{{ID: "t2", Title: "New"}}}, false, false)
	if len(updated.tasks) != 1 || updated.tasks[0].ID != "t2" {
		t.Error("expected tasks to be updated")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestHandleTasksLoaded_NoRestoreWithRefresh(t *testing.T) {
	m := Model{
		tasks:       []domain.Task{{ID: "t1", Title: "Old", ColumnID: strPtr("c1")}},
		selected:    0,
		columns:     []domain.Column{{ID: "c1", Name: "Todo"}},
		width:       80,
		height:      24,
		showDetails: true,
	}
	updated, cmd := m.handleTasksLoaded(tasksLoadedMsg{tasks: []domain.Task{{ID: "t2", Title: "New", ColumnID: strPtr("c1")}}}, false, true)
	if len(updated.tasks) != 1 || updated.tasks[0].ID != "t2" {
		t.Error("expected tasks to be updated")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd for details refresh")
	}
}

// --- loadCommentsIfVisible tests ---

func TestLoadCommentsIfVisible_ShowDetailsWithTask(t *testing.T) {
	m := Model{
		tasks:       []domain.Task{{ID: "t1", Title: "Task"}},
		selected:    0,
		showDetails: true,
	}
	cmd := m.loadCommentsIfVisible()
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestLoadCommentsIfVisible_ShowDetailsNoTask(t *testing.T) {
	m := Model{
		tasks:       []domain.Task{},
		showDetails: true,
	}
	cmd := m.loadCommentsIfVisible()
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestLoadCommentsIfVisible_HideDetails(t *testing.T) {
	m := Model{
		tasks:       []domain.Task{{ID: "t1", Title: "Task"}},
		selected:    0,
		showDetails: false,
		comments:    []domain.Comment{{ID: "c1"}},
	}
	cmd := m.loadCommentsIfVisible()
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	if m.comments == nil {
		t.Error("expected comments to be preserved when showDetails is false")
	}
}

// --- refreshDetails tests ---

func TestRefreshDetails_ShowDetailsWithTask(t *testing.T) {
	m := Model{
		tasks:       []domain.Task{{ID: "t1", Title: "Task"}},
		selected:    0,
		showDetails: true,
	}
	cmd := m.refreshDetails()
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestRefreshDetails_ShowDetailsNoTask(t *testing.T) {
	m := Model{
		tasks:       []domain.Task{},
		showDetails: true,
	}
	cmd := m.refreshDetails()
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	if m.comments != nil {
		t.Error("expected comments to be cleared")
	}
}

func TestRefreshDetails_HideDetails(t *testing.T) {
	m := Model{
		tasks:       []domain.Task{{ID: "t1", Title: "Task"}},
		selected:    0,
		showDetails: false,
		comments:    []domain.Comment{{ID: "c1"}},
	}
	cmd := m.refreshDetails()
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	if m.comments != nil {
		t.Error("expected comments to be cleared when showDetails is false")
	}
}

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
