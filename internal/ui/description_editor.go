package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tiagokriok/kanji/internal/domain"
)

// startExternalDescriptionEdit prepares the model for external description editing
// and returns a command that opens the user's preferred editor.
func (m *Model) startExternalDescriptionEdit(task domain.Task) tea.Cmd {
	m.editingDescTask = task.ID
	m.statusLine = ""
	return openDescriptionEditorCmd(task.DescriptionMD)
}

// summarizeDescription returns a short preview of the first line, truncated to 80 chars.
func summarizeDescription(description string) string {
	trimmed := strings.TrimSpace(description)
	if trimmed == "" {
		return ""
	}
	firstLine := strings.Split(trimmed, "\n")[0]
	if len(firstLine) > 80 {
		return firstLine[:77] + "..."
	}
	return firstLine
}

// chooseEditor returns the user's preferred editor from $EDITOR, falling back
// to nvim, vim, or vi in PATH. Returns empty string if none is found.
func chooseEditor() string {
	if editor := strings.TrimSpace(os.Getenv("EDITOR")); editor != "" {
		return editor
	}
	for _, candidate := range []string{"nvim", "vim", "vi"} {
		if _, err := exec.LookPath(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

// openDescriptionEditorCmd writes initial content to a temp file and opens it
// in the user's preferred editor. The returned message contains the edited content
// or an error.
func openDescriptionEditorCmd(initial string) tea.Cmd {
	editor := chooseEditor()
	if editor == "" {
		return func() tea.Msg {
			return descriptionEditedMsg{err: fmt.Errorf("no editor found (set $EDITOR or install nvim/vim/vi)")}
		}
	}

	tmpFile, err := os.CreateTemp("", "kanji-description-*.md")
	if err != nil {
		return func() tea.Msg {
			return descriptionEditedMsg{err: err}
		}
	}

	path := tmpFile.Name()
	if _, err := tmpFile.WriteString(initial); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(path)
		return func() tea.Msg {
			return descriptionEditedMsg{err: err}
		}
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(path)
		return func() tea.Msg {
			return descriptionEditedMsg{err: err}
		}
	}

	parts := strings.Fields(editor)
	if len(parts) == 0 {
		_ = os.Remove(path)
		return func() tea.Msg {
			return descriptionEditedMsg{err: fmt.Errorf("invalid editor command")}
		}
	}

	cmd := exec.Command(parts[0], append(parts[1:], path)...)
	return tea.ExecProcess(cmd, func(execErr error) tea.Msg {
		defer os.Remove(path)
		if execErr != nil {
			return descriptionEditedMsg{err: execErr}
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return descriptionEditedMsg{err: readErr}
		}
		return descriptionEditedMsg{content: strings.TrimRight(string(content), "\n")}
	})
}

// handleExternalDescriptionEdited processes a descriptionEditedMsg for external editor
// sessions. It clears the pending task ID and either updates the description or
// surfaces the error.
func (m Model) handleExternalDescriptionEdited(msg descriptionEditedMsg) (tea.Model, tea.Cmd) {
	if m.editingDescTask == "" {
		return m, nil
	}
	taskID := m.editingDescTask
	m.editingDescTask = ""
	if msg.err != nil {
		m.err = msg.err
		m.statusLine = fmt.Sprintf("editor error: %v", msg.err)
		return m, nil
	}
	return m, m.updateTaskDescriptionCmd(taskID, msg.content)
}
