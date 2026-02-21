package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	liptable "github.com/charmbracelet/lipgloss/table"
)

func (m Model) renderListScreen() string {
	containerWidth := max(40, m.width-2)

	filterBar := m.renderListFilterBar(containerWidth)
	footer := m.renderListFooter(containerWidth)

	mainHeight := m.height - lipgloss.Height(filterBar) - lipgloss.Height(footer)
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

	left := lipgloss.NewStyle().
		Width(mainWidth).
		Height(mainHeight).
		MaxWidth(mainWidth).
		MaxHeight(mainHeight).
		Render(m.renderListView(mainWidth, mainHeight))
	center := left
	if detailWidth > 0 {
		right := lipgloss.NewStyle().
			Width(detailWidth).
			Height(mainHeight).
			MaxWidth(detailWidth).
			MaxHeight(mainHeight).
			Render(m.renderDetailView(detailWidth, mainHeight))
		center = lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", gap), right)
	}
	center = lipgloss.NewStyle().Width(containerWidth).Height(mainHeight).Render(center)

	page := lipgloss.JoinVertical(lipgloss.Left, filterBar, center, footer)
	return lipgloss.NewStyle().Padding(0, 1).Render(page)
}

func (m Model) renderListFilterBar(width int) string {
	contentWidth := boxContentWidth(width, 1, true)
	statusFilterValue := m.statusFilterLabel()
	if m.filterIndex >= 0 && strings.TrimSpace(m.columnFilter) != "" {
		statusFilterValue = lipgloss.NewStyle().
			Foreground(m.colorForColumnID(m.columnFilter)).
			Bold(true).
			Render(statusFilterValue)
	}
	filters := []string{
		fmt.Sprintf("Status: %s", statusFilterValue),
		fmt.Sprintf("Due: %s", m.dueFilterLabel()),
	}
	if m.priorityFilter >= 0 {
		filters = append(filters, fmt.Sprintf("Priority: p%d", m.priorityFilter))
	}
	filterLabel := strings.Join(filters, " + ")

	content := fmt.Sprintf("View: List | Sort: %s | Filter: %s", m.sortModeLabel(), filterLabel)
	if strings.TrimSpace(m.titleFilter) != "" {
		content = fmt.Sprintf("%s | Search: %s", content, m.titleFilter)
	}
	panel := lipgloss.NewStyle().
		Width(contentWidth).
		Padding(0, 1).
		Foreground(lipgloss.Color("253")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("250")).
		Render(content)
	return withTopBorderLabel(panel, m.renderWorkspaceBoardStrip())
}

func (m Model) renderWorkspaceBoardStrip() string {
	workspace := truncate(strings.TrimSpace(m.workspaceName), 24)
	if workspace == "" {
		workspace = "-"
	}
	workspaceLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("151")).Bold(true).Render("[" + workspace + "]")

	if len(m.boards) == 0 {
		return workspaceLabel + " " + lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render("-")
	}

	current := boardIndexByID(m.boards, m.boardID)
	if current < 0 {
		current = 0
	}
	prev := (current - 1 + len(m.boards)) % len(m.boards)
	next := (current + 1) % len(m.boards)

	prevStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	currStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("114")).Bold(true)
	nextStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	type boardSlot struct {
		index int
		style lipgloss.Style
	}
	slots := []boardSlot{
		{index: prev, style: prevStyle},
		{index: current, style: currStyle},
		{index: next, style: nextStyle},
	}

	seen := map[string]struct{}{}
	segments := make([]string, 0, 3)
	for _, slot := range slots {
		board := m.boards[slot.index]
		if _, ok := seen[board.ID]; ok {
			continue
		}
		seen[board.ID] = struct{}{}

		name := truncate(strings.TrimSpace(board.Name), 20)
		if name == "" {
			name = "-"
		}
		segments = append(segments, slot.style.Render(name))
	}

	return workspaceLabel + " " + strings.Join(segments, " | ")
}

