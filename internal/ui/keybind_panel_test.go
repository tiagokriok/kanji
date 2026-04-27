package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newKeybindTestModel() Model {
	m := Model{
		keys:   newKeyMap(),
		width:  80,
		height: 24,
	}
	m.keyFilter = textinput.New()
	m.keyFilter.Placeholder = "Filter keybindings..."
	m.keyFilter.Prompt = "Filter: "
	m.keyFilter.CharLimit = 128
	return m
}

func TestOpenKeybindPanel(t *testing.T) {
	m := newKeybindTestModel()
	m.keyFilter.SetValue("old")
	m.keyFilter.Blur()

	m.openKeybindPanel()

	if !m.showKeybinds {
		t.Error("expected showKeybinds to be true")
	}
	if m.keyFilter.Value() != "" {
		t.Errorf("keyFilter value = %q, want empty", m.keyFilter.Value())
	}
	if !m.keyFilter.Focused() {
		t.Error("expected keyFilter to be focused")
	}
}

func TestCloseKeybindPanel(t *testing.T) {
	m := newKeybindTestModel()
	m.overlayState.openKeybinds()
	m.keyFilter.Focus()

	m.closeKeybindPanel()

	if m.showKeybinds {
		t.Error("expected showKeybinds to be false")
	}
	if m.keyFilter.Focused() {
		t.Error("expected keyFilter to be blurred")
	}
}

func TestClampKeybindSelectionEmpty(t *testing.T) {
	m := newKeybindTestModel()
	m.keySelected = 5
	m.keyFilter.SetValue("nonexistentxyz")

	m.clampKeybindSelection()

	if m.keySelected != 0 {
		t.Errorf("keySelected = %d, want 0", m.keySelected)
	}
}

func TestClampKeybindSelectionInRange(t *testing.T) {
	m := newKeybindTestModel()
	m.keySelected = 2

	m.clampKeybindSelection()

	if m.keySelected != 2 {
		t.Errorf("keySelected = %d, want 2", m.keySelected)
	}
}

func TestClampKeybindSelectionAboveRange(t *testing.T) {
	m := newKeybindTestModel()
	m.keySelected = 999

	m.clampKeybindSelection()

	entries := m.filteredKeybindEntries()
	want := len(entries) - 1
	if m.keySelected != want {
		t.Errorf("keySelected = %d, want %d", m.keySelected, want)
	}
}

func TestUpdateKeybindPanel_Cancel(t *testing.T) {
	m := newKeybindTestModel()
	m.overlayState.openKeybinds()

	updated, cmd := m.updateKeybindPanel(tea.KeyMsg{Type: tea.KeyEscape})
	um := updated.(Model)

	if um.showKeybinds {
		t.Error("expected showKeybinds to be false after cancel")
	}
	if cmd != nil {
		t.Error("expected nil cmd after cancel")
	}
}

