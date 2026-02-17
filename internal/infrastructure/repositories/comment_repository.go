package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/tiagokriok/lazytask/internal/domain"
	"github.com/tiagokriok/lazytask/internal/infrastructure/db"
	"github.com/tiagokriok/lazytask/internal/infrastructure/db/sqlc"
)

type CommentRepository struct {
	db      db.Adapter
	queries *sqlc.Queries
}

func NewCommentRepository(adapter db.Adapter) *CommentRepository {
	return &CommentRepository{
		db:      adapter,
		queries: adapter.Queries(),
	}
}

func (r *CommentRepository) Create(ctx context.Context, comment domain.Comment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx for create comment: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	qtx := r.queries.WithTx(tx)
	err = qtx.CreateComment(ctx, sqlc.CreateCommentParams{
		ID:         comment.ID,
		TaskID:     comment.TaskID,
		ProviderID: comment.ProviderID,
		RemoteID:   nullString(comment.RemoteID),
		BodyMd:     comment.BodyMD,
		Author:     nullString(comment.Author),
		CreatedAt:  comment.CreatedAt.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("create comment: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit create comment: %w", err)
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
