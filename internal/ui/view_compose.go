package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// renderBaseView renders the base screen without overlays.
func (m Model) renderBaseView() string {
	if m.viewMode == viewList {
		return m.renderListScreen()
	}

	containerWidth := max(40, m.width-2)
	header := m.renderHeader(containerWidth)
	footer := lipgloss.NewStyle().Width(containerWidth).Render(m.renderFooter())
	bodyHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
	if bodyHeight < 5 {
		bodyHeight = 5
	}

	detailWidth := 0
	mainWidth := containerWidth
	if m.showDetails {
		detailWidth = m.width / 3
		if detailWidth < 34 {
			detailWidth = 34
		}
		mainWidth = containerWidth - detailWidth - 1
		if mainWidth < 20 {
			mainWidth = containerWidth
			detailWidth = 0
		}
	}

	mainPane := m.renderKanbanView(mainWidth, bodyHeight)
	if detailWidth > 0 {
		detailPane := m.renderDetailView(detailWidth, bodyHeight)
		mainPane = lipgloss.JoinHorizontal(lipgloss.Top, mainPane, detailPane)
	}
	mainPane = lipgloss.NewStyle().Width(containerWidth).Height(bodyHeight).Render(mainPane)

	content := lipgloss.JoinVertical(lipgloss.Left, header, mainPane, footer)
	return lipgloss.NewStyle().Padding(0, 1).Render(content)
}

// wrapOverlays applies active overlays on top of the base view.
func (m Model) wrapOverlays(base string) string {
	if m.showKeybinds {
		return m.renderKeybindPanel(base)
	}
	if m.showFilters {
		return m.renderFilterPanel(base)
	}
	if m.showContexts {
		return m.renderContextPanel(base)
	}
	base = m.renderTaskFormOverlay(base)
	if m.showTaskView {
		return m.renderTaskViewerPanel(base)
	}
	return base
}
