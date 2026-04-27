package repositories

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tiagokriok/kanji/internal/domain"
	"github.com/tiagokriok/kanji/internal/infrastructure/db/sqlc"
	"github.com/tiagokriok/kanji/internal/infrastructure/store"
)

func seedProviderWorkspaceBoardColumn(t *testing.T, ctx context.Context, q *sqlc.Queries) (providerID, workspaceID, boardID, columnID string) {
	t.Helper()
	providerID = "p-task"
	workspaceID = "w-task"
	boardID = "b-task"
	columnID = "c-task"

	if err := q.CreateProvider(ctx, sqlc.CreateProviderParams{
		ID:        providerID,
		Type:      "local",
		Name:      "Test Provider",
		CreatedAt: "2024-01-01T00:00:00Z",
	}); err != nil {
		t.Fatalf("create provider: %v", err)
	}
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         workspaceID,
		ProviderID: providerID,
		Name:       "Test Workspace",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          boardID,
		WorkspaceID: workspaceID,
		Name:        "Test Board",
		ViewDefault: "kanban",
	}); err != nil {
		t.Fatalf("create board: %v", err)
	}
	if err := q.CreateColumn(ctx, sqlc.CreateColumnParams{
		ID:       columnID,
		BoardID:  boardID,
		Name:     "To Do",
		Color:    "#6B7280",
		Position: 1,
	}); err != nil {
		t.Fatalf("create column: %v", err)
	}
	return
}

