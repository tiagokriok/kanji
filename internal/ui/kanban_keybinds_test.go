package ui

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tiagokriok/kanji/internal/application"
	"github.com/tiagokriok/kanji/internal/domain"
)

func TestShiftRightMovesTaskToNextColumnInKanban(t *testing.T) {
	repo := &kanbanMoveRepo{}
	model, secondColumnID := kanbanMoveTestModel(repo)

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyShiftRight})
	if cmd == nil {
		t.Fatal("expected shift+right to produce a move command")
	}
	_ = cmd()

	if repo.lastMove == nil {
		t.Fatal("expected task move to be recorded")
	}
	if repo.lastMove.TaskID != "task-1" {
		t.Fatalf("expected task-1 to move, got %q", repo.lastMove.TaskID)
	}
	if repo.lastMove.ColumnID == nil || *repo.lastMove.ColumnID != secondColumnID {
		t.Fatalf("expected move to column %q, got %#v", secondColumnID, repo.lastMove.ColumnID)
	}
	if repo.lastMove.Status == nil || *repo.lastMove.Status != "doing" {
		t.Fatalf("expected status %q, got %#v", "doing", repo.lastMove.Status)
	}
}

func TestShiftLeftMovesTaskToPreviousColumnInKanban(t *testing.T) {
	repo := &kanbanMoveRepo{}
	model, _ := kanbanMoveTestModel(repo)
	model.activeColumn = 1
	model.tasks[0].ColumnID = &model.columns[1].ID

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyShiftLeft})
	if cmd == nil {
		t.Fatal("expected shift+left to produce a move command")
	}
	_ = cmd()

	if repo.lastMove == nil {
		t.Fatal("expected task move to be recorded")
	}
	if repo.lastMove.ColumnID == nil || *repo.lastMove.ColumnID != model.columns[0].ID {
		t.Fatalf("expected move to column %q, got %#v", model.columns[0].ID, repo.lastMove.ColumnID)
	}
	if repo.lastMove.Status == nil || *repo.lastMove.Status != "todo" {
		t.Fatalf("expected status %q, got %#v", "todo", repo.lastMove.Status)
	}
}

func TestShiftArrowsDoNothingOutsideKanban(t *testing.T) {
	repo := &kanbanMoveRepo{}
	model, _ := kanbanMoveTestModel(repo)
	model.viewMode = viewList

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyShiftRight})
	if cmd != nil {
		t.Fatal("expected shift+right to be ignored outside kanban")
	}
	if repo.lastMove != nil {
		t.Fatal("expected no move outside kanban")
	}
}

func TestMovedKanbanCardKeepsSelectionAfterReload(t *testing.T) {
	repo := &kanbanMoveRepo{}
	model, secondColumnID := kanbanMoveTestModel(repo)

	nextModel, _ := model.Update(opResultMsg{status: "moved to Doing", taskID: "task-1", columnID: secondColumnID})
	updated := nextModel.(Model)

	updatedTask := updated.tasks[0]
	updatedTask.ColumnID = &secondColumnID
	nextModel, _ = updated.Update(tasksLoadedMsg{tasks: []domain.Task{updatedTask}})
	updated = nextModel.(Model)

	if updated.activeColumn != 1 {
		t.Fatalf("expected active column 1, got %d", updated.activeColumn)
	}
	if updated.kanbanRow != 0 {
		t.Fatalf("expected kanban row 0, got %d", updated.kanbanRow)
	}
	currentTask, ok := updated.currentTask()
	if !ok {
		t.Fatal("expected moved task to stay selected")
	}
	if currentTask.ID != "task-1" {
		t.Fatalf("expected selected task task-1, got %q", currentTask.ID)
	}
}

func kanbanMoveTestModel(repo *kanbanMoveRepo) (Model, string) {
	service := application.NewTaskService(repo)
	firstColumnID := "col-1"
	secondColumnID := "col-2"

	model := Model{
		taskService: service,
		columns: []domain.Column{
			{ID: firstColumnID, Name: "Todo"},
			{ID: secondColumnID, Name: "Doing"},
		},
		tasks: []domain.Task{{
			ID:        "task-1",
			Title:     "Task 1",
			ColumnID:  &firstColumnID,
			Position:  1,
			UpdatedAt: time.Now().UTC(),
		}},
		viewMode: viewKanban,
		keys:     newKeyMap(),
	}

	return model, secondColumnID
}

type kanbanMoveRepo struct {
	lastMove *domain.MoveTaskInput
}

func (r *kanbanMoveRepo) Create(context.Context, domain.Task) error              { return nil }
func (r *kanbanMoveRepo) Update(context.Context, string, domain.TaskPatch) error { return nil }
func (r *kanbanMoveRepo) GetByID(context.Context, string) (domain.Task, error) {
	return domain.Task{}, nil
}
func (r *kanbanMoveRepo) List(context.Context, domain.TaskFilter) ([]domain.Task, error) {
	return nil, nil
}
func (r *kanbanMoveRepo) Move(_ context.Context, input domain.MoveTaskInput) error {
	r.lastMove = &input
	return nil
}
func (r *kanbanMoveRepo) Delete(context.Context, string) error { return nil }
func (r *kanbanMoveRepo) ListColumns(context.Context, string) ([]domain.Column, error) {
	return nil, nil
}
func (r *kanbanMoveRepo) ListBoards(context.Context, string) ([]domain.Board, error) { return nil, nil }
