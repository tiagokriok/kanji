package repositories

import (
	"context"
	"time"

	"github.com/tiagokriok/kanji/internal/domain"
	"github.com/tiagokriok/kanji/internal/infrastructure/db/sqlc"
	"github.com/tiagokriok/kanji/internal/infrastructure/store"
)

type CommentRepository struct {
	store store.Store
}

func NewCommentRepository(s store.Store) *CommentRepository {
	return &CommentRepository{store: s}
}

func (r *CommentRepository) Create(ctx context.Context, comment domain.Comment) error {
	return r.store.Write(ctx, "create comment", func(tx store.Tx) error {
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
	})
}

func (r *CommentRepository) ListByTask(ctx context.Context, taskID string) ([]domain.Comment, error) {
	items, err := r.store.Queries().ListComments(ctx, taskID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Comment, 0, len(items))
	for _, item := range items {
		result = append(result, fromSQLComment(item))
	}
	return result, nil
}

func (r *CommentRepository) Update(ctx context.Context, commentID string, bodyMD string) error {
	return r.store.Write(ctx, "update comment", func(tx store.Tx) error {
		return tx.Queries().UpdateComment(ctx, sqlc.UpdateCommentParams{
			ID:     commentID,
			BodyMd: bodyMD,
		})
	})
}

func (r *CommentRepository) Delete(ctx context.Context, commentID string) error {
	return r.store.Write(ctx, "delete comment", func(tx store.Tx) error {
		return tx.Queries().DeleteComment(ctx, commentID)
	})
}
