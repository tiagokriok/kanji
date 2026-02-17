package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderListScreen() string {
	containerWidth := max(40, m.width-2)

	topRow := m.renderListTopRow(containerWidth)
	filterBar := m.renderListFilterBar(containerWidth)
	footer := m.renderListFooter(containerWidth)

	mainHeight := m.height - lipgloss.Height(topRow) - lipgloss.Height(filterBar) - lipgloss.Height(footer)
	if mainHeight < 8 {
		mainHeight = 8
	}

	detailWidth := 0
	gap := 1
	if m.showDetails {
		detailWidth = max(36, containerWidth/4)
		if detailWidth > containerWidth-28 {
			detailWidth = containerWidth - 28
		}
	}
	mainWidth := containerWidth
	if detailWidth > 0 {
		mainWidth = containerWidth - detailWidth - gap
	}
	if mainWidth < 24 {
		mainWidth = containerWidth
		detailWidth = 0
	}

	left := m.renderListView(mainWidth, mainHeight)
	center := left
	if detailWidth > 0 {
		right := m.renderDetailView(detailWidth, mainHeight)
		center = lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", gap), right)
	}

	page := lipgloss.JoinVertical(lipgloss.Left, topRow, filterBar, center, footer)
	return lipgloss.NewStyle().Padding(0, 1).Render(page)
}

func (m Model) renderListTopRow(width int) string {
	bar := lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Foreground(lipgloss.Color("255"))

	icon := lipgloss.NewStyle().
		Foreground(lipgloss.Color("195")).
		Render("~")

	user := os.Getenv("USER")
	if strings.TrimSpace(user) == "" {
		user = "there"
	}
	greeting := lipgloss.NewStyle().Foreground(lipgloss.Color("195")).Render("Hello, ")
	name := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("191")).Render(user + "!")
	left := fmt.Sprintf("%s  %s%s", icon, greeting, name)

	clock := lipgloss.NewStyle().Foreground(lipgloss.Color("183")).Render(time.Now().Format("Mon Jan 2 15:04:05 MST"))
	return bar.Render(lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(max(1, width-28)).Render(left),
		lipgloss.NewStyle().Width(28).Align(lipgloss.Right).Render(clock),
	))
}

func (m Model) renderListFilterBar(width int) string {
	content := "View: List | Order: Priority"
	if strings.TrimSpace(m.titleFilter) != "" {
		content = fmt.Sprintf("%s | Search: %s", content, m.titleFilter)
	}
	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Foreground(lipgloss.Color("253")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("250")).
		Render(content)
}

func (m Model) renderListView(width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255")).
		Padding(0, 1)
	rowStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color("252"))
	selectedStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("62"))

	lines := make([]string, 0, len(m.tasks)+1)
	lines = append(lines, titleStyle.Render("TASK LIST VIEW"))
	if len(m.tasks) == 0 {
		empty := lipgloss.NewStyle().
			Width(max(1, width-4)).
			Height(max(1, height-4)).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("245")).
			Render("No tasks yet.\nPress n to create one.")
		return lipgloss.NewStyle().
			Width(width).
			Height(height).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("250")).
			Render(empty)
	}

	for i, task := range m.tasks {
		column := "-"
		if task.ColumnID != nil {
			column = m.columnName(*task.ColumnID)
		}
		due := "-"
		if task.DueAt != nil {
			due = task.DueAt.Format("2006-01-02")
		}
		priority := fmt.Sprintf("p%d", task.Priority)
		label := truncate(fmt.Sprintf("%s  [%s]  due:%s  %s", task.Title, column, due, priority), max(12, width-8))
		if i == m.selected && m.viewMode == viewList {
			lines = append(lines, selectedStyle.Render(label))
		} else {
			lines = append(lines, rowStyle.Render(label))
		}
	}

	if len(lines) > height {
		start := 0
		if m.selected >= height-2 {
			start = m.selected - (height - 3)
		}
		end := min(len(lines), start+height)
		lines = lines[start:end]
	}

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0, 0).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("250")).
		Render(strings.Join(lines, "\n"))
}

func (m Model) renderListFooter(width int) string {
	shortcuts := "N: Create task | E: Edit task | D: Toggle details | /: Search | Enter: Open/Move | j k: Up/Down | h l: Left/Right"
	helpLine := lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Foreground(lipgloss.Color("248")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("250")).
		Render(shortcuts)

	lines := []string{}
	if strings.TrimSpace(m.statusLine) != "" {
		statusLine := lipgloss.NewStyle().
			Width(width).
			Padding(0, 1).
			Foreground(lipgloss.Color("222")).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("250")).
			Render(m.statusLine)
		lines = append(lines, statusLine)
	}
	lines = append(lines, helpLine)

	inputLine := m.renderInlineInput(width)
	if inputLine != "" {
		lines = append(lines, inputLine)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderInlineInput(width int) string {
	switch m.inputMode {
	case inputSearch, inputAddComment, inputTaskForm:
		return lipgloss.NewStyle().
			Width(width).
			Padding(0, 1).
			Foreground(lipgloss.Color("221")).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("250")).
			Render(m.textInput.View())
	case inputEditDescription:
		return lipgloss.NewStyle().
			Width(width).
			Padding(0, 1).
			Foreground(lipgloss.Color("221")).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("250")).
			Render(m.textArea.View())
	default:
		return ""
	}
}

func (m Model) columnName(columnID string) string {
	for _, c := range m.columns {
		if c.ID == columnID {
			return c.Name
		}
	}
	return columnID
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncate(input string, maxLen int) string {
	if maxLen <= 0 || len(input) <= maxLen {
		return input
	}
	if maxLen < 4 {
		return input[:maxLen]
	}
	return input[:maxLen-3] + "..."
}
