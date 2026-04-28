package ui

import (
	"testing"
	"time"

	"github.com/tiagokriok/kanji/internal/domain"
)

func makeModelSelectionTestModel() Model {
	return Model{}
}

func TestModelToSelectionState(t *testing.T) {
	col1 := "col-1"
	col2 := "col-2"
	columns := []domain.Column{
		{ID: col1, Name: "Todo"},
		{ID: col2, Name: "Doing"},
	}
	tasks := []domain.Task{{ID: "t1"}, {ID: "t2"}}

	m := makeModelSelectionTestModel()
	m.viewMode = viewKanban
	m.columns = columns
	m.tasks = tasks
	m.selected = 3
	m.activeColumn = 1
	m.kanbanRow = 2
	m.pendingKanbanTaskID = "pending-task"
	m.pendingKanbanColumnID = "pending-col"

	s := m.toSelectionState()

	if s.viewMode != viewKanban {
		t.Errorf("viewMode = %v, want viewKanban", s.viewMode)
	}
	if len(s.columns) != 2 || s.columns[0].ID != col1 {
		t.Errorf("columns not copied correctly")
	}
	if len(s.tasks) != 2 || s.tasks[0].ID != "t1" {
		t.Errorf("tasks not copied correctly")
	}
	if s.selected != 3 {
		t.Errorf("selected = %d, want 3", s.selected)
	}
	if s.activeColumn != 1 {
		t.Errorf("activeColumn = %d, want 1", s.activeColumn)
	}
	if s.kanbanRow != 2 {
		t.Errorf("kanbanRow = %d, want 2", s.kanbanRow)
	}
	if s.pendingKanbanTaskID != "pending-task" {
		t.Errorf("pendingKanbanTaskID = %q, want pending-task", s.pendingKanbanTaskID)
	}
	if s.pendingKanbanColumnID != "pending-col" {
		t.Errorf("pendingKanbanColumnID = %q, want pending-col", s.pendingKanbanColumnID)
	}
}

func TestModelApplySelectionState(t *testing.T) {
	m := makeModelSelectionTestModel()
	s := selectionState{
		selected:              5,
		activeColumn:          2,
		kanbanRow:             3,
		pendingKanbanTaskID:   "task-id",
		pendingKanbanColumnID: "col-id",
	}

	m.applySelectionState(s)

	if m.selected != 5 {
		t.Errorf("selected = %d, want 5", m.selected)
	}
	if m.activeColumn != 2 {
		t.Errorf("activeColumn = %d, want 2", m.activeColumn)
	}
	if m.kanbanRow != 3 {
		t.Errorf("kanbanRow = %d, want 3", m.kanbanRow)
	}
	if m.pendingKanbanTaskID != "task-id" {
		t.Errorf("pendingKanbanTaskID = %q, want task-id", m.pendingKanbanTaskID)
	}
	if m.pendingKanbanColumnID != "col-id" {
		t.Errorf("pendingKanbanColumnID = %q, want col-id", m.pendingKanbanColumnID)
	}
}

func TestModelEnsureSelectionListView(t *testing.T) {
	tests := []struct {
		name     string
		tasks    []domain.Task
		selected int
		want     int
	}{
		{
			name:     "empty tasks resets selection to 0",
			tasks:    []domain.Task{},
			selected: 5,
			want:     0,
		},
		{
			name:     "negative selection clamps to 0",
			tasks:    []domain.Task{{ID: "t1"}, {ID: "t2"}},
			selected: -5,
			want:     0,
		},
		{
			name:     "selection within bounds stays",
			tasks:    []domain.Task{{ID: "t1"}, {ID: "t2"}, {ID: "t3"}},
			selected: 1,
			want:     1,
		},
		{
			name:     "selection beyond length clamps to last index",
			tasks:    []domain.Task{{ID: "t1"}, {ID: "t2"}},
			selected: 10,
			want:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := makeModelSelectionTestModel()
			m.viewMode = viewList
			m.tasks = tt.tasks
			m.selected = tt.selected
			m.ensureSelection()
			if m.selected != tt.want {
				t.Errorf("selected = %d, want %d", m.selected, tt.want)
			}
		})
	}
}

