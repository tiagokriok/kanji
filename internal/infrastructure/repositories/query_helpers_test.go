package repositories

import (
	"context"
	"testing"

	"github.com/tiagokriok/kanji/internal/infrastructure/db/sqlc"
)

func TestQueryListBoards(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()

	providerID := seedProvider(t, ctx, q)
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         "w-qhb",
		ProviderID: providerID,
		Name:       "Workspace",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          "b-qhb",
		WorkspaceID: "w-qhb",
		Name:        "Board",
		ViewDefault: "kanban",
	}); err != nil {
		t.Fatalf("create board: %v", err)
	}

	boards, err := queryListBoards(ctx, q, "w-qhb")
	if err != nil {
		t.Fatalf("queryListBoards: %v", err)
	}
	if len(boards) != 1 {
		t.Fatalf("len(boards) = %d, want 1", len(boards))
	}
	if boards[0].ID != "b-qhb" {
		t.Errorf("ID = %q, want b-qhb", boards[0].ID)
	}
	if boards[0].Name != "Board" {
		t.Errorf("Name = %q, want Board", boards[0].Name)
	}
}

func TestQueryListColumns(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()

	providerID := seedProvider(t, ctx, q)
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         "w-qhc",
		ProviderID: providerID,
		Name:       "Workspace",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          "b-qhc",
		WorkspaceID: "w-qhc",
		Name:        "Board",
		ViewDefault: "kanban",
	}); err != nil {
		t.Fatalf("create board: %v", err)
	}
	if err := q.CreateColumn(ctx, sqlc.CreateColumnParams{
		ID:       "c-qhc",
		BoardID:  "b-qhc",
		Name:     "To Do",
		Color:    "#6B7280",
		Position: 1,
	}); err != nil {
		t.Fatalf("create column: %v", err)
	}

	cols, err := queryListColumns(ctx, q, "b-qhc")
	if err != nil {
		t.Fatalf("queryListColumns: %v", err)
	}
	if len(cols) != 1 {
		t.Fatalf("len(columns) = %d, want 1", len(cols))
	}
	if cols[0].ID != "c-qhc" {
		t.Errorf("ID = %q, want c-qhc", cols[0].ID)
	}
	if cols[0].Name != "To Do" {
		t.Errorf("Name = %q, want To Do", cols[0].Name)
	}
}
