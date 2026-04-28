package ui

import (
	"context"
	"testing"

	"github.com/tiagokriok/kanji/internal/domain"
)

// --- fakes that override List / ListByTask ---

type fakeTaskRepoForLoad struct {
	fakeTaskRepoForCommands
	listTasks []domain.Task
	listErr   error
}

func (r *fakeTaskRepoForLoad) List(ctx context.Context, filter domain.TaskFilter) ([]domain.Task, error) {
	return r.listTasks, r.listErr
}

type fakeCommentRepoForLoad struct {
	fakeCommentRepoForCommands
	listComments []domain.Comment
	listErr      error
}

func (r *fakeCommentRepoForLoad) ListByTask(ctx context.Context, taskID string) ([]domain.Comment, error) {
	return r.listComments, r.listErr
}

func newTestModelForLoad(taskRepo domain.TaskRepository, commentRepo domain.CommentRepository) Model {
	m := newTestModelWithServices(taskRepo, commentRepo)
	m.workspaceID = "ws-1"
	m.boardID = "board-1"
	return m
}

// --- loadTasksCmd ---

func TestLoadTasksCmd_Success(t *testing.T) {
	repo := &fakeTaskRepoForLoad{listTasks: []domain.Task{{ID: "t1", Title: "Task 1"}}}
	m := newTestModelForLoad(repo, &fakeCommentRepoForLoad{})
	m.titleFilter = "search"
	m.columnFilter = "col-1"

	cmd := m.loadTasksCmd()
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	loaded, ok := msg.(tasksLoadedMsg)
	if !ok {
		t.Fatalf("expected tasksLoadedMsg, got %T", msg)
	}
	if loaded.err != nil {
		t.Errorf("unexpected err: %v", loaded.err)
	}
	if len(loaded.tasks) != 1 || loaded.tasks[0].ID != "t1" {
		t.Errorf("tasks = %v, want [{t1}]", loaded.tasks)
	}
}

func TestLoadTasksCmd_Error(t *testing.T) {
	repo := &fakeTaskRepoForLoad{listErr: errTest("list failed")}
	m := newTestModelForLoad(repo, &fakeCommentRepoForLoad{})

	cmd := m.loadTasksCmd()
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	loaded, ok := msg.(tasksLoadedMsg)
	if !ok {
		t.Fatalf("expected tasksLoadedMsg, got %T", msg)
	}
	if loaded.err == nil || loaded.err.Error() != "list failed" {
		t.Errorf("err = %v, want list failed", loaded.err)
	}
}

// --- loadCommentsCmd ---

func TestLoadCommentsCmd_Success(t *testing.T) {
	repo := &fakeCommentRepoForLoad{listComments: []domain.Comment{{ID: "c1", BodyMD: "hello"}}}
	m := newTestModelForLoad(&fakeTaskRepoForLoad{}, repo)

	cmd := m.loadCommentsCmd("task-1")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	loaded, ok := msg.(commentsLoadedMsg)
	if !ok {
		t.Fatalf("expected commentsLoadedMsg, got %T", msg)
	}
	if loaded.err != nil {
		t.Errorf("unexpected err: %v", loaded.err)
	}
	if len(loaded.comments) != 1 || loaded.comments[0].ID != "c1" {
		t.Errorf("comments = %v, want [{c1}]", loaded.comments)
	}
}

func TestLoadCommentsCmd_Error(t *testing.T) {
	repo := &fakeCommentRepoForLoad{listErr: errTest("comment list failed")}
	m := newTestModelForLoad(&fakeTaskRepoForLoad{}, repo)

	cmd := m.loadCommentsCmd("task-1")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	loaded, ok := msg.(commentsLoadedMsg)
	if !ok {
		t.Fatalf("expected commentsLoadedMsg, got %T", msg)
	}
	if loaded.err == nil || loaded.err.Error() != "comment list failed" {
		t.Errorf("err = %v, want comment list failed", loaded.err)
	}
}

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