func TestModelEnsureSelectionKanbanView(t *testing.T) {
	col1 := "col-1"
	col2 := "col-2"
	columns := []domain.Column{
		{ID: col1, Name: "Todo"},
		{ID: col2, Name: "Doing"},
	}

	tests := []struct {
		name         string
		columns      []domain.Column
		tasks        []domain.Task
		activeColumn int
		kanbanRow    int
		wantColumn   int
		wantRow      int
	}{
		{
			name:         "empty columns resets to 0,0",
			columns:      []domain.Column{},
			tasks:        []domain.Task{},
			activeColumn: 1,
			kanbanRow:    5,
			wantColumn:   0,
			wantRow:      0,
		},
		{
			name:         "negative column clamps to 0",
			columns:      columns,
			tasks:        []domain.Task{},
			activeColumn: -1,
			kanbanRow:    0,
			wantColumn:   0,
			wantRow:      0,
		},
		{
			name:         "column beyond length clamps to last",
			columns:      columns,
			tasks:        []domain.Task{},
			activeColumn: 10,
			kanbanRow:    0,
			wantColumn:   1,
			wantRow:      0,
		},
		{
			name:         "valid column with tasks ensures row bounds",
			columns:      columns,
			tasks:        []domain.Task{{ID: "t1", ColumnID: &col1}, {ID: "t2", ColumnID: &col1}},
			activeColumn: 0,
			kanbanRow:    5,
			wantColumn:   0,
			wantRow:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := makeModelSelectionTestModel()
			m.viewMode = viewKanban
			m.columns = tt.columns
			m.tasks = tt.tasks
			m.activeColumn = tt.activeColumn
			m.kanbanRow = tt.kanbanRow
			m.ensureSelection()
			if m.activeColumn != tt.wantColumn {
				t.Errorf("activeColumn = %d, want %d", m.activeColumn, tt.wantColumn)
			}
			if m.kanbanRow != tt.wantRow {
				t.Errorf("kanbanRow = %d, want %d", m.kanbanRow, tt.wantRow)
			}
		})
	}
}

func TestModelSetActiveColumnByID(t *testing.T) {
	columns := []domain.Column{
		{ID: "col-1", Name: "Todo"},
		{ID: "col-2", Name: "Doing"},
		{ID: "col-3", Name: "Done"},
	}

	tests := []struct {
		name       string
		columnID   string
		wantColumn int
	}{
		{
			name:       "finds first column",
			columnID:   "col-1",
			wantColumn: 0,
		},
		{
			name:       "finds middle column",
			columnID:   "col-2",
			wantColumn: 1,
		},
		{
			name:       "finds last column",
			columnID:   "col-3",
			wantColumn: 2,
		},
		{
			name:       "empty id does nothing",
			columnID:   "",
			wantColumn: 0,
		},
		{
			name:       "unknown id does nothing",
			columnID:   "col-unknown",
			wantColumn: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := makeModelSelectionTestModel()
			m.columns = columns
			m.activeColumn = 0
			m.setActiveColumnByID(tt.columnID)
			if m.activeColumn != tt.wantColumn {
				t.Errorf("activeColumn = %d, want %d", m.activeColumn, tt.wantColumn)
			}
		})
	}
}

