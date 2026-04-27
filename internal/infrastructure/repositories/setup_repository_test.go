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

func seedProvider(t *testing.T, ctx context.Context, q *sqlc.Queries) string {
	t.Helper()
	providerID := "p-setup"
	if err := q.CreateProvider(ctx, sqlc.CreateProviderParams{
		ID:        providerID,
		Type:      "local",
		Name:      "Test Provider",
		CreatedAt: "2024-01-01T00:00:00Z",
	}); err != nil {
		t.Fatalf("create provider: %v", err)
	}
	return providerID
}

func TestSetupRepository_CreateProvider(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	repo := NewSetupRepository(store.New(adapter))

	provider := domain.Provider{
		ID:        "p-create",
		Type:      "local",
		Name:      "Create Test",
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := repo.CreateProvider(ctx, provider); err != nil {
		t.Fatalf("create provider: %v", err)
	}

	got, err := repo.ListProviders(ctx)
	if err != nil {
		t.Fatalf("list providers: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(providers) = %d, want 1", len(got))
	}
	if got[0].ID != provider.ID {
		t.Errorf("ID = %q, want %q", got[0].ID, provider.ID)
	}
	if got[0].Name != provider.Name {
		t.Errorf("Name = %q, want %q", got[0].Name, provider.Name)
	}
}

func TestSetupRepository_CreateProvider_ErrorContext(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	repo := NewSetupRepository(store.New(adapter))

	provider := domain.Provider{
		ID:        "p-dup",
		Type:      "local",
		Name:      "Dup Test",
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := repo.CreateProvider(ctx, provider); err != nil {
		t.Fatalf("first create: %v", err)
	}

	err := repo.CreateProvider(ctx, provider)
	if err == nil {
		t.Fatal("expected error for duplicate ID, got nil")
	}
	if !strings.Contains(err.Error(), "create provider:") {
		t.Errorf("error = %q, want 'create provider:' prefix", err.Error())
	}

	got, err := repo.ListProviders(ctx)
	if err != nil {
		t.Fatalf("list providers: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len(providers) = %d, want 1", len(got))
	}
}

func TestSetupRepository_ListProviders(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	seedProvider(t, ctx, q)

	repo := NewSetupRepository(store.New(adapter))
	got, err := repo.ListProviders(ctx)
	if err != nil {
		t.Fatalf("list providers: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(providers) = %d, want 1", len(got))
	}
	if got[0].ID != "p-setup" {
		t.Errorf("ID = %q, want %q", got[0].ID, "p-setup")
	}
}

func TestSetupRepository_CreateWorkspace(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID := seedProvider(t, ctx, q)

	repo := NewSetupRepository(store.New(adapter))
	workspace := domain.Workspace{
		ID:         "w-create",
		ProviderID: providerID,
		Name:       "Create Workspace",
	}
	if err := repo.CreateWorkspace(ctx, workspace); err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	got, err := repo.ListWorkspaces(ctx)
	if err != nil {
		t.Fatalf("list workspaces: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(workspaces) = %d, want 1", len(got))
	}
	if got[0].ID != workspace.ID {
		t.Errorf("ID = %q, want %q", got[0].ID, workspace.ID)
	}
}

func TestSetupRepository_CreateWorkspace_ErrorContext(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID := seedProvider(t, ctx, q)

	repo := NewSetupRepository(store.New(adapter))
	workspace := domain.Workspace{
		ID:         "w-dup",
		ProviderID: providerID,
		Name:       "Dup Workspace",
	}
	if err := repo.CreateWorkspace(ctx, workspace); err != nil {
		t.Fatalf("first create: %v", err)
	}

	err := repo.CreateWorkspace(ctx, workspace)
	if err == nil {
		t.Fatal("expected error for duplicate ID, got nil")
	}
	if !strings.Contains(err.Error(), "create workspace:") {
		t.Errorf("error = %q, want 'create workspace:' prefix", err.Error())
	}

	got, err := repo.ListWorkspaces(ctx)
	if err != nil {
		t.Fatalf("list workspaces: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len(workspaces) = %d, want 1", len(got))
	}
}

func TestSetupRepository_ListWorkspaces(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID := seedProvider(t, ctx, q)
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         "w-list",
		ProviderID: providerID,
		Name:       "List Workspace",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	repo := NewSetupRepository(store.New(adapter))
	got, err := repo.ListWorkspaces(ctx)
	if err != nil {
		t.Fatalf("list workspaces: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(workspaces) = %d, want 1", len(got))
	}
	if got[0].ID != "w-list" {
		t.Errorf("ID = %q, want %q", got[0].ID, "w-list")
	}
}

func TestSetupRepository_RenameWorkspace(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID := seedProvider(t, ctx, q)
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         "w-rename",
		ProviderID: providerID,
		Name:       "Old Name",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	repo := NewSetupRepository(store.New(adapter))
	if err := repo.RenameWorkspace(ctx, "w-rename", "New Name"); err != nil {
		t.Fatalf("rename workspace: %v", err)
	}

	got, err := repo.ListWorkspaces(ctx)
	if err != nil {
		t.Fatalf("list workspaces: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(workspaces) = %d, want 1", len(got))
	}
	if got[0].Name != "New Name" {
		t.Errorf("Name = %q, want %q", got[0].Name, "New Name")
	}
}

func TestSetupRepository_RenameWorkspace_Validation(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	repo := NewSetupRepository(store.New(adapter))

	if err := repo.RenameWorkspace(ctx, "", "Name"); err == nil {
		t.Error("expected error for empty workspace id, got nil")
	}
	if err := repo.RenameWorkspace(ctx, "id", ""); err == nil {
		t.Error("expected error for empty name, got nil")
	}
}

func TestSetupRepository_RenameWorkspace_ErrorContext(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID := seedProvider(t, ctx, q)
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         "w-rename-err",
		ProviderID: providerID,
		Name:       "Old Name",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	repo := NewSetupRepository(store.New(adapter))
	if err := repo.RenameWorkspace(ctx, "w-rename-err", "New Name"); err != nil {
		t.Fatalf("rename workspace: %v", err)
	}

	got, err := repo.ListWorkspaces(ctx)
	if err != nil {
		t.Fatalf("list workspaces: %v", err)
	}
	if len(got) != 1 || got[0].Name != "New Name" {
		t.Errorf("Name = %q, want %q", got[0].Name, "New Name")
	}
}

func TestSetupRepository_CreateBoard(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID := seedProvider(t, ctx, q)
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         "w-board",
		ProviderID: providerID,
		Name:       "Workspace",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	repo := NewSetupRepository(store.New(adapter))
	board := domain.Board{
		ID:          "b-create",
		WorkspaceID: "w-board",
		Name:        "Create Board",
		ViewDefault: "kanban",
	}
	if err := repo.CreateBoard(ctx, board); err != nil {
		t.Fatalf("create board: %v", err)
	}

	got, err := repo.ListBoards(ctx, "w-board")
	if err != nil {
		t.Fatalf("list boards: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(boards) = %d, want 1", len(got))
	}
	if got[0].ID != board.ID {
		t.Errorf("ID = %q, want %q", got[0].ID, board.ID)
	}
	if got[0].ViewDefault != board.ViewDefault {
		t.Errorf("ViewDefault = %q, want %q", got[0].ViewDefault, board.ViewDefault)
	}
}

func TestSetupRepository_CreateBoard_ErrorContext(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID := seedProvider(t, ctx, q)
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         "w-board-dup",
		ProviderID: providerID,
		Name:       "Workspace",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	repo := NewSetupRepository(store.New(adapter))
	board := domain.Board{
		ID:          "b-dup",
		WorkspaceID: "w-board-dup",
		Name:        "Dup Board",
		ViewDefault: "list",
	}
	if err := repo.CreateBoard(ctx, board); err != nil {
		t.Fatalf("first create: %v", err)
	}

	err := repo.CreateBoard(ctx, board)
	if err == nil {
		t.Fatal("expected error for duplicate ID, got nil")
	}
	if !strings.Contains(err.Error(), "create board:") {
		t.Errorf("error = %q, want 'create board:' prefix", err.Error())
	}

	got, err := repo.ListBoards(ctx, "w-board-dup")
	if err != nil {
		t.Fatalf("list boards: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len(boards) = %d, want 1", len(got))
	}
}

func TestSetupRepository_ListBoards(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID := seedProvider(t, ctx, q)
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         "w-list-boards",
		ProviderID: providerID,
		Name:       "Workspace",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          "b-list",
		WorkspaceID: "w-list-boards",
		Name:        "List Board",
		ViewDefault: "list",
	}); err != nil {
		t.Fatalf("create board: %v", err)
	}

	repo := NewSetupRepository(store.New(adapter))
	got, err := repo.ListBoards(ctx, "w-list-boards")
	if err != nil {
		t.Fatalf("list boards: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(boards) = %d, want 1", len(got))
	}
	if got[0].ID != "b-list" {
		t.Errorf("ID = %q, want %q", got[0].ID, "b-list")
	}
}

