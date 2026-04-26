package ui

import (
	"testing"
	"time"

	"github.com/tiagokriok/kanji/internal/domain"
)

func TestSelectionEnsureSelectionListView(t *testing.T) {
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
			s := &selectionState{
				viewMode: viewList,
				tasks:    tt.tasks,
				selected: tt.selected,
			}
			s.ensureSelection()
			if s.selected != tt.want {
				t.Errorf("selected = %d, want %d", s.selected, tt.want)
			}
		})
	}
}

func TestSelectionEnsureSelectionKanbanView(t *testing.T) {
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
			s := &selectionState{
				viewMode:     viewKanban,
				columns:      tt.columns,
				tasks:        tt.tasks,
				activeColumn: tt.activeColumn,
				kanbanRow:    tt.kanbanRow,
			}
			s.ensureSelection()
			if s.activeColumn != tt.wantColumn {
				t.Errorf("activeColumn = %d, want %d", s.activeColumn, tt.wantColumn)
			}
			if s.kanbanRow != tt.wantRow {
				t.Errorf("kanbanRow = %d, want %d", s.kanbanRow, tt.wantRow)
			}
		})
	}
}

func TestSelectionEnsureKanbanRow(t *testing.T) {
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
			s := &selectionState{
				viewMode:     viewKanban,
				columns:      columns,
				tasks:        tt.tasks,
				activeColumn: tt.activeColumn,
				kanbanRow:    tt.kanbanRow,
			}
			s.ensureKanbanRow()
			if s.kanbanRow != tt.want {
				t.Errorf("kanbanRow = %d, want %d", s.kanbanRow, tt.want)
			}
		})
	}
}

func TestSelectionSetActiveColumnByID(t *testing.T) {
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
		{
			name:       "whitespace id does nothing",
			columnID:   "   ",
			wantColumn: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &selectionState{
				columns:      columns,
				activeColumn: 0,
			}
			s.setActiveColumnByID(tt.columnID)
			if s.activeColumn != tt.wantColumn {
				t.Errorf("activeColumn = %d, want %d", s.activeColumn, tt.wantColumn)
			}
		})
	}
}

func TestSelectionRestorePendingKanbanSelection(t *testing.T) {
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
			name:            "empty pending column returns false",
			viewMode:        viewKanban,
			pendingTaskID:   task1,
			pendingColID:    "",
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
			name:            "task not found in column clears pending and returns false",
			viewMode:        viewKanban,
			pendingTaskID:   "task-unknown",
			pendingColID:    col1,
			tasks:           tasks,
			wantOK:          false,
			wantColumn:      0, // setActiveColumnByID still runs
			wantRow:         0,
			wantPendingTask: "",
			wantPendingCol:  "",
		},
		{
			name:            "column not found still clears pending",
			viewMode:        viewKanban,
			pendingTaskID:   task1,
			pendingColID:    "col-unknown",
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
			s := &selectionState{
				viewMode:              tt.viewMode,
				columns:               columns,
				tasks:                 tt.tasks,
				activeColumn:          0,
				kanbanRow:             0,
				pendingKanbanTaskID:   tt.pendingTaskID,
				pendingKanbanColumnID: tt.pendingColID,
			}
			got := s.restorePendingKanbanSelection()
			if got != tt.wantOK {
				t.Errorf("restorePendingKanbanSelection() = %v, want %v", got, tt.wantOK)
			}
			if s.activeColumn != tt.wantColumn {
				t.Errorf("activeColumn = %d, want %d", s.activeColumn, tt.wantColumn)
			}
			if s.kanbanRow != tt.wantRow {
				t.Errorf("kanbanRow = %d, want %d", s.kanbanRow, tt.wantRow)
			}
			if s.pendingKanbanTaskID != tt.wantPendingTask {
				t.Errorf("pendingKanbanTaskID = %q, want %q", s.pendingKanbanTaskID, tt.wantPendingTask)
			}
			if s.pendingKanbanColumnID != tt.wantPendingCol {
				t.Errorf("pendingKanbanColumnID = %q, want %q", s.pendingKanbanColumnID, tt.wantPendingCol)
			}
		})
	}
}

func TestSelectionMoveUpDownListView(t *testing.T) {
	tasks := []domain.Task{
		{ID: "t1"}, {ID: "t2"}, {ID: "t3"},
	}

	t.Run("moveUp decrements selection", func(t *testing.T) {
		s := &selectionState{
			viewMode: viewList,
			tasks:    tasks,
			selected: 2,
		}
		s.moveUp()
		if s.selected != 1 {
			t.Errorf("selected = %d, want 1", s.selected)
		}
	})

	t.Run("moveDown increments selection", func(t *testing.T) {
		s := &selectionState{
			viewMode: viewList,
			tasks:    tasks,
			selected: 0,
		}
		s.moveDown()
		if s.selected != 1 {
			t.Errorf("selected = %d, want 1", s.selected)
		}
	})

	t.Run("moveUp clamps to 0", func(t *testing.T) {
		s := &selectionState{
			viewMode: viewList,
			tasks:    tasks,
			selected: 0,
		}
		s.moveUp()
		if s.selected != 0 {
			t.Errorf("selected = %d, want 0", s.selected)
		}
	})

	t.Run("moveDown clamps to last index", func(t *testing.T) {
		s := &selectionState{
			viewMode: viewList,
			tasks:    tasks,
			selected: 2,
		}
		s.moveDown()
		if s.selected != 2 {
			t.Errorf("selected = %d, want 2", s.selected)
		}
	})
}

