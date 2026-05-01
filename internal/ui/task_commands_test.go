package ui

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tiagokriok/kanji/internal/application"
	"github.com/tiagokriok/kanji/internal/domain"
)

// --- fake repositories ---

type fakeTaskRepoForCommands struct {
	createErr     error
	updateErr     error
	deleteErr     error
	moveErr       error
	lastCreated   domain.Task
	lastUpdateID  string
	lastUpdate    domain.TaskPatch
	lastMove      domain.MoveTaskInput
	lastDeletedID string
}

func (r *fakeTaskRepoForCommands) Create(ctx context.Context, task domain.Task) error {
	r.lastCreated = task
	return r.createErr
}
func (r *fakeTaskRepoForCommands) Update(ctx context.Context, taskID string, patch domain.TaskPatch) error {
	r.lastUpdateID = taskID
	r.lastUpdate = patch
	return r.updateErr
}
func (r *fakeTaskRepoForCommands) GetByID(ctx context.Context, taskID string) (domain.Task, error) {
	return domain.Task{}, nil
}
func (r *fakeTaskRepoForCommands) List(ctx context.Context, filter domain.TaskFilter) ([]domain.Task, error) {
	return nil, nil
}
func (r *fakeTaskRepoForCommands) Move(ctx context.Context, input domain.MoveTaskInput) error {
	r.lastMove = input
	return r.moveErr
}
func (r *fakeTaskRepoForCommands) Delete(ctx context.Context, id string) error {
	r.lastDeletedID = id
	return r.deleteErr
}
func (r *fakeTaskRepoForCommands) ListColumns(ctx context.Context, boardID string) ([]domain.Column, error) {
	return nil, nil
}
func (r *fakeTaskRepoForCommands) ListBoards(ctx context.Context, workspaceID string) ([]domain.Board, error) {
	return nil, nil
}

type fakeCommentRepoForCommands struct {
	createErr   error
	lastCreated domain.Comment
}

func (r *fakeCommentRepoForCommands) Create(ctx context.Context, comment domain.Comment) error {
	r.lastCreated = comment
	return r.createErr
}
func (r *fakeCommentRepoForCommands) ListByTask(ctx context.Context, taskID string) ([]domain.Comment, error) {
	return nil, nil
}
func (r *fakeCommentRepoForCommands) Update(ctx context.Context, commentID string, bodyMD string) error {
	return nil
}
func (r *fakeCommentRepoForCommands) Delete(ctx context.Context, commentID string) error {
	return nil
}

// --- helpers ---

func intPtr(i int) *int {
	return &i
}

func newTestModelWithServices(taskRepo domain.TaskRepository, commentRepo domain.CommentRepository) Model {
	ts := application.NewTaskService(taskRepo)
	tf := application.NewTaskFlow(taskRepo)
	cs := application.NewCommentService(commentRepo)
	return Model{
		taskService:    ts,
		taskFlow:       tf,
		commentService: cs,
		providerID:     "provider-1",
		workspaceID:    "workspace-1",
		keys:           newKeyMap(),
	}
}

func assertOpResultStatus(t *testing.T, cmd tea.Cmd, wantStatus string) {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	result, ok := msg.(opResultMsg)
	if !ok {
		t.Fatalf("expected opResultMsg, got %T", msg)
	}
	if result.status != wantStatus {
		t.Errorf("status = %q, want %q", result.status, wantStatus)
	}
	if result.err != nil {
		t.Errorf("unexpected err: %v", result.err)
	}
}

func assertOpResultError(t *testing.T, cmd tea.Cmd, wantErr string) {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	result, ok := msg.(opResultMsg)
	if !ok {
		t.Fatalf("expected opResultMsg, got %T", msg)
	}
	if result.err == nil || result.err.Error() != wantErr {
		t.Errorf("err = %v, want %q", result.err, wantErr)
	}
}

// --- createTaskWithDetailsCmd ---

