package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"

	"github.com/tiagokriok/kanji/internal/domain"
)

func TestStartSearch(t *testing.T) {
	m := Model{titleFilter: "hello"}
	m.textInput = textinput.New()
	cmd := m.startSearch()

	if m.inputMode != inputSearch {
		t.Errorf("inputMode = %v, want inputSearch", m.inputMode)
	}
	if m.textInput.Value() != "hello" {
		t.Errorf("value = %q, want %q", m.textInput.Value(), "hello")
	}
	if m.textInput.Placeholder != "Search title" {
		t.Errorf("placeholder = %q, want %q", m.textInput.Placeholder, "Search title")
	}
	if !m.textInput.Focused() {
		t.Error("expected textInput to be focused")
	}
	if m.statusLine != "Search by title" {
		t.Errorf("statusLine = %q, want %q", m.statusLine, "Search by title")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestConfirmSearch(t *testing.T) {
	m := Model{titleFilter: "old", overlayState: overlayState{inputMode: inputSearch}}
	m.textInput = textinput.New()
	m.textInput.SetValue("new query")
	cmd := m.confirmSearch()

	if m.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", m.inputMode)
	}
	if m.titleFilter != "new query" {
		t.Errorf("titleFilter = %q, want %q", m.titleFilter, "new query")
	}
	if m.statusLine != "" {
		t.Errorf("statusLine = %q, want empty", m.statusLine)
	}
	if m.textInput.Focused() {
		t.Error("expected textInput to be blurred")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestCancelInput(t *testing.T) {
	m := Model{statusLine: "something", overlayState: overlayState{inputMode: inputSearch}}
	m.textInput = textinput.New()
	m.textInput.Focus()
	m.textArea = textarea.New()
	m.textArea.Focus()

	m.cancelInput()

	if m.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", m.inputMode)
	}
	if m.textInput.Focused() {
		t.Error("expected textInput to be blurred")
	}
	if m.textArea.Focused() {
		t.Error("expected textArea to be blurred")
	}
	if m.statusLine != "" {
		t.Errorf("statusLine = %q, want empty", m.statusLine)
	}
}

func TestStartAddComment(t *testing.T) {
	m := Model{}
	m.textInput = textinput.New()
	cmd := m.startAddComment()

	if m.inputMode != inputAddComment {
		t.Errorf("inputMode = %v, want inputAddComment", m.inputMode)
	}
	if m.textInput.Value() != "" {
		t.Errorf("value = %q, want empty", m.textInput.Value())
	}
	if m.textInput.Placeholder != "Comment body" {
		t.Errorf("placeholder = %q, want %q", m.textInput.Placeholder, "Comment body")
	}
	if !m.textInput.Focused() {
		t.Error("expected textInput to be focused")
	}
	if m.statusLine != "Add comment" {
		t.Errorf("statusLine = %q, want %q", m.statusLine, "Add comment")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestConfirmAddComment_Success(t *testing.T) {
	m := Model{
		viewMode:     viewList,
		tasks:        []domain.Task{{ID: "t1", Title: "Task"}},
		selected:     0,
		statusLine:   "Add comment",
		overlayState: overlayState{inputMode: inputAddComment},
	}
	m.textInput = textinput.New()
	m.textInput.SetValue("nice comment")
	cmd := m.confirmAddComment()

	if m.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", m.inputMode)
	}
	if m.statusLine != "Add comment" {
		t.Errorf("statusLine = %q, want %q", m.statusLine, "Add comment")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestConfirmAddComment_EmptyReopens(t *testing.T) {
	m := Model{
		viewMode:     viewList,
		tasks:        []domain.Task{{ID: "t1", Title: "Task"}},
		selected:     0,
		overlayState: overlayState{inputMode: inputAddComment},
	}
	m.textInput = textinput.New()
	m.textInput.SetValue("   ")
	cmd := m.confirmAddComment()

	if m.inputMode != inputAddComment {
		t.Errorf("inputMode = %v, want inputAddComment", m.inputMode)
	}
	if !m.textInput.Focused() {
		t.Error("expected textInput to be focused")
	}
	if m.textInput.Value() != "   " {
		t.Errorf("value = %q, want %q", m.textInput.Value(), "   ")
	}
	if m.statusLine != "comment is required" {
		t.Errorf("statusLine = %q, want %q", m.statusLine, "comment is required")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestConfirmAddComment_NoTask(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputAddComment}}
	m.textInput = textinput.New()
	m.textInput.SetValue("comment")
	cmd := m.confirmAddComment()

	if m.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", m.inputMode)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestCancelAddComment_WithReturn(t *testing.T) {
	m := Model{
		overlayState: overlayState{
			inputMode:      inputAddComment,
			returnTaskView: true,
			returnTaskID:   "task-99",
		},
	}
	m.textInput = textinput.New()
	m.textInput.Focus()
	cmd := m.cancelAddComment()

	if m.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", m.inputMode)
	}
	if m.returnTaskView {
		t.Error("expected returnTaskView to be cleared")
	}
	if m.returnTaskID != "" {
		t.Errorf("returnTaskID = %q, want empty", m.returnTaskID)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestCancelAddComment_WithoutReturn(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputAddComment}}
	m.textInput = textinput.New()
	m.textInput.Focus()
	cmd := m.cancelAddComment()

	if m.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", m.inputMode)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestStartEditDescription(t *testing.T) {
	m := Model{}
	m.textArea = textarea.New()
	task := domain.Task{ID: "t1", DescriptionMD: "# hello"}
	m.startEditDescription(task)

	if m.inputMode != inputEditDescription {
		t.Errorf("inputMode = %v, want inputEditDescription", m.inputMode)
	}
	if m.textArea.Value() != "# hello" {
		t.Errorf("value = %q, want %q", m.textArea.Value(), "# hello")
	}
	if !m.textArea.Focused() {
		t.Error("expected textArea to be focused")
	}
	if m.statusLine != "Edit description (ctrl+s to save, esc to cancel)" {
		t.Errorf("statusLine = %q", m.statusLine)
	}
}

func TestSubmitEditDescription_Success(t *testing.T) {
	m := Model{
		viewMode:     viewList,
		tasks:        []domain.Task{{ID: "t1", Title: "Task"}},
		selected:     0,
		statusLine:   "Edit description (ctrl+s to save, esc to cancel)",
		overlayState: overlayState{inputMode: inputEditDescription},
	}
	m.textArea = textarea.New()
	m.textArea.SetValue("updated desc")
	cmd := m.submitEditDescription()

	if m.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", m.inputMode)
	}
	if m.statusLine != "Edit description (ctrl+s to save, esc to cancel)" {
		t.Errorf("statusLine = %q", m.statusLine)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestSubmitEditDescription_NoTask(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputEditDescription}}
	m.textArea = textarea.New()
	m.textArea.SetValue("updated desc")
	cmd := m.submitEditDescription()

	if m.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", m.inputMode)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}
