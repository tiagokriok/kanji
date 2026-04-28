package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderHeader(width int) string {
	viewLabel := "List"
	if m.viewMode == viewKanban {
		viewLabel = "Kanban"
	}
	filterParts := []string{fmt.Sprintf("status:%s", m.statusFilterLabel()), fmt.Sprintf("due:%s", strings.ToLower(m.dueFilterLabel()))}
	if m.priorityFilter >= 0 {
		filterParts = append(filterParts, fmt.Sprintf("priority:p%d", m.priorityFilter))
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))

	left := headerStyle.Render(fmt.Sprintf("%s / %s", m.workspaceName, m.boardName))
	right := metaStyle.Render(fmt.Sprintf("view:%s  sort:%s  filter:%s  search:%q", viewLabel, strings.ToLower(m.sortModeLabel()), strings.Join(filterParts, ","), m.titleFilter))
	if width > 20 {
		return lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(width/2).Render(left),
			lipgloss.NewStyle().Width(max(1, width-width/2-1)).Align(lipgloss.Right).Render(right),
		)
	}
	return left + " " + right
}

func (m Model) renderFooter() string {
	inputLine := ""
	switch m.inputMode {
	case inputSearch, inputAddComment:
		inputLine = lipgloss.NewStyle().Foreground(lipgloss.Color("221")).Render(m.textInput.View())
	case inputEditDescription:
		inputLine = lipgloss.NewStyle().Foreground(lipgloss.Color("221")).Render(m.textArea.View())
	}

	shortcuts := "?:help  n:new  /:search  enter:open  w:workspaces  b:boards  f:filters  q:quit"
	if strings.TrimSpace(m.titleFilter) != "" {
		shortcuts += " x:clear-search"
	}
	lines := make([]string, 0, 3)
	if strings.TrimSpace(m.statusLine) != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("222")).Render(m.statusLine))
	}
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(shortcuts))
	if inputLine != "" {
		lines = append(lines, inputLine)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
