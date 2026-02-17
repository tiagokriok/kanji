package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/tiagokriok/lazytask/internal/domain"
	"github.com/tiagokriok/lazytask/internal/infrastructure/db"
	"github.com/tiagokriok/lazytask/internal/infrastructure/db/sqlc"
)

type SetupRepository struct {
	db      db.Adapter
	queries *sqlc.Queries
}

func NewSetupRepository(adapter db.Adapter) *SetupRepository {
	return &SetupRepository{
		db:      adapter,
		queries: adapter.Queries(),
	}
}

func (r *SetupRepository) ListProviders(ctx context.Context) ([]domain.Provider, error) {
	items, err := r.queries.ListProviders(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Provider, 0, len(items))
	for _, item := range items {
		result = append(result, fromSQLProvider(item))
	}
	return result, nil
}

func (r *SetupRepository) CreateProvider(ctx context.Context, provider domain.Provider) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx for create provider: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	qtx := r.queries.WithTx(tx)
	err = qtx.CreateProvider(ctx, sqlc.CreateProviderParams{
		ID:        provider.ID,
		Type:      provider.Type,
		Name:      provider.Name,
		AuthJSON:  nullString(provider.AuthJSON),
		CreatedAt: provider.CreatedAt.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SetupRepository) ListWorkspaces(ctx context.Context) ([]domain.Workspace, error) {
	items, err := r.queries.ListWorkspaces(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Workspace, 0, len(items))
	for _, item := range items {
		result = append(result, fromSQLWorkspace(item))
	}
	return result, nil
}

func (r *SetupRepository) CreateWorkspace(ctx context.Context, workspace domain.Workspace) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx for create workspace: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	qtx := r.queries.WithTx(tx)
	err = qtx.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         workspace.ID,
		ProviderID: workspace.ProviderID,
		RemoteID:   nullString(workspace.RemoteID),
		Name:       workspace.Name,
	})
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SetupRepository) ListBoards(ctx context.Context, workspaceID string) ([]domain.Board, error) {
	items, err := r.queries.ListBoards(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Board, 0, len(items))
	for _, item := range items {
		result = append(result, fromSQLBoard(item))
	}
	return result, nil
}

func (r *SetupRepository) CreateBoard(ctx context.Context, board domain.Board) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx for create board: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	qtx := r.queries.WithTx(tx)
	err = qtx.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          board.ID,
		WorkspaceID: board.WorkspaceID,
		RemoteID:    nullString(board.RemoteID),
		Name:        board.Name,
		ViewDefault: board.ViewDefault,
	})
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SetupRepository) ListColumns(ctx context.Context, boardID string) ([]domain.Column, error) {
	items, err := r.queries.ListColumns(ctx, boardID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Column, 0, len(items))
	for _, item := range items {
		result = append(result, fromSQLColumn(item))
	}
	return result, nil
}

func (r *SetupRepository) CreateColumn(ctx context.Context, column domain.Column) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx for create column: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	qtx := r.queries.WithTx(tx)
	var wip sql.NullInt64
	if column.WIPLimit != nil {
		wip = sql.NullInt64{Int64: int64(*column.WIPLimit), Valid: true}
	}
	err = qtx.CreateColumn(ctx, sqlc.CreateColumnParams{
		ID:       column.ID,
		BoardID:  column.BoardID,
		RemoteID: nullString(column.RemoteID),
		Name:     column.Name,
		Position: int64(column.Position),
		WipLimit: wip,
	})
	if err != nil {
		return err
	}
	return tx.Commit()
}