func TestSelectionMoveUpDownKanbanView(t *testing.T) {
	col1 := "col-1"
	columns := []domain.Column{{ID: col1, Name: "Todo"}}
	tasks := []domain.Task{
		{ID: "t1", ColumnID: &col1},
		{ID: "t2", ColumnID: &col1},
		{ID: "t3", ColumnID: &col1},
	}

	t.Run("moveUp decrements kanban row", func(t *testing.T) {
		s := &selectionState{
			viewMode:     viewKanban,
			columns:      columns,
			tasks:        tasks,
			activeColumn: 0,
			kanbanRow:    2,
		}
		s.moveUp()
		if s.kanbanRow != 1 {
			t.Errorf("kanbanRow = %d, want 1", s.kanbanRow)
		}
	})

	t.Run("moveDown increments kanban row", func(t *testing.T) {
		s := &selectionState{
			viewMode:     viewKanban,
			columns:      columns,
			tasks:        tasks,
			activeColumn: 0,
			kanbanRow:    0,
		}
		s.moveDown()
		if s.kanbanRow != 1 {
			t.Errorf("kanbanRow = %d, want 1", s.kanbanRow)
		}
	})

	t.Run("moveUp clamps to 0", func(t *testing.T) {
		s := &selectionState{
			viewMode:     viewKanban,
			columns:      columns,
			tasks:        tasks,
			activeColumn: 0,
			kanbanRow:    0,
		}
		s.moveUp()
		if s.kanbanRow != 0 {
			t.Errorf("kanbanRow = %d, want 0", s.kanbanRow)
		}
	})

	t.Run("moveDown clamps to last row", func(t *testing.T) {
		s := &selectionState{
			viewMode:     viewKanban,
			columns:      columns,
			tasks:        tasks,
			activeColumn: 0,
			kanbanRow:    2,
		}
		s.moveDown()
		if s.kanbanRow != 2 {
			t.Errorf("kanbanRow = %d, want 2", s.kanbanRow)
		}
	})
}

func TestSelectionCurrentTaskListView(t *testing.T) {
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
			taskList := tasks
			if tt.name == "empty tasks returns false" {
				taskList = []domain.Task{}
			}
			s := &selectionState{
				viewMode: viewList,
				tasks:    taskList,
				selected: tt.selected,
			}
			task, ok := s.currentTask()
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && task.ID != tt.wantID {
				t.Errorf("task.ID = %q, want %q", task.ID, tt.wantID)
			}
		})
	}
}

func TestSelectionCurrentTaskKanbanView(t *testing.T) {
	col1 := "col-1"
	col2 := "col-2"
	columns := []domain.Column{
		{ID: col1, Name: "Todo"},
		{ID: col2, Name: "Doing"},
	}

	now := time.Now().UTC()
	// Tasks with explicit positions - t1 at position 1, t2 at position 2
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
			cols := columns
			if tt.name == "empty columns returns false" {
				cols = []domain.Column{}
			}
			s := &selectionState{
				viewMode:     viewKanban,
				columns:      cols,
				tasks:        tasks,
				activeColumn: tt.activeColumn,
				kanbanRow:    tt.kanbanRow,
			}
			task, ok := s.currentTask()
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && task.ID != tt.wantID {
				t.Errorf("task.ID = %q, want %q", task.ID, tt.wantID)
			}
		})
	}
}

func TestSelectionTasksForColumn(t *testing.T) {
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

	s := &selectionState{tasks: tasks}

	t.Run("returns tasks for column sorted by position", func(t *testing.T) {
		got := s.tasksForColumn(col1)
		if len(got) != 2 {
			t.Fatalf("expected 2 tasks, got %d", len(got))
		}
		if got[0].ID != "t1" || got[1].ID != "t2" {
			t.Errorf("expected [t1, t2] by position, got [%s, %s]", got[0].ID, got[1].ID)
		}
	})

	t.Run("returns empty for column with no tasks", func(t *testing.T) {
		got := s.tasksForColumn(col3)
		if len(got) != 0 {
			t.Errorf("expected 0 tasks, got %d", len(got))
		}
	})

	t.Run("excludes tasks with nil column", func(t *testing.T) {
		got := s.tasksForColumn(col2)
		if len(got) != 1 || got[0].ID != "t3" {
			t.Errorf("expected [t3], got %v", got)
		}
	})
}
