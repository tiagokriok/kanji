package ui

import (
	"sort"
	"strings"

	"github.com/tiagokriok/kanji/internal/domain"
)

// selectionState holds the selection-related state for the UI.
// This encapsulates the logic for managing task selection in both list and kanban views.
type selectionState struct {
	viewMode viewMode
	columns  []domain.Column
	tasks    []domain.Task

	// List view selection
	selected int

	// Kanban view selection
	activeColumn int
	kanbanRow    int

	// Pending kanban restoration
	pendingKanbanTaskID   string
	pendingKanbanColumnID string
}

// ensureSelection validates and adjusts selection indices based on current data.
// For list view: ensures selected index is within bounds of tasks.
// For kanban view: ensures activeColumn is within bounds and kanbanRow is valid for that column.
func (s *selectionState) ensureSelection() {
	if s.viewMode == viewKanban {
		if len(s.columns) == 0 {
			s.activeColumn = 0
			s.kanbanRow = 0
			return
		}
		if s.activeColumn < 0 {
			s.activeColumn = 0
		}
		if s.activeColumn >= len(s.columns) {
			s.activeColumn = len(s.columns) - 1
		}
		s.ensureKanbanRow()
		return
	}
	if len(s.tasks) == 0 {
		s.selected = 0
		return
	}
	if s.selected < 0 {
		s.selected = 0
	}
	if s.selected >= len(s.tasks) {
		s.selected = len(s.tasks) - 1
	}
}

// setActiveColumnByID sets the active column index by searching for the given column ID.
// Does nothing if columnID is empty or not found.
func (s *selectionState) setActiveColumnByID(columnID string) {
	if strings.TrimSpace(columnID) == "" {
		return
	}
	for i, col := range s.columns {
		if col.ID == columnID {
			s.activeColumn = i
			return
		}
	}
}

// restorePendingKanbanSelection attempts to restore the kanban selection after a task move.
// It looks for the pending task in the pending column and sets the kanban row if found.
// Returns true if the selection was successfully restored, false otherwise.
// Always clears the pending IDs after attempting restoration.
func (s *selectionState) restorePendingKanbanSelection() bool {
	// Always clear pending IDs on exit
	defer func() {
		s.pendingKanbanTaskID = ""
		s.pendingKanbanColumnID = ""
	}()

	if s.viewMode != viewKanban {
		return false
	}
	if strings.TrimSpace(s.pendingKanbanTaskID) == "" || strings.TrimSpace(s.pendingKanbanColumnID) == "" {
		return false
	}

	s.setActiveColumnByID(s.pendingKanbanColumnID)
	tasks := s.tasksForColumn(s.pendingKanbanColumnID)
	for i, task := range tasks {
		if task.ID == s.pendingKanbanTaskID {
			s.kanbanRow = i
			return true
		}
	}
	return false
}

// ensureKanbanRow validates and adjusts the kanban row index for the current column.
// Ensures kanbanRow is within bounds of tasks in the active column.
func (s *selectionState) ensureKanbanRow() {
	if len(s.columns) == 0 {
		s.kanbanRow = 0
		return
	}
	colID := s.columns[s.activeColumn].ID
	tasks := s.tasksForColumn(colID)
	if len(tasks) == 0 {
		s.kanbanRow = 0
		return
	}
	if s.kanbanRow < 0 {
		s.kanbanRow = 0
	}
	if s.kanbanRow >= len(tasks) {
		s.kanbanRow = len(tasks) - 1
	}
}

// moveUp moves the selection up (decrements index) based on current view mode.
// In list view: decrements selected index.
// In kanban view: decrements kanbanRow for the active column.
func (s *selectionState) moveUp() {
	if s.viewMode == viewKanban {
		s.kanbanRow--
		s.ensureKanbanRow()
		return
	}
	s.selected--
	s.ensureSelection()
}

// moveDown moves the selection down (increments index) based on current view mode.
// In list view: increments selected index.
// In kanban view: increments kanbanRow for the active column.
func (s *selectionState) moveDown() {
	if s.viewMode == viewKanban {
		s.kanbanRow++
		s.ensureKanbanRow()
		return
	}
	s.selected++
	s.ensureSelection()
}

// currentTask returns the currently selected task based on the view mode.
// In list view: returns the task at the selected index.
// In kanban view: returns the task at kanbanRow in the active column.
func (s *selectionState) currentTask() (domain.Task, bool) {
	if len(s.tasks) == 0 {
		return domain.Task{}, false
	}
	if s.viewMode == viewKanban {
		if len(s.columns) == 0 {
			return domain.Task{}, false
		}
		col := s.columns[s.activeColumn]
		tasks := s.tasksForColumn(col.ID)
		if len(tasks) == 0 || s.kanbanRow < 0 || s.kanbanRow >= len(tasks) {
			return domain.Task{}, false
		}
		return tasks[s.kanbanRow], true
	}
	if s.selected < 0 || s.selected >= len(s.tasks) {
		return domain.Task{}, false
	}
	return s.tasks[s.selected], true
}

// tasksForColumn returns all tasks belonging to the specified column, sorted by position.
// Tasks are sorted by position ascending, with ties broken by updated time (newest first).
func (s *selectionState) tasksForColumn(columnID string) []domain.Task {
	result := make([]domain.Task, 0)
	for _, t := range s.tasks {
		if t.ColumnID != nil && *t.ColumnID == columnID {
			result = append(result, t)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Position == result[j].Position {
			return result[i].UpdatedAt.After(result[j].UpdatedAt)
		}
		return result[i].Position < result[j].Position
	})
	return result
}