func TestSetupRepository_RenameBoard(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID := seedProvider(t, ctx, q)
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         "w-rename-board",
		ProviderID: providerID,
		Name:       "Workspace",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          "b-rename",
		WorkspaceID: "w-rename-board",
		Name:        "Old Name",
		ViewDefault: "list",
	}); err != nil {
		t.Fatalf("create board: %v", err)
	}

	repo := NewSetupRepository(store.New(adapter))
	if err := repo.RenameBoard(ctx, "b-rename", "New Name"); err != nil {
		t.Fatalf("rename board: %v", err)
	}

	got, err := repo.ListBoards(ctx, "w-rename-board")
	if err != nil {
		t.Fatalf("list boards: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(boards) = %d, want 1", len(got))
	}
	if got[0].Name != "New Name" {
		t.Errorf("Name = %q, want %q", got[0].Name, "New Name")
	}
}

func TestSetupRepository_RenameBoard_Validation(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	repo := NewSetupRepository(store.New(adapter))

	if err := repo.RenameBoard(ctx, "", "Name"); err == nil {
		t.Error("expected error for empty board id, got nil")
	}
	if err := repo.RenameBoard(ctx, "id", ""); err == nil {
		t.Error("expected error for empty name, got nil")
	}
}

func TestSetupRepository_RenameBoard_ErrorContext(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID := seedProvider(t, ctx, q)
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         "w-rename-board-err",
		ProviderID: providerID,
		Name:       "Workspace",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          "b-rename-err",
		WorkspaceID: "w-rename-board-err",
		Name:        "Old Name",
		ViewDefault: "list",
	}); err != nil {
		t.Fatalf("create board: %v", err)
	}

	repo := NewSetupRepository(store.New(adapter))
	if err := repo.RenameBoard(ctx, "b-rename-err", "New Name"); err != nil {
		t.Fatalf("rename board: %v", err)
	}

	got, err := repo.ListBoards(ctx, "w-rename-board-err")
	if err != nil {
		t.Fatalf("list boards: %v", err)
	}
	if len(got) != 1 || got[0].Name != "New Name" {
		t.Errorf("Name = %q, want %q", got[0].Name, "New Name")
	}
}

