package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/tiagokriok/kanji/internal/domain"
)

func (m Model) renderTaskViewerPanel(base string) string {
	_ = base

	task, ok := m.viewerTask()
	if !ok {
		return base
	}

	panelWidth := m.width - 6
	if panelWidth < 80 {
		panelWidth = 80
	}
	if panelWidth > m.width-2 {
		panelWidth = max(20, m.width-2)
	}

	panelHeight := m.height - 4
	if panelHeight < 18 {
		panelHeight = 18
	}
	if panelHeight > m.height-2 {
		panelHeight = max(10, m.height-2)
	}

	contentWidth := boxContentWidth(panelWidth, 1, true)
	contentHeight := boxContentHeight(panelHeight, true)

	rightWidth := max(24, contentWidth/4)
	if rightWidth > contentWidth-24 {
		rightWidth = contentWidth - 24
	}
	if rightWidth < 16 {
		rightWidth = 16
	}
	leftWidth := contentWidth - rightWidth - 1
	if leftWidth < 20 {
		leftWidth = 20
		rightWidth = max(16, contentWidth-leftWidth-1)
	}

	leftLines := m.renderTaskViewerLeftLines(task, leftWidth, contentHeight)
	rightLines := m.renderTaskViewerRightLines(rightWidth, contentHeight)

	if len(leftLines) < contentHeight {
		leftLines = append(leftLines, makeViewerBlankLines(leftWidth, contentHeight-len(leftLines))...)
	}
	if len(rightLines) < contentHeight {
		rightLines = append(rightLines, makeViewerBlankLines(rightWidth, contentHeight-len(rightLines))...)
	}

	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	rows := make([]string, 0, contentHeight)
	for i := 0; i < contentHeight; i++ {
		rows = append(rows, leftLines[i]+separatorStyle.Render("│")+rightLines[i])
	}

	panel := lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("250")).
		Render(strings.Join(rows, "\n"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
}

func (m Model) renderTaskViewerLeftLines(task domain.Task, width, height int) []string {
	titleStyle := lipgloss.NewStyle().Width(width).Foreground(lipgloss.Color("231")).Bold(true)
	metaStyle := lipgloss.NewStyle().Width(width).Foreground(lipgloss.Color("246"))
	descStyle := lipgloss.NewStyle().Width(width).Foreground(lipgloss.Color("252"))

	dueText := "-"
	dueColor := dueColorDefault
	if task.DueAt != nil {
		dueText, dueColor = m.dueDisplay(*task.DueAt)
	}
	dueValue := lipgloss.NewStyle().Foreground(dueColor).Bold(true).Render(dueText)
	priorityValue := lipgloss.NewStyle().
		Foreground(priorityColor(normalizePriority(task.Priority))).
		Bold(true).
		Render(fmt.Sprintf("p%d", task.Priority))
	statusValue := lipgloss.NewStyle().
		Foreground(m.statusColorForTask(task)).
		Bold(true).
		Render(m.statusLabelForTask(task))

	lines := []string{
		titleStyle.Render(truncate(task.Title, max(1, width))),
		metaStyle.Render(fmt.Sprintf("%s | %s | %s", dueValue, priorityValue, statusValue)),
		descStyle.Render(""),
	}

	description := strings.TrimSpace(task.DescriptionMD)
	if description == "" {
		description = "(empty)"
	}
	description = normalizeViewerText(description)
	for _, line := range wrapViewerText(description, width) {
		lines = append(lines, descStyle.Render(line))
		if len(lines) >= height {
			return lines[:height]
		}
	}
	return lines
}

func (m Model) renderTaskViewerRightLines(width, height int) []string {
	headerStyle := lipgloss.NewStyle().Width(width).Foreground(lipgloss.Color("231")).Bold(true).Align(lipgloss.Center)
	commentMetaStyle := lipgloss.NewStyle().Width(width).Foreground(lipgloss.Color("245"))
	commentBodyStyle := lipgloss.NewStyle().Width(width).Foreground(lipgloss.Color("252"))
	emptyStyle := lipgloss.NewStyle().Width(width).Foreground(lipgloss.Color("245"))

	lines := []string{
		headerStyle.Render("Comments"),
		emptyStyle.Render(""),
	}

	if m.comments == nil {
		lines = append(lines, emptyStyle.Render("(loading...)"))
		return trimViewerLines(lines, width, height)
	}
	if len(m.comments) == 0 {
		lines = append(lines, emptyStyle.Render("(none)"))
		return trimViewerLines(lines, width, height)
	}

	for _, comment := range m.comments {
		author := ""
		if comment.Author != nil {
			author = strings.TrimSpace(*comment.Author)
		}
		if author == "" {
			author = "unknown"
		}
		meta := fmt.Sprintf("%s  %s", author, m.formatCommentDateTime(comment.CreatedAt))
		lines = append(lines, commentMetaStyle.Render(truncate(meta, max(1, width))))

		body := strings.TrimSpace(comment.BodyMD)
		if body == "" {
			body = "(empty)"
		}
		body = normalizeViewerText(body)
		for _, bodyLine := range wrapViewerText(body, width-2) {
			lines = append(lines, commentBodyStyle.Render("  "+bodyLine))
			if len(lines) >= height {
				return lines[:height]
			}
		}
		lines = append(lines, emptyStyle.Render(""))
		if len(lines) >= height {
			return lines[:height]
		}
	}

	return trimViewerLines(lines, width, height)
}

func trimViewerLines(lines []string, width, height int) []string {
	if len(lines) > height {
		return lines[:height]
	}
	return append(lines, makeViewerBlankLines(width, height-len(lines))...)
}

func makeViewerBlankLines(width, count int) []string {
	style := lipgloss.NewStyle().Width(width)
	lines := make([]string, 0, count)
	for i := 0; i < count; i++ {
		lines = append(lines, style.Render(""))
	}
	return lines
}

func wrapViewerText(text string, width int) []string {
	text = normalizeViewerText(text)
	if width < 1 {
		return []string{text}
	}

	rawLines := strings.Split(text, "\n")
	wrapped := make([]string, 0, len(rawLines))
	for _, raw := range rawLines {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			wrapped = append(wrapped, "")
			continue
		}

		remaining := raw
		for len([]rune(remaining)) > width {
			runes := []rune(remaining)
			chunk := runes[:width]
			breakAt := -1
			for i := len(chunk) - 1; i >= 0; i-- {
				if chunk[i] == ' ' || chunk[i] == '\t' {
					breakAt = i
					break
				}
			}

			if breakAt <= 0 {
				wrapped = append(wrapped, string(chunk))
				remaining = strings.TrimLeft(string(runes[width:]), " \t")
				continue
			}

			wrapped = append(wrapped, strings.TrimSpace(string(chunk[:breakAt])))
			remaining = strings.TrimLeft(string(runes[breakAt:]), " \t")
		}
		if remaining != "" {
			wrapped = append(wrapped, remaining)
		}
	}
	return wrapped
}

func normalizeViewerText(text string) string {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	return normalized
}