func TestTaskRepository_Create(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID, workspaceID, boardID, columnID := seedProviderWorkspaceBoardColumn(t, ctx, q)

	repo := NewTaskRepository(store.New(adapter))
	task := domain.Task{
		ID:            "t-create",
		ProviderID:    providerID,
		WorkspaceID:   workspaceID,
		BoardID:       &boardID,
		ColumnID:      &columnID,
		Title:         "Create Test",
		DescriptionMD: "desc",
		Priority:      2,
		Labels:        []string{"a", "b"},
		Position:      1,
		CreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	got, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.ID != task.ID {
		t.Errorf("ID = %q, want %q", got.ID, task.ID)
	}
	if got.Title != task.Title {
		t.Errorf("Title = %q, want %q", got.Title, task.Title)
	}
	if got.DescriptionMD != task.DescriptionMD {
		t.Errorf("DescriptionMD = %q, want %q", got.DescriptionMD, task.DescriptionMD)
	}
	if got.Priority != task.Priority {
		t.Errorf("Priority = %d, want %d", got.Priority, task.Priority)
	}
	if got.ColumnID == nil || *got.ColumnID != columnID {
		t.Errorf("ColumnID = %v, want %q", got.ColumnID, columnID)
	}
	if len(got.Labels) != 2 || got.Labels[0] != "a" || got.Labels[1] != "b" {
		t.Errorf("Labels = %v, want [a b]", got.Labels)
	}
}

func TestTaskRepository_Create_WithEstimateMinutes(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID, workspaceID, boardID, columnID := seedProviderWorkspaceBoardColumn(t, ctx, q)

	repo := NewTaskRepository(store.New(adapter))
	estimate := 120
	task := domain.Task{
		ID:              "t-create-est",
		ProviderID:      providerID,
		WorkspaceID:     workspaceID,
		BoardID:         &boardID,
		ColumnID:        &columnID,
		Title:           "Estimate Test",
		Priority:        0,
		EstimateMinutes: &estimate,
		Labels:          []string{},
		Position:        1,
		CreatedAt:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	got, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.EstimateMinutes == nil || *got.EstimateMinutes != estimate {
		t.Errorf("EstimateMinutes = %v, want %d", got.EstimateMinutes, estimate)
	}
}

func TestTaskRepository_Move(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID, workspaceID, boardID, columnID := seedProviderWorkspaceBoardColumn(t, ctx, q)

	repo := NewTaskRepository(store.New(adapter))
	task := domain.Task{
		ID:          "t-move",
		ProviderID:  providerID,
		WorkspaceID: workspaceID,
		BoardID:     &boardID,
		ColumnID:    &columnID,
		Title:       "Move Test",
		Priority:    0,
		Labels:      []string{},
		Position:    1,
		CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	newColumnID := "c-new"
	if err := q.CreateColumn(ctx, sqlc.CreateColumnParams{
		ID:       newColumnID,
		BoardID:  boardID,
		Name:     "Done",
		Color:    "#00FF00",
		Position: 2,
	}); err != nil {
		t.Fatalf("create column: %v", err)
	}

	newStatus := "done"
	newPosition := float64(99)
	err := repo.Move(ctx, domain.MoveTaskInput{
		TaskID:    task.ID,
		ColumnID:  &newColumnID,
		Status:    &newStatus,
		Position:  newPosition,
		UpdatedAt: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("move task: %v", err)
	}

	got, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("get by id after move: %v", err)
	}
	if got.ColumnID == nil || *got.ColumnID != newColumnID {
		t.Errorf("ColumnID = %v, want %q", got.ColumnID, newColumnID)
	}
	if got.Status == nil || *got.Status != newStatus {
		t.Errorf("Status = %v, want %q", got.Status, newStatus)
	}
	if got.Position != newPosition {
		t.Errorf("Position = %f, want %f", got.Position, newPosition)
	}
}

func TestTaskRepository_Update(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID, workspaceID, boardID, columnID := seedProviderWorkspaceBoardColumn(t, ctx, q)

	repo := NewTaskRepository(store.New(adapter))
	task := domain.Task{
		ID:            "t-update",
		ProviderID:    providerID,
		WorkspaceID:   workspaceID,
		BoardID:       &boardID,
		ColumnID:      &columnID,
		Title:         "Update Test",
		DescriptionMD: "original",
		Priority:      0,
		Labels:        []string{},
		Position:      1,
		CreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	newTitle := "Updated Title"
	newDesc := "Updated Description"
	newPriority := 3
	newLabels := []string{"x"}
	err := repo.Update(ctx, task.ID, domain.TaskPatch{
		Title:         &newTitle,
		DescriptionMD: &newDesc,
		Priority:      &newPriority,
		Labels:        &newLabels,
	})
	if err != nil {
		t.Fatalf("update task: %v", err)
	}

	got, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("get by id after update: %v", err)
	}
	if got.Title != newTitle {
		t.Errorf("Title = %q, want %q", got.Title, newTitle)
	}
	if got.DescriptionMD != newDesc {
		t.Errorf("DescriptionMD = %q, want %q", got.DescriptionMD, newDesc)
	}
	if got.Priority != newPriority {
		t.Errorf("Priority = %d, want %d", got.Priority, newPriority)
	}
	if len(got.Labels) != 1 || got.Labels[0] != "x" {
		t.Errorf("Labels = %v, want [x]", got.Labels)
	}
}

func TestTaskRepository_Create_ErrorContext(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID, workspaceID, boardID, columnID := seedProviderWorkspaceBoardColumn(t, ctx, q)

	repo := NewTaskRepository(store.New(adapter))
	task := domain.Task{
		ID:          "t-dup",
		ProviderID:  providerID,
		WorkspaceID: workspaceID,
		BoardID:     &boardID,
		ColumnID:    &columnID,
		Title:       "Dup Test",
		Priority:    0,
		Labels:      []string{},
		Position:    1,
		CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("first create: %v", err)
	}

	// duplicate ID should fail with wrapped error
	err := repo.Create(ctx, task)
	if err == nil {
		t.Fatal("expected error for duplicate ID, got nil")
	}
	if !strings.Contains(err.Error(), "create task:") {
		t.Errorf("error = %q, want 'create task:' prefix", err.Error())
	}

	// verify only one row exists (tx was rolled back)
	got, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Title != task.Title {
		t.Errorf("Title = %q, want %q", got.Title, task.Title)
	}
}

func TestTaskRepository_Update_ErrorContext(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID, workspaceID, boardID, columnID := seedProviderWorkspaceBoardColumn(t, ctx, q)

	repo := NewTaskRepository(store.New(adapter))
	task := domain.Task{
		ID:          "t-update-err",
		ProviderID:  providerID,
		WorkspaceID: workspaceID,
		BoardID:     &boardID,
		ColumnID:    &columnID,
		Title:       "Update Err Test",
		Priority:    0,
		Labels:      []string{},
		Position:    1,
		CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	// foreign key violation: column does not exist
	badColumn := "bad-column"
	err := repo.Update(ctx, task.ID, domain.TaskPatch{ColumnID: &badColumn})
	if err == nil {
		t.Fatal("expected error for invalid column_id, got nil")
	}
	if !strings.Contains(err.Error(), "update task:") {
		t.Errorf("error = %q, want 'update task:' prefix", err.Error())
	}

	// verify original task unchanged (tx rolled back)
	got, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.ColumnID == nil || *got.ColumnID != columnID {
		t.Errorf("ColumnID = %v, want %q", got.ColumnID, columnID)
	}
}

func TestTaskRepository_Delete(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID, workspaceID, boardID, columnID := seedProviderWorkspaceBoardColumn(t, ctx, q)

	repo := NewTaskRepository(store.New(adapter))
	task := domain.Task{
		ID:          "t-delete",
		ProviderID:  providerID,
		WorkspaceID: workspaceID,
		BoardID:     &boardID,
		ColumnID:    &columnID,
		Title:       "Delete Test",
		Priority:    0,
		Labels:      []string{},
		Position:    1,
		CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	if err := repo.Delete(ctx, task.ID); err != nil {
		t.Fatalf("delete task: %v", err)
	}

	_, err := repo.GetByID(ctx, task.ID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestTaskRepository_List(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID, workspaceID, boardID, columnID := seedProviderWorkspaceBoardColumn(t, ctx, q)

	repo := NewTaskRepository(store.New(adapter))
	for i, title := range []string{"Alpha", "Beta", "Gamma"} {
		task := domain.Task{
			ID:          "t-list-" + title,
			ProviderID:  providerID,
			WorkspaceID: workspaceID,
			BoardID:     &boardID,
			ColumnID:    &columnID,
			Title:       title,
			Priority:    i,
			Labels:      []string{},
			Position:    float64(i + 1),
			CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2024, 1, int(1+i), 0, 0, 0, 0, time.UTC),
		}
		if err := repo.Create(ctx, task); err != nil {
			t.Fatalf("create task %s: %v", title, err)
		}
	}

	got, err := repo.List(ctx, domain.TaskFilter{WorkspaceID: workspaceID})
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len(tasks) = %d, want 3", len(got))
	}
	// Ordered by updated_at DESC
	if got[0].Title != "Gamma" {
		t.Errorf("first title = %q, want Gamma", got[0].Title)
	}
}

func TestTaskRepository_List_FilterBoard(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID, workspaceID, boardID, columnID := seedProviderWorkspaceBoardColumn(t, ctx, q)

	otherBoardID := "b-other"
	if err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          otherBoardID,
		WorkspaceID: workspaceID,
		Name:        "Other Board",
		ViewDefault: "list",
	}); err != nil {
		t.Fatalf("create other board: %v", err)
	}

	repo := NewTaskRepository(store.New(adapter))
	for _, tc := range []struct {
		id      string
		boardID string
		title   string
	}{
		{"t-fb-1", boardID, "In Board"},
		{"t-fb-2", otherBoardID, "Other Board"},
	} {
		task := domain.Task{
			ID:          tc.id,
			ProviderID:  providerID,
			WorkspaceID: workspaceID,
			BoardID:     &tc.boardID,
			ColumnID:    &columnID,
			Title:       tc.title,
			Priority:    0,
			Labels:      []string{},
			Position:    1,
			CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		if err := repo.Create(ctx, task); err != nil {
			t.Fatalf("create task %s: %v", tc.title, err)
		}
	}

	got, err := repo.List(ctx, domain.TaskFilter{WorkspaceID: workspaceID, BoardID: boardID})
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(got))
	}
	if got[0].Title != "In Board" {
		t.Errorf("Title = %q, want In Board", got[0].Title)
	}
}

func TestTaskRepository_List_FilterTitle(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID, workspaceID, boardID, columnID := seedProviderWorkspaceBoardColumn(t, ctx, q)

	repo := NewTaskRepository(store.New(adapter))
	for _, tc := range []struct {
		id    string
		title string
	}{
		{"t-ft-1", "Search Me"},
		{"t-ft-2", "Another"},
	} {
		task := domain.Task{
			ID:          tc.id,
			ProviderID:  providerID,
			WorkspaceID: workspaceID,
			BoardID:     &boardID,
			ColumnID:    &columnID,
			Title:       tc.title,
			Priority:    0,
			Labels:      []string{},
			Position:    1,
			CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		if err := repo.Create(ctx, task); err != nil {
			t.Fatalf("create task %s: %v", tc.title, err)
		}
	}

	got, err := repo.List(ctx, domain.TaskFilter{WorkspaceID: workspaceID, TitleQuery: "search"})
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(got))
	}
	if got[0].Title != "Search Me" {
		t.Errorf("Title = %q, want Search Me", got[0].Title)
	}
}

func TestTaskRepository_List_FilterDueSoon(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID, workspaceID, boardID, columnID := seedProviderWorkspaceBoardColumn(t, ctx, q)

	repo := NewTaskRepository(store.New(adapter))
	soon := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	later := time.Date(2024, 6, 20, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		id    string
		title string
		due   *time.Time
	}{
		{"t-fd-1", "Due Soon", &soon},
		{"t-fd-2", "Later", &later},
		{"t-fd-3", "No Due", nil},
	} {
		task := domain.Task{
			ID:          tc.id,
			ProviderID:  providerID,
			WorkspaceID: workspaceID,
			BoardID:     &boardID,
			ColumnID:    &columnID,
			Title:       tc.title,
			Priority:    0,
			DueAt:       tc.due,
			Labels:      []string{},
			Position:    1,
			CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		if err := repo.Create(ctx, task); err != nil {
			t.Fatalf("create task %s: %v", tc.title, err)
		}
	}

	cutoff := time.Date(2024, 6, 17, 0, 0, 0, 0, time.UTC)
	got, err := repo.List(ctx, domain.TaskFilter{WorkspaceID: workspaceID, DueSoonBy: &cutoff})
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(got))
	}
	if got[0].Title != "Due Soon" {
		t.Errorf("Title = %q, want Due Soon", got[0].Title)
	}
}

