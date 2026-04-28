package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
)

func TestRenderHeader_ListView(t *testing.T) {
	m := Model{
		workspaceName:  "Dev",
		boardName:      "Board",
		viewMode:       viewList,
		sortMode:       sortByPriority,
		filterIndex:    -1,
		dueFilter:      dueFilterAny,
		priorityFilter: -1,
	}
	out := m.renderHeader(80)
	for _, want := range []string{"Dev / Board", "view:List", "sort:priority", "status:All", "due:any", "search:\"\""} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestRenderHeader_KanbanView(t *testing.T) {
	m := Model{
		workspaceName:  "Dev",
		boardName:      "Board",
		viewMode:       viewKanban,
		sortMode:       sortByTitle,
		filterIndex:    -1,
		dueFilter:      dueFilterSoon,
		priorityFilter: -1,
	}
	out := m.renderHeader(80)
	if !strings.Contains(out, "Kanban") {
		t.Errorf("expected output to contain 'Kanban', got:\n%s", out)
	}
	if !strings.Contains(out, "due in 7d") {
		t.Errorf("expected output to contain 'due in 7d', got:\n%s", out)
	}
}

func TestRenderHeader_PriorityFilter(t *testing.T) {
	m := Model{
		workspaceName:  "Dev",
		boardName:      "Board",
		viewMode:       viewList,
		sortMode:       sortByPriority,
		filterIndex:    -1,
		dueFilter:      dueFilterAny,
		priorityFilter: 2,
	}
	out := m.renderHeader(80)
	if !strings.Contains(out, "priority:p2") {
		t.Errorf("expected output to contain 'priority:p2', got:\n%s", out)
	}
}

func TestRenderHeader_SmallWidth(t *testing.T) {
	m := Model{
		workspaceName: "Dev",
		boardName:     "Board",
		viewMode:      viewList,
		sortMode:      sortByPriority,
		filterIndex:   -1,
		dueFilter:     dueFilterAny,
	}
	out := m.renderHeader(10)
	if !strings.Contains(out, "Dev / Board") {
		t.Errorf("expected output to contain 'Dev / Board', got:\n%s", out)
	}
}

func TestRenderFooter_Default(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputNone}}
	out := m.renderFooter()
	if !strings.Contains(out, "?:help") {
		t.Errorf("expected output to contain '?:help', got:\n%s", out)
	}
	if strings.Contains(out, "x:clear-search") {
		t.Error("expected no clear-search hint when titleFilter is empty")
	}
}

func TestRenderFooter_WithStatusLine(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputNone}, statusLine: "saved"}
	out := m.renderFooter()
	if !strings.Contains(out, "saved") {
		t.Errorf("expected output to contain 'saved', got:\n%s", out)
	}
}

func TestRenderFooter_WithSearchFilter(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputNone}, titleFilter: "query"}
	out := m.renderFooter()
	if !strings.Contains(out, "x:clear-search") {
		t.Errorf("expected output to contain 'x:clear-search', got:\n%s", out)
	}
}

func TestRenderFooter_InputSearch(t *testing.T) {
	ti := textinput.New()
	ti.SetValue("find me")
	m := Model{overlayState: overlayState{inputMode: inputSearch}, textInput: ti}
	out := m.renderFooter()
	if !strings.Contains(out, "find me") {
		t.Errorf("expected output to contain 'find me', got:\n%s", out)
	}
}

func TestRenderFooter_InputEditDescription(t *testing.T) {
	ta := textarea.New()
	ta.SetValue("desc text")
	m := Model{overlayState: overlayState{inputMode: inputEditDescription}, textArea: ta}
	out := m.renderFooter()
	if !strings.Contains(out, "desc text") {
		t.Errorf("expected output to contain 'desc text', got:\n%s", out)
	}
}