func TestSetupRepository_CreateColumn(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID := seedProvider(t, ctx, q)
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         "w-col",
		ProviderID: providerID,
		Name:       "Workspace",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          "b-col",
		WorkspaceID: "w-col",
		Name:        "Board",
		ViewDefault: "kanban",
	}); err != nil {
		t.Fatalf("create board: %v", err)
	}

	repo := NewSetupRepository(store.New(adapter))
	column := domain.Column{
		ID:       "c-create",
		BoardID:  "b-col",
		Name:     "To Do",
		Color:    "#FF0000",
		Position: 1,
	}
	if err := repo.CreateColumn(ctx, column); err != nil {
		t.Fatalf("create column: %v", err)
	}

	got, err := repo.ListColumns(ctx, "b-col")
	if err != nil {
		t.Fatalf("list columns: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(columns) = %d, want 1", len(got))
	}
	if got[0].ID != column.ID {
		t.Errorf("ID = %q, want %q", got[0].ID, column.ID)
	}
	if got[0].Color != "#FF0000" {
		t.Errorf("Color = %q, want %q", got[0].Color, "#FF0000")
	}
}

func TestSetupRepository_CreateColumn_ErrorContext(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID := seedProvider(t, ctx, q)
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         "w-col-dup",
		ProviderID: providerID,
		Name:       "Workspace",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          "b-col-dup",
		WorkspaceID: "w-col-dup",
		Name:        "Board",
		ViewDefault: "kanban",
	}); err != nil {
		t.Fatalf("create board: %v", err)
	}

	repo := NewSetupRepository(store.New(adapter))
	column := domain.Column{
		ID:       "c-dup",
		BoardID:  "b-col-dup",
		Name:     "Dup Column",
		Color:    "#00FF00",
		Position: 1,
	}
	if err := repo.CreateColumn(ctx, column); err != nil {
		t.Fatalf("first create: %v", err)
	}

	err := repo.CreateColumn(ctx, column)
	if err == nil {
		t.Fatal("expected error for duplicate ID, got nil")
	}
	if !strings.Contains(err.Error(), "create column:") {
		t.Errorf("error = %q, want 'create column:' prefix", err.Error())
	}

	got, err := repo.ListColumns(ctx, "b-col-dup")
	if err != nil {
		t.Fatalf("list columns: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len(columns) = %d, want 1", len(got))
	}
}