func TestCreateTaskWithDetailsCmd_Success(t *testing.T) {
	repo := &fakeTaskRepoForCommands{}
	m := newTestModelWithServices(repo, &fakeCommentRepoForCommands{})

	cmd := m.createTaskWithDetailsCmd("title", "desc", 1, nil, strPtr("board-1"), strPtr("col-1"), strPtr("todo"))
	assertOpResultStatus(t, cmd, "task created")
	if repo.lastCreated.ProviderID != "provider-1" {
		t.Errorf("ProviderID = %q, want provider-1", repo.lastCreated.ProviderID)
	}
	if repo.lastCreated.WorkspaceID != "workspace-1" {
		t.Errorf("WorkspaceID = %q, want workspace-1", repo.lastCreated.WorkspaceID)
	}
	if repo.lastCreated.Title != "title" {
		t.Errorf("Title = %q, want title", repo.lastCreated.Title)
	}
	if repo.lastCreated.DescriptionMD != "desc" {
		t.Errorf("DescriptionMD = %q, want desc", repo.lastCreated.DescriptionMD)
	}
	if repo.lastCreated.BoardID == nil || *repo.lastCreated.BoardID != "board-1" {
		t.Errorf("BoardID = %v, want board-1", repo.lastCreated.BoardID)
	}
	if repo.lastCreated.ColumnID == nil || *repo.lastCreated.ColumnID != "col-1" {
		t.Errorf("ColumnID = %v, want col-1", repo.lastCreated.ColumnID)
	}
	if repo.lastCreated.Status == nil || *repo.lastCreated.Status != "todo" {
		t.Errorf("Status = %v, want todo", repo.lastCreated.Status)
	}
	if repo.lastCreated.Priority != 1 {
		t.Errorf("Priority = %d, want 1", repo.lastCreated.Priority)
	}
	if repo.lastCreated.Labels == nil || len(repo.lastCreated.Labels) != 0 {
		t.Errorf("Labels = %#v, want empty slice", repo.lastCreated.Labels)
	}
}

func TestCreateTaskWithDetailsCmd_Error(t *testing.T) {
	repo := &fakeTaskRepoForCommands{createErr: errors.New("create failed")}
	m := newTestModelWithServices(repo, &fakeCommentRepoForCommands{})

	cmd := m.createTaskWithDetailsCmd("title", "desc", 1, nil, strPtr("board-1"), strPtr("col-1"), strPtr("todo"))
	assertOpResultError(t, cmd, "create failed")
}

// --- updateTaskWithDetailsCmd ---

func TestUpdateTaskWithDetailsCmd_Success(t *testing.T) {
	repo := &fakeTaskRepoForCommands{}
	m := newTestModelWithServices(repo, &fakeCommentRepoForCommands{})

	cmd := m.updateTaskWithDetailsCmd("task-1", strPtr("new title"), strPtr("new desc"), intPtr(2), nil, strPtr("col-2"), strPtr("doing"))
	assertOpResultStatus(t, cmd, "task updated")
	if repo.lastUpdateID != "task-1" {
		t.Errorf("taskID = %q, want task-1", repo.lastUpdateID)
	}
	if repo.lastUpdate.Title == nil || *repo.lastUpdate.Title != "new title" {
		t.Errorf("Title = %v, want new title", repo.lastUpdate.Title)
	}
	if repo.lastUpdate.DescriptionMD == nil || *repo.lastUpdate.DescriptionMD != "new desc" {
		t.Errorf("DescriptionMD = %v, want new desc", repo.lastUpdate.DescriptionMD)
	}
	if repo.lastUpdate.Priority == nil || *repo.lastUpdate.Priority != 2 {
		t.Errorf("Priority = %v, want 2", repo.lastUpdate.Priority)
	}
	if repo.lastUpdate.ColumnID == nil || *repo.lastUpdate.ColumnID != "col-2" {
		t.Errorf("ColumnID = %v, want col-2", repo.lastUpdate.ColumnID)
	}
	if repo.lastUpdate.Status == nil || *repo.lastUpdate.Status != "doing" {
		t.Errorf("Status = %v, want doing", repo.lastUpdate.Status)
	}
}

