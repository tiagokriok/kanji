package application

import (
	"context"
	"strings"
	"testing"

	"github.com/tiagokriok/kanji/internal/domain"
	"github.com/tiagokriok/kanji/internal/state"
)

// diagFakeRepo is a test double for domain.SetupRepository.
type diagFakeRepo struct {
	workspaces []domain.Workspace
	boards     map[string][]domain.Board
	columns    map[string][]domain.Column
}

func (r *diagFakeRepo) ListProviders(ctx context.Context) ([]domain.Provider, error) {
	return nil, nil
}
func (r *diagFakeRepo) CreateProvider(ctx context.Context, provider domain.Provider) error {
	return nil
}
func (r *diagFakeRepo) ListWorkspaces(ctx context.Context) ([]domain.Workspace, error) {
	return r.workspaces, nil
}
func (r *diagFakeRepo) CreateWorkspace(ctx context.Context, workspace domain.Workspace) error {
	return nil
}
func (r *diagFakeRepo) RenameWorkspace(ctx context.Context, workspaceID, name string) error {
	return nil
}
func (r *diagFakeRepo) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	return nil
}
func (r *diagFakeRepo) ListBoards(ctx context.Context, workspaceID string) ([]domain.Board, error) {
	return r.boards[workspaceID], nil
}
func (r *diagFakeRepo) CreateBoard(ctx context.Context, board domain.Board) error {
	return nil
}
func (r *diagFakeRepo) RenameBoard(ctx context.Context, boardID, name string) error {
	return nil
}
func (r *diagFakeRepo) DeleteBoard(ctx context.Context, boardID string) error {
	return nil
}
func (r *diagFakeRepo) ListColumns(ctx context.Context, boardID string) ([]domain.Column, error) {
	return r.columns[boardID], nil
}
func (r *diagFakeRepo) CreateColumn(ctx context.Context, column domain.Column) error {
	return nil
}
func (r *diagFakeRepo) UpdateColumn(ctx context.Context, columnID string, name, color *string, wipLimit *int, clearWIP bool) error {
	return nil
}
func (r *diagFakeRepo) ReorderColumns(ctx context.Context, boardID string, orderedColumnIDs []string) error {
	return nil
}
func (r *diagFakeRepo) DeleteColumn(ctx context.Context, columnID string) error {
	return nil
}

func TestFindDuplicateWorkspaceNames_Found(t *testing.T) {
	repo := &diagFakeRepo{
		workspaces: []domain.Workspace{
			{ID: "ws-1", Name: "Foo"},
			{ID: "ws-2", Name: "foo"},
			{ID: "ws-3", Name: "Bar"},
		},
	}

	dups, err := FindDuplicateWorkspaceNames(context.Background(), repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dups) != 1 {
		t.Fatalf("expected 1 duplicate, got %d", len(dups))
	}
	if dups[0].Scope != "workspace" {
		t.Errorf("Scope = %q, want workspace", dups[0].Scope)
	}
	if dups[0].Name != "foo" {
		t.Errorf("Name = %q, want foo", dups[0].Name)
	}
	if dups[0].Count != 2 {
		t.Errorf("Count = %d, want 2", dups[0].Count)
	}
	if len(dups[0].IDs) != 2 {
		t.Errorf("len(IDs) = %d, want 2", len(dups[0].IDs))
	}
}

func TestFindDuplicateWorkspaceNames_None(t *testing.T) {
	repo := &diagFakeRepo{
		workspaces: []domain.Workspace{
			{ID: "ws-1", Name: "Foo"},
			{ID: "ws-2", Name: "Bar"},
		},
	}

	dups, err := FindDuplicateWorkspaceNames(context.Background(), repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dups) != 0 {
		t.Fatalf("expected 0 duplicates, got %d", len(dups))
	}
}

func TestFindDuplicateWorkspaceNames_TrimAndLower(t *testing.T) {
	repo := &diagFakeRepo{
		workspaces: []domain.Workspace{
			{ID: "ws-1", Name: "  Foo  "},
			{ID: "ws-2", Name: "FOO"},
		},
	}

	dups, err := FindDuplicateWorkspaceNames(context.Background(), repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dups) != 1 {
		t.Fatalf("expected 1 duplicate, got %d", len(dups))
	}
	if dups[0].Name != "foo" {
		t.Errorf("Name = %q, want foo", dups[0].Name)
	}
}

func TestFindDuplicateBoardNames_Found(t *testing.T) {
	repo := &diagFakeRepo{
		workspaces: []domain.Workspace{
			{ID: "ws-1"},
			{ID: "ws-2"},
		},
		boards: map[string][]domain.Board{
			"ws-1": {
				{ID: "b-1", WorkspaceID: "ws-1", Name: "Main"},
				{ID: "b-2", WorkspaceID: "ws-1", Name: "main"},
			},
			"ws-2": {
				{ID: "b-3", WorkspaceID: "ws-2", Name: "Main"},
			},
		},
	}

	dups, err := FindDuplicateBoardNames(context.Background(), repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dups) != 1 {
		t.Fatalf("expected 1 duplicate, got %d", len(dups))
	}
	if dups[0].Scope != "board" {
		t.Errorf("Scope = %q, want board", dups[0].Scope)
	}
	if dups[0].Name != "main" {
		t.Errorf("Name = %q, want main", dups[0].Name)
	}
	if dups[0].Count != 2 {
		t.Errorf("Count = %d, want 2", dups[0].Count)
	}
}