func TestTaskRepository_ListBoards(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	_, workspaceID, _, _ := seedProviderWorkspaceBoardColumn(t, ctx, q)

	if err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          "b-list-2",
		WorkspaceID: workspaceID,
		Name:        "Second Board",
		ViewDefault: "list",
	}); err != nil {
		t.Fatalf("create second board: %v", err)
	}

	repo := NewTaskRepository(store.New(adapter))
	got, err := repo.ListBoards(ctx, workspaceID)
	if err != nil {
		t.Fatalf("list boards: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(boards) = %d, want 2", len(got))
	}
}

func TestTaskRepository_ListColumns(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	_, _, boardID, _ := seedProviderWorkspaceBoardColumn(t, ctx, q)

	if err := q.CreateColumn(ctx, sqlc.CreateColumnParams{
		ID:       "c-list-2",
		BoardID:  boardID,
		Name:     "Done",
		Color:    "#00FF00",
		Position: 2,
	}); err != nil {
		t.Fatalf("create second column: %v", err)
	}

	repo := NewTaskRepository(store.New(adapter))
	got, err := repo.ListColumns(ctx, boardID)
	if err != nil {
		t.Fatalf("list columns: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(columns) = %d, want 2", len(got))
	}
}

func TestTaskRepository_GetByID_NotFound(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	repo := NewTaskRepository(store.New(adapter))

	_, err := repo.GetByID(ctx, "missing-task")
	if err == nil {
		t.Fatal("expected error for missing task, got nil")
	}
}

