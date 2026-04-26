package ui

import (
	"fmt"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/tiagokriok/kanji/internal/domain"
)

func TestKanbanViewRespectsTerminalHeight(t *testing.T) {
	model := kanbanTestModel(80, 14)

	rendered := model.View()
	if got := lipgloss.Height(rendered); got > model.height {
		t.Fatalf("expected kanban view height <= %d, got %d\n%s", model.height, got, rendered)
	}
}

func TestKanbanViewRespectsTerminalWidth(t *testing.T) {
	model := kanbanTestModel(80, 14)

	rendered := model.View()
	if got := lipgloss.Width(rendered); got > model.width {
		t.Fatalf("expected kanban view width <= %d, got %d\n%s", model.width, got, rendered)
	}
}

func kanbanTestModel(width, height int) Model {
	columnID := "col-1"
	now := time.Date(2026, time.April, 25, 12, 0, 0, 0, time.UTC)
	tasks := make([]domain.Task, 0, 8)
	for i := range 8 {
		title := fmt.Sprintf("Task %02d", i+1)
		tasks = append(tasks, domain.Task{
			ID:        fmt.Sprintf("task-%02d", i+1),
			Title:     title,
			ColumnID:  &columnID,
			Position:  float64(i + 1),
			UpdatedAt: now.Add(time.Duration(i) * time.Minute),
		})
	}

	return Model{
		workspaceName: "Workspace",
		boardName:     "Board",
		columns: []domain.Column{{
			ID:    columnID,
			Name:  "Todo",
			Color: "#333333",
		}},
		tasks:    tasks,
		viewMode: viewKanban,
		width:    width,
		height:   height,
	}
}
