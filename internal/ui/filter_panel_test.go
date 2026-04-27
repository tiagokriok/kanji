package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tiagokriok/kanji/internal/domain"
)

// --- label helper tests ---

func TestStatusFilterLabel_All(t *testing.T) {
	m := Model{filterIndex: -1, columnFilter: ""}
	if got := m.statusFilterLabel(); got != "All" {
		t.Errorf("statusFilterLabel() = %q, want %q", got, "All")
	}
}

func TestStatusFilterLabel_CustomColumnFilter(t *testing.T) {
	m := Model{
		filterIndex:  -1,
		columnFilter: "c1",
		columns:      []domain.Column{{ID: "c1", Name: "Todo"}},
	}
	if got := m.statusFilterLabel(); got != "Todo" {
		t.Errorf("statusFilterLabel() = %q, want %q", got, "Todo")
	}
}

func TestStatusFilterLabel_ByIndex(t *testing.T) {
	m := Model{
		filterIndex: 0,
		columns:     []domain.Column{{ID: "c1", Name: "Done"}},
	}
	if got := m.statusFilterLabel(); got != "Done" {
		t.Errorf("statusFilterLabel() = %q, want %q", got, "Done")
	}
}

func TestDueFilterLabel_Any(t *testing.T) {
	m := Model{dueFilter: dueFilterAny}
	if got := m.dueFilterLabel(); got != "Any" {
		t.Errorf("dueFilterLabel() = %q, want %q", got, "Any")
	}
}

func TestDueFilterLabel_Soon(t *testing.T) {
	m := Model{dueFilter: dueFilterSoon}
	if got := m.dueFilterLabel(); got != "Due in 7d" {
		t.Errorf("dueFilterLabel() = %q, want %q", got, "Due in 7d")
	}
}

func TestDueFilterLabel_Overdue(t *testing.T) {
	m := Model{dueFilter: dueFilterOverdue}
	if got := m.dueFilterLabel(); got != "Overdue" {
		t.Errorf("dueFilterLabel() = %q, want %q", got, "Overdue")
	}
}

func TestDueFilterLabel_NoDate(t *testing.T) {
	m := Model{dueFilter: dueFilterNoDate}
	if got := m.dueFilterLabel(); got != "No due date" {
		t.Errorf("dueFilterLabel() = %q, want %q", got, "No due date")
	}
}

func TestPriorityFilterLabel_All(t *testing.T) {
	m := Model{priorityFilter: -1}
	if got := m.priorityFilterLabel(); got != "All" {
		t.Errorf("priorityFilterLabel() = %q, want %q", got, "All")
	}
}

func TestPriorityFilterLabel_Critical(t *testing.T) {
	m := Model{priorityFilter: 0}
	if got := m.priorityFilterLabel(); got != "Critical (0)" {
		t.Errorf("priorityFilterLabel() = %q, want %q", got, "Critical (0)")
	}
}

func TestPriorityFilterLabel_None(t *testing.T) {
	m := Model{priorityFilter: 5}
	if got := m.priorityFilterLabel(); got != "None (5)" {
		t.Errorf("priorityFilterLabel() = %q, want %q", got, "None (5)")
	}
}

func TestSortModeLabel_Priority(t *testing.T) {
	m := Model{sortMode: sortByPriority}
	if got := m.sortModeLabel(); got != "Priority" {
		t.Errorf("sortModeLabel() = %q, want %q", got, "Priority")
	}
}

func TestSortModeLabel_DueDate(t *testing.T) {
	m := Model{sortMode: sortByDueDate}
	if got := m.sortModeLabel(); got != "Due date" {
		t.Errorf("sortModeLabel() = %q, want %q", got, "Due date")
	}
}

func TestSortModeLabel_Title(t *testing.T) {
	m := Model{sortMode: sortByTitle}
	if got := m.sortModeLabel(); got != "Title" {
		t.Errorf("sortModeLabel() = %q, want %q", got, "Title")
	}
}

func TestSortModeLabel_Updated(t *testing.T) {
	m := Model{sortMode: sortByUpdated}
	if got := m.sortModeLabel(); got != "Updated" {
		t.Errorf("sortModeLabel() = %q, want %q", got, "Updated")
	}
}

func TestSortModeLabel_Created(t *testing.T) {
	m := Model{sortMode: sortByCreated}
	if got := m.sortModeLabel(); got != "Created" {
		t.Errorf("sortModeLabel() = %q, want %q", got, "Created")
	}
}

