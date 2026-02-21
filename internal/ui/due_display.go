package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	dueColorToday    = lipgloss.Color("220")
	dueColorOverdue  = lipgloss.Color("203")
	dueColorDefault  = lipgloss.Color("252")
	dueColorNoDueSet = lipgloss.Color("245")
)

func dateOnlyUTC(t time.Time) time.Time {
	v := t.UTC()
	return time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, time.UTC)
}

func (m Model) dueDisplay(dueAt time.Time) (string, lipgloss.Color) {
	today := dateOnlyUTC(time.Now())
	due := dateOnlyUTC(dueAt)
	deltaDays := int(due.Sub(today).Hours() / 24)

	switch {
	case deltaDays == 0:
		return "Today", dueColorToday
	case deltaDays == 1:
		return "Tomorrow", dueColorDefault
	case deltaDays < 0:
		overdueDays := -deltaDays
		if overdueDays <= 7 {
			return fmt.Sprintf("%dd", overdueDays), dueColorOverdue
		}
		return m.formatDueDate(dueAt), dueColorOverdue
	default:
		return m.formatDueDate(dueAt), dueColorDefault
	}
}
