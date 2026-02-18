package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	liptable "github.com/charmbracelet/lipgloss/table"
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

	innerWidth := max(12, width-4)
	visibleRows := max(2, height-4) // Includes table header row.
	visibleTaskRows := max(1, visibleRows-1)

	offset := 0
	if m.selected >= visibleTaskRows {
		offset = m.selected - visibleTaskRows + 1
	}
	maxOffset := max(0, len(m.tasks)-visibleTaskRows)
	if offset > maxOffset {
		offset = maxOffset
	}

	taskColWidth := max(16, innerWidth-31)
	rows := make([][]string, 0, len(m.tasks))
	for _, task := range m.tasks {
		column := "-"
		if task.ColumnID != nil {
			column = m.columnName(*task.ColumnID)
		}
		due := "-"
		if task.DueAt != nil {
			due = task.DueAt.Format("2006-01-02")
		}
		rows = append(rows, []string{
			truncate(task.Title, taskColWidth),
			truncate(column, 12),
			truncate(due, 10),
			fmt.Sprintf("p%d", task.Priority),
		})
	}

	selectedTableRow := m.selected - offset
	t := liptable.New().
		Headers("Task", "Status", "Due", "Pri").
		Rows(rows...).
		Border(lipgloss.HiddenBorder()).
		Width(innerWidth).
		Offset(offset).
		StyleFunc(func(row, col int) lipgloss.Style {
			style := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("252"))
			if row == liptable.HeaderRow {
				style = lipgloss.NewStyle().Padding(0, 1).Bold(true).Foreground(lipgloss.Color("245"))
			} else if row == selectedTableRow {
				style = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62"))
			}
			switch col {
			case 0:
				return style.MaxWidth(taskColWidth)
			case 1:
				return style.Width(12)
			case 2:
				return style.Width(10)
			case 3:
				return style.Width(4)
			default:
				return style
			}
		})

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0, 0).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("250")).
		Render(t.String())
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

func truncate(input string, maxLen int) string {
	if maxLen <= 0 || len(input) <= maxLen {
		return input
	}
	if maxLen < 4 {
		return input[:maxLen]
	}
	return input[:maxLen-3] + "..."
}