// --- filter state mutator tests ---

func TestSetStatusFilterByIndex_Valid(t *testing.T) {
	m := Model{columns: []domain.Column{{ID: "c1", Name: "Todo"}}, filterIndex: -1}
	m.setStatusFilterByIndex(0)
	if m.filterIndex != 0 || m.columnFilter != "c1" {
		t.Errorf("filterIndex = %d, columnFilter = %q; want 0, c1", m.filterIndex, m.columnFilter)
	}
}

func TestSetStatusFilterByIndex_Negative(t *testing.T) {
	m := Model{columns: []domain.Column{{ID: "c1"}}, filterIndex: 0, columnFilter: "c1"}
	m.setStatusFilterByIndex(-1)
	if m.filterIndex != -1 || m.columnFilter != "" {
		t.Errorf("filterIndex = %d, columnFilter = %q; want -1, empty", m.filterIndex, m.columnFilter)
	}
}

func TestSetStatusFilterByIndex_OutOfRange(t *testing.T) {
	m := Model{columns: []domain.Column{{ID: "c1"}}, filterIndex: 0, columnFilter: "c1"}
	m.setStatusFilterByIndex(5)
	if m.filterIndex != -1 || m.columnFilter != "" {
		t.Errorf("filterIndex = %d, columnFilter = %q; want -1, empty", m.filterIndex, m.columnFilter)
	}
}

func TestCycleDueFilter(t *testing.T) {
	m := Model{dueFilter: dueFilterAny}
	m.cycleDueFilter()
	if m.dueFilter != dueFilterSoon {
		t.Errorf("dueFilter = %v, want dueFilterSoon", m.dueFilter)
	}
	m.cycleDueFilter()
	if m.dueFilter != dueFilterOverdue {
		t.Errorf("dueFilter = %v, want dueFilterOverdue", m.dueFilter)
	}
	m.cycleDueFilter()
	if m.dueFilter != dueFilterNoDate {
		t.Errorf("dueFilter = %v, want dueFilterNoDate", m.dueFilter)
	}
	m.cycleDueFilter()
	if m.dueFilter != dueFilterAny {
		t.Errorf("dueFilter = %v, want dueFilterAny", m.dueFilter)
	}
}

func TestCycleSortMode(t *testing.T) {
	m := Model{sortMode: sortByPriority}
	m.cycleSortMode()
	if m.sortMode != sortByDueDate {
		t.Errorf("sortMode = %v, want sortByDueDate", m.sortMode)
	}
	m.cycleSortMode()
	if m.sortMode != sortByTitle {
		t.Errorf("sortMode = %v, want sortByTitle", m.sortMode)
	}
	m.cycleSortMode()
	if m.sortMode != sortByUpdated {
		t.Errorf("sortMode = %v, want sortByUpdated", m.sortMode)
	}
	m.cycleSortMode()
	if m.sortMode != sortByCreated {
		t.Errorf("sortMode = %v, want sortByCreated", m.sortMode)
	}
	m.cycleSortMode()
	if m.sortMode != sortByPriority {
		t.Errorf("sortMode = %v, want sortByPriority", m.sortMode)
	}
}

func TestCycleColumnFilter_WithColumns(t *testing.T) {
	m := Model{columns: []domain.Column{{ID: "c1"}, {ID: "c2"}}, filterIndex: -1}
	m.cycleColumnFilter()
	if m.filterIndex != 0 || m.columnFilter != "c1" {
		t.Errorf("filterIndex = %d, columnFilter = %q; want 0, c1", m.filterIndex, m.columnFilter)
	}
	m.cycleColumnFilter()
	if m.filterIndex != 1 || m.columnFilter != "c2" {
		t.Errorf("filterIndex = %d, columnFilter = %q; want 1, c2", m.filterIndex, m.columnFilter)
	}
	m.cycleColumnFilter()
	if m.filterIndex != -1 || m.columnFilter != "" {
		t.Errorf("filterIndex = %d, columnFilter = %q; want -1, empty", m.filterIndex, m.columnFilter)
	}
}

func TestCycleColumnFilter_EmptyColumns(t *testing.T) {
	m := Model{columns: []domain.Column{}, filterIndex: 0, columnFilter: "c1"}
	m.cycleColumnFilter()
	if m.filterIndex != -1 || m.columnFilter != "" {
		t.Errorf("filterIndex = %d, columnFilter = %q; want -1, empty", m.filterIndex, m.columnFilter)
	}
}