func TestModelRestorePendingKanbanSelection(t *testing.T) {
	col1 := "col-1"
	col2 := "col-2"
	task1 := "task-1"
	task2 := "task-2"

	columns := []domain.Column{
		{ID: col1, Name: "Todo"},
		{ID: col2, Name: "Doing"},
	}

	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: task1, ColumnID: &col1, UpdatedAt: now},
		{ID: task2, ColumnID: &col2, UpdatedAt: now.Add(time.Minute)},
	}

	tests := []struct {
		name            string
		viewMode        viewMode
		pendingTaskID   string
		pendingColID    string
		tasks           []domain.Task
		wantOK          bool
		wantColumn      int
		wantRow         int
		wantPendingTask string
		wantPendingCol  string
	}{
		{
			name:            "list view clears pending and returns false",
			viewMode:        viewList,
			pendingTaskID:   task1,
			pendingColID:    col1,
			tasks:           tasks,
			wantOK:          false,
			wantColumn:      0,
			wantRow:         0,
			wantPendingTask: "",
			wantPendingCol:  "",
		},
		{
			name:            "empty pending task returns false",
			viewMode:        viewKanban,
			pendingTaskID:   "",
			pendingColID:    col1,
			tasks:           tasks,
			wantOK:          false,
			wantColumn:      0,
			wantRow:         0,
			wantPendingTask: "",
			wantPendingCol:  "",
		},
		{
			name:            "finds task in column",
			viewMode:        viewKanban,
			pendingTaskID:   task2,
			pendingColID:    col2,
			tasks:           tasks,
			wantOK:          true,
			wantColumn:      1,
			wantRow:         0,
			wantPendingTask: "",
			wantPendingCol:  "",
		},
		{
			name:            "task not found returns false",
			viewMode:        viewKanban,
			pendingTaskID:   "task-unknown",
			pendingColID:    col1,
			tasks:           tasks,
			wantOK:          false,
			wantColumn:      0,
			wantRow:         0,
			wantPendingTask: "",
			wantPendingCol:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := makeModelSelectionTestModel()
			m.viewMode = tt.viewMode
			m.columns = columns
			m.tasks = tt.tasks
			m.activeColumn = 0
			m.kanbanRow = 0
			m.pendingKanbanTaskID = tt.pendingTaskID
			m.pendingKanbanColumnID = tt.pendingColID

			got := m.restorePendingKanbanSelection()
			if got != tt.wantOK {
				t.Errorf("restorePendingKanbanSelection() = %v, want %v", got, tt.wantOK)
			}
			if m.activeColumn != tt.wantColumn {
				t.Errorf("activeColumn = %d, want %d", m.activeColumn, tt.wantColumn)
			}
			if m.kanbanRow != tt.wantRow {
				t.Errorf("kanbanRow = %d, want %d", m.kanbanRow, tt.wantRow)
			}
			if m.pendingKanbanTaskID != tt.wantPendingTask {
				t.Errorf("pendingKanbanTaskID = %q, want %q", m.pendingKanbanTaskID, tt.wantPendingTask)
			}
			if m.pendingKanbanColumnID != tt.wantPendingCol {
				t.Errorf("pendingKanbanColumnID = %q, want %q", m.pendingKanbanColumnID, tt.wantPendingCol)
			}
		})
	}
}

func TestModelEnsureKanbanRow(t *testing.T) {
	col1 := "col-1"
	col2 := "col-2"
	columns := []domain.Column{
		{ID: col1, Name: "Todo"},
		{ID: col2, Name: "Doing"},
	}

	tests := []struct {
		name         string
		activeColumn int
		tasks        []domain.Task
		kanbanRow    int
		want         int
	}{
		{
			name:         "empty column tasks resets row to 0",
			activeColumn: 1,
			tasks:        []domain.Task{{ID: "t1", ColumnID: &col1}},
			kanbanRow:    5,
			want:         0,
		},
		{
			name:         "negative row clamps to 0",
			activeColumn: 0,
			tasks:        []domain.Task{{ID: "t1", ColumnID: &col1}, {ID: "t2", ColumnID: &col1}},
			kanbanRow:    -5,
			want:         0,
		},
		{
			name:         "row within bounds stays",
			activeColumn: 0,
			tasks:        []domain.Task{{ID: "t1", ColumnID: &col1}, {ID: "t2", ColumnID: &col1}},
			kanbanRow:    1,
			want:         1,
		},
		{
			name:         "row beyond tasks clamps to last",
			activeColumn: 0,
			tasks:        []domain.Task{{ID: "t1", ColumnID: &col1}},
			kanbanRow:    10,
			want:         0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := makeModelSelectionTestModel()
			m.viewMode = viewKanban
			m.columns = columns
			m.tasks = tt.tasks
			m.activeColumn = tt.activeColumn
			m.kanbanRow = tt.kanbanRow
			m.ensureKanbanRow()
			if m.kanbanRow != tt.want {
				t.Errorf("kanbanRow = %d, want %d", m.kanbanRow, tt.want)
			}
		})
	}
}

