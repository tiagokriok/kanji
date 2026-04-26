package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderKanbanView(width, height int) string {
	if len(m.columns) == 0 {
		return "No columns"
	}

	panelContentHeight := boxContentHeight(height, true)
	columnOuterWidth := max(24, width/max(1, len(m.columns)))
	panelContentWidth := boxContentWidth(columnOuterWidth, 0, true)
	cardContentWidth := boxContentWidth(panelContentWidth, 1, true)
	cards := make([]string, 0, len(m.columns))
	for ci, col := range m.columns {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(contrastingTextColorFromHexOrDefault(col.Color, "255")).
			Background(colorFromHexOrDefault(col.Color, "#333333")).
			Width(max(1, panelContentWidth-2)).
			Padding(0, 1)

		header := headerStyle.Render(fmt.Sprintf("%s", col.Name))
		colTasks := m.tasksForColumn(col.ID)
		cardRows := make([]string, 0, max(1, len(colTasks)))
		if len(colTasks) == 0 {
			cardRows = append(cardRows, lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(0, 1).Render("(empty)"))
		}
		for ri, task := range colTasks {
			isActive := ci == m.activeColumn && ri == m.kanbanRow

			p := normalizePriority(task.Priority)
			prefix := ""
			if p > 0 {
				prefix = lipgloss.NewStyle().Foreground(priorityColor(p)).Render("●") + " "
			}

			titleMax := max(4, cardContentWidth-4)
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
				Width(cardContentWidth).
				Padding(0, 1).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(borderColor).
				Render(content)
			cardRows = append(cardRows, card)
		}

		activeRow := -1
		if ci == m.activeColumn {
			activeRow = m.kanbanRow
		}
		content := renderKanbanColumnContent(header, cardRows, panelContentHeight, activeRow)
		panel := lipgloss.NewStyle().Width(panelContentWidth).Height(panelContentHeight).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("235")).Render(content)
		cards = append(cards, panel)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cards...)
}

func renderKanbanColumnContent(header string, rows []string, maxHeight, activeRow int) string {
	if maxHeight < 1 {
		return ""
	}

	content := []string{header}
	remainingHeight := maxHeight - lipgloss.Height(header)
	if remainingHeight < 1 || len(rows) == 0 {
		return strings.Join(content, "\n")
	}

	start, end := kanbanVisibleRange(rows, remainingHeight, activeRow)
	for _, row := range rows[start:end] {
		content = append(content, row)
	}
	return strings.Join(content, "\n")
}

func kanbanVisibleRange(rows []string, maxHeight, activeRow int) (int, int) {
	if len(rows) == 0 || maxHeight < 1 {
		return 0, 0
	}
	if activeRow < 0 || activeRow >= len(rows) {
		return kanbanFillForward(rows, maxHeight, 0)
	}

	usedHeight := lipgloss.Height(rows[activeRow])
	if usedHeight > maxHeight {
		return 0, 0
	}

	start := activeRow
	end := activeRow + 1
	for start > 0 {
		nextHeight := lipgloss.Height(rows[start-1])
		if usedHeight+nextHeight > maxHeight {
			break
		}
		start--
		usedHeight += nextHeight
	}
	for end < len(rows) {
		nextHeight := lipgloss.Height(rows[end])
		if usedHeight+nextHeight > maxHeight {
			break
		}
		usedHeight += nextHeight
		end++
	}
	return start, end
}

func kanbanFillForward(rows []string, maxHeight, start int) (int, int) {
	usedHeight := 0
	end := start
	for end < len(rows) {
		nextHeight := lipgloss.Height(rows[end])
		if usedHeight+nextHeight > maxHeight {
			break
		}
		usedHeight += nextHeight
		end++
	}
	return start, end
}
