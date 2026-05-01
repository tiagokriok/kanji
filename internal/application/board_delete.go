package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/tiagokriok/kanji/internal/domain"
)

// BoardDeleteImpact holds a summary of entities that would be removed
// when deleting a board.
type BoardDeleteImpact struct {
	BoardID  string
	Columns  int
	Tasks    int
	Comments int
}

// BoardDeleteService handles board deletion and impact analysis.
type BoardDeleteService struct {
	setupRepo   domain.SetupRepository
	taskRepo    domain.TaskRepository
	commentRepo domain.CommentRepository
}

// NewBoardDeleteService creates a new BoardDeleteService.
func NewBoardDeleteService(setup domain.SetupRepository, task domain.TaskRepository, comment domain.CommentRepository) *BoardDeleteService {
	return &BoardDeleteService{setupRepo: setup, taskRepo: task, commentRepo: comment}
}

// BoardDeleteImpact returns a summary of all entities inside the board.
func (s *BoardDeleteService) BoardDeleteImpact(ctx context.Context, workspaceID, boardID string) (BoardDeleteImpact, error) {
	boardID = strings.TrimSpace(boardID)
	if boardID == "" {
		return BoardDeleteImpact{}, fmt.Errorf("board id is required")
	}

	columns, err := s.setupRepo.ListColumns(ctx, boardID)
	if err != nil {
		return BoardDeleteImpact{}, err
	}
	impact := BoardDeleteImpact{BoardID: boardID, Columns: len(columns)}

	tasks, err := s.taskRepo.List(ctx, domain.TaskFilter{WorkspaceID: workspaceID, BoardID: boardID})
	if err != nil {
		return BoardDeleteImpact{}, err
	}
	impact.Tasks = len(tasks)

	for _, task := range tasks {
		comments, err := s.commentRepo.ListByTask(ctx, task.ID)
		if err != nil {
			return BoardDeleteImpact{}, err
		}
		impact.Comments += len(comments)
	}

	return impact, nil
}

// DeleteBoard removes a board and all its dependent entities.
func (s *BoardDeleteService) DeleteBoard(ctx context.Context, boardID string) error {
	boardID = strings.TrimSpace(boardID)
	if boardID == "" {
		return fmt.Errorf("board id is required")
	}

	return s.setupRepo.DeleteBoard(ctx, boardID)
}