func TestTaskRepository_Delete_NotFound(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	repo := NewTaskRepository(store.New(adapter))

	// Deleting a non-existent task should not error (sqlite exec with no rows)
	if err := repo.Delete(ctx, "missing-task"); err != nil {
		t.Fatalf("delete non-existent task: %v", err)
	}
}

func TestTaskRepository_Move_ErrorContext(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID, workspaceID, boardID, columnID := seedProviderWorkspaceBoardColumn(t, ctx, q)

	repo := NewTaskRepository(store.New(adapter))
	task := domain.Task{
		ID:          "t-move-err",
		ProviderID:  providerID,
		WorkspaceID: workspaceID,
		BoardID:     &boardID,
		ColumnID:    &columnID,
		Title:       "Move Err Test",
		Priority:    0,
		Labels:      []string{},
		Position:    1,
		CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	// foreign key violation: column does not exist
	badColumn := "bad-column"
	err := repo.Move(ctx, domain.MoveTaskInput{
		TaskID:    task.ID,
		ColumnID:  &badColumn,
		Position:  99,
		UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err == nil {
		t.Fatal("expected error for invalid column_id, got nil")
	}
	if !strings.Contains(err.Error(), "move task:") {
		t.Errorf("error = %q, want 'move task:' prefix", err.Error())
	}

	// verify original task unchanged (tx rolled back)
	got, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.ColumnID == nil || *got.ColumnID != columnID {
		t.Errorf("ColumnID = %v, want %q", got.ColumnID, columnID)
	}
}
