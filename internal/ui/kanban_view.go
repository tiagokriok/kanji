package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderKanbanView(height int) string {
	if len(m.columns) == 0 {
		return "No columns"
	}

	columnWidth := max(24, (m.width-4)/max(1, len(m.columns)))
	cards := make([]string, 0, len(m.columns))
	for ci, col := range m.columns {
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("221")).Padding(0, 1)
		if ci == m.activeColumn {
			headerStyle = headerStyle.Background(lipgloss.Color("58")).Foreground(lipgloss.Color("230"))
		}

		rows := []string{headerStyle.Width(columnWidth - 2).Render(fmt.Sprintf("%s", col.Name))}
		colTasks := m.tasksForColumn(col.ID)
		if len(colTasks) == 0 {
			rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(0, 1).Render("(empty)"))
		}
		for ri, task := range colTasks {
			line := task.Title
			if len(line) > columnWidth-6 {
				line = line[:columnWidth-9] + "..."
			}
			style := lipgloss.NewStyle().Padding(0, 1)
			if ci == m.activeColumn && ri == m.kanbanRow {
				style = style.Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62"))
			}
			rows = append(rows, style.Render(line))
		}

		panel := lipgloss.NewStyle().Width(columnWidth).Height(height).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).Render(strings.Join(rows, "\n"))
		cards = append(cards, panel)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cards...)
}