// --- adjustFilterSelection tests ---

func TestAdjustFilterSelection_Status(t *testing.T) {
	m := Model{
		overlayState: overlayState{filterFocus: 2},
		columns:      []domain.Column{{ID: "c1"}},
		filterIndex:  -1,
	}
	changed, err := m.adjustFilterSelection(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Error("expected changed=true")
	}
	if m.filterIndex != 0 {
		t.Errorf("filterIndex = %d, want 0", m.filterIndex)
	}
}

func TestAdjustFilterSelection_Due(t *testing.T) {
	m := Model{overlayState: overlayState{filterFocus: 3}, dueFilter: dueFilterAny}
	changed, err := m.adjustFilterSelection(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Error("expected changed=true")
	}
	if m.dueFilter != dueFilterSoon {
		t.Errorf("dueFilter = %v, want dueFilterSoon", m.dueFilter)
	}
}

func TestAdjustFilterSelection_Priority(t *testing.T) {
	m := Model{overlayState: overlayState{filterFocus: 4}, priorityFilter: -1}
	changed, err := m.adjustFilterSelection(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Error("expected changed=true")
	}
	if m.priorityFilter != 0 {
		t.Errorf("priorityFilter = %d, want 0", m.priorityFilter)
	}
}

func TestAdjustFilterSelection_Sort(t *testing.T) {
	m := Model{overlayState: overlayState{filterFocus: 5}, sortMode: sortByPriority}
	changed, err := m.adjustFilterSelection(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Error("expected changed=true")
	}
	if m.sortMode != sortByDueDate {
		t.Errorf("sortMode = %v, want sortByDueDate", m.sortMode)
	}
}

