package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/tiagokriok/kanji/internal/domain"
	"github.com/tiagokriok/kanji/internal/infrastructure/db/sqlc"
	"github.com/tiagokriok/kanji/internal/infrastructure/store"
)

type SetupRepository struct {
	store store.Store
}

func NewSetupRepository(s store.Store) *SetupRepository {
	return &SetupRepository{store: s}
}

func (r *SetupRepository) ListProviders(ctx context.Context) ([]domain.Provider, error) {
	items, err := r.store.Queries().ListProviders(ctx)
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
	if err := r.store.InTx(ctx, func(tx store.Tx) error {
		qtx := tx.Queries()
		return qtx.CreateProvider(ctx, sqlc.CreateProviderParams{
			ID:        provider.ID,
			Type:      provider.Type,
			Name:      provider.Name,
			AuthJSON:  nullString(provider.AuthJSON),
			CreatedAt: provider.CreatedAt.UTC().Format(time.RFC3339),
		})
	}); err != nil {
		return fmt.Errorf("create provider: %w", err)
	}
	return nil
}

func (r *SetupRepository) ListWorkspaces(ctx context.Context) ([]domain.Workspace, error) {
	items, err := r.store.Queries().ListWorkspaces(ctx)
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
	if err := r.store.InTx(ctx, func(tx store.Tx) error {
		qtx := tx.Queries()
		return qtx.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
			ID:         workspace.ID,
			ProviderID: workspace.ProviderID,
			RemoteID:   nullString(workspace.RemoteID),
			Name:       workspace.Name,
		})
	}); err != nil {
		return fmt.Errorf("create workspace: %w", err)
	}
	return nil
}

func (r *SetupRepository) RenameWorkspace(ctx context.Context, workspaceID, name string) error {
	workspaceID = strings.TrimSpace(workspaceID)
	name = strings.TrimSpace(name)
	if workspaceID == "" {
		return fmt.Errorf("workspace id is required")
	}
	if name == "" {
		return fmt.Errorf("workspace name is required")
	}

	if err := r.store.InTx(ctx, func(tx store.Tx) error {
		qtx := tx.Queries()
		return qtx.UpdateWorkspaceName(ctx, sqlc.UpdateWorkspaceNameParams{
			Name: name,
			ID:   workspaceID,
		})
	}); err != nil {
		return fmt.Errorf("rename workspace: %w", err)
	}
	return nil
}

func (r *SetupRepository) ListBoards(ctx context.Context, workspaceID string) ([]domain.Board, error) {
	items, err := r.store.Queries().ListBoards(ctx, workspaceID)
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
	if err := r.store.InTx(ctx, func(tx store.Tx) error {
		qtx := tx.Queries()
		return qtx.CreateBoard(ctx, sqlc.CreateBoardParams{
			ID:          board.ID,
			WorkspaceID: board.WorkspaceID,
			RemoteID:    nullString(board.RemoteID),
			Name:        board.Name,
			ViewDefault: board.ViewDefault,
		})
	}); err != nil {
		return fmt.Errorf("create board: %w", err)
	}
	return nil
}

func (r *SetupRepository) RenameBoard(ctx context.Context, boardID, name string) error {
	boardID = strings.TrimSpace(boardID)
	name = strings.TrimSpace(name)
	if boardID == "" {
		return fmt.Errorf("board id is required")
	}
	if name == "" {
		return fmt.Errorf("board name is required")
	}

	if err := r.store.InTx(ctx, func(tx store.Tx) error {
		qtx := tx.Queries()
		return qtx.UpdateBoardName(ctx, sqlc.UpdateBoardNameParams{
			Name: name,
			ID:   boardID,
		})
	}); err != nil {
		return fmt.Errorf("rename board: %w", err)
	}
	return nil
}

func (r *SetupRepository) ListColumns(ctx context.Context, boardID string) ([]domain.Column, error) {
	items, err := r.store.Queries().ListColumns(ctx, boardID)
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
	if err := r.store.InTx(ctx, func(tx store.Tx) error {
		qtx := tx.Queries()
		var wip sql.NullInt64
		if column.WIPLimit != nil {
			wip = sql.NullInt64{Int64: int64(*column.WIPLimit), Valid: true}
		}
		return qtx.CreateColumn(ctx, sqlc.CreateColumnParams{
			ID:       column.ID,
			BoardID:  column.BoardID,
			RemoteID: nullString(column.RemoteID),
			Name:     column.Name,
			Color:    normalizeHexColor(column.Color),
			Position: int64(column.Position),
			WipLimit: wip,
		})
	}); err != nil {
		return fmt.Errorf("create column: %w", err)
	}
	return nil
}

func (r *SetupRepository) ReorderColumns(ctx context.Context, boardID string, orderedColumnIDs []string) error {
	boardID = strings.TrimSpace(boardID)
	if boardID == "" {
		return fmt.Errorf("board id is required")
	}
	if len(orderedColumnIDs) == 0 {
		return fmt.Errorf("at least one column id is required")
	}

	return r.store.InTx(ctx, func(tx store.Tx) error {
		qtx := tx.Queries()
		for idx, columnID := range orderedColumnIDs {
			columnID = strings.TrimSpace(columnID)
			if columnID == "" {
				return fmt.Errorf("column id at position %d is empty", idx+1)
			}
			affected, execErr := qtx.UpdateColumnPosition(ctx, sqlc.UpdateColumnPositionParams{
				Position: int64(idx + 1),
				ID:       columnID,
				BoardID:  boardID,
			})
			if execErr != nil {
				return fmt.Errorf("update position for column %s: %w", columnID, execErr)
			}
			if affected == 0 {
				return fmt.Errorf("column %s not found in board", columnID)
			}
		}
		return nil
	})
}
