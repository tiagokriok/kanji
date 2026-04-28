package ui

import "github.com/tiagokriok/kanji/internal/domain"

// toSelectionState creates a selectionState from the Model's current selection-related fields.
func (m Model) toSelectionState() selectionState {
	return selectionState{
		viewMode:              m.viewMode,
		columns:               m.columns,
		tasks:                 m.tasks,
		selected:              m.selected,
		activeColumn:          m.activeColumn,
		kanbanRow:             m.kanbanRow,
		pendingKanbanTaskID:   m.pendingKanbanTaskID,
		pendingKanbanColumnID: m.pendingKanbanColumnID,
	}
}

// applySelectionState copies the selection-related fields from selectionState back to Model.
func (m *Model) applySelectionState(s selectionState) {
	m.selected = s.selected
	m.activeColumn = s.activeColumn
	m.kanbanRow = s.kanbanRow
	m.pendingKanbanTaskID = s.pendingKanbanTaskID
	m.pendingKanbanColumnID = s.pendingKanbanColumnID
}

func (m *Model) ensureSelection() {
	s := m.toSelectionState()
	s.ensureSelection()
	m.applySelectionState(s)
}

func (m *Model) setActiveColumnByID(columnID string) {
	s := m.toSelectionState()
	s.setActiveColumnByID(columnID)
	m.applySelectionState(s)
}

func (m *Model) restorePendingKanbanSelection() bool {
	s := m.toSelectionState()
	ok := s.restorePendingKanbanSelection()
	m.applySelectionState(s)
	return ok
}

func (m *Model) ensureKanbanRow() {
	s := m.toSelectionState()
	s.ensureKanbanRow()
	m.applySelectionState(s)
}

func (m *Model) moveUp() {
	s := m.toSelectionState()
	s.moveUp()
	m.applySelectionState(s)
}

func (m *Model) moveDown() {
	s := m.toSelectionState()
	s.moveDown()
	m.applySelectionState(s)
}

func (m Model) currentTask() (domain.Task, bool) {
	s := m.toSelectionState()
	return s.currentTask()
}

func (m Model) tasksForColumn(columnID string) []domain.Task {
	s := m.toSelectionState()
	return s.tasksForColumn(columnID)
}
