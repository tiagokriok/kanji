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
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(colorFromHexOrDefault(col.Color, "#333333")).
			Width(columnWidth-2).
			Padding(0, 1)

		rows := []string{headerStyle.Render(fmt.Sprintf("%s", col.Name))}
		colTasks := m.tasksForColumn(col.ID)
		if len(colTasks) == 0 {
			rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(0, 1).Render("(empty)"))
		}
		for ri, task := range colTasks {
			isActive := ci == m.activeColumn && ri == m.kanbanRow

			p := normalizePriority(task.Priority)
			prefix := ""
			if p > 0 {
				prefix = lipgloss.NewStyle().Foreground(priorityColor(p)).Render("●") + " "
			}

			titleMax := columnWidth - 8
			title := task.Title
			if len([]rune(title)) > titleMax {
				title = string([]rune(title)[:titleMax-3]) + "..."
			}
			titleLine := prefix + title

			content := titleLine
			if task.DueAt != nil {
				dueText, dueColor := m.dueDisplay(*task.DueAt)
				meta := lipgloss.NewStyle().Foreground(dueColor).Render("  due: " + dueText)
				content += "\n" + meta
			}

			borderColor := lipgloss.Color("238")
			if isActive {
				borderColor = lipgloss.Color("62")
			}

			card := lipgloss.NewStyle().
				Width(columnWidth-4).
				Padding(0, 1).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(borderColor).
				Render(content)
			rows = append(rows, card)
		}

		panel := lipgloss.NewStyle().Width(columnWidth).Height(height).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("235")).Render(strings.Join(rows, "\n"))
		cards = append(cards, panel)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cards...)
}