func TestUpdateKeybindPanel_ToggleKeybinds(t *testing.T) {
	m := newKeybindTestModel()
	m.overlayState.openKeybinds()

	updated, cmd := m.updateKeybindPanel(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	um := updated.(Model)

	if um.showKeybinds {
		t.Error("expected showKeybinds to be false after toggle key")
	}
	if cmd != nil {
		t.Error("expected nil cmd after toggle key")
	}
}

func TestUpdateKeybindPanel_Up(t *testing.T) {
	m := newKeybindTestModel()
	m.overlayState.openKeybinds()
	m.keySelected = 2

	updated, cmd := m.updateKeybindPanel(tea.KeyMsg{Type: tea.KeyUp})
	um := updated.(Model)

	if um.keySelected != 1 {
		t.Errorf("keySelected = %d, want 1", um.keySelected)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateKeybindPanel_Down(t *testing.T) {
	m := newKeybindTestModel()
	m.overlayState.openKeybinds()
	m.keySelected = 0

	updated, cmd := m.updateKeybindPanel(tea.KeyMsg{Type: tea.KeyDown})
	um := updated.(Model)

	if um.keySelected != 1 {
		t.Errorf("keySelected = %d, want 1", um.keySelected)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateKeybindPanel_Confirm(t *testing.T) {
	m := newKeybindTestModel()
	m.overlayState.openKeybinds()

	updated, cmd := m.updateKeybindPanel(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(Model)

	if um.showKeybinds {
		t.Error("expected showKeybinds to be false after confirm")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd after confirm")
	}
	msg := cmd()
	if _, ok := msg.(executeActionMsg); !ok {
		t.Errorf("expected executeActionMsg, got %T", msg)
	}
}

func TestUpdateKeybindPanel_ConfirmEmptyFilter(t *testing.T) {
	m := newKeybindTestModel()
	m.overlayState.openKeybinds()
	m.keyFilter.SetValue("nonexistentxyz")

	updated, cmd := m.updateKeybindPanel(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(Model)

	if !um.showKeybinds {
		t.Error("expected showKeybinds to stay true when no entries match")
	}
	if cmd != nil {
		t.Error("expected nil cmd when no entries match")
	}
}

func TestUpdateKeybindPanel_WindowSize(t *testing.T) {
	m := newKeybindTestModel()
	m.overlayState.openKeybinds()

	updated, cmd := m.updateKeybindPanel(tea.WindowSizeMsg{Width: 120, Height: 40})
	um := updated.(Model)

	if um.width != 120 {
		t.Errorf("width = %d, want 120", um.width)
	}
	if um.height != 40 {
		t.Errorf("height = %d, want 40", um.height)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestFilteredKeybindEntries_NoFilter(t *testing.T) {
	m := newKeybindTestModel()
	entries := m.filteredKeybindEntries()
	if len(entries) == 0 {
		t.Fatal("expected non-empty entries")
	}
	foundQuit := false
	for _, e := range entries {
		if e.ID == "quit" {
			foundQuit = true
			break
		}
	}
	if !foundQuit {
		t.Error("expected 'quit' entry in unfiltered list")
	}
}

func TestFilteredKeybindEntries_FilterByKey(t *testing.T) {
	m := newKeybindTestModel()
	m.keyFilter.SetValue("q")
	entries := m.filteredKeybindEntries()
	if len(entries) == 0 {
		t.Fatal("expected non-empty entries")
	}
	for _, e := range entries {
		if e.ID == "quit" {
			return
		}
	}
	t.Error("expected 'quit' entry when filtering by 'q'")
}

func TestFilteredKeybindEntries_FilterByLabel(t *testing.T) {
	m := newKeybindTestModel()
	m.keyFilter.SetValue("quit")
	entries := m.filteredKeybindEntries()
	if len(entries) == 0 {
		t.Fatal("expected non-empty entries")
	}
	for _, e := range entries {
		if e.ID == "quit" {
			return
		}
	}
	t.Error("expected 'quit' entry when filtering by label 'quit'")
}

func TestFilteredKeybindEntries_ClearSearchAppended(t *testing.T) {
	m := newKeybindTestModel()
	m.titleFilter = "search"
	entries := m.filteredKeybindEntries()
	foundClear := false
	for _, e := range entries {
		if e.ID == "clear_search" {
			foundClear = true
			break
		}
	}
	if !foundClear {
		t.Error("expected 'clear_search' entry when titleFilter is set")
	}
}

func TestRenderKeybindPanel(t *testing.T) {
	m := newKeybindTestModel()
	m.overlayState.openKeybinds()

	rendered := m.renderKeybindPanel("")
	if rendered == "" {
		t.Error("expected non-empty rendered panel")
	}
}

func TestRenderKeybindPanel_EmptyFilter(t *testing.T) {
	m := newKeybindTestModel()
	m.overlayState.openKeybinds()
	m.keyFilter.SetValue("zzzzzz")

	rendered := m.renderKeybindPanel("")
	if rendered == "" {
		t.Error("expected non-empty rendered panel even with no matches")
	}
	if !contains(rendered, "no keybindings match filter") {
		t.Error("expected '(no keybindings match filter)' message")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