func withTopBorderLabel(panel, label string) string {
	if strings.TrimSpace(label) == "" {
		return panel
	}

	lines := strings.Split(panel, "\n")
	if len(lines) == 0 {
		return panel
	}

	panelWidth := lipgloss.Width(lines[0])
	if panelWidth < 6 {
		return panel
	}

	innerWidth := panelWidth - 2 // between corner glyphs
	labelText := " " + strings.TrimSpace(label) + " "
	labelWidth := lipgloss.Width(labelText)
	// Keep at least one line segment after the label before top-right corner.
	if labelWidth >= innerWidth {
		return panel
	}

	fillWidth := innerWidth - 1 - labelWidth
	if fillWidth < 0 {
		fillWidth = 0
	}

	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	topLine := borderStyle.Render("╭─") + labelText + borderStyle.Render(strings.Repeat("─", fillWidth)+"╮")
	lines[0] = topLine
	return strings.Join(lines, "\n")
}

func (m Model) renderListView(width, height int) string {
	panelContentWidth := boxContentWidth(width, 1, true)
	panelContentHeight := boxContentHeight(height, true)
	innerWidth := max(12, panelContentWidth-2)

	selectedCount := 0
	totalCount := len(m.tasks)
	if totalCount > 0 {
		selectedCount = m.selected
		if selectedCount < 0 {
			selectedCount = 0
		}
		if selectedCount >= totalCount {
			selectedCount = totalCount - 1
		}
		selectedCount++
	}
	counterText := fmt.Sprintf("%d of %d", selectedCount, totalCount)

	if len(m.tasks) == 0 {
		empty := lipgloss.NewStyle().
			Width(innerWidth).
			Height(panelContentHeight).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("245")).
			Render("No tasks yet.\nPress n to create one.")
		panel := lipgloss.NewStyle().
			Width(panelContentWidth).
			Height(panelContentHeight).
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("250")).
			Render(empty)
		return withBottomCounter(panel, counterText)
	}

	visibleRows := max(2, panelContentHeight) // Includes table header row.
	visibleTaskRows := max(1, visibleRows-1)

	offset := 0
	if m.selected >= visibleTaskRows {
		offset = m.selected - visibleTaskRows + 1
	}
	maxOffset := max(0, len(m.tasks)-visibleTaskRows)
	if offset > maxOffset {
		offset = maxOffset
	}

	const (
		statusContentWidth = 10
		dueContentWidth    = 10
		priContentWidth    = 3

		cellHorizontalPadding = 2 // left + right
		statusCellWidth       = statusContentWidth + cellHorizontalPadding
		dueCellWidth          = dueContentWidth + cellHorizontalPadding
		priCellWidth          = priContentWidth + cellHorizontalPadding
	)

	tableWidth := innerWidth
	fixedTailWidth := statusCellWidth + dueCellWidth + priCellWidth
	const tableGutterReserve = 4
	taskContentWidth := max(8, tableWidth-fixedTailWidth-tableGutterReserve)
	rows := make([][]string, 0, len(m.tasks))
	for _, task := range m.tasks {
		status := m.statusLabelForTask(task)
		due := "-"
		if task.DueAt != nil {
			due, _ = m.dueDisplay(*task.DueAt)
		}
		rows = append(rows, []string{
			truncate(task.Title, taskContentWidth),
			truncate(status, statusContentWidth),
			truncate(due, dueContentWidth),
			fmt.Sprintf("p%d", task.Priority),
		})
	}

	selectedTableRow := m.selected - offset
	t := liptable.New().
		Headers("Task", "Status", "Due", "Pri").
		Rows(rows...).
		Border(lipgloss.HiddenBorder()).
		BorderLeft(false).
		BorderRight(false).
		BorderTop(false).
		BorderBottom(false).
		BorderColumn(false).
		BorderRow(false).
		BorderHeader(false).
		Width(tableWidth).
		Height(visibleRows).
		Wrap(false).
		Offset(offset).
		StyleFunc(func(row, col int) lipgloss.Style {
			style := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("252"))
			if row == liptable.HeaderRow {
				style = lipgloss.NewStyle().Padding(0, 1).Bold(true).Foreground(lipgloss.Color("245"))
			} else if row == selectedTableRow {
				style = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62"))
			}

			if row >= 0 {
				taskIndex := row + offset
				if taskIndex >= 0 && taskIndex < len(m.tasks) {
					task := m.tasks[taskIndex]
					if col == 1 {
						style = style.Foreground(m.statusColorForTask(task)).Bold(true)
					}
					priority := normalizePriority(task.Priority)
					priorityColor := priorityColor(priority)
					if col == 3 {
						style = style.Foreground(priorityColor).Bold(true)
					}
					if col == 2 {
						if task.DueAt == nil {
							style = style.Foreground(dueColorNoDueSet)
						} else {
							_, dueColor := m.dueDisplay(*task.DueAt)
							style = style.Foreground(dueColor).Bold(true)
						}
					}
				}
			}

			switch col {
			case 0:
				return style.MaxWidth(taskContentWidth)
			case 1:
				return style.Width(statusCellWidth)
			case 2:
				return style.Width(dueCellWidth)
			case 3:
				return style.Width(priCellWidth)
			default:
				return style
			}
		})

	panel := lipgloss.NewStyle().
		Width(panelContentWidth).
		Height(panelContentHeight).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("250")).
		Render(t.String())

	return withBottomCounter(panel, counterText)
}