func TestSetupRepository_ListColumns(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID := seedProvider(t, ctx, q)
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         "w-list-cols",
		ProviderID: providerID,
		Name:       "Workspace",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          "b-list-cols",
		WorkspaceID: "w-list-cols",
		Name:        "Board",
		ViewDefault: "kanban",
	}); err != nil {
		t.Fatalf("create board: %v", err)
	}
	if err := q.CreateColumn(ctx, sqlc.CreateColumnParams{
		ID:       "c-list",
		BoardID:  "b-list-cols",
		Name:     "Listed",
		Color:    "#6B7280",
		Position: 1,
	}); err != nil {
		t.Fatalf("create column: %v", err)
	}

	repo := NewSetupRepository(store.New(adapter))
	got, err := repo.ListColumns(ctx, "b-list-cols")
	if err != nil {
		t.Fatalf("list columns: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(columns) = %d, want 1", len(got))
	}
	if got[0].ID != "c-list" {
		t.Errorf("ID = %q, want %q", got[0].ID, "c-list")
	}
}

func TestSetupRepository_ReorderColumns(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID := seedProvider(t, ctx, q)
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         "w-reorder",
		ProviderID: providerID,
		Name:       "Workspace",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          "b-reorder",
		WorkspaceID: "w-reorder",
		Name:        "Board",
		ViewDefault: "kanban",
	}); err != nil {
		t.Fatalf("create board: %v", err)
	}
	for _, c := range []struct {
		id       string
		position int64
	}{
		{"c-reorder-1", 1},
		{"c-reorder-2", 2},
		{"c-reorder-3", 3},
	} {
		if err := q.CreateColumn(ctx, sqlc.CreateColumnParams{
			ID:       c.id,
			BoardID:  "b-reorder",
			Name:     c.id,
			Color:    "#6B7280",
			Position: c.position,
		}); err != nil {
			t.Fatalf("create column %s: %v", c.id, err)
		}
	}

	repo := NewSetupRepository(store.New(adapter))
	if err := repo.ReorderColumns(ctx, "b-reorder", []string{"c-reorder-3", "c-reorder-1", "c-reorder-2"}); err != nil {
		t.Fatalf("reorder columns: %v", err)
	}

	got, err := repo.ListColumns(ctx, "b-reorder")
	if err != nil {
		t.Fatalf("list columns: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len(columns) = %d, want 3", len(got))
	}
	// ListColumns order is by position
	if got[0].ID != "c-reorder-3" || got[0].Position != 1 {
		t.Errorf("first: ID=%q position=%d, want c-reorder-3/1", got[0].ID, got[0].Position)
	}
	if got[1].ID != "c-reorder-1" || got[1].Position != 2 {
		t.Errorf("second: ID=%q position=%d, want c-reorder-1/2", got[1].ID, got[1].Position)
	}
	if got[2].ID != "c-reorder-2" || got[2].Position != 3 {
		t.Errorf("third: ID=%q position=%d, want c-reorder-2/3", got[2].ID, got[2].Position)
	}
}

func TestSetupRepository_ReorderColumns_Validation(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	repo := NewSetupRepository(store.New(adapter))

	if err := repo.ReorderColumns(ctx, "", []string{"c1"}); err == nil {
		t.Error("expected error for empty board id, got nil")
	}
	if err := repo.ReorderColumns(ctx, "b1", []string{}); err == nil {
		t.Error("expected error for empty column ids, got nil")
	}
}

