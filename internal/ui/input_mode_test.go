package ui

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

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
	m := Model{
		statusLine:   "Add comment",
		overlayState: overlayState{inputMode: inputAddComment},
	}
	m.textInput = textinput.New()
	m.textInput.SetValue("comment")
	m.textInput.Focus()
	cmd := m.confirmAddComment()

	if m.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", m.inputMode)
	}
	if m.statusLine != "Add comment" {
		t.Errorf("statusLine = %q, want %q", m.statusLine, "Add comment")
	}
	if m.textInput.Focused() {
		t.Error("expected textInput to be blurred")
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

// --- Update-level integration tests ---

func TestUpdateSearchKeyEntersSearchMode(t *testing.T) {
	m := Model{keys: newKeyMap()}
	m.textInput = textinput.New()

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	updated := newM.(Model)

	if updated.inputMode != inputSearch {
		t.Errorf("inputMode = %v, want inputSearch", updated.inputMode)
	}
	if updated.statusLine != "Search by title" {
		t.Errorf("statusLine = %q, want %q", updated.statusLine, "Search by title")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestUpdateAddCommentKeyEntersCommentMode(t *testing.T) {
	m := Model{
		keys:     newKeyMap(),
		tasks:    []domain.Task{{ID: "t1", Title: "Task"}},
		selected: 0,
	}
	m.textInput = textinput.New()

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	updated := newM.(Model)

	if updated.inputMode != inputAddComment {
		t.Errorf("inputMode = %v, want inputAddComment", updated.inputMode)
	}
	if updated.statusLine != "Add comment" {
		t.Errorf("statusLine = %q, want %q", updated.statusLine, "Add comment")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestUpdateAddCommentKeyNoTask(t *testing.T) {
	m := Model{keys: newKeyMap()}
	m.textInput = textinput.New()

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	updated := newM.(Model)

	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateConfirmSearchSubmitsQuery(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{inputMode: inputSearch},
		titleFilter:  "old",
	}
	m.textInput = textinput.New()
	m.textInput.SetValue("new query")
	m.textInput.Focus()

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := newM.(Model)

	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if updated.titleFilter != "new query" {
		t.Errorf("titleFilter = %q, want %q", updated.titleFilter, "new query")
	}
	if updated.statusLine != "" {
		t.Errorf("statusLine = %q, want empty", updated.statusLine)
	}
	if updated.textInput.Focused() {
		t.Error("expected textInput to be blurred")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestUpdateCancelSearchExitsInput(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{inputMode: inputSearch},
		statusLine:   "Search by title",
	}
	m.textInput = textinput.New()
	m.textInput.Focus()

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	updated := newM.(Model)

	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if updated.statusLine != "" {
		t.Errorf("statusLine = %q, want empty", updated.statusLine)
	}
	if updated.textInput.Focused() {
		t.Error("expected textInput to be blurred")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateConfirmAddCommentSubmitsComment(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		statusLine:   "Add comment",
		overlayState: overlayState{inputMode: inputAddComment},
		tasks:        []domain.Task{{ID: "t1", Title: "Task"}},
		selected:     0,
	}
	m.textInput = textinput.New()
	m.textInput.SetValue("nice comment")
	m.textInput.Focus()

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := newM.(Model)

	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if updated.statusLine != "Add comment" {
		t.Errorf("statusLine = %q, want %q", updated.statusLine, "Add comment")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestUpdateConfirmAddCommentEmptyReopens(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{inputMode: inputAddComment},
		tasks:        []domain.Task{{ID: "t1", Title: "Task"}},
		selected:     0,
	}
	m.textInput = textinput.New()
	m.textInput.SetValue("   ")
	m.textInput.Focus()

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := newM.(Model)

	if updated.inputMode != inputAddComment {
		t.Errorf("inputMode = %v, want inputAddComment", updated.inputMode)
	}
	if updated.statusLine != "comment is required" {
		t.Errorf("statusLine = %q, want %q", updated.statusLine, "comment is required")
	}
	if !updated.textInput.Focused() {
		t.Error("expected textInput to be focused")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestUpdateCancelAddCommentReturnsToViewer(t *testing.T) {
	m := Model{
		keys: newKeyMap(),
		overlayState: overlayState{
			inputMode:      inputAddComment,
			returnTaskView: true,
			returnTaskID:   "task-99",
		},
	}
	m.textInput = textinput.New()
	m.textInput.Focus()

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	updated := newM.(Model)

	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if updated.returnTaskView {
		t.Error("expected returnTaskView to be cleared")
	}
	if updated.returnTaskID != "" {
		t.Errorf("returnTaskID = %q, want empty", updated.returnTaskID)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd to reopen task viewer")
	}
}

func TestUpdateCancelAddCommentWithoutReturn(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{inputMode: inputAddComment},
	}
	m.textInput = textinput.New()
	m.textInput.Focus()

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	updated := newM.(Model)

	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskViewerAddCommentEntersCommentMode(t *testing.T) {
	m := Model{
		keys: newKeyMap(),
		overlayState: overlayState{
			showTaskView: true,
			viewTaskID:   "t1",
		},
		tasks:    []domain.Task{{ID: "t1", Title: "Task"}},
		selected: 0,
	}
	m.textInput = textinput.New()

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	updated := newM.(Model)

	if updated.inputMode != inputAddComment {
		t.Errorf("inputMode = %v, want inputAddComment", updated.inputMode)
	}
	if updated.showTaskView {
		t.Error("expected task viewer to be closed")
	}
	if !updated.returnTaskView {
		t.Error("expected returnTaskView to be set")
	}
	if updated.returnTaskID != "t1" {
		t.Errorf("returnTaskID = %q, want %q", updated.returnTaskID, "t1")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestExecuteActionSearchEntersSearchMode(t *testing.T) {
	m := Model{keys: newKeyMap()}
	m.textInput = textinput.New()

	newM, cmd := m.executeAction("search")
	updated := newM.(Model)

	if updated.inputMode != inputSearch {
		t.Errorf("inputMode = %v, want inputSearch", updated.inputMode)
	}
	if updated.statusLine != "Search by title" {
		t.Errorf("statusLine = %q, want %q", updated.statusLine, "Search by title")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestExecuteActionAddCommentEntersCommentMode(t *testing.T) {
	m := Model{
		keys:     newKeyMap(),
		tasks:    []domain.Task{{ID: "t1", Title: "Task"}},
		selected: 0,
	}
	m.textInput = textinput.New()

	newM, cmd := m.executeAction("add_comment")
	updated := newM.(Model)

	if updated.inputMode != inputAddComment {
		t.Errorf("inputMode = %v, want inputAddComment", updated.inputMode)
	}
	if updated.statusLine != "Add comment" {
		t.Errorf("statusLine = %q, want %q", updated.statusLine, "Add comment")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestExecuteActionAddCommentNoTask(t *testing.T) {
	m := Model{keys: newKeyMap()}
	m.textInput = textinput.New()

	newM, cmd := m.executeAction("add_comment")
	updated := newM.(Model)

	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

// --- handleDescriptionEditedMsg tests ---

func TestHandleDescriptionEditedMsg_TaskFormSuccess(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}}}
	m.taskForm.descriptionFull = ""
	m.taskForm.description = textinput.New()

	model, cmd, handled := m.handleDescriptionEditedMsg(descriptionEditedMsg{content: "full desc"})
	updated := model.(Model)

	if !handled {
		t.Error("expected handled")
	}
	if updated.taskForm.descriptionFull != "full desc" {
		t.Errorf("descriptionFull = %q, want %q", updated.taskForm.descriptionFull, "full desc")
	}
	if updated.taskForm.description.Value() != "full desc" {
		t.Errorf("description = %q, want %q", updated.taskForm.description.Value(), "full desc")
	}
	if updated.statusLine != "" {
		t.Errorf("statusLine = %q, want empty", updated.statusLine)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestHandleDescriptionEditedMsg_TaskFormError(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}}}
	m.taskForm.description = textinput.New()

	model, cmd, handled := m.handleDescriptionEditedMsg(descriptionEditedMsg{err: fmt.Errorf("boom")})
	updated := model.(Model)

	if !handled {
		t.Error("expected handled")
	}
	if updated.statusLine != "editor error: boom" {
		t.Errorf("statusLine = %q, want %q", updated.statusLine, "editor error: boom")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestHandleDescriptionEditedMsg_NotTaskForm(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputSearch}}

	_, _, handled := m.handleDescriptionEditedMsg(descriptionEditedMsg{content: "x"})
	if handled {
		t.Error("expected not handled")
	}
}

// --- cancelInputMode tests ---

func TestCancelInputMode_TaskForm(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}}}
	m.statusLine = "Create task"

	model, cmd := m.cancelInputMode()
	updated := model.(Model)

	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if updated.taskForm != nil {
		t.Error("expected taskForm to be nil")
	}
	if updated.statusLine != "" {
		t.Errorf("statusLine = %q, want empty", updated.statusLine)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestCancelInputMode_TaskFormWithReturn(t *testing.T) {
	m := Model{
		overlayState: overlayState{
			inputMode:      inputTaskForm,
			taskForm:       &taskForm{},
			returnTaskView: true,
			returnTaskID:   "task-99",
		},
		statusLine: "Edit task",
	}

	model, cmd := m.cancelInputMode()
	updated := model.(Model)

	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if updated.returnTaskView {
		t.Error("expected returnTaskView to be cleared")
	}
	if updated.returnTaskID != "" {
		t.Errorf("returnTaskID = %q, want empty", updated.returnTaskID)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestCancelInputMode_AddCommentWithReturn(t *testing.T) {
	m := Model{
		overlayState: overlayState{
			inputMode:      inputAddComment,
			returnTaskView: true,
			returnTaskID:   "task-99",
		},
	}
	m.textInput = textinput.New()
	m.textInput.Focus()

	model, cmd := m.cancelInputMode()
	updated := model.(Model)

	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if updated.returnTaskView {
		t.Error("expected returnTaskView to be cleared")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestCancelInputMode_Search(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputSearch}, statusLine: "Search"}
	m.textInput = textinput.New()
	m.textInput.Focus()

	model, cmd := m.cancelInputMode()
	updated := model.(Model)

	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if updated.statusLine != "" {
		t.Errorf("statusLine = %q, want empty", updated.statusLine)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

// --- saveEditDescription tests ---

func TestSaveEditDescription_Success(t *testing.T) {
	m := Model{
		viewMode:     viewList,
		tasks:        []domain.Task{{ID: "t1", Title: "Task"}},
		selected:     0,
		overlayState: overlayState{inputMode: inputEditDescription},
	}
	m.textArea = textarea.New()
	m.textArea.SetValue("new desc")

	model, cmd := m.saveEditDescription()
	updated := model.(Model)

	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if updated.textArea.Focused() {
		t.Error("expected textArea to be blurred")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestSaveEditDescription_NoTask(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputEditDescription}}
	m.textArea = textarea.New()
	m.textArea.SetValue("new desc")

	model, cmd := m.saveEditDescription()
	updated := model.(Model)

	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

// --- confirmInputMode tests ---

func TestConfirmInputMode_Search(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputSearch}, titleFilter: "old"}
	m.textInput = textinput.New()
	m.textInput.SetValue("new")

	model, cmd, handled := m.confirmInputMode()
	updated := model.(Model)

	if !handled {
		t.Error("expected handled")
	}
	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if updated.titleFilter != "new" {
		t.Errorf("titleFilter = %q, want %q", updated.titleFilter, "new")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestConfirmInputMode_AddComment(t *testing.T) {
	m := Model{
		viewMode:     viewList,
		tasks:        []domain.Task{{ID: "t1", Title: "Task"}},
		selected:     0,
		overlayState: overlayState{inputMode: inputAddComment},
	}
	m.textInput = textinput.New()
	m.textInput.SetValue("comment")

	model, cmd, handled := m.confirmInputMode()
	updated := model.(Model)

	if !handled {
		t.Error("expected handled")
	}
	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestConfirmInputMode_EditDescription(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputEditDescription}}

	_, _, handled := m.confirmInputMode()
	if handled {
		t.Error("expected not handled")
	}
}

func TestConfirmInputMode_TaskForm(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}}}

	_, _, handled := m.confirmInputMode()
	if handled {
		t.Error("expected not handled")
	}
}

// --- updateInputModeWidgets tests ---

func TestUpdateInputModeWidgets_EditDescription(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputEditDescription}}
	m.textArea = textarea.New()
	m.textArea.Focus()

	model, cmd := m.updateInputModeWidgets(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	updated := model.(Model)

	if !updated.textArea.Focused() {
		t.Error("expected textArea to remain focused")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestUpdateTaskFormWidgets_RoutesToTitle(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}}}
	m.taskForm.title = textinput.New()
	m.taskForm.title.Focus()
	m.taskForm.focus = taskFieldTitle

	model, cmd := m.updateTaskFormWidgets(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	updated := model.(Model)

	if updated.taskForm.title.Value() != "x" {
		t.Errorf("title = %q, want %q", updated.taskForm.title.Value(), "x")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestUpdateInputModeWidgets_Default(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputSearch}}
	m.textInput = textinput.New()
	m.textInput.Focus()

	model, cmd := m.updateInputModeWidgets(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	updated := model.(Model)

	if updated.textInput.Value() != "z" {
		t.Errorf("value = %q, want %q", updated.textInput.Value(), "z")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

// --- updateInputMode integration tests for remaining branches ---

func TestUpdateDescriptionEditedMsg_InTaskForm(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}}}
	m.taskForm.description = textinput.New()
	m.taskForm.descriptionFull = ""

	newM, cmd := m.Update(descriptionEditedMsg{content: "edited"})
	updated := newM.(Model)

	if updated.taskForm.descriptionFull != "edited" {
		t.Errorf("descriptionFull = %q, want %q", updated.taskForm.descriptionFull, "edited")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateCancelTaskForm(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}},
		statusLine:   "Create task",
	}

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	updated := newM.(Model)

	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if updated.taskForm != nil {
		t.Error("expected taskForm to be nil")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateCancelTaskFormWithReturn(t *testing.T) {
	m := Model{
		keys: newKeyMap(),
		overlayState: overlayState{
			inputMode:      inputTaskForm,
			taskForm:       &taskForm{},
			returnTaskView: true,
			returnTaskID:   "task-99",
		},
		statusLine: "Edit task",
	}

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	updated := newM.(Model)

	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if updated.returnTaskView {
		t.Error("expected returnTaskView to be cleared")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestUpdateCtrlSInEditDescription(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		viewMode:     viewList,
		tasks:        []domain.Task{{ID: "t1", Title: "Task"}},
		selected:     0,
		overlayState: overlayState{inputMode: inputEditDescription},
	}
	m.textArea = textarea.New()
	m.textArea.SetValue("new desc")

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	updated := newM.(Model)

	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if updated.textArea.Focused() {
		t.Error("expected textArea to be blurred")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestUpdateCtrlSInEditDescriptionNoTask(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{inputMode: inputEditDescription},
	}
	m.textArea = textarea.New()
	m.textArea.SetValue("new desc")

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	updated := newM.(Model)

	if updated.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", updated.inputMode)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskFormKeyRoutesToWidgets(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}}}
	m.taskForm.title = textinput.New()
	m.taskForm.title.Focus()
	m.taskForm.focus = taskFieldTitle

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	updated := newM.(Model)

	if updated.taskForm.title.Value() != "a" {
		t.Errorf("title = %q, want %q", updated.taskForm.title.Value(), "a")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestUpdateEditDescriptionKeyRoutesToTextArea(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{inputMode: inputEditDescription},
	}
	m.textArea = textarea.New()
	m.textArea.Focus()

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	updated := newM.(Model)

	if !updated.textArea.Focused() {
		t.Error("expected textArea to remain focused")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

// --- updateTaskForm tests ---

func TestUpdateTaskForm_HandledKey(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}}}
	m.taskForm.title = textinput.New()
	m.taskForm.description = textinput.New()
	m.taskForm.dueDate = textinput.New()
	m.taskForm.title.Focus()
	m.taskForm.focus = taskFieldTitle

	// Tab should be handled by handleTaskFormKey (moves focus)
	newM, cmd := m.updateTaskForm(tea.KeyMsg{Type: tea.KeyTab})
	updated := newM.(Model)

	if updated.taskForm.focus != taskFieldDescription {
		t.Errorf("focus = %d, want taskFieldDescription", updated.taskForm.focus)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestUpdateTaskForm_UnhandledKeyRoutesToWidgets(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}}}
	m.taskForm.title = textinput.New()
	m.taskForm.title.Focus()
	m.taskForm.focus = taskFieldTitle

	newM, cmd := m.updateTaskForm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	updated := newM.(Model)

	if updated.taskForm.title.Value() != "a" {
		t.Errorf("title = %q, want %q", updated.taskForm.title.Value(), "a")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestUpdateTaskForm_NonKeyMsgRoutesToWidgets(t *testing.T) {
	m := Model{
		keys:         newKeyMap(),
		overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}}}
	m.taskForm.title = textinput.New()
	m.taskForm.title.Focus()
	m.taskForm.focus = taskFieldTitle

	newM, cmd := m.updateTaskForm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	updated := newM.(Model)

	if updated.taskForm.title.Value() != "b" {
		t.Errorf("title = %q, want %q", updated.taskForm.title.Value(), "b")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

// --- updateTaskFormWidgets tests ---

func TestUpdateTaskFormWidgets_UpdatesActiveField(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}}}
	m.taskForm.title = textinput.New()
	m.taskForm.title.Focus()
	m.taskForm.focus = taskFieldTitle

	model, cmd := m.updateTaskFormWidgets(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	updated := model.(Model)

	if updated.taskForm.title.Value() != "x" {
		t.Errorf("title = %q, want %q", updated.taskForm.title.Value(), "x")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestUpdateTaskFormWidgets_SyncsDescriptionFull(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}}}
	m.taskForm.description = textinput.New()
	m.taskForm.description.Focus()
	m.taskForm.focus = taskFieldDescription
	m.taskForm.descriptionFull = "old"

	model, cmd := m.updateTaskFormWidgets(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	updated := model.(Model)

	if updated.taskForm.descriptionFull != "n" {
		t.Errorf("descriptionFull = %q, want %q", updated.taskForm.descriptionFull, "n")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestUpdateTaskFormWidgets_NoForm(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputTaskForm}}

	model, cmd := m.updateTaskFormWidgets(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	updated := model.(Model)

	if updated.taskForm != nil {
		t.Error("expected nil taskForm")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestUpdateTaskFormWidgets_NoField(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}}}
	m.taskForm.focus = taskFieldPriority

	model, cmd := m.updateTaskFormWidgets(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	updated := model.(Model)

	if updated.taskForm.focus != taskFieldPriority {
		t.Errorf("focus = %d, want taskFieldPriority", updated.taskForm.focus)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

// --- task form lifecycle tests ---

func TestCloseTaskForm(t *testing.T) {
	m := Model{
		overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}},
		statusLine:   "Create task",
	}

	m.closeTaskForm()

	if m.inputMode != inputNone {
		t.Errorf("inputMode = %v, want inputNone", m.inputMode)
	}
	if m.taskForm != nil {
		t.Error("expected taskForm to be nil")
	}
	if m.statusLine != "" {
		t.Errorf("statusLine = %q, want empty", m.statusLine)
	}
}

func TestStartCreateTaskForm(t *testing.T) {
	m := Model{
		viewMode: viewList,
		columns: []domain.Column{
			{ID: "c1", Name: "Todo", Color: "#ff0000"},
		},
	}
	m.textInput = textinput.New()

	m.startCreateTaskForm()

	if m.inputMode != inputTaskForm {
		t.Errorf("inputMode = %v, want inputTaskForm", m.inputMode)
	}
	if m.taskForm == nil {
		t.Fatal("expected taskForm to be non-nil")
	}
	if m.taskForm.mode != taskFormCreate {
		t.Errorf("mode = %v, want taskFormCreate", m.taskForm.mode)
	}
	if m.statusLine != "Create task" {
		t.Errorf("statusLine = %q, want %q", m.statusLine, "Create task")
	}
	if m.taskForm.title.Placeholder != "Title" {
		t.Errorf("title placeholder = %q, want %q", m.taskForm.title.Placeholder, "Title")
	}
}

func TestStartEditTaskForm(t *testing.T) {
	m := Model{
		viewMode: viewList,
		columns: []domain.Column{
			{ID: "c1", Name: "Todo", Color: "#ff0000"},
		},
	}
	m.textInput = textinput.New()

	task := domain.Task{
		ID:       "t1",
		Title:    "Test Task",
		Priority: 2,
	}
	m.startEditTaskForm(task)

	if m.inputMode != inputTaskForm {
		t.Errorf("inputMode = %v, want inputTaskForm", m.inputMode)
	}
	if m.taskForm == nil {
		t.Fatal("expected taskForm to be non-nil")
	}
	if m.taskForm.mode != taskFormEdit {
		t.Errorf("mode = %v, want taskFormEdit", m.taskForm.mode)
	}
	if m.taskForm.taskID != "t1" {
		t.Errorf("taskID = %q, want %q", m.taskForm.taskID, "t1")
	}
	if m.statusLine != "Edit task" {
		t.Errorf("statusLine = %q, want %q", m.statusLine, "Edit task")
	}
}

func TestHandleTaskFormKey_TabMovesFocus(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}}}
	m.taskForm.title = textinput.New()
	m.taskForm.description = textinput.New()
	m.taskForm.dueDate = textinput.New()
	m.taskForm.title.Focus()
	m.taskForm.focus = taskFieldTitle

	model, cmd, handled := m.handleTaskFormKey(tea.KeyMsg{Type: tea.KeyTab})
	updated := model.(Model)

	if !handled {
		t.Error("expected handled")
	}
	if updated.taskForm.focus != taskFieldDescription {
		t.Errorf("focus = %d, want taskFieldDescription", updated.taskForm.focus)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestHandleTaskFormKey_CtrlGOpensEditorFromAnyField(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{descriptionFull: "long description"}}}
	m.taskForm.title = textinput.New()
	m.taskForm.description = textinput.New()
	m.taskForm.dueDate = textinput.New()
	m.taskForm.title.Focus()
	m.taskForm.focus = taskFieldTitle

	model, cmd, handled := m.handleTaskFormKey(tea.KeyMsg{Type: tea.KeyCtrlG})
	updated := model.(Model)

	if !handled {
		t.Fatal("expected handled")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	if updated.taskForm.focus != taskFieldTitle {
		t.Errorf("focus = %d, want taskFieldTitle", updated.taskForm.focus)
	}
}

func TestHandleTaskFormKey_EscNotHandled(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}}}
	m.taskForm.title = textinput.New()
	m.taskForm.description = textinput.New()
	m.taskForm.dueDate = textinput.New()
	m.taskForm.title.Focus()
	m.taskForm.focus = taskFieldTitle

	_, _, handled := m.handleTaskFormKey(tea.KeyMsg{Type: tea.KeyEscape})
	if handled {
		t.Error("expected not handled")
	}
}

func TestSubmitTaskForm_NilForm(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputTaskForm}}

	model, cmd := m.submitTaskForm()
	updated := model.(Model)

	if updated.inputMode != inputTaskForm {
		t.Errorf("inputMode = %v, want inputTaskForm", updated.inputMode)
	}
	if updated.statusLine == "" {
		t.Error("expected statusLine to be set")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestSubmitTaskFormCmd_EmptyTitle(t *testing.T) {
	m := Model{overlayState: overlayState{inputMode: inputTaskForm, taskForm: &taskForm{}}}
	m.taskForm.title = textinput.New()
	m.taskForm.title.SetValue("")

	_, err := m.submitTaskFormCmd()
	if err == nil {
		t.Error("expected error for empty title")
	}
}

func TestBuildTaskStatusOptions_NoTask(t *testing.T) {
	m := Model{
		columns: []domain.Column{
			{ID: "c1", Name: "Todo", Color: "#ff0000"},
			{ID: "c2", Name: "Done", Color: "#00ff00"},
		},
	}

	options, selected := m.buildTaskStatusOptions(nil)
	if len(options) != 2 {
		t.Errorf("len(options) = %d, want 2", len(options))
	}
	if selected != 0 {
		t.Errorf("selected = %d, want 0", selected)
	}
	if options[0].Label != "Todo" {
		t.Errorf("options[0].Label = %q, want %q", options[0].Label, "Todo")
	}
}

func TestBuildTaskStatusOptions_WithTask(t *testing.T) {
	m := Model{
		columns: []domain.Column{
			{ID: "c1", Name: "Todo", Color: "#ff0000"},
			{ID: "c2", Name: "Done", Color: "#00ff00"},
		},
	}
	colID := "c2"
	task := domain.Task{ColumnID: &colID}

	options, selected := m.buildTaskStatusOptions(&task)
	if len(options) != 3 { // (none) + 2 columns
		t.Errorf("len(options) = %d, want 3", len(options))
	}
	if selected != 2 {
		t.Errorf("selected = %d, want 2", selected)
	}
}
