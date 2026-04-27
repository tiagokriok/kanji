package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keybindEntry struct {
	ID    string
	Key   string
	Label string
}

func (m Model) keybindEntries() []keybindEntry {
	entries := []keybindEntry{
		{ID: "new_task", Key: "n", Label: "Create task"},
		{ID: "edit_task", Key: "e", Label: "Edit selected task"},
		{ID: "edit_description", Key: "E", Label: "Edit description"},
		{ID: "add_comment", Key: "c", Label: "Add comment"},
		{ID: "search", Key: "/", Label: "Search"},
		{ID: "open_filters", Key: "f", Label: "Open filter/sort panel"},
		{ID: "open_workspaces", Key: "w", Label: "Open workspace switcher"},
		{ID: "open_board_panel", Key: "b", Label: "Open board manager"},
		{ID: "prev_board", Key: "[", Label: "Previous board"},
		{ID: "next_board", Key: "]", Label: "Next board"},
		{ID: "toggle_details", Key: "d", Label: "Toggle details pane"},
		{ID: "open_move", Key: "Enter", Label: "Open task viewer"},
		{ID: "move_task", Key: "m", Label: "Move task to next status"},
		{ID: "toggle_view", Key: "Tab", Label: "Switch list/kanban"},
		{ID: "cycle_status", Key: "s", Label: "Cycle status filter"},
		{ID: "cycle_due_filter", Key: "z", Label: "Cycle due filter"},
		{ID: "cycle_sort", Key: "o", Label: "Cycle sort mode"},
		{ID: "move_up", Key: "↑", Label: "Move selection up"},
		{ID: "move_down", Key: "↓", Label: "Move selection down"},
		{ID: "move_left", Key: "←", Label: "Move selection left (kanban)"},
		{ID: "move_right", Key: "→", Label: "Move selection right (kanban)"},
		{ID: "move_task_left", Key: "Shift+←", Label: "Move card to left column (kanban)"},
		{ID: "move_task_right", Key: "Shift+→", Label: "Move card to right column (kanban)"},
		{ID: "quit", Key: "q", Label: "Quit"},
	}
	if strings.TrimSpace(m.titleFilter) != "" {
		entries = append(entries, keybindEntry{ID: "clear_search", Key: "x", Label: "Clear search"})
	}
	return entries
}

func (m Model) filteredKeybindEntries() []keybindEntry {
	query := strings.ToLower(strings.TrimSpace(m.keyFilter.Value()))
	entries := m.keybindEntries()
	if query == "" {
		return entries
	}

	filtered := make([]keybindEntry, 0, len(entries))
	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Key), query) || strings.Contains(strings.ToLower(entry.Label), query) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func (m *Model) openKeybindPanel() {
	m.overlayState.openKeybinds()
	m.keyFilter.SetValue("")
	m.keyFilter.Width = max(24, m.width/3)
	m.keyFilter.Focus()
}

func (m *Model) closeKeybindPanel() {
	m.overlayState.closeKeybinds()
	m.keyFilter.Blur()
}

func (m *Model) clampKeybindSelection() {
	entries := m.filteredKeybindEntries()
	if len(entries) == 0 {
		m.keySelected = 0
		return
	}
	if m.keySelected < 0 {
		m.keySelected = 0
	}
	if m.keySelected >= len(entries) {
		m.keySelected = len(entries) - 1
	}
}

func (m Model) updateKeybindPanel(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.keyFilter.Width = max(24, msg.Width/3)
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Cancel), key.Matches(msg, m.keys.ShowKeybinds):
			m.closeKeybindPanel()
			return m, nil
		case key.Matches(msg, m.keys.Up):
			m.keySelected--
			m.clampKeybindSelection()
			return m, nil
		case key.Matches(msg, m.keys.Down):
			m.keySelected++
			m.clampKeybindSelection()
			return m, nil
		case key.Matches(msg, m.keys.Confirm):
			entries := m.filteredKeybindEntries()
			if len(entries) == 0 {
				return m, nil
			}
			action := entries[m.keySelected].ID
			m.closeKeybindPanel()
			return m, func() tea.Msg { return executeActionMsg{action: action} }
		}
	}

	var cmd tea.Cmd
	m.keyFilter, cmd = m.keyFilter.Update(msg)
	m.clampKeybindSelection()
	return m, cmd
}

func (m Model) renderKeybindPanel(base string) string {
	_ = base

	panelWidth := m.width * 3 / 4
	if panelWidth < 64 {
		panelWidth = 64
	}
	if panelWidth > m.width-2 {
		panelWidth = max(20, m.width-2)
	}

	panelHeight := m.height * 3 / 4
	if panelHeight < 12 {
		panelHeight = 12
	}
	if panelHeight > m.height-2 {
		panelHeight = max(8, m.height-2)
	}

	contentWidth := boxContentWidth(panelWidth, 1, true)
	contentHeight := boxContentHeight(panelHeight, true)
	listHeight := max(1, contentHeight-5)

	entries := m.filteredKeybindEntries()
	offset := 0
	if m.keySelected >= listHeight {
		offset = m.keySelected - listHeight + 1
	}
	maxOffset := max(0, len(entries)-listHeight)
	if offset > maxOffset {
		offset = maxOffset
	}

	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("151")).Render("Keybindings"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("Type to filter | Enter: execute | Esc: close"),
		m.keyFilter.View(),
		"",
	}

	if len(entries) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("(no keybindings match filter)"))
	} else {
		end := offset + listHeight
		if end > len(entries) {
			end = len(entries)
		}
		for i := offset; i < end; i++ {
			entry := entries[i]
			keyLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("221")).Width(12).Render(entry.Key)
			descLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(entry.Label)
			line := keyLabel + " " + descLabel
			if i == m.keySelected {
				line = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62")).Render(line)
			}
			lines = append(lines, line)
		}
	}

	panel := lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("151")).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
}