func withBottomCounter(panel, counter string) string {
	if strings.TrimSpace(counter) == "" {
		return panel
	}

	lines := strings.Split(panel, "\n")
	if len(lines) == 0 {
		return panel
	}

	panelWidth := lipgloss.Width(lines[len(lines)-1])
	if panelWidth < 4 {
		return panel
	}

	counterText := " " + strings.TrimSpace(counter) + " "
	innerWidth := panelWidth - 2 // width between rounded corners
	if lipgloss.Width(counterText) > innerWidth {
		trimmed := truncate(strings.TrimSpace(counter), max(1, innerWidth-2))
		counterText = " " + trimmed + " "
		if lipgloss.Width(counterText) > innerWidth {
			counterText = truncate(strings.TrimSpace(counter), innerWidth)
		}
	}

	fillWidth := innerWidth - lipgloss.Width(counterText)
	if fillWidth < 0 {
		fillWidth = 0
	}

	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	counterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	bottomLine := borderStyle.Render("╰"+strings.Repeat("─", fillWidth)) +
		counterStyle.Render(counterText) +
		borderStyle.Render("╯")

	lines[len(lines)-1] = bottomLine
	return strings.Join(lines, "\n")
}

func (m Model) renderListFooter(width int) string {
	contentWidth := boxContentWidth(width, 1, true)
	shortcuts := "?: Keybinds | W: Workspaces | b: Board manager | [: Prev board | ]: Next board | F: Filter/Sort | S: Status | Z: Due | O: Sort | N: Create | /: Search | Enter: Open/Move | Q: Quit"
	if strings.TrimSpace(m.titleFilter) != "" {
		shortcuts += " | X: Clear search"
	}
	helpLine := lipgloss.NewStyle().
		Width(contentWidth).
		Padding(0, 1).
		Foreground(lipgloss.Color("248")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("250")).
		Render(shortcuts)

	lines := []string{}
	if strings.TrimSpace(m.statusLine) != "" {
		statusLine := lipgloss.NewStyle().
			Width(contentWidth).
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
	contentWidth := boxContentWidth(width, 1, true)
	switch m.inputMode {
	case inputSearch, inputAddComment, inputTaskForm:
		return lipgloss.NewStyle().
			Width(contentWidth).
			Padding(0, 1).
			Foreground(lipgloss.Color("221")).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("250")).
			Render(m.textInput.View())
	case inputEditDescription:
		return lipgloss.NewStyle().
			Width(contentWidth).
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

func boxContentWidth(outerWidth, horizontalPadding int, bordered bool) int {
	width := outerWidth - (horizontalPadding * 2)
	if bordered {
		width -= 2
	}
	if width < 1 {
		return 1
	}
	return width
}

func boxContentHeight(outerHeight int, bordered bool) int {
	height := outerHeight
	if bordered {
		height -= 2
	}
	if height < 1 {
		return 1
	}
	return height
}

func priorityColor(priority int) lipgloss.Color {
	switch priority {
	case 0: // Critical
		return lipgloss.Color("203")
	case 1: // Urgent
		return lipgloss.Color("208")
	case 2: // High
		return lipgloss.Color("220")
	case 3: // Medium
		return lipgloss.Color("117")
	case 4: // Low
		return lipgloss.Color("114")
	case 5: // None
		return lipgloss.Color("245")
	default:
		return lipgloss.Color("240")
	}
}
