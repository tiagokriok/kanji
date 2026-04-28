package ui

import (
	"sort"
	"strings"
	"time"

	"github.com/tiagokriok/kanji/internal/domain"
)

// taskFilterState holds the filter and sort configuration for tasks.
type taskFilterState struct {
	columnFilter   string
	priorityFilter int
	dueFilter      dueFilterMode
	sortMode       taskSortMode
}

// applyActiveFilters returns a slice of tasks matching the current filter state.
// Filters are combined with AND logic: column, priority, and due date.
func (fs taskFilterState) applyActiveFilters(tasks []domain.Task) []domain.Task {
	if len(tasks) == 0 {
		return tasks
	}
	now := time.Now().UTC()
	soonLimit := now.AddDate(0, 0, 7)
	filtered := make([]domain.Task, 0, len(tasks))
	for _, task := range tasks {
		if fs.columnFilter != "" {
			if task.ColumnID == nil || *task.ColumnID != fs.columnFilter {
				continue
			}
		}
		if fs.priorityFilter >= 0 && normalizePriority(task.Priority) != fs.priorityFilter {
			continue
		}
		switch fs.dueFilter {
		case dueFilterSoon:
			if task.DueAt == nil {
				continue
			}
			due := task.DueAt.UTC()
			if due.Before(now) || due.After(soonLimit) {
				continue
			}
		case dueFilterOverdue:
			if task.DueAt == nil || !task.DueAt.UTC().Before(now) {
				continue
			}
		case dueFilterNoDate:
			if task.DueAt != nil {
				continue
			}
		}
		filtered = append(filtered, task)
	}
	return filtered
}

// sortTasks sorts tasks in place according to the current sort mode.
func (fs taskFilterState) sortTasks(tasks []domain.Task) {
	switch fs.sortMode {
	case sortByDueDate:
		sort.SliceStable(tasks, func(i, j int) bool {
			if tasks[i].DueAt != nil && tasks[j].DueAt == nil {
				return true
			}
			if tasks[i].DueAt == nil && tasks[j].DueAt != nil {
				return false
			}
			if tasks[i].DueAt != nil && tasks[j].DueAt != nil && !tasks[i].DueAt.Equal(*tasks[j].DueAt) {
				return tasks[i].DueAt.Before(*tasks[j].DueAt)
			}
			return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
		})
	case sortByTitle:
		sort.SliceStable(tasks, func(i, j int) bool {
			ti := strings.ToLower(strings.TrimSpace(tasks[i].Title))
			tj := strings.ToLower(strings.TrimSpace(tasks[j].Title))
			if ti != tj {
				return ti < tj
			}
			return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
		})
	case sortByUpdated:
		sort.SliceStable(tasks, func(i, j int) bool {
			return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
		})
	case sortByCreated:
		sort.SliceStable(tasks, func(i, j int) bool {
			return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
		})
	default:
		fs.sortTasksByPriority(tasks)
	}
}

// sortTasksByPriority sorts tasks by normalized priority ascending,
// with ties broken by due date (sooner first) then updated time (newer first).
func (fs taskFilterState) sortTasksByPriority(tasks []domain.Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		pi := normalizePriority(tasks[i].Priority)
		pj := normalizePriority(tasks[j].Priority)
		if pi != pj {
			return pi < pj
		}
		if tasks[i].DueAt != nil && tasks[j].DueAt == nil {
			return true
		}
		if tasks[i].DueAt == nil && tasks[j].DueAt != nil {
			return false
		}
		if tasks[i].DueAt != nil && tasks[j].DueAt != nil && !tasks[i].DueAt.Equal(*tasks[j].DueAt) {
			return tasks[i].DueAt.Before(*tasks[j].DueAt)
		}
		return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
	})
}

// normalizePriority maps out-of-range priorities to 6, leaving 0..5 unchanged.
func normalizePriority(priority int) int {
	if priority < 0 || priority > 5 {
		return 6
	}
	return priority
}