func TestModelMoveUpDownListView(t *testing.T) {
	tasks := []domain.Task{
		{ID: "t1"}, {ID: "t2"}, {ID: "t3"},
	}

	t.Run("moveUp decrements selection", func(t *testing.T) {
		m := makeModelSelectionTestModel()
		m.viewMode = viewList
		m.tasks = tasks
		m.selected = 2
		m.moveUp()
		if m.selected != 1 {
			t.Errorf("selected = %d, want 1", m.selected)
		}
	})

	t.Run("moveDown increments selection", func(t *testing.T) {
		m := makeModelSelectionTestModel()
		m.viewMode = viewList
		m.tasks = tasks
		m.selected = 0
		m.moveDown()
		if m.selected != 1 {
			t.Errorf("selected = %d, want 1", m.selected)
		}
	})

	t.Run("moveUp clamps to 0", func(t *testing.T) {
		m := makeModelSelectionTestModel()
		m.viewMode = viewList
		m.tasks = tasks
		m.selected = 0
		m.moveUp()
		if m.selected != 0 {
			t.Errorf("selected = %d, want 0", m.selected)
		}
	})

	t.Run("moveDown clamps to last index", func(t *testing.T) {
		m := makeModelSelectionTestModel()
		m.viewMode = viewList
		m.tasks = tasks
		m.selected = 2
		m.moveDown()
		if m.selected != 2 {
			t.Errorf("selected = %d, want 2", m.selected)
		}
	})
}

func TestModelMoveUpDownKanbanView(t *testing.T) {
	col1 := "col-1"
	columns := []domain.Column{{ID: col1, Name: "Todo"}}
	tasks := []domain.Task{
		{ID: "t1", ColumnID: &col1},
		{ID: "t2", ColumnID: &col1},
		{ID: "t3", ColumnID: &col1},
	}

	t.Run("moveUp decrements kanban row", func(t *testing.T) {
		m := makeModelSelectionTestModel()
		m.viewMode = viewKanban
		m.columns = columns
		m.tasks = tasks
		m.activeColumn = 0
		m.kanbanRow = 2
		m.moveUp()
		if m.kanbanRow != 1 {
			t.Errorf("kanbanRow = %d, want 1", m.kanbanRow)
		}
	})

	t.Run("moveDown increments kanban row", func(t *testing.T) {
		m := makeModelSelectionTestModel()
		m.viewMode = viewKanban
		m.columns = columns
		m.tasks = tasks
		m.activeColumn = 0
		m.kanbanRow = 0
		m.moveDown()
		if m.kanbanRow != 1 {
			t.Errorf("kanbanRow = %d, want 1", m.kanbanRow)
		}
	})

	t.Run("moveUp clamps to 0", func(t *testing.T) {
		m := makeModelSelectionTestModel()
		m.viewMode = viewKanban
		m.columns = columns
		m.tasks = tasks
		m.activeColumn = 0
		m.kanbanRow = 0
		m.moveUp()
		if m.kanbanRow != 0 {
			t.Errorf("kanbanRow = %d, want 0", m.kanbanRow)
		}
	})

	t.Run("moveDown clamps to last row", func(t *testing.T) {
		m := makeModelSelectionTestModel()
		m.viewMode = viewKanban
		m.columns = columns
		m.tasks = tasks
		m.activeColumn = 0
		m.kanbanRow = 2
		m.moveDown()
		if m.kanbanRow != 2 {
			t.Errorf("kanbanRow = %d, want 2", m.kanbanRow)
		}
	})
}

