package ui

import (
	"testing"

	"github.com/tiagokriok/kanji/internal/domain"
)

// --- summarizeDescription tests ---

func TestSummarizeDescription_Empty(t *testing.T) {
	got := summarizeDescription("")
	if got != "" {
		t.Errorf("summarizeDescription(\"\") = %q, want empty", got)
	}
}

func TestSummarizeDescription_Whitespace(t *testing.T) {
	got := summarizeDescription("   \n\n  ")
	if got != "" {
		t.Errorf("summarizeDescription(whitespace) = %q, want empty", got)
	}
}

func TestSummarizeDescription_Short(t *testing.T) {
	got := summarizeDescription("hello world")
	if got != "hello world" {
		t.Errorf("summarizeDescription = %q, want %q", got, "hello world")
	}
}

func TestSummarizeDescription_FirstLineOnly(t *testing.T) {
	got := summarizeDescription("first line\nsecond line")
	if got != "first line" {
		t.Errorf("summarizeDescription = %q, want %q", got, "first line")
	}
}

func TestSummarizeDescription_TruncatesLongLine(t *testing.T) {
	input := make([]byte, 120)
	for i := range input {
		input[i] = 'a'
	}
	got := summarizeDescription(string(input))
	want := string(input[:77]) + "..."
	if got != want {
		t.Errorf("summarizeDescription len = %d, want %d", len(got), len(want))
	}
}

// --- chooseEditor tests ---

func TestChooseEditor_RespectsEnv(t *testing.T) {
	t.Setenv("EDITOR", "my-editor")

	got := chooseEditor()
	if got != "my-editor" {
		t.Errorf("chooseEditor() = %q, want %q", got, "my-editor")
	}
}

func TestChooseEditor_TrimsWhitespace(t *testing.T) {
	t.Setenv("EDITOR", "  code  ")

	got := chooseEditor()
	if got != "code" {
		t.Errorf("chooseEditor() = %q, want %q", got, "code")
	}
}

// --- openDescriptionEditorCmd tests ---

func TestOpenDescriptionEditorCmd_NoEditor(t *testing.T) {
	t.Setenv("EDITOR", "")
	t.Setenv("PATH", "")

	cmd := openDescriptionEditorCmd("hello")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	edited, ok := msg.(descriptionEditedMsg)
	if !ok {
		t.Fatalf("expected descriptionEditedMsg, got %T", msg)
	}
	if edited.err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestOpenDescriptionEditorCmd_ReturnsNonNil(t *testing.T) {
	cmd := openDescriptionEditorCmd("initial")
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

// --- startExternalDescriptionEdit tests ---

func TestStartExternalDescriptionEdit(t *testing.T) {
	m := Model{}
	task := domain.Task{ID: "t1", DescriptionMD: "# hello"}
	cmd := m.startExternalDescriptionEdit(task)

	if m.editingDescTask != "t1" {
		t.Errorf("editingDescTask = %q, want %q", m.editingDescTask, "t1")
	}
	if m.statusLine != "" {
		t.Errorf("statusLine = %q, want empty", m.statusLine)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

// --- handleExternalDescriptionEdited tests ---

func TestHandleExternalDescriptionEdited_Success(t *testing.T) {
	m := Model{editingDescTask: "t1"}
	model, cmd := m.handleExternalDescriptionEdited(descriptionEditedMsg{content: "new desc"})
	updated := model.(Model)

	if updated.editingDescTask != "" {
		t.Errorf("editingDescTask = %q, want empty", updated.editingDescTask)
	}
	if updated.statusLine != "" {
		t.Errorf("statusLine = %q, want empty", updated.statusLine)
	}
	if updated.err != nil {
		t.Errorf("err = %v, want nil", updated.err)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestHandleExternalDescriptionEdited_Error(t *testing.T) {
	m := Model{editingDescTask: "t1"}
	model, cmd := m.handleExternalDescriptionEdited(descriptionEditedMsg{err: errTest("boom")})
	updated := model.(Model)

	if updated.editingDescTask != "" {
		t.Errorf("editingDescTask = %q, want empty", updated.editingDescTask)
	}
	if updated.statusLine != "editor error: boom" {
		t.Errorf("statusLine = %q, want %q", updated.statusLine, "editor error: boom")
	}
	if updated.err == nil {
		t.Error("expected err to be set")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestHandleExternalDescriptionEdited_NoPendingTask(t *testing.T) {
	m := Model{editingDescTask: ""}
	model, cmd := m.handleExternalDescriptionEdited(descriptionEditedMsg{content: "new desc"})
	updated := model.(Model)

	if updated.editingDescTask != "" {
		t.Errorf("editingDescTask = %q, want empty", updated.editingDescTask)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}
