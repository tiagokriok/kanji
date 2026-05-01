package application

import (
	"context"
	"strings"
	"testing"

	"github.com/tiagokriok/kanji/internal/domain"
)

func TestPreflightMigrations_Clean(t *testing.T) {
	repo := &diagFakeRepo{
		workspaces: []domain.Workspace{
			{ID: "ws-1", Name: "Foo"},
		},
		boards: map[string][]domain.Board{
			"ws-1": {{ID: "b-1", WorkspaceID: "ws-1", Name: "Main"}},
		},
		columns: map[string][]domain.Column{
			"b-1": {{ID: "c-1", BoardID: "b-1", Name: "Todo"}},
		},
	}

	err := PreflightMigrations(context.Background(), repo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestPreflightMigrations_DuplicateWorkspace(t *testing.T) {
	repo := &diagFakeRepo{
		workspaces: []domain.Workspace{
			{ID: "ws-1", Name: "Foo"},
			{ID: "ws-2", Name: "foo"},
		},
	}

	err := PreflightMigrations(context.Background(), repo)
	if err == nil {
		t.Fatal("expected error for duplicate workspace names, got nil")
	}
	if !strings.Contains(err.Error(), `workspace name "foo" has 2 duplicates`) {
		t.Errorf("error = %q, want duplicate workspace message", err.Error())
	}
	if !strings.Contains(err.Error(), "kanji db doctor") {
		t.Errorf("error = %q, want actionable message", err.Error())
	}
}

func TestPreflightMigrations_DuplicateBoard(t *testing.T) {
	repo := &diagFakeRepo{
		workspaces: []domain.Workspace{{ID: "ws-1"}},
		boards: map[string][]domain.Board{
			"ws-1": {
				{ID: "b-1", WorkspaceID: "ws-1", Name: "Main"},
				{ID: "b-2", WorkspaceID: "ws-1", Name: "main"},
			},
		},
	}

	err := PreflightMigrations(context.Background(), repo)
	if err == nil {
		t.Fatal("expected error for duplicate board names, got nil")
	}
	if !strings.Contains(err.Error(), `board name "main" has 2 duplicates`) {
		t.Errorf("error = %q, want duplicate board message", err.Error())
	}
}

func TestPreflightMigrations_DuplicateColumn(t *testing.T) {
	repo := &diagFakeRepo{
		workspaces: []domain.Workspace{{ID: "ws-1"}},
		boards:     map[string][]domain.Board{"ws-1": {{ID: "b-1", WorkspaceID: "ws-1"}}},
		columns: map[string][]domain.Column{
			"b-1": {
				{ID: "c-1", BoardID: "b-1", Name: "Todo"},
				{ID: "c-2", BoardID: "b-1", Name: "TODO"},
			},
		},
	}

	err := PreflightMigrations(context.Background(), repo)
	if err == nil {
		t.Fatal("expected error for duplicate column names, got nil")
	}
	if !strings.Contains(err.Error(), `column name "todo" has 2 duplicates`) {
		t.Errorf("error = %q, want duplicate column message", err.Error())
	}
}

func TestPreflightMigrations_MultipleIssues(t *testing.T) {
	repo := &diagFakeRepo{
		workspaces: []domain.Workspace{
			{ID: "ws-1", Name: "Foo"},
			{ID: "ws-2", Name: "foo"},
		},
		boards: map[string][]domain.Board{
			"ws-1": {
				{ID: "b-1", WorkspaceID: "ws-1", Name: "Main"},
				{ID: "b-2", WorkspaceID: "ws-1", Name: "main"},
			},
		},
	}

	err := PreflightMigrations(context.Background(), repo)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), `workspace name "foo" has 2 duplicates`) {
		t.Errorf("missing workspace duplicate in error: %q", err.Error())
	}
	if !strings.Contains(err.Error(), `board name "main" has 2 duplicates`) {
		t.Errorf("missing board duplicate in error: %q", err.Error())
	}
}
