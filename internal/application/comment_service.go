package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/tiagokriok/kanji/internal/domain"
)

type AddCommentInput struct {
	TaskID     string
	ProviderID string
	BodyMD     string
	Author     *string
}

type CommentService struct {
	repo domain.CommentRepository
}

func NewCommentService(repo domain.CommentRepository) *CommentService {
	return &CommentService{repo: repo}
}

func (s *CommentService) AddComment(ctx context.Context, input AddCommentInput) (domain.Comment, error) {
	if strings.TrimSpace(input.TaskID) == "" {
		return domain.Comment{}, errors.New("task id is required")
	}
	if strings.TrimSpace(input.ProviderID) == "" {
		return domain.Comment{}, errors.New("provider id is required")
	}
	if strings.TrimSpace(input.BodyMD) == "" {
		return domain.Comment{}, errors.New("comment body is required")
	}

	now := time.Now().UTC()
	comment := domain.Comment{
		ID:         uuid.NewString(),
		TaskID:     input.TaskID,
		ProviderID: input.ProviderID,
		BodyMD:     input.BodyMD,
		Author:     input.Author,
		CreatedAt:  now,
	}
	if err := s.repo.Create(ctx, comment); err != nil {
		return domain.Comment{}, err
	}
	return comment, nil
}

func (s *CommentService) ListComments(ctx context.Context, taskID string) ([]domain.Comment, error) {
	if strings.TrimSpace(taskID) == "" {
		return nil, errors.New("task id is required")
	}
	return s.repo.ListByTask(ctx, taskID)
}