func TestSetupRepository_ReorderColumns_ErrorContext(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID := seedProvider(t, ctx, q)
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         "w-reorder-err-ctx",
		ProviderID: providerID,
		Name:       "Workspace",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          "b-reorder-err-ctx",
		WorkspaceID: "w-reorder-err-ctx",
		Name:        "Board",
		ViewDefault: "kanban",
	}); err != nil {
		t.Fatalf("create board: %v", err)
	}
	for _, c := range []struct {
		id       string
		position int64
	}{
		{"c-reorder-err-1", 1},
		{"c-reorder-err-2", 2},
	} {
		if err := q.CreateColumn(ctx, sqlc.CreateColumnParams{
			ID:       c.id,
			BoardID:  "b-reorder-err-ctx",
			Name:     c.id,
			Color:    "#6B7280",
			Position: c.position,
		}); err != nil {
			t.Fatalf("create column %s: %v", c.id, err)
		}
	}

	repo := NewSetupRepository(store.New(adapter))
	if err := repo.ReorderColumns(ctx, "b-reorder-err-ctx", []string{"c-reorder-err-2", "c-reorder-err-1"}); err != nil {
		t.Fatalf("reorder columns: %v", err)
	}

	got, err := repo.ListColumns(ctx, "b-reorder-err-ctx")
	if err != nil {
		t.Fatalf("list columns: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(columns) = %d, want 2", len(got))
	}
	if got[0].ID != "c-reorder-err-2" || got[0].Position != 1 {
		t.Errorf("first: ID=%q position=%d, want c-reorder-err-2/1", got[0].ID, got[0].Position)
	}
	if got[1].ID != "c-reorder-err-1" || got[1].Position != 2 {
		t.Errorf("second: ID=%q position=%d, want c-reorder-err-1/2", got[1].ID, got[1].Position)
	}
}

func TestSetupRepository_RenameWorkspace_NotFound(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	repo := NewSetupRepository(store.New(adapter))

	// Renaming a non-existent workspace returns nil because UpdateWorkspaceName uses :exec
	if err := repo.RenameWorkspace(ctx, "missing-ws", "New Name"); err != nil {
		t.Fatalf("rename non-existent workspace: %v", err)
	}
}

func TestSetupRepository_RenameBoard_NotFound(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	repo := NewSetupRepository(store.New(adapter))

	// Renaming a non-existent board returns nil because UpdateBoardName uses :exec
	if err := repo.RenameBoard(ctx, "missing-board", "New Name"); err != nil {
		t.Fatalf("rename non-existent board: %v", err)
	}
}

func TestSetupRepository_ReorderColumns_EmptyIDInSlice(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID := seedProvider(t, ctx, q)
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         "w-reorder-empty",
		ProviderID: providerID,
		Name:       "Workspace",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          "b-reorder-empty",
		WorkspaceID: "w-reorder-empty",
		Name:        "Board",
		ViewDefault: "kanban",
	}); err != nil {
		t.Fatalf("create board: %v", err)
	}
	if err := q.CreateColumn(ctx, sqlc.CreateColumnParams{
		ID:       "c-reorder-empty",
		BoardID:  "b-reorder-empty",
		Name:     "Column",
		Color:    "#6B7280",
		Position: 1,
	}); err != nil {
		t.Fatalf("create column: %v", err)
	}

	repo := NewSetupRepository(store.New(adapter))
	err := repo.ReorderColumns(ctx, "b-reorder-empty", []string{"c-reorder-empty", "", "c-reorder-empty"})
	if err == nil {
		t.Fatal("expected error for empty column id in slice, got nil")
	}
	if !strings.Contains(err.Error(), "column id at position 2 is empty") {
		t.Errorf("error = %q, want 'column id at position 2 is empty'", err.Error())
	}
}

func TestSetupRepository_ReorderColumns_NotFound(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID := seedProvider(t, ctx, q)
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         "w-reorder-err",
		ProviderID: providerID,
		Name:       "Workspace",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          "b-reorder-err",
		WorkspaceID: "w-reorder-err",
		Name:        "Board",
		ViewDefault: "kanban",
	}); err != nil {
		t.Fatalf("create board: %v", err)
	}
	if err := q.CreateColumn(ctx, sqlc.CreateColumnParams{
		ID:       "c-reorder-err",
		BoardID:  "b-reorder-err",
		Name:     "Column",
		Color:    "#6B7280",
		Position: 1,
	}); err != nil {
		t.Fatalf("create column: %v", err)
	}

	repo := NewSetupRepository(store.New(adapter))
	err := repo.ReorderColumns(ctx, "b-reorder-err", []string{"c-reorder-err", "missing"})
	if err == nil {
		t.Fatal("expected error for missing column, got nil")
	}
	if !strings.Contains(err.Error(), "missing not found in board") {
		t.Errorf("error = %q, want 'missing not found in board'", err.Error())
	}
}