func TestModelCurrentTaskListView(t *testing.T) {
	tasks := []domain.Task{
		{ID: "t1", Title: "Task 1"},
		{ID: "t2", Title: "Task 2"},
		{ID: "t3", Title: "Task 3"},
	}

	tests := []struct {
		name     string
		selected int
		wantID   string
		wantOK   bool
	}{
		{
			name:     "returns selected task",
			selected: 1,
			wantID:   "t2",
			wantOK:   true,
		},
		{
			name:     "empty tasks returns false",
			selected: 0,
			wantID:   "",
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := makeModelSelectionTestModel()
			m.viewMode = viewList
			if tt.name == "empty tasks returns false" {
				m.tasks = []domain.Task{}
			} else {
				m.tasks = tasks
			}
			m.selected = tt.selected
			task, ok := m.currentTask()
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && task.ID != tt.wantID {
				t.Errorf("task.ID = %q, want %q", task.ID, tt.wantID)
			}
		})
	}
}

func TestModelCurrentTaskKanbanView(t *testing.T) {
	col1 := "col-1"
	col2 := "col-2"
	columns := []domain.Column{
		{ID: col1, Name: "Todo"},
		{ID: col2, Name: "Doing"},
	}

	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", Title: "Task 1", ColumnID: &col1, Position: 1, UpdatedAt: now},
		{ID: "t2", Title: "Task 2", ColumnID: &col1, Position: 2, UpdatedAt: now.Add(time.Minute)},
		{ID: "t3", Title: "Task 3", ColumnID: &col2, Position: 1, UpdatedAt: now},
	}

	tests := []struct {
		name         string
		activeColumn int
		kanbanRow    int
		wantID       string
		wantOK       bool
	}{
		{
			name:         "returns task in first column",
			activeColumn: 0,
			kanbanRow:    0,
			wantID:       "t1",
			wantOK:       true,
		},
		{
			name:         "returns task in first column second row",
			activeColumn: 0,
			kanbanRow:    1,
			wantID:       "t2",
			wantOK:       true,
		},
		{
			name:         "returns task in second column",
			activeColumn: 1,
			kanbanRow:    0,
			wantID:       "t3",
			wantOK:       true,
		},
		{
			name:         "empty column returns false",
			activeColumn: 1,
			kanbanRow:    5,
			wantID:       "",
			wantOK:       false,
		},
		{
			name:         "empty columns returns false",
			activeColumn: 0,
			kanbanRow:    0,
			wantID:       "",
			wantOK:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := makeModelSelectionTestModel()
			m.viewMode = viewKanban
			if tt.name == "empty columns returns false" {
				m.columns = []domain.Column{}
			} else {
				m.columns = columns
			}
			m.tasks = tasks
			m.activeColumn = tt.activeColumn
			m.kanbanRow = tt.kanbanRow
			task, ok := m.currentTask()
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && task.ID != tt.wantID {
				t.Errorf("task.ID = %q, want %q", task.ID, tt.wantID)
			}
		})
	}
}

func TestModelTasksForColumn(t *testing.T) {
	col1 := "col-1"
	col2 := "col-2"
	col3 := "col-3"

	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", ColumnID: &col1, Position: 1, UpdatedAt: now.Add(time.Minute)},
		{ID: "t2", ColumnID: &col1, Position: 2, UpdatedAt: now},
		{ID: "t3", ColumnID: &col2, Position: 1, UpdatedAt: now},
		{ID: "t4", ColumnID: nil, Position: 1, UpdatedAt: now},
	}

	m := makeModelSelectionTestModel()
	m.tasks = tasks

	t.Run("returns tasks for column sorted by position", func(t *testing.T) {
		got := m.tasksForColumn(col1)
		if len(got) != 2 {
			t.Fatalf("expected 2 tasks, got %d", len(got))
		}
		if got[0].ID != "t1" || got[1].ID != "t2" {
			t.Errorf("expected [t1, t2] by position, got [%s, %s]", got[0].ID, got[1].ID)
		}
	})

	t.Run("returns empty for column with no tasks", func(t *testing.T) {
		got := m.tasksForColumn(col3)
		if len(got) != 0 {
			t.Errorf("expected 0 tasks, got %d", len(got))
		}
	})

	t.Run("excludes tasks with nil column", func(t *testing.T) {
		got := m.tasksForColumn(col2)
		if len(got) != 1 || got[0].ID != "t3" {
			t.Errorf("expected [t3], got %v", got)
		}
	})
}