func TestUpdateTaskWithDetailsCmd_Error(t *testing.T) {
	repo := &fakeTaskRepoForCommands{updateErr: errors.New("update failed")}
	m := newTestModelWithServices(repo, &fakeCommentRepoForCommands{})

	cmd := m.updateTaskWithDetailsCmd("task-1", strPtr("new title"), nil, nil, nil, nil, nil)
	assertOpResultError(t, cmd, "update failed")
}

// --- updateTaskDescriptionCmd ---

func TestUpdateTaskDescriptionCmd_Success(t *testing.T) {
	repo := &fakeTaskRepoForCommands{}
	m := newTestModelWithServices(repo, &fakeCommentRepoForCommands{})

	cmd := m.updateTaskDescriptionCmd("task-1", "new description")
	assertOpResultStatus(t, cmd, "description updated")
	if repo.lastUpdateID != "task-1" {
		t.Errorf("taskID = %q, want task-1", repo.lastUpdateID)
	}
	if repo.lastUpdate.DescriptionMD == nil || *repo.lastUpdate.DescriptionMD != "new description" {
		t.Errorf("DescriptionMD = %v, want new description", repo.lastUpdate.DescriptionMD)
	}
}

func TestUpdateTaskDescriptionCmd_Error(t *testing.T) {
	repo := &fakeTaskRepoForCommands{updateErr: errors.New("update failed")}
	m := newTestModelWithServices(repo, &fakeCommentRepoForCommands{})

	cmd := m.updateTaskDescriptionCmd("task-1", "new description")
	assertOpResultError(t, cmd, "update failed")
}

// --- addCommentCmd ---

func TestAddCommentCmd_Success(t *testing.T) {
	repo := &fakeCommentRepoForCommands{}
	m := newTestModelWithServices(&fakeTaskRepoForCommands{}, repo)

	cmd := m.addCommentCmd("task-1", "comment body")
	assertOpResultStatus(t, cmd, "comment added")
	if repo.lastCreated.TaskID != "task-1" {
		t.Errorf("TaskID = %q, want task-1", repo.lastCreated.TaskID)
	}
	if repo.lastCreated.ProviderID != "provider-1" {
		t.Errorf("ProviderID = %q, want provider-1", repo.lastCreated.ProviderID)
	}
	if repo.lastCreated.BodyMD != "comment body" {
		t.Errorf("BodyMD = %q, want comment body", repo.lastCreated.BodyMD)
	}
}

func TestAddCommentCmd_Error(t *testing.T) {
	repo := &fakeCommentRepoForCommands{createErr: errors.New("comment failed")}
	m := newTestModelWithServices(&fakeTaskRepoForCommands{}, repo)

	cmd := m.addCommentCmd("task-1", "comment body")
	assertOpResultError(t, cmd, "comment failed")
}

// --- moveToNextColumnCmd ---

func TestMoveToNextColumnCmd_NoColumns(t *testing.T) {
	m := newTestModelWithServices(&fakeTaskRepoForCommands{}, &fakeCommentRepoForCommands{})
	m.columns = []domain.Column{}

	task := domain.Task{ID: "task-1", ColumnID: strPtr("c1")}
	cmd := m.moveToNextColumnCmd(task)
	if cmd != nil {
		t.Error("expected nil cmd when no columns")
	}
}

func TestMoveToNextColumnCmd_Success(t *testing.T) {
	repo := &fakeTaskRepoForCommands{}
	m := newTestModelWithServices(repo, &fakeCommentRepoForCommands{})
	m.columns = []domain.Column{{ID: "c1", Name: "Todo"}, {ID: "c2", Name: "Doing"}}

	task := domain.Task{ID: "task-1", ColumnID: strPtr("c1")}
	cmd := m.moveToNextColumnCmd(task)
	assertOpResultStatus(t, cmd, "moved to Doing")
	if repo.lastMove.TaskID != "task-1" {
		t.Errorf("TaskID = %q, want task-1", repo.lastMove.TaskID)
	}
	if repo.lastMove.ColumnID == nil || *repo.lastMove.ColumnID != "c2" {
		t.Errorf("ColumnID = %v, want c2", repo.lastMove.ColumnID)
	}
	if repo.lastMove.Status == nil || *repo.lastMove.Status != "doing" {
		t.Errorf("Status = %v, want doing", repo.lastMove.Status)
	}
}

