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
	maxDescLines := max(2, contentHeight-5)
	maxDescWidth := max(12, contentWidth-4)
	descPreview := previewMarkdown(task.DescriptionMD, maxDescLines, maxDescWidth)
	desc := renderMarkdownMinimal(descPreview)
	if strings.TrimSpace(desc) == "" {
		desc = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("(empty)")
	}

	content := []string{header, metaLine, "", descTitle, desc}
	joined := strings.Join(content, "\n")
	return panelStyle.Render(joined)
}

func previewMarkdown(md string, maxLines, maxLineWidth int) string {
	if strings.TrimSpace(md) == "" {
		return ""
	}
	if maxLines < 1 {
		maxLines = 1
	}
	if maxLineWidth < 4 {
		maxLineWidth = 4
	}

	source := strings.Split(md, "\n")
	out := make([]string, 0, min(len(source), maxLines))
	for i := 0; i < len(source) && len(out) < maxLines; i++ {
		out = append(out, truncate(source[i], maxLineWidth))
	}

	if len(source) > maxLines && len(out) > 0 {
		last := strings.TrimSpace(out[len(out)-1])
		if last == "" {
			last = "..."
		} else if !strings.HasSuffix(last, "...") {
			last = truncate(last, max(4, maxLineWidth-3)) + "..."
		}
		out[len(out)-1] = last
	}

	return strings.Join(out, "\n")
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
