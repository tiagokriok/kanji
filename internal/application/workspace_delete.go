package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/tiagokriok/kanji/internal/domain"
)

// WorkspaceDeleteImpact holds a summary of entities that would be removed
// when deleting a workspace.
type WorkspaceDeleteImpact struct {
	WorkspaceID string
	Boards      int
	Columns     int
	Tasks       int
	Comments    int
}

// WorkspaceDeleteService handles workspace deletion and impact analysis.
type WorkspaceDeleteService struct {
	setupRepo   domain.SetupRepository
	taskRepo    domain.TaskRepository
	commentRepo domain.CommentRepository
}

// NewWorkspaceDeleteService creates a new WorkspaceDeleteService.
func NewWorkspaceDeleteService(
	setup domain.SetupRepository,
	task domain.TaskRepository,
	comment domain.CommentRepository,
) *WorkspaceDeleteService {
	return &WorkspaceDeleteService{
		setupRepo:   setup,
		taskRepo:    task,
		commentRepo: comment,
	}
}

// Impact returns a summary of all entities inside the workspace.
func (s *WorkspaceDeleteService) Impact(ctx context.Context, workspaceID string) (WorkspaceDeleteImpact, error) {
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return WorkspaceDeleteImpact{}, fmt.Errorf("workspace id is required")
	}

	boards, err := s.setupRepo.ListBoards(ctx, workspaceID)
	if err != nil {
		return WorkspaceDeleteImpact{}, err
	}

	impact := WorkspaceDeleteImpact{
		WorkspaceID: workspaceID,
		Boards:      len(boards),
	}

	for _, board := range boards {
		columns, err := s.setupRepo.ListColumns(ctx, board.ID)
		if err != nil {
			return WorkspaceDeleteImpact{}, err
		}
		impact.Columns += len(columns)
	}

	tasks, err := s.taskRepo.List(ctx, domain.TaskFilter{WorkspaceID: workspaceID})
	if err != nil {
		return WorkspaceDeleteImpact{}, err
	}
	impact.Tasks = len(tasks)

	for _, task := range tasks {
		comments, err := s.commentRepo.ListByTask(ctx, task.ID)
		if err != nil {
			return WorkspaceDeleteImpact{}, err
		}
		impact.Comments += len(comments)
	}

	return impact, nil
}

// Delete removes a workspace and all its dependent entities.
func (s *WorkspaceDeleteService) Delete(ctx context.Context, workspaceID string) error {
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return fmt.Errorf("workspace id is required")
	}

	return s.setupRepo.DeleteWorkspace(ctx, workspaceID)
}
