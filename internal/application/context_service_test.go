package application

import (
	"context"
	"errors"
	"testing"

	"github.com/tiagokriok/kanji/internal/domain"
)

type fakeSetupRepo struct {
	workspaces        []domain.Workspace
	boards            []domain.Board
	columns           []domain.Column
	createColumnErr   error
	reorderColumnsErr error
}

func (r *fakeSetupRepo) ListProviders(ctx context.Context) ([]domain.Provider, error) {
	return nil, nil
}
func (r *fakeSetupRepo) CreateProvider(ctx context.Context, provider domain.Provider) error {
	return nil
}
func (r *fakeSetupRepo) ListWorkspaces(ctx context.Context) ([]domain.Workspace, error) {
	return r.workspaces, nil
}
func (r *fakeSetupRepo) CreateWorkspace(ctx context.Context, workspace domain.Workspace) error {
	return nil
}
func (r *fakeSetupRepo) RenameWorkspace(ctx context.Context, workspaceID, name string) error {
	return nil
}
func (r *fakeSetupRepo) ListBoards(ctx context.Context, workspaceID string) ([]domain.Board, error) {
	return r.boards, nil
}
func (r *fakeSetupRepo) CreateBoard(ctx context.Context, board domain.Board) error {
	return nil
}
func (r *fakeSetupRepo) RenameBoard(ctx context.Context, boardID, name string) error {
	return nil
}
func (r *fakeSetupRepo) ListColumns(ctx context.Context, boardID string) ([]domain.Column, error) {
	return r.columns, nil
}
func (r *fakeSetupRepo) CreateColumn(ctx context.Context, column domain.Column) error {
	if r.createColumnErr != nil {
		return r.createColumnErr
	}
	r.columns = append(r.columns, column)
	return nil
}
func (r *fakeSetupRepo) ReorderColumns(ctx context.Context, boardID string, orderedColumnIDs []string) error {
	if r.reorderColumnsErr != nil {
		return r.reorderColumnsErr
	}
	posMap := make(map[string]int, len(orderedColumnIDs))
	for i, id := range orderedColumnIDs {
		posMap[id] = i + 1
	}
	for i := range r.columns {
		if p, ok := posMap[r.columns[i].ID]; ok {
			r.columns[i].Position = p
		}
	}
	return nil
}

