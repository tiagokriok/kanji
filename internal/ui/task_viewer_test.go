package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tiagokriok/kanji/internal/domain"
)

// --- scroll helper tests ---

func TestScrollUp_BoundedAtZero(t *testing.T) {
	if got := scrollUp(0); got != 0 {
		t.Errorf("scrollUp(0) = %d, want 0", got)
	}
}

func TestScrollUp_Decrements(t *testing.T) {
	if got := scrollUp(5); got != 4 {
		t.Errorf("scrollUp(5) = %d, want 4", got)
	}
}

func TestScrollDown_BoundedAtMax(t *testing.T) {
	if got := scrollDown(5, 5); got != 5 {
		t.Errorf("scrollDown(5, 5) = %d, want 5", got)
	}
}

func TestScrollDown_Increments(t *testing.T) {
	if got := scrollDown(2, 10); got != 3 {
		t.Errorf("scrollDown(2, 10) = %d, want 3", got)
	}
}

func TestPageUp_BoundedAtZero(t *testing.T) {
	if got := pageUp(0, 4); got != 0 {
		t.Errorf("pageUp(0, 4) = %d, want 0", got)
	}
}

func TestPageUp_HalvesViewport(t *testing.T) {
	if got := pageUp(10, 4); got != 8 {
		t.Errorf("pageUp(10, 4) = %d, want 8", got)
	}
}

func TestPageUp_ClampedToZero(t *testing.T) {
	if got := pageUp(1, 4); got != 0 {
		t.Errorf("pageUp(1, 4) = %d, want 0", got)
	}
}

func TestPageDown_BoundedAtMax(t *testing.T) {
	if got := pageDown(10, 10, 4); got != 10 {
		t.Errorf("pageDown(10, 10, 4) = %d, want 10", got)
	}
}

func TestPageDown_HalvesViewport(t *testing.T) {
	if got := pageDown(2, 20, 4); got != 4 {
		t.Errorf("pageDown(2, 20, 4) = %d, want 4", got)
	}
}

func TestPageDown_ClampedToMax(t *testing.T) {
	if got := pageDown(18, 20, 4); got != 20 {
		t.Errorf("pageDown(18, 20, 4) = %d, want 20", got)
	}
}

// --- open / close / transition tests ---

func TestOpenTaskViewer_NoTask(t *testing.T) {
	m := Model{tasks: []domain.Task{}}
	cmd := m.openTaskViewer()
	if cmd != nil {
		t.Error("expected nil cmd when no task")
	}
	if m.showTaskView {
		t.Error("expected showTaskView to be false")
	}
}

func TestOpenTaskViewer_WithTask(t *testing.T) {
	m := Model{
		tasks:    []domain.Task{{ID: "t1", Title: "Task"}},
		selected: 0,
	}
	cmd := m.openTaskViewer()
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
	if !m.showTaskView {
		t.Error("expected showTaskView to be true")
	}
	if m.viewTaskID != "t1" {
		t.Errorf("viewTaskID = %q, want %q", m.viewTaskID, "t1")
	}
	if m.viewDescScroll != 0 {
		t.Errorf("viewDescScroll = %d, want 0", m.viewDescScroll)
	}
	if m.comments != nil {
		t.Error("expected comments to be cleared")
	}
}

func TestOpenTaskViewerByID_Empty(t *testing.T) {
	m := Model{}
	cmd := m.openTaskViewerByID("  ")
	if cmd != nil {
		t.Error("expected nil cmd for empty id")
	}
}

func TestCloseTaskViewer(t *testing.T) {
	m := Model{overlayState: overlayState{showTaskView: true, viewTaskID: "t1", viewDescScroll: 3}}
	m.closeTaskViewer()
	if m.showTaskView {
		t.Error("expected showTaskView to be false")
	}
	if m.viewTaskID != "" {
		t.Errorf("viewTaskID = %q, want empty", m.viewTaskID)
	}
	if m.viewDescScroll != 0 {
		t.Errorf("viewDescScroll = %d, want 0", m.viewDescScroll)
	}
}

func TestSetTaskViewerReturn(t *testing.T) {
	m := Model{}
	m.setTaskViewerReturn("t99")
	if !m.returnTaskView {
		t.Error("expected returnTaskView to be true")
	}
	if m.returnTaskID != "t99" {
		t.Errorf("returnTaskID = %q, want %q", m.returnTaskID, "t99")
	}
}

func TestClearTaskViewerReturn(t *testing.T) {
	m := Model{overlayState: overlayState{returnTaskView: true, returnTaskID: "t99"}}
	m.clearTaskViewerReturn()
	if m.returnTaskView {
		t.Error("expected returnTaskView to be false")
	}
	if m.returnTaskID != "" {
		t.Errorf("returnTaskID = %q, want empty", m.returnTaskID)
	}
}

