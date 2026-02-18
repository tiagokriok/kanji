package domain

import "context"

type TaskRepository interface {
	Create(ctx context.Context, task Task) error
	Update(ctx context.Context, taskID string, patch TaskPatch) error
	GetByID(ctx context.Context, taskID string) (Task, error)
	List(ctx context.Context, filter TaskFilter) ([]Task, error)
	Move(ctx context.Context, input MoveTaskInput) error
	ListColumns(ctx context.Context, boardID string) ([]Column, error)
	ListBoards(ctx context.Context, workspaceID string) ([]Board, error)
}

type CommentRepository interface {
	Create(ctx context.Context, comment Comment) error
	ListByTask(ctx context.Context, taskID string) ([]Comment, error)
}

type SetupRepository interface {
	ListProviders(ctx context.Context) ([]Provider, error)
	CreateProvider(ctx context.Context, provider Provider) error
	ListWorkspaces(ctx context.Context) ([]Workspace, error)
	CreateWorkspace(ctx context.Context, workspace Workspace) error
	RenameWorkspace(ctx context.Context, workspaceID, name string) error
	ListBoards(ctx context.Context, workspaceID string) ([]Board, error)
	CreateBoard(ctx context.Context, board Board) error
	RenameBoard(ctx context.Context, boardID, name string) error
	ListColumns(ctx context.Context, boardID string) ([]Column, error)
	CreateColumn(ctx context.Context, column Column) error
}