func TestAdjustFilterSelection_WorkspaceEmpty(t *testing.T) {
	m := Model{overlayState: overlayState{filterFocus: 0}, workspaces: []domain.Workspace{}}
	changed, err := m.adjustFilterSelection(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if changed {
		t.Error("expected changed=false for empty workspaces")
	}
}

func TestAdjustFilterSelection_BoardEmpty(t *testing.T) {
	m := Model{overlayState: overlayState{filterFocus: 1}, boards: []domain.Board{}}
	changed, err := m.adjustFilterSelection(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if changed {
		t.Error("expected changed=false for empty boards")
	}
}

// --- panel lifecycle tests ---

func TestOpenFilterPanel(t *testing.T) {
	m := Model{}
	m.openFilterPanel()
	if !m.showFilters {
		t.Error("expected showFilters to be true")
	}
	if m.filterFocus != 0 {
		t.Errorf("filterFocus = %d, want 0", m.filterFocus)
	}
}

func TestCloseFilterPanel(t *testing.T) {
	m := Model{overlayState: overlayState{showFilters: true, filterFocus: 3}}
	m.closeFilterPanel()
	if m.showFilters {
		t.Error("expected showFilters to be false")
	}
}

// --- updateFilterPanel tests ---

func TestUpdateFilterPanel_Cancel(t *testing.T) {
	m := Model{overlayState: overlayState{showFilters: true}, keys: newKeyMap()}
	updated, cmd := m.updateFilterPanel(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("esc")})
	um := updated.(Model)
	if um.showFilters {
		t.Error("expected showFilters to be false after cancel")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateFilterPanel_ShowFiltersKey(t *testing.T) {
	m := Model{overlayState: overlayState{showFilters: true}, keys: newKeyMap()}
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")}
	if !key.Matches(msg, m.keys.ShowFilters) {
		t.Fatal("test setup error: key message does not match ShowFilters")
	}
	updated, cmd := m.updateFilterPanel(msg)
	um := updated.(Model)
	if um.showFilters {
		t.Error("expected showFilters to be false")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateFilterPanel_Up(t *testing.T) {
	m := Model{overlayState: overlayState{showFilters: true, filterFocus: 2}, keys: newKeyMap()}
	updated, cmd := m.updateFilterPanel(tea.KeyMsg{Type: tea.KeyUp})
	um := updated.(Model)
	if um.filterFocus != 1 {
		t.Errorf("filterFocus = %d, want 1", um.filterFocus)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateFilterPanel_Down(t *testing.T) {
	m := Model{overlayState: overlayState{showFilters: true, filterFocus: 5}, keys: newKeyMap()}
	updated, cmd := m.updateFilterPanel(tea.KeyMsg{Type: tea.KeyDown})
	um := updated.(Model)
	if um.filterFocus != 0 {
		t.Errorf("filterFocus = %d, want 0", um.filterFocus)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateFilterPanel_Left(t *testing.T) {
	m := Model{
		overlayState: overlayState{showFilters: true, filterFocus: 3},
		dueFilter:    dueFilterSoon,
		keys:         newKeyMap(),
	}
	updated, cmd := m.updateFilterPanel(tea.KeyMsg{Type: tea.KeyLeft})
	um := updated.(Model)
	if um.dueFilter != dueFilterAny {
		t.Errorf("dueFilter = %v, want dueFilterAny", um.dueFilter)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd after left adjustment")
	}
}

func TestUpdateFilterPanel_Right(t *testing.T) {
	m := Model{
		overlayState: overlayState{showFilters: true, filterFocus: 3},
		dueFilter:    dueFilterAny,
		keys:         newKeyMap(),
	}
	updated, cmd := m.updateFilterPanel(tea.KeyMsg{Type: tea.KeyRight})
	um := updated.(Model)
	if um.dueFilter != dueFilterSoon {
		t.Errorf("dueFilter = %v, want dueFilterSoon", um.dueFilter)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd after right adjustment")
	}
}

func TestUpdateFilterPanel_Confirm(t *testing.T) {
	m := Model{overlayState: overlayState{showFilters: true}, keys: newKeyMap()}
	updated, cmd := m.updateFilterPanel(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(Model)
	if um.showFilters {
		t.Error("expected showFilters to be false after confirm")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateFilterPanel_TasksLoaded(t *testing.T) {
	m := Model{
		overlayState: overlayState{showFilters: true},
		tasks:        []domain.Task{{ID: "t1", Title: "Old"}},
		selected:     0,
		columns:      []domain.Column{{ID: "c1", Name: "Todo"}},
		width:        80,
		height:       24,
	}
	updated, cmd := m.updateFilterPanel(tasksLoadedMsg{tasks: []domain.Task{{ID: "t2", Title: "New"}}})
	um := updated.(Model)
	if len(um.tasks) != 1 || um.tasks[0].ID != "t2" {
		t.Error("expected tasks to be updated")
	}
	if cmd != nil {
		t.Error("expected nil cmd when refreshDetails is false")
	}
}

func TestUpdateFilterPanel_OpResultError(t *testing.T) {
	m := Model{overlayState: overlayState{showFilters: true}, statusLine: ""}
	updated, cmd := m.updateFilterPanel(opResultMsg{err: errTest("boom")})
	um := updated.(Model)
	if um.statusLine != "boom" {
		t.Errorf("statusLine = %q, want %q", um.statusLine, "boom")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateFilterPanel_WindowSize(t *testing.T) {
	m := Model{overlayState: overlayState{showFilters: true}, width: 80, height: 24}
	updated, cmd := m.updateFilterPanel(tea.WindowSizeMsg{Width: 100, Height: 40})
	um := updated.(Model)
	if um.width != 100 || um.height != 40 {
		t.Errorf("width=%d height=%d; want 100, 40", um.width, um.height)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

// --- renderFilterPanel tests ---

func TestRenderFilterPanel_ContainsTitle(t *testing.T) {
	m := Model{
		overlayState:  overlayState{showFilters: true},
		width:         80,
		height:        24,
		workspaceName: "WS",
		boardName:     "Board",
		columns:       []domain.Column{{ID: "c1", Name: "Todo"}},
		keys:          newKeyMap(),
	}
	out := m.renderFilterPanel("base")
	if !strings.Contains(out, "Filter & Sort") {
		t.Error("expected output to contain 'Filter & Sort'")
	}
}

func TestRenderFilterPanel_ContainsRows(t *testing.T) {
	m := Model{
		overlayState:   overlayState{showFilters: true},
		width:          80,
		height:         24,
		workspaceName:  "Dev",
		boardName:      "Main",
		columns:        []domain.Column{{ID: "c1", Name: "Todo"}},
		filterIndex:    0,
		dueFilter:      dueFilterSoon,
		priorityFilter: 1,
		sortMode:       sortByDueDate,
		keys:           newKeyMap(),
	}
	out := m.renderFilterPanel("base")
	for _, want := range []string{"Workspace", "Board", "Status", "Due", "Priority", "Sort", "Dev", "Main", "Todo", "Due in 7d", "Urgent (1)", "Due date"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q", want)
		}
	}
}
