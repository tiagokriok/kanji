package repositories

import (
	"context"
	"errors"
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
	return r.store.Write(ctx, "create provider", func(tx store.Tx) error {
		qtx := tx.Queries()
		return qtx.CreateProvider(ctx, sqlc.CreateProviderParams{
			ID:        provider.ID,
			Type:      provider.Type,
			Name:      provider.Name,
			AuthJSON:  nullString(provider.AuthJSON),
			CreatedAt: provider.CreatedAt.UTC().Format(time.RFC3339),
		})
	})
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
	return r.store.Write(ctx, "create workspace", func(tx store.Tx) error {
		qtx := tx.Queries()
		return qtx.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
			ID:         workspace.ID,
			ProviderID: workspace.ProviderID,
			RemoteID:   nullString(workspace.RemoteID),
			Name:       workspace.Name,
		})
	})
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

	return r.store.Write(ctx, "rename workspace", func(tx store.Tx) error {
		qtx := tx.Queries()
		return qtx.UpdateWorkspaceName(ctx, sqlc.UpdateWorkspaceNameParams{
			Name: name,
			ID:   workspaceID,
		})
	})
}

func (r *SetupRepository) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return fmt.Errorf("workspace id is required")
	}

	return r.store.Write(ctx, "delete workspace", func(tx store.Tx) error {
		qtx := tx.Queries()
		if err := qtx.DeleteCommentsByWorkspace(ctx, workspaceID); err != nil {
			return fmt.Errorf("delete comments: %w", err)
		}
		if err := qtx.DeleteTasksByWorkspace(ctx, workspaceID); err != nil {
			return fmt.Errorf("delete tasks: %w", err)
		}
		if err := qtx.DeleteColumnsByWorkspace(ctx, workspaceID); err != nil {
			return fmt.Errorf("delete columns: %w", err)
		}
		if err := qtx.DeleteBoardsByWorkspace(ctx, workspaceID); err != nil {
			return fmt.Errorf("delete boards: %w", err)
		}
		if err := qtx.DeleteWorkspace(ctx, workspaceID); err != nil {
			return fmt.Errorf("delete workspace: %w", err)
		}
		return nil
	})
}

func (r *SetupRepository) ListBoards(ctx context.Context, workspaceID string) ([]domain.Board, error) {
	return queryListBoards(ctx, r.store.Queries(), workspaceID)
}

func (r *SetupRepository) CreateBoard(ctx context.Context, board domain.Board) error {
	return r.store.Write(ctx, "create board", func(tx store.Tx) error {
		qtx := tx.Queries()
		return qtx.CreateBoard(ctx, sqlc.CreateBoardParams{
			ID:          board.ID,
			WorkspaceID: board.WorkspaceID,
			RemoteID:    nullString(board.RemoteID),
			Name:        board.Name,
			ViewDefault: board.ViewDefault,
		})
	})
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

	return r.store.Write(ctx, "rename board", func(tx store.Tx) error {
		qtx := tx.Queries()
		return qtx.UpdateBoardName(ctx, sqlc.UpdateBoardNameParams{
			Name: name,
			ID:   boardID,
		})
	})
}

func (r *SetupRepository) DeleteBoard(ctx context.Context, boardID string) error {
	boardID = strings.TrimSpace(boardID)
	if boardID == "" {
		return fmt.Errorf("board id is required")
	}

	return r.store.Write(ctx, "delete board", func(tx store.Tx) error {
		qtx := tx.Queries()
		if err := qtx.DeleteCommentsByBoard(ctx, boardID); err != nil {
			return fmt.Errorf("delete comments: %w", err)
		}
		if err := qtx.DeleteTasksByBoard(ctx, boardID); err != nil {
			return fmt.Errorf("delete tasks: %w", err)
		}
		if err := qtx.DeleteColumnsByBoard(ctx, boardID); err != nil {
			return fmt.Errorf("delete columns: %w", err)
		}
		if err := qtx.DeleteBoard(ctx, boardID); err != nil {
			return fmt.Errorf("delete board: %w", err)
		}
		return nil
	})
}

func (r *SetupRepository) ListColumns(ctx context.Context, boardID string) ([]domain.Column, error) {
	return queryListColumns(ctx, r.store.Queries(), boardID)
}

func (r *SetupRepository) CreateColumn(ctx context.Context, column domain.Column) error {
	return r.store.Write(ctx, "create column", func(tx store.Tx) error {
		qtx := tx.Queries()
		return qtx.CreateColumn(ctx, sqlc.CreateColumnParams{
			ID:       column.ID,
			BoardID:  column.BoardID,
			RemoteID: nullString(column.RemoteID),
			Name:     column.Name,
			Color:    normalizeHexColor(column.Color),
			Position: int64(column.Position),
			WipLimit: nullInt(column.WIPLimit),
		})
	})
}

func (r *SetupRepository) UpdateColumn(ctx context.Context, columnID string, name, color *string, wipLimit *int, clearWIP bool) error {
	columnID = strings.TrimSpace(columnID)
	if columnID == "" {
		return fmt.Errorf("column id is required")
	}

	return r.store.Write(ctx, "update column", func(tx store.Tx) error {
		qtx := tx.Queries()
		if name != nil {
			if strings.TrimSpace(*name) == "" {
				return errors.New("column name cannot be empty")
			}
			if err := qtx.UpdateColumnName(ctx, sqlc.UpdateColumnNameParams{
				Name: strings.TrimSpace(*name),
				ID:   columnID,
			}); err != nil {
				return fmt.Errorf("update column name: %w", err)
			}
		}
		if color != nil {
			if err := qtx.UpdateColumnColor(ctx, sqlc.UpdateColumnColorParams{
				Color: normalizeHexColor(*color),
				ID:    columnID,
			}); err != nil {
				return fmt.Errorf("update column color: %w", err)
			}
		}
		if clearWIP {
			if err := qtx.UpdateColumnWIPLimit(ctx, sqlc.UpdateColumnWIPLimitParams{
				WipLimit: nullInt(nil),
				ID:       columnID,
			}); err != nil {
				return fmt.Errorf("clear column wip limit: %w", err)
			}
		}
		if wipLimit != nil {
			if err := qtx.UpdateColumnWIPLimit(ctx, sqlc.UpdateColumnWIPLimitParams{
				WipLimit: nullInt(wipLimit),
				ID:       columnID,
			}); err != nil {
				return fmt.Errorf("update column wip limit: %w", err)
			}
		}
		return nil
	})
}

func (r *SetupRepository) ReorderColumns(ctx context.Context, boardID string, orderedColumnIDs []string) error {
	boardID = strings.TrimSpace(boardID)
	if boardID == "" {
		return fmt.Errorf("board id is required")
	}
	if len(orderedColumnIDs) == 0 {
		return fmt.Errorf("at least one column id is required")
	}

	return r.store.Write(ctx, "reorder columns", func(tx store.Tx) error {
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

func (r *SetupRepository) DeleteColumn(ctx context.Context, columnID string) error {
	columnID = strings.TrimSpace(columnID)
	if columnID == "" {
		return fmt.Errorf("column id is required")
	}

	return r.store.Write(ctx, "delete column", func(tx store.Tx) error {
		return tx.Queries().DeleteColumn(ctx, columnID)
	})
}
