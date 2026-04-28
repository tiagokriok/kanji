package repositories

import (
	"context"

	"github.com/tiagokriok/kanji/internal/domain"
	"github.com/tiagokriok/kanji/internal/infrastructure/db/sqlc"
)

func queryListBoards(ctx context.Context, q *sqlc.Queries, workspaceID string) ([]domain.Board, error) {
	items, err := q.ListBoards(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Board, 0, len(items))
	for _, item := range items {
		result = append(result, fromSQLBoard(item))
	}
	return result, nil
}

func queryListColumns(ctx context.Context, q *sqlc.Queries, boardID string) ([]domain.Column, error) {
	items, err := q.ListColumns(ctx, boardID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Column, 0, len(items))
	for _, item := range items {
		result = append(result, fromSQLColumn(item))
	}
	return result, nil
}
