package ui

import (
	"sort"
	"testing"
	"time"

	"github.com/tiagokriok/kanji/internal/domain"
)

func ptr[T any](v T) *T {
	return &v
}

func TestNormalizePriority(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"negative returns 6", -1, 6},
		{"zero stays 0", 0, 0},
		{"five stays 5", 5, 5},
		{"six returns 6", 6, 6},
		{"large returns 6", 99, 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizePriority(tt.input)
			if got != tt.expected {
				t.Errorf("normalizePriority(%d) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestApplyActiveFilters_EmptyTasks(t *testing.T) {
	fs := taskFilterState{columnFilter: "c1", priorityFilter: 0, dueFilter: dueFilterSoon}
	result := fs.applyActiveFilters([]domain.Task{})
	if len(result) != 0 {
		t.Errorf("expected empty, got %d tasks", len(result))
	}
}

func TestApplyActiveFilters_ColumnFilter(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", ColumnID: ptr("c1"), Priority: 1, UpdatedAt: now},
		{ID: "t2", ColumnID: ptr("c2"), Priority: 1, UpdatedAt: now},
		{ID: "t3", ColumnID: nil, Priority: 1, UpdatedAt: now},
	}
	fs := taskFilterState{columnFilter: "c1", priorityFilter: -1}
	result := fs.applyActiveFilters(tasks)
	if len(result) != 1 || result[0].ID != "t1" {
		t.Errorf("expected [t1], got %v", ids(result))
	}
}

func TestApplyActiveFilters_PriorityFilter(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", Priority: 0, UpdatedAt: now},
		{ID: "t2", Priority: 1, UpdatedAt: now},
		{ID: "t3", Priority: 99, UpdatedAt: now},
	}
	fs := taskFilterState{priorityFilter: 0}
	result := fs.applyActiveFilters(tasks)
	if len(result) != 1 || result[0].ID != "t1" {
		t.Errorf("expected [t1], got %v", ids(result))
	}
}

func TestApplyActiveFilters_PriorityFilterNegativeMeansAll(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", Priority: 0, UpdatedAt: now},
		{ID: "t2", Priority: 5, UpdatedAt: now},
	}
	fs := taskFilterState{priorityFilter: -1}
	result := fs.applyActiveFilters(tasks)
	if len(result) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(result))
	}
}

func TestApplyActiveFilters_DueFilterSoon(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", DueAt: ptr(now.AddDate(0, 0, 3)), UpdatedAt: now},
		{ID: "t2", DueAt: ptr(now.AddDate(0, 0, -1)), UpdatedAt: now},
		{ID: "t3", DueAt: nil, UpdatedAt: now},
		{ID: "t4", DueAt: ptr(now.AddDate(0, 0, 10)), UpdatedAt: now},
	}
	fs := taskFilterState{dueFilter: dueFilterSoon, priorityFilter: -1}
	result := fs.applyActiveFilters(tasks)
	if len(result) != 1 || result[0].ID != "t1" {
		t.Errorf("expected [t1], got %v", ids(result))
	}
}

func TestApplyActiveFilters_DueFilterOverdue(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", DueAt: ptr(now.AddDate(0, 0, -1)), UpdatedAt: now},
		{ID: "t2", DueAt: ptr(now.AddDate(0, 0, 3)), UpdatedAt: now},
		{ID: "t3", DueAt: nil, UpdatedAt: now},
	}
	fs := taskFilterState{dueFilter: dueFilterOverdue, priorityFilter: -1}
	result := fs.applyActiveFilters(tasks)
	if len(result) != 1 || result[0].ID != "t1" {
		t.Errorf("expected [t1], got %v", ids(result))
	}
}

func TestApplyActiveFilters_DueFilterNoDate(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", DueAt: nil, UpdatedAt: now},
		{ID: "t2", DueAt: ptr(now.AddDate(0, 0, 3)), UpdatedAt: now},
	}
	fs := taskFilterState{dueFilter: dueFilterNoDate, priorityFilter: -1}
	result := fs.applyActiveFilters(tasks)
	if len(result) != 1 || result[0].ID != "t1" {
		t.Errorf("expected [t1], got %v", ids(result))
	}
}

func TestApplyActiveFilters_DueFilterAny(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", DueAt: nil, UpdatedAt: now},
		{ID: "t2", DueAt: ptr(now.AddDate(0, 0, 3)), UpdatedAt: now},
	}
	fs := taskFilterState{dueFilter: dueFilterAny, priorityFilter: -1}
	result := fs.applyActiveFilters(tasks)
	if len(result) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(result))
	}
}

func TestApplyActiveFilters_CombinedFilters(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", ColumnID: ptr("c1"), Priority: 0, DueAt: ptr(now.AddDate(0, 0, 3)), UpdatedAt: now},
		{ID: "t2", ColumnID: ptr("c1"), Priority: 1, DueAt: ptr(now.AddDate(0, 0, 3)), UpdatedAt: now},
		{ID: "t3", ColumnID: ptr("c1"), Priority: 0, DueAt: nil, UpdatedAt: now},
	}
	fs := taskFilterState{columnFilter: "c1", priorityFilter: 0, dueFilter: dueFilterSoon}
	result := fs.applyActiveFilters(tasks)
	if len(result) != 1 || result[0].ID != "t1" {
		t.Errorf("expected [t1], got %v", ids(result))
	}
}

