package ui

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/tiagokriok/kanji/internal/domain"
)

var uiHexColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func (m Model) statusLabelForTask(task domain.Task) string {
	if task.ColumnID != nil {
		return m.columnName(*task.ColumnID)
	}
	if task.Status != nil && strings.TrimSpace(*task.Status) != "" {
		return strings.TrimSpace(*task.Status)
	}
	return "-"
}

func (m Model) statusColorForTask(task domain.Task) lipgloss.Color {
	if task.ColumnID != nil {
		return m.colorForColumnID(*task.ColumnID)
	}
	if task.Status != nil {
		return m.colorForColumnName(*task.Status)
	}
	return lipgloss.Color("252")
}

func (m Model) colorForColumnID(columnID string) lipgloss.Color {
	for _, c := range m.columns {
		if c.ID == columnID {
			return colorFromHexOrDefault(c.Color, "252")
		}
	}
	return lipgloss.Color("252")
}

func (m Model) colorForColumnName(name string) lipgloss.Color {
	needle := strings.TrimSpace(name)
	if needle == "" {
		return lipgloss.Color("252")
	}
	for _, c := range m.columns {
		if strings.EqualFold(strings.TrimSpace(c.Name), needle) {
			return colorFromHexOrDefault(c.Color, "252")
		}
	}
	return lipgloss.Color("252")
}

func colorFromHexOrDefault(hex, fallback string) lipgloss.Color {
	normalized := strings.TrimSpace(hex)
	if uiHexColorPattern.MatchString(normalized) {
		return lipgloss.Color(normalized)
	}
	return lipgloss.Color(fallback)
}

func contrastingTextColorFromHexOrDefault(hex, fallback string) lipgloss.Color {
	normalized := strings.TrimSpace(hex)
	if !uiHexColorPattern.MatchString(normalized) {
		return lipgloss.Color(fallback)
	}

	red, _ := strconv.ParseInt(normalized[1:3], 16, 64)
	green, _ := strconv.ParseInt(normalized[3:5], 16, 64)
	blue, _ := strconv.ParseInt(normalized[5:7], 16, 64)
	brightness := (red*299 + green*587 + blue*114) / 1000
	if brightness >= 150 {
		return lipgloss.Color("232")
	}
	return lipgloss.Color("255")
}
