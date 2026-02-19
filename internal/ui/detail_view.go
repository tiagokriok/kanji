package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderDetailView(width, height int) string {
	contentWidth := boxContentWidth(width, 1, true)
	contentHeight := boxContentHeight(height, true)
	panelStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("250"))
	task, ok := m.currentTask()
	if !ok {
		return panelStyle.Render("No task selected")
	}

	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("213")).Render(task.Title)
	meta := []string{}
	priorityValue := lipgloss.NewStyle().
		Foreground(priorityColor(normalizePriority(task.Priority))).
		Bold(true).
		Render(fmt.Sprintf("%d", task.Priority))
	meta = append(meta, fmt.Sprintf("Priority: %s", priorityValue))
	if task.DueAt != nil {
		meta = append(meta, fmt.Sprintf("Due: %s", m.formatDueDate(*task.DueAt)))
	}
	if task.ColumnID != nil || (task.Status != nil && strings.TrimSpace(*task.Status) != "") {
		statusValue := lipgloss.NewStyle().
			Foreground(m.statusColorForTask(task)).
			Bold(true).
			Render(m.statusLabelForTask(task))
		meta = append(meta, fmt.Sprintf("Status: %s", statusValue))
	}
	metaLine := strings.Join(meta, " | ")

	descTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("221")).Render("Description")
	desc := renderMarkdownMinimal(task.DescriptionMD)
	if strings.TrimSpace(desc) == "" {
		desc = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("(empty)")
	}

	commentsTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("221")).Render("Comments")
	commentLines := make([]string, 0, len(m.comments))
	if len(m.comments) == 0 {
		commentLines = append(commentLines, lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("(none)"))
	}
	for _, c := range m.comments {
		author := "anon"
		if c.Author != nil && *c.Author != "" {
			author = *c.Author
		}
		timestamp := m.formatCommentDateTime(c.CreatedAt)
		body := renderMarkdownMinimal(c.BodyMD)
		commentLines = append(commentLines, fmt.Sprintf("%s  %s\n%s", author, timestamp, indentLines(body, "  ")))
	}

	content := []string{header, metaLine, "", descTitle, desc, "", commentsTitle, strings.Join(commentLines, "\n\n")}
	joined := strings.Join(content, "\n")
	return panelStyle.Render(joined)
}

func renderMarkdownMinimal(md string) string {
	if strings.TrimSpace(md) == "" {
		return ""
	}
	lines := strings.Split(md, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "### "):
			out = append(out, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("110")).Render(strings.TrimPrefix(trimmed, "### ")))
		case strings.HasPrefix(trimmed, "## "):
			out = append(out, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117")).Render(strings.TrimPrefix(trimmed, "## ")))
		case strings.HasPrefix(trimmed, "# "):
			out = append(out, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("153")).Render(strings.TrimPrefix(trimmed, "# ")))
		case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
			out = append(out, "• "+trimmed[2:])
		default:
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

func indentLines(s, prefix string) string {
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}