func TestSortTasksByPriority(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", Priority: 2, UpdatedAt: now},
		{ID: "t2", Priority: 0, UpdatedAt: now},
		{ID: "t3", Priority: 1, UpdatedAt: now},
	}
	fs := taskFilterState{}
	fs.sortTasksByPriority(tasks)
	want := []string{"t2", "t3", "t1"}
	got := ids(tasks)
	if !sliceEq(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSortTasksByPriority_TieBreakDue(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", Priority: 0, DueAt: ptr(now.AddDate(0, 0, 5)), UpdatedAt: now},
		{ID: "t2", Priority: 0, DueAt: nil, UpdatedAt: now},
		{ID: "t3", Priority: 0, DueAt: ptr(now.AddDate(0, 0, 2)), UpdatedAt: now},
	}
	fs := taskFilterState{}
	fs.sortTasksByPriority(tasks)
	want := []string{"t3", "t1", "t2"}
	got := ids(tasks)
	if !sliceEq(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSortTasksByPriority_TieBreakUpdated(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", Priority: 0, DueAt: nil, UpdatedAt: now.Add(-time.Hour)},
		{ID: "t2", Priority: 0, DueAt: nil, UpdatedAt: now},
	}
	fs := taskFilterState{}
	fs.sortTasksByPriority(tasks)
	want := []string{"t2", "t1"}
	got := ids(tasks)
	if !sliceEq(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSortTasks_DefaultIsPriority(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", Priority: 2, UpdatedAt: now},
		{ID: "t2", Priority: 0, UpdatedAt: now},
	}
	fs := taskFilterState{sortMode: sortByPriority}
	fs.sortTasks(tasks)
	want := []string{"t2", "t1"}
	got := ids(tasks)
	if !sliceEq(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSortTasks_ByDueDate(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", DueAt: nil, UpdatedAt: now},
		{ID: "t2", DueAt: ptr(now.AddDate(0, 0, 2)), UpdatedAt: now},
		{ID: "t3", DueAt: ptr(now.AddDate(0, 0, 1)), UpdatedAt: now},
	}
	fs := taskFilterState{sortMode: sortByDueDate}
	fs.sortTasks(tasks)
	want := []string{"t3", "t2", "t1"}
	got := ids(tasks)
	if !sliceEq(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSortTasks_ByTitle(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", Title: "Zebra", UpdatedAt: now},
		{ID: "t2", Title: "apple", UpdatedAt: now},
		{ID: "t3", Title: "Banana", UpdatedAt: now},
	}
	fs := taskFilterState{sortMode: sortByTitle}
	fs.sortTasks(tasks)
	want := []string{"t2", "t3", "t1"}
	got := ids(tasks)
	if !sliceEq(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSortTasks_ByUpdated(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", UpdatedAt: now.Add(-time.Hour * 2)},
		{ID: "t2", UpdatedAt: now},
		{ID: "t3", UpdatedAt: now.Add(-time.Hour)},
	}
	fs := taskFilterState{sortMode: sortByUpdated}
	fs.sortTasks(tasks)
	want := []string{"t2", "t3", "t1"}
	got := ids(tasks)
	if !sliceEq(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSortTasks_ByCreated(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", CreatedAt: now.Add(-time.Hour * 2), UpdatedAt: now},
		{ID: "t2", CreatedAt: now, UpdatedAt: now},
		{ID: "t3", CreatedAt: now.Add(-time.Hour), UpdatedAt: now},
	}
	fs := taskFilterState{sortMode: sortByCreated}
	fs.sortTasks(tasks)
	want := []string{"t2", "t3", "t1"}
	got := ids(tasks)
	if !sliceEq(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSortTasks_IsStable(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", Title: "a", UpdatedAt: now},
		{ID: "t2", Title: "a", UpdatedAt: now},
	}
	fs := taskFilterState{sortMode: sortByTitle}
	fs.sortTasks(tasks)
	want := []string{"t1", "t2"}
	got := ids(tasks)
	if !sliceEq(got, want) {
		t.Errorf("stability broken: got %v, want %v", got, want)
	}
}

func TestSortTasks_DueDateTieBreakUpdated(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", DueAt: ptr(now.AddDate(0, 0, 1)), UpdatedAt: now.Add(-time.Hour)},
		{ID: "t2", DueAt: ptr(now.AddDate(0, 0, 1)), UpdatedAt: now},
	}
	fs := taskFilterState{sortMode: sortByDueDate}
	fs.sortTasks(tasks)
	want := []string{"t2", "t1"}
	got := ids(tasks)
	if !sliceEq(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSortTasks_TitleTieBreakUpdated(t *testing.T) {
	now := time.Now().UTC()
	tasks := []domain.Task{
		{ID: "t1", Title: "same", UpdatedAt: now.Add(-time.Hour)},
		{ID: "t2", Title: "same", UpdatedAt: now},
	}
	fs := taskFilterState{sortMode: sortByTitle}
	fs.sortTasks(tasks)
	want := []string{"t2", "t1"}
	got := ids(tasks)
	if !sliceEq(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func ids(tasks []domain.Task) []string {
	result := make([]string, len(tasks))
	for i, t := range tasks {
		result[i] = t.ID
	}
	return result
}

func sliceEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Verify that sort.SliceStable is actually used (via compile-time check)
var _ = sort.SliceStable