func TestViewerTask_ByID(t *testing.T) {
	m := Model{
		overlayState: overlayState{viewTaskID: "t2"},
		tasks:        []domain.Task{{ID: "t1"}, {ID: "t2", Title: "Two"}},
		selected:     0,
	}
	task, ok := m.viewerTask()
	if !ok {
		t.Fatal("expected task to be found")
	}
	if task.ID != "t2" {
		t.Errorf("task.ID = %q, want %q", task.ID, "t2")
	}
}

func TestViewerTask_FallbackToCurrent(t *testing.T) {
	m := Model{
		tasks:    []domain.Task{{ID: "t1", Title: "One"}},
		selected: 0,
	}
	task, ok := m.viewerTask()
	if !ok {
		t.Fatal("expected task to be found")
	}
	if task.ID != "t1" {
		t.Errorf("task.ID = %q, want %q", task.ID, "t1")
	}
}

// --- updateTaskViewer integration tests ---

func TestUpdateTaskViewer_CancelCloses(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{showTaskView: true, viewTaskID: "t1"},
	}
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	updated := newM.(Model)
	if updated.showTaskView {
		t.Error("expected task viewer to be closed")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskViewer_ConfirmCloses(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{showTaskView: true, viewTaskID: "t1"},
	}
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := newM.(Model)
	if updated.showTaskView {
		t.Error("expected task viewer to be closed")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskViewer_QuitCloses(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{showTaskView: true, viewTaskID: "t1"},
	}
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	updated := newM.(Model)
	if updated.showTaskView {
		t.Error("expected task viewer to be closed")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskViewer_EditTitleEntersForm(t *testing.T) {
	m := Model{
		keys: newKeyMap(),
		overlayState: overlayState{
			showTaskView: true,
			viewTaskID:   "t1",
		},
		tasks: []domain.Task{
			{ID: "t1", Title: "Task", ColumnID: strPtr("c1")},
		},
		columns: []domain.Column{{ID: "c1", Name: "Todo"}},
	}
	m.textInput = newTaskFormInput("Title", "", 512)

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	updated := newM.(Model)

	if updated.showTaskView {
		t.Error("expected task viewer to be closed")
	}
	if !updated.returnTaskView {
		t.Error("expected returnTaskView to be set")
	}
	if updated.returnTaskID != "t1" {
		t.Errorf("returnTaskID = %q, want %q", updated.returnTaskID, "t1")
	}
	if updated.inputMode != inputTaskForm {
		t.Errorf("inputMode = %v, want inputTaskForm", updated.inputMode)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestUpdateTaskViewer_AddCommentEntersMode(t *testing.T) {
	m := Model{
		keys: newKeyMap(),
		overlayState: overlayState{
			showTaskView: true,
			viewTaskID:   "t1",
		},
		tasks: []domain.Task{{ID: "t1", Title: "Task"}},
	}
	m.textInput = newTaskFormInput("Comment", "", 512)

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	updated := newM.(Model)

	if updated.showTaskView {
		t.Error("expected task viewer to be closed")
	}
	if !updated.returnTaskView {
		t.Error("expected returnTaskView to be set")
	}
	if updated.returnTaskID != "t1" {
		t.Errorf("returnTaskID = %q, want %q", updated.returnTaskID, "t1")
	}
	if updated.inputMode != inputAddComment {
		t.Errorf("inputMode = %v, want inputAddComment", updated.inputMode)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestUpdateTaskViewer_ScrollUp(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{showTaskView: true, viewTaskID: "t1", viewDescScroll: 3},
		tasks:        []domain.Task{{ID: "t1", Title: "Task"}},
		width:        120,
		height:       40,
	}
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	updated := newM.(Model)
	if updated.viewDescScroll != 2 {
		t.Errorf("viewDescScroll = %d, want 2", updated.viewDescScroll)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskViewer_ScrollUp_Bounded(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{showTaskView: true, viewTaskID: "t1", viewDescScroll: 0},
		tasks:        []domain.Task{{ID: "t1", Title: "Task"}},
		width:        120,
		height:       40,
	}
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	updated := newM.(Model)
	if updated.viewDescScroll != 0 {
		t.Errorf("viewDescScroll = %d, want 0", updated.viewDescScroll)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskViewer_ScrollDown(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{showTaskView: true, viewTaskID: "t1", viewDescScroll: 0},
		tasks:        []domain.Task{{ID: "t1", Title: "Task", DescriptionMD: longDescription(50)}},
		width:        120,
		height:       20,
	}
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated := newM.(Model)
	if updated.viewDescScroll != 1 {
		t.Errorf("viewDescScroll = %d, want 1", updated.viewDescScroll)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskViewer_PageUp(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{showTaskView: true, viewTaskID: "t1", viewDescScroll: 10},
		tasks:        []domain.Task{{ID: "t1", Title: "Task"}},
		width:        120,
		height:       40,
	}
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	updated := newM.(Model)
	if updated.viewDescScroll >= 10 {
		t.Errorf("viewDescScroll = %d, expected less than 10", updated.viewDescScroll)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskViewer_PageDown(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{showTaskView: true, viewTaskID: "t1", viewDescScroll: 0},
		tasks:        []domain.Task{{ID: "t1", Title: "Task", DescriptionMD: longDescription(50)}},
		width:        120,
		height:       20,
	}
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	updated := newM.(Model)
	if updated.viewDescScroll <= 0 {
		t.Errorf("viewDescScroll = %d, expected greater than 0", updated.viewDescScroll)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskViewer_Home(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{showTaskView: true, viewTaskID: "t1", viewDescScroll: 5},
		tasks:        []domain.Task{{ID: "t1", Title: "Task"}},
		width:        120,
		height:       40,
	}
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyHome})
	updated := newM.(Model)
	if updated.viewDescScroll != 0 {
		t.Errorf("viewDescScroll = %d, want 0", updated.viewDescScroll)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskViewer_End(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{showTaskView: true, viewTaskID: "t1", viewDescScroll: 0},
		tasks:        []domain.Task{{ID: "t1", Title: "Task", DescriptionMD: longDescription(50)}},
		width:        120,
		height:       20,
	}
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	updated := newM.(Model)
	if updated.viewDescScroll <= 0 {
		t.Errorf("viewDescScroll = %d, expected greater than 0", updated.viewDescScroll)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskViewer_WindowSize(t *testing.T) {
	m := Model{
		overlayState: overlayState{showTaskView: true, viewTaskID: "t1"},
		width:        80,
		height:       24,
	}
	newM, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	updated := newM.(Model)
	if updated.width != 100 {
		t.Errorf("width = %d, want 100", updated.width)
	}
	if updated.height != 30 {
		t.Errorf("height = %d, want 30", updated.height)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskViewer_TasksLoaded(t *testing.T) {
	m := Model{
		overlayState: overlayState{showTaskView: true, viewTaskID: "t1"},
		tasks:        []domain.Task{{ID: "t1", Title: "Old"}},
		columns:      []domain.Column{{ID: "c1", Name: "Todo"}},
		width:        120,
		height:       40,
	}
	newM, cmd := m.Update(tasksLoadedMsg{tasks: []domain.Task{{ID: "t1", Title: "New"}}})
	updated := newM.(Model)
	if len(updated.tasks) != 1 || updated.tasks[0].Title != "New" {
		t.Error("expected tasks to be updated")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskViewer_TasksLoadedError(t *testing.T) {
	m := Model{
		overlayState: overlayState{showTaskView: true, viewTaskID: "t1"},
		width:        120,
		height:       40,
	}
	newM, cmd := m.Update(tasksLoadedMsg{err: errTest("fail")})
	updated := newM.(Model)
	if updated.err == nil {
		t.Error("expected err to be set")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskViewer_CommentsLoaded(t *testing.T) {
	m := Model{
		overlayState: overlayState{showTaskView: true, viewTaskID: "t1"},
		width:        120,
		height:       40,
	}
	newM, cmd := m.Update(commentsLoadedMsg{comments: []domain.Comment{{ID: "c1", BodyMD: "hi"}}})
	updated := newM.(Model)
	if len(updated.comments) != 1 {
		t.Error("expected comments to be set")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskViewer_CommentsLoadedError(t *testing.T) {
	m := Model{
		overlayState: overlayState{showTaskView: true, viewTaskID: "t1"},
		width:        120,
		height:       40,
	}
	newM, cmd := m.Update(commentsLoadedMsg{err: errTest("fail")})
	updated := newM.(Model)
	if updated.err == nil {
		t.Error("expected err to be set")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskViewer_EditTitleNoTask(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{showTaskView: true, viewTaskID: "missing"},
		tasks:        []domain.Task{},
		width:        120,
		height:       40,
	}
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	updated := newM.(Model)
	if !updated.showTaskView {
		t.Error("expected task viewer to remain open")
	}
	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskViewer_AddCommentNoTask(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{showTaskView: true, viewTaskID: "missing"},
		tasks:        []domain.Task{},
		width:        120,
		height:       40,
	}
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	updated := newM.(Model)
	if !updated.showTaskView {
		t.Error("expected task viewer to remain open")
	}
	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

// helpers

func strPtr(s string) *string {
	return &s
}

func longDescription(lines int) string {
	sb := "```\n"
	for i := 0; i < lines; i++ {
		sb += "line "
		if i < lines-1 {
			sb += "\n"
		}
	}
	sb += "\n```"
	return sb
}

type errTest string

func (e errTest) Error() string {
	return string(e)
}