func TestMoveToNextColumnCmd_Error(t *testing.T) {
	repo := &fakeTaskRepoForCommands{moveErr: errors.New("move failed")}
	m := newTestModelWithServices(repo, &fakeCommentRepoForCommands{})
	m.columns = []domain.Column{{ID: "c1", Name: "Todo"}, {ID: "c2", Name: "Doing"}}

	task := domain.Task{ID: "task-1", ColumnID: strPtr("c1")}
	cmd := m.moveToNextColumnCmd(task)
	assertOpResultError(t, cmd, "move failed")
}

// --- moveToPrevColumnCmd ---

func TestMoveToPrevColumnCmd_NoColumns(t *testing.T) {
	m := newTestModelWithServices(&fakeTaskRepoForCommands{}, &fakeCommentRepoForCommands{})
	m.columns = []domain.Column{}

	task := domain.Task{ID: "task-1", ColumnID: strPtr("c2")}
	cmd := m.moveToPrevColumnCmd(task)
	if cmd != nil {
		t.Error("expected nil cmd when no columns")
	}
}

func TestMoveToPrevColumnCmd_Success(t *testing.T) {
	repo := &fakeTaskRepoForCommands{}
	m := newTestModelWithServices(repo, &fakeCommentRepoForCommands{})
	m.columns = []domain.Column{{ID: "c1", Name: "Todo"}, {ID: "c2", Name: "Doing"}}

	task := domain.Task{ID: "task-1", ColumnID: strPtr("c2")}
	cmd := m.moveToPrevColumnCmd(task)
	assertOpResultStatus(t, cmd, "moved to Todo")
	if repo.lastMove.TaskID != "task-1" {
		t.Errorf("TaskID = %q, want task-1", repo.lastMove.TaskID)
	}
	if repo.lastMove.ColumnID == nil || *repo.lastMove.ColumnID != "c1" {
		t.Errorf("ColumnID = %v, want c1", repo.lastMove.ColumnID)
	}
	if repo.lastMove.Status == nil || *repo.lastMove.Status != "todo" {
		t.Errorf("Status = %v, want todo", repo.lastMove.Status)
	}
}

func TestMoveToPrevColumnCmd_Error(t *testing.T) {
	repo := &fakeTaskRepoForCommands{moveErr: errors.New("move failed")}
	m := newTestModelWithServices(repo, &fakeCommentRepoForCommands{})
	m.columns = []domain.Column{{ID: "c1", Name: "Todo"}, {ID: "c2", Name: "Doing"}}

	task := domain.Task{ID: "task-1", ColumnID: strPtr("c2")}
	cmd := m.moveToPrevColumnCmd(task)
	assertOpResultError(t, cmd, "move failed")
}

// --- deleteTaskCmd ---

func TestDeleteTaskCmd_Success(t *testing.T) {
	repo := &fakeTaskRepoForCommands{}
	m := newTestModelWithServices(repo, &fakeCommentRepoForCommands{})

	cmd := m.deleteTaskCmd("task-1")
	assertOpResultStatus(t, cmd, "task deleted")
	if repo.lastDeletedID != "task-1" {
		t.Errorf("deleted id = %q, want task-1", repo.lastDeletedID)
	}
}

func TestDeleteTaskCmd_Error(t *testing.T) {
	repo := &fakeTaskRepoForCommands{deleteErr: errors.New("delete failed")}
	m := newTestModelWithServices(repo, &fakeCommentRepoForCommands{})

	cmd := m.deleteTaskCmd("task-1")
	assertOpResultError(t, cmd, "delete failed")
}