func TestFindDuplicateBoardNames_None(t *testing.T) {
	repo := &diagFakeRepo{
		workspaces: []domain.Workspace{{ID: "ws-1"}},
		boards: map[string][]domain.Board{
			"ws-1": {
				{ID: "b-1", WorkspaceID: "ws-1", Name: "Main"},
				{ID: "b-2", WorkspaceID: "ws-1", Name: "Other"},
			},
		},
	}

	dups, err := FindDuplicateBoardNames(context.Background(), repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dups) != 0 {
		t.Fatalf("expected 0 duplicates, got %d", len(dups))
	}
}

func TestFindDuplicateColumnNames_Found(t *testing.T) {
	repo := &diagFakeRepo{
		workspaces: []domain.Workspace{{ID: "ws-1"}},
		boards: map[string][]domain.Board{
			"ws-1": {{ID: "b-1", WorkspaceID: "ws-1"}},
		},
		columns: map[string][]domain.Column{
			"b-1": {
				{ID: "c-1", BoardID: "b-1", Name: "Todo"},
				{ID: "c-2", BoardID: "b-1", Name: "TODO"},
			},
		},
	}

	dups, err := FindDuplicateColumnNames(context.Background(), repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dups) != 1 {
		t.Fatalf("expected 1 duplicate, got %d", len(dups))
	}
	if dups[0].Scope != "column" {
		t.Errorf("Scope = %q, want column", dups[0].Scope)
	}
	if dups[0].Name != "todo" {
		t.Errorf("Name = %q, want todo", dups[0].Name)
	}
	if dups[0].Count != 2 {
		t.Errorf("Count = %d, want 2", dups[0].Count)
	}
}

func TestFindDuplicateColumnNames_None(t *testing.T) {
	repo := &diagFakeRepo{
		workspaces: []domain.Workspace{{ID: "ws-1"}},
		boards: map[string][]domain.Board{
			"ws-1": {{ID: "b-1", WorkspaceID: "ws-1"}},
		},
		columns: map[string][]domain.Column{
			"b-1": {
				{ID: "c-1", BoardID: "b-1", Name: "Todo"},
				{ID: "c-2", BoardID: "b-1", Name: "Doing"},
			},
		},
	}

	dups, err := FindDuplicateColumnNames(context.Background(), repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dups) != 0 {
		t.Fatalf("expected 0 duplicates, got %d", len(dups))
	}
}

func TestFindDanglingContextRefs_NoDangling(t *testing.T) {
	repo := &diagFakeRepo{
		workspaces: []domain.Workspace{{ID: "ws-1"}},
		boards: map[string][]domain.Board{
			"ws-1": {{ID: "b-1", WorkspaceID: "ws-1"}},
		},
	}

	tmpFile := t.TempDir() + "/state.json"
	store := state.NewStore(tmpFile)
	if err := store.SetCLIContext("ns-1", state.CLIContext{
		WorkspaceID: "ws-1",
		BoardID:     "b-1",
	}); err != nil {
		t.Fatalf("set context: %v", err)
	}

	refs, err := FindDanglingContextRefs(context.Background(), repo, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refs) != 0 {
		t.Fatalf("expected 0 dangling refs, got %d: %v", len(refs), refs)
	}
}

func TestFindDanglingContextRefs_DanglingWorkspaceAndBoard(t *testing.T) {
	repo := &diagFakeRepo{
		workspaces: []domain.Workspace{{ID: "ws-1"}},
		boards:     map[string][]domain.Board{"ws-1": {{ID: "b-1", WorkspaceID: "ws-1"}}},
	}

	tmpFile := t.TempDir() + "/state.json"
	store := state.NewStore(tmpFile)
	if err := store.SetCLIContext("ns-1", state.CLIContext{
		WorkspaceID: "missing-ws",
		BoardID:     "missing-board",
	}); err != nil {
		t.Fatalf("set context: %v", err)
	}

	refs, err := FindDanglingContextRefs(context.Background(), repo, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refs) != 2 {
		t.Fatalf("expected 2 dangling refs, got %d: %v", len(refs), refs)
	}
	foundWS := false
	foundBoard := false
	for _, r := range refs {
		if contains(r, "missing-ws") {
			foundWS = true
		}
		if contains(r, "missing-board") {
			foundBoard = true
		}
	}
	if !foundWS {
		t.Error("expected dangling workspace reference")
	}
	if !foundBoard {
		t.Error("expected dangling board reference")
	}
}

func TestFindDanglingContextRefs_MissingStateFile(t *testing.T) {
	repo := &diagFakeRepo{}

	tmpFile := t.TempDir() + "/nonexistent/state.json"
	store := state.NewStore(tmpFile)

	refs, err := FindDanglingContextRefs(context.Background(), repo, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refs) != 0 {
		t.Fatalf("expected 0 dangling refs for missing file, got %d", len(refs))
	}
}

func TestFindDanglingContextRefs_MultipleNamespaces(t *testing.T) {
	repo := &diagFakeRepo{
		workspaces: []domain.Workspace{{ID: "ws-1"}},
		boards:     map[string][]domain.Board{"ws-1": {{ID: "b-1", WorkspaceID: "ws-1"}}},
	}

	tmpFile := t.TempDir() + "/state.json"
	store := state.NewStore(tmpFile)
	if err := store.SetCLIContext("ns-ok", state.CLIContext{WorkspaceID: "ws-1", BoardID: "b-1"}); err != nil {
		t.Fatalf("set context: %v", err)
	}
	if err := store.SetCLIContext("ns-bad", state.CLIContext{WorkspaceID: "bad-ws"}); err != nil {
		t.Fatalf("set context: %v", err)
	}

	refs, err := FindDanglingContextRefs(context.Background(), repo, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("expected 1 dangling ref, got %d: %v", len(refs), refs)
	}
	if !contains(refs[0], "ns-bad") || !contains(refs[0], "bad-ws") {
		t.Errorf("unexpected ref: %q", refs[0])
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
