package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/tiagokriok/kanji/internal/domain"
	"github.com/tiagokriok/kanji/internal/infrastructure/db"
	"github.com/tiagokriok/kanji/internal/infrastructure/db/sqlc"
	"github.com/tiagokriok/kanji/internal/infrastructure/store"
)

type CommentRepository struct {
	store   store.Store
	queries *sqlc.Queries
}

func NewCommentRepository(adapter db.Adapter) *CommentRepository {
	return &CommentRepository{
		store:   store.New(adapter),
		queries: adapter.Queries(),
	}
}

func (r *CommentRepository) Create(ctx context.Context, comment domain.Comment) error {
	if err := r.store.InTx(ctx, func(tx store.Tx) error {
		qtx := tx.Queries()
		return qtx.CreateComment(ctx, sqlc.CreateCommentParams{
			ID:         comment.ID,
			TaskID:     comment.TaskID,
			ProviderID: comment.ProviderID,
			RemoteID:   nullString(comment.RemoteID),
			BodyMd:     comment.BodyMD,
			Author:     nullString(comment.Author),
			CreatedAt:  comment.CreatedAt.UTC().Format(time.RFC3339),
		})
	}); err != nil {
		return fmt.Errorf("create comment: %w", err)
	}
	return nil
}

func (r *CommentRepository) ListByTask(ctx context.Context, taskID string) ([]domain.Comment, error) {
	items, err := r.queries.ListComments(ctx, taskID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Comment, 0, len(items))
	for _, item := range items {
		result = append(result, fromSQLComment(item))
	}
	return result, nil
}