func TestContextService_CreateColumn_Success(t *testing.T) {
	repo := &fakeSetupRepo{}
	svc := NewContextService(repo)

	col, err := svc.CreateColumn(context.Background(), "board-1", "Review", "#FF0000", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.BoardID != "board-1" {
		t.Errorf("BoardID = %q, want %q", col.BoardID, "board-1")
	}
	if col.Name != "Review" {
		t.Errorf("Name = %q, want %q", col.Name, "Review")
	}
	if col.Color != "#FF0000" {
		t.Errorf("Color = %q, want %q", col.Color, "#FF0000")
	}
	if col.Position != 1 {
		t.Errorf("Position = %d, want %d", col.Position, 1)
	}
	if col.WIPLimit != nil {
		t.Errorf("WIPLimit = %v, want nil", col.WIPLimit)
	}
}

func TestContextService_CreateColumn_DefaultColor(t *testing.T) {
	repo := &fakeSetupRepo{
		columns: []domain.Column{
			{ID: "c1", BoardID: "board-1", Name: "Todo", Color: "#60A5FA", Position: 1},
			{ID: "c2", BoardID: "board-1", Name: "Doing", Color: "#F59E0B", Position: 2},
		},
	}
	svc := NewContextService(repo)

	col, err := svc.CreateColumn(context.Background(), "board-1", "Review", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.Color != "#22C55E" {
		t.Errorf("Color = %q, want %q", col.Color, "#22C55E")
	}
	if col.Position != 3 {
		t.Errorf("Position = %d, want %d", col.Position, 3)
	}
}

func TestContextService_CreateColumn_WithWIPLimit(t *testing.T) {
	repo := &fakeSetupRepo{}
	svc := NewContextService(repo)

	wip := 5
	col, err := svc.CreateColumn(context.Background(), "board-1", "Review", "#FF0000", &wip)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.WIPLimit == nil || *col.WIPLimit != 5 {
		t.Errorf("WIPLimit = %v, want 5", col.WIPLimit)
	}
}

func TestContextService_CreateColumn_Validation(t *testing.T) {
	repo := &fakeSetupRepo{}
	svc := NewContextService(repo)

	_, err := svc.CreateColumn(context.Background(), "", "Review", "#FF0000", nil)
	if err == nil {
		t.Fatal("expected error for empty board id, got nil")
	}

	_, err = svc.CreateColumn(context.Background(), "board-1", "", "#FF0000", nil)
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestContextService_CreateColumn_InvalidColor(t *testing.T) {
	repo := &fakeSetupRepo{}
	svc := NewContextService(repo)

	_, err := svc.CreateColumn(context.Background(), "board-1", "Review", "red", nil)
	if err == nil {
		t.Fatal("expected error for invalid color, got nil")
	}
}

func TestContextService_CreateColumn_PropagatesRepoError(t *testing.T) {
	repo := &fakeSetupRepo{createColumnErr: errors.New("db down")}
	svc := NewContextService(repo)

	_, err := svc.CreateColumn(context.Background(), "board-1", "Review", "#FF0000", nil)
	if err == nil || err.Error() != "db down" {
		t.Fatalf("expected 'db down' error, got %v", err)
	}
}

func TestContextService_ReorderColumns_Success(t *testing.T) {
	repo := &fakeSetupRepo{
		columns: []domain.Column{
			{ID: "c1", BoardID: "board-1", Name: "Todo", Position: 1},
			{ID: "c2", BoardID: "board-1", Name: "Doing", Position: 2},
			{ID: "c3", BoardID: "board-1", Name: "Done", Position: 3},
		},
	}
	svc := NewContextService(repo)

	err := svc.ReorderColumns(context.Background(), "board-1", []string{"c3", "c1", "c2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.columns[0].Position != 2 {
		t.Errorf("c1 position = %d, want 2", repo.columns[0].Position)
	}
	if repo.columns[1].Position != 3 {
		t.Errorf("c2 position = %d, want 3", repo.columns[1].Position)
	}
	if repo.columns[2].Position != 1 {
		t.Errorf("c3 position = %d, want 1", repo.columns[2].Position)
	}
}

func TestContextService_ReorderColumns_MissingColumn(t *testing.T) {
	repo := &fakeSetupRepo{
		columns: []domain.Column{
			{ID: "c1", BoardID: "board-1", Name: "Todo", Position: 1},
			{ID: "c2", BoardID: "board-1", Name: "Doing", Position: 2},
			{ID: "c3", BoardID: "board-1", Name: "Done", Position: 3},
		},
	}
	svc := NewContextService(repo)

	err := svc.ReorderColumns(context.Background(), "board-1", []string{"c3", "c1"})
	if err == nil {
		t.Fatal("expected error for missing column, got nil")
	}
	if err.Error() != "column c2 not included in reorder" {
		t.Errorf("error = %q, want %q", err.Error(), "column c2 not included in reorder")
	}
}

func TestContextService_ReorderColumns_ExtraColumn(t *testing.T) {
	repo := &fakeSetupRepo{
		columns: []domain.Column{
			{ID: "c1", BoardID: "board-1", Name: "Todo", Position: 1},
			{ID: "c2", BoardID: "board-1", Name: "Doing", Position: 2},
		},
	}
	svc := NewContextService(repo)

	err := svc.ReorderColumns(context.Background(), "board-1", []string{"c2", "c1", "c3"})
	if err == nil {
		t.Fatal("expected error for extra column, got nil")
	}
	if err.Error() != "column c3 not found in board" {
		t.Errorf("error = %q, want %q", err.Error(), "column c3 not found in board")
	}
}

func TestContextService_ReorderColumns_EmptyBoardID(t *testing.T) {
	repo := &fakeSetupRepo{}
	svc := NewContextService(repo)

	err := svc.ReorderColumns(context.Background(), "", []string{"c1"})
	if err == nil {
		t.Fatal("expected error for empty board id, got nil")
	}
}

func TestContextService_ReorderColumns_EmptyColumnIDs(t *testing.T) {
	repo := &fakeSetupRepo{}
	svc := NewContextService(repo)

	err := svc.ReorderColumns(context.Background(), "board-1", []string{})
	if err == nil {
		t.Fatal("expected error for empty column ids, got nil")
	}
}

func TestContextService_ReorderColumns_DuplicateColumnID(t *testing.T) {
	repo := &fakeSetupRepo{
		columns: []domain.Column{
			{ID: "c1", BoardID: "board-1", Name: "Todo", Position: 1},
			{ID: "c2", BoardID: "board-1", Name: "Doing", Position: 2},
		},
	}
	svc := NewContextService(repo)

	err := svc.ReorderColumns(context.Background(), "board-1", []string{"c1", "c1"})
	if err == nil {
		t.Fatal("expected error for duplicate column id, got nil")
	}
	if err.Error() != "duplicate column id at position 2" {
		t.Errorf("error = %q, want %q", err.Error(), "duplicate column id at position 2")
	}
}
