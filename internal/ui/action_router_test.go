package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tiagokriok/kanji/internal/domain"
)

func TestExecuteAction_Quit(t *testing.T) {
	m := Model{}
	_, cmd := m.executeAction("quit")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestExecuteAction_ToggleView(t *testing.T) {
	m := Model{viewMode: viewList, tasks: []domain.Task{{ID: "t1"}}, selected: 0}
	updated, cmd := m.executeAction("toggle_view")
	um := updated.(Model)
	if um.viewMode != viewKanban {
		t.Errorf("viewMode = %v, want viewKanban", um.viewMode)
	}
	if cmd != nil {
		t.Error("expected nil cmd when showDetails is false")
	}
}

func TestExecuteAction_ToggleDetails(t *testing.T) {
	m := Model{showDetails: false}
	updated, cmd := m.executeAction("toggle_details")
	um := updated.(Model)
	if !um.showDetails {
		t.Error("expected showDetails to be true")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestExecuteAction_OpenFilters(t *testing.T) {
	m := Model{}
	updated, cmd := m.executeAction("open_filters")
	um := updated.(Model)
	if !um.showFilters {
		t.Error("expected showFilters to be true")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestExecuteAction_ClearSearch(t *testing.T) {
	m := Model{titleFilter: "query"}
	updated, cmd := m.executeAction("clear_search")
	um := updated.(Model)
	if um.titleFilter != "" {
		t.Errorf("titleFilter = %q, want empty", um.titleFilter)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestExecuteAction_ClearSearchEmpty(t *testing.T) {
	m := Model{titleFilter: ""}
	updated, cmd := m.executeAction("clear_search")
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	_ = updated
}

func TestExecuteAction_CycleStatus(t *testing.T) {
	m := Model{columns: []domain.Column{{ID: "c1", Name: "Todo"}}, filterIndex: 0, columnFilter: "c1"}
	updated, cmd := m.executeAction("cycle_status")
	um := updated.(Model)
	if um.filterIndex != -1 {
		t.Errorf("filterIndex = %d, want -1", um.filterIndex)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestExecuteAction_MoveUp(t *testing.T) {
	m := Model{
		tasks:    []domain.Task{{ID: "t1"}, {ID: "t2"}},
		selected: 1,
	}
	updated, cmd := m.executeAction("move_up")
	um := updated.(Model)
	if um.selected != 0 {
		t.Errorf("selected = %d, want 0", um.selected)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestExecuteAction_MoveTask(t *testing.T) {
	m := Model{
		tasks:    []domain.Task{{ID: "t1", ColumnID: strPtr("c1")}},
		selected: 0,
		columns:  []domain.Column{{ID: "c1", Name: "Todo"}, {ID: "c2", Name: "Doing"}},
	}
	updated, cmd := m.executeAction("move_task")
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
	_ = updated
}

func TestExecuteAction_MoveTaskNoTask(t *testing.T) {
	m := Model{tasks: []domain.Task{}}
	updated, cmd := m.executeAction("move_task")
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	_ = updated
}

func TestExecuteAction_Default(t *testing.T) {
	m := Model{}
	updated, cmd := m.executeAction("unknown_action")
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	_ = updated
}

func TestExecuteAction_Search(t *testing.T) {
	m := Model{}
	m.textInput = textinput.New()
	updated, cmd := m.executeAction("search")
	um := updated.(Model)
	if um.inputMode != inputSearch {
		t.Errorf("inputMode = %v, want inputSearch", um.inputMode)
	}
	if um.statusLine != "Search by title" {
		t.Errorf("statusLine = %q, want %q", um.statusLine, "Search by title")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestExecuteAction_OpenWorkspaces(t *testing.T) {
	m := Model{}
	m.contextFilter = textinput.New()
	updated, cmd := m.executeAction("open_workspaces")
	um := updated.(Model)
	if um.activeOverlay() != overlayContexts {
		t.Errorf("activeOverlay = %v, want overlayContexts", um.activeOverlay())
	}
	if um.contextMode != contextWorkspace {
		t.Errorf("contextMode = %v, want contextWorkspace", um.overlayState.contextMode)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestExecuteAction_OpenBoardPanel(t *testing.T) {
	m := Model{}
	m.contextFilter = textinput.New()
	updated, cmd := m.executeAction("open_board_panel")
	um := updated.(Model)
	if um.activeOverlay() != overlayContexts {
		t.Errorf("activeOverlay = %v, want overlayContexts", um.activeOverlay())
	}
	if um.contextMode != contextBoard {
		t.Errorf("contextMode = %v, want contextBoard", um.overlayState.contextMode)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestExecuteAction_PrevBoardNoChange(t *testing.T) {
	m := Model{boards: []domain.Board{{ID: "b1"}}, boardID: "b1"}
	updated, cmd := m.executeAction("prev_board")
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	_ = updated
}

func TestExecuteAction_NextBoardNoChange(t *testing.T) {
	m := Model{boards: []domain.Board{{ID: "b1"}}, boardID: "b1"}
	updated, cmd := m.executeAction("next_board")
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	_ = updated
}

func TestExecuteAction_NewTask(t *testing.T) {
	m := Model{}
	m.textInput = textinput.New()
	updated, cmd := m.executeAction("new_task")
	um := updated.(Model)
	if um.inputMode != inputTaskForm {
		t.Errorf("inputMode = %v, want inputTaskForm", um.inputMode)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestExecuteAction_EditTask(t *testing.T) {
	m := Model{tasks: []domain.Task{{ID: "t1", Title: "Task"}}, selected: 0}
	m.textInput = textinput.New()
	updated, cmd := m.executeAction("edit_task")
	um := updated.(Model)
	if um.inputMode != inputTaskForm {
		t.Errorf("inputMode = %v, want inputTaskForm", um.inputMode)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestExecuteAction_EditTaskNoTask(t *testing.T) {
	m := Model{tasks: []domain.Task{}}
	updated, cmd := m.executeAction("edit_task")
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	_ = updated
}

func TestExecuteAction_AddComment(t *testing.T) {
	m := Model{tasks: []domain.Task{{ID: "t1"}}, selected: 0}
	m.textInput = textinput.New()
	updated, cmd := m.executeAction("add_comment")
	um := updated.(Model)
	if um.inputMode != inputAddComment {
		t.Errorf("inputMode = %v, want inputAddComment", um.inputMode)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestExecuteAction_AddCommentNoTask(t *testing.T) {
	m := Model{tasks: []domain.Task{}}
	updated, cmd := m.executeAction("add_comment")
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	_ = updated
}

func TestExecuteAction_EditDescription(t *testing.T) {
	m := Model{tasks: []domain.Task{{ID: "t1"}}, selected: 0}
	updated, cmd := m.executeAction("edit_description")
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
	_ = updated
}

func TestExecuteAction_EditDescriptionNoTask(t *testing.T) {
	m := Model{tasks: []domain.Task{}}
	updated, cmd := m.executeAction("edit_description")
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	_ = updated
}

func TestExecuteAction_CycleDueFilter(t *testing.T) {
	m := Model{dueFilter: dueFilterAny}
	updated, cmd := m.executeAction("cycle_due_filter")
	um := updated.(Model)
	if um.dueFilter != dueFilterSoon {
		t.Errorf("dueFilter = %v, want dueFilterSoon", um.dueFilter)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestExecuteAction_CycleSort(t *testing.T) {
	m := Model{sortMode: sortByPriority}
	updated, cmd := m.executeAction("cycle_sort")
	um := updated.(Model)
	if um.sortMode != sortByDueDate {
		t.Errorf("sortMode = %v, want sortByDueDate", um.sortMode)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestExecuteAction_MoveDown(t *testing.T) {
	m := Model{tasks: []domain.Task{{ID: "t1"}, {ID: "t2"}}, selected: 0}
	updated, cmd := m.executeAction("move_down")
	um := updated.(Model)
	if um.selected != 1 {
		t.Errorf("selected = %d, want 1", um.selected)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestExecuteAction_MoveLeftKanban(t *testing.T) {
	m := Model{viewMode: viewKanban, activeColumn: 1, columns: []domain.Column{{ID: "c1"}, {ID: "c2"}}}
	updated, cmd := m.executeAction("move_left")
	um := updated.(Model)
	if um.activeColumn != 0 {
		t.Errorf("activeColumn = %d, want 0", um.activeColumn)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestExecuteAction_MoveLeftNotKanban(t *testing.T) {
	m := Model{viewMode: viewList}
	updated, cmd := m.executeAction("move_left")
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	_ = updated
}

func TestExecuteAction_MoveRightKanban(t *testing.T) {
	m := Model{viewMode: viewKanban, activeColumn: 0, columns: []domain.Column{{ID: "c1"}, {ID: "c2"}}}
	updated, cmd := m.executeAction("move_right")
	um := updated.(Model)
	if um.activeColumn != 1 {
		t.Errorf("activeColumn = %d, want 1", um.activeColumn)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestExecuteAction_MoveRightNotKanban(t *testing.T) {
	m := Model{viewMode: viewList}
	updated, cmd := m.executeAction("move_right")
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	_ = updated
}

func TestExecuteAction_MoveTaskLeftKanban(t *testing.T) {
	m := Model{
		viewMode:     viewKanban,
		activeColumn: 1,
		tasks:        []domain.Task{{ID: "t1", ColumnID: strPtr("c2")}},
		selected:     0,
		columns:      []domain.Column{{ID: "c1"}, {ID: "c2"}},
	}
	updated, cmd := m.executeAction("move_task_left")
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
	_ = updated
}

func TestExecuteAction_MoveTaskLeftNotKanban(t *testing.T) {
	m := Model{viewMode: viewList}
	updated, cmd := m.executeAction("move_task_left")
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	_ = updated
}

func TestExecuteAction_MoveTaskRightKanban(t *testing.T) {
	m := Model{
		viewMode: viewKanban,
		tasks:    []domain.Task{{ID: "t1", ColumnID: strPtr("c1")}},
		selected: 0,
		columns:  []domain.Column{{ID: "c1"}, {ID: "c2"}},
	}
	updated, cmd := m.executeAction("move_task_right")
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
	_ = updated
}

func TestExecuteAction_MoveTaskRightNotKanban(t *testing.T) {
	m := Model{viewMode: viewList}
	updated, cmd := m.executeAction("move_task_right")
	if cmd != nil {
		t.Error("expected nil cmd")
	}
	_ = updated
}

func TestExecuteAction_OpenMove(t *testing.T) {
	m := Model{tasks: []domain.Task{{ID: "t1"}}, selected: 0}
	updated, cmd := m.executeAction("open_move")
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
	_ = updated
}
