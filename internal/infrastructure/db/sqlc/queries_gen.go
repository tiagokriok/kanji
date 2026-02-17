package sqlc

import (
	"context"
	"database/sql"
)

const createProvider = `-- name: CreateProvider :exec
INSERT INTO providers (id, type, name, auth_json, created_at)
VALUES (?, ?, ?, ?, ?)
`

type CreateProviderParams struct {
	ID        string
	Type      string
	Name      string
	AuthJSON  sql.NullString
	CreatedAt string
}

func (q *Queries) CreateProvider(ctx context.Context, arg CreateProviderParams) error {
	_, err := q.db.ExecContext(ctx, createProvider,
		arg.ID,
		arg.Type,
		arg.Name,
		arg.AuthJSON,
		arg.CreatedAt,
	)
	return err
}

const listProviders = `-- name: ListProviders :many
SELECT id, type, name, auth_json, created_at
FROM providers
ORDER BY created_at ASC
`

func (q *Queries) ListProviders(ctx context.Context) ([]Provider, error) {
	rows, err := q.db.QueryContext(ctx, listProviders)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Provider, 0)
	for rows.Next() {
		var i Provider
		if err := rows.Scan(&i.ID, &i.Type, &i.Name, &i.AuthJSON, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const createWorkspace = `-- name: CreateWorkspace :exec
INSERT INTO workspaces (id, provider_id, remote_id, name)
VALUES (?, ?, ?, ?)
`

type CreateWorkspaceParams struct {
	ID         string
	ProviderID string
	RemoteID   sql.NullString
	Name       string
}

func (q *Queries) CreateWorkspace(ctx context.Context, arg CreateWorkspaceParams) error {
	_, err := q.db.ExecContext(ctx, createWorkspace, arg.ID, arg.ProviderID, arg.RemoteID, arg.Name)
	return err
}

const listWorkspaces = `-- name: ListWorkspaces :many
SELECT id, provider_id, remote_id, name
FROM workspaces
ORDER BY name ASC
`

func (q *Queries) ListWorkspaces(ctx context.Context) ([]Workspace, error) {
	rows, err := q.db.QueryContext(ctx, listWorkspaces)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Workspace, 0)
	for rows.Next() {
		var i Workspace
		if err := rows.Scan(&i.ID, &i.ProviderID, &i.RemoteID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const createBoard = `-- name: CreateBoard :exec
INSERT INTO boards (id, workspace_id, remote_id, name, view_default)
VALUES (?, ?, ?, ?, ?)
`

type CreateBoardParams struct {
	ID          string
	WorkspaceID string
	RemoteID    sql.NullString
	Name        string
	ViewDefault string
}

func (q *Queries) CreateBoard(ctx context.Context, arg CreateBoardParams) error {
	_, err := q.db.ExecContext(ctx, createBoard, arg.ID, arg.WorkspaceID, arg.RemoteID, arg.Name, arg.ViewDefault)
	return err
}

const listBoards = `-- name: ListBoards :many
SELECT id, workspace_id, remote_id, name, view_default
FROM boards
WHERE workspace_id = ?
ORDER BY name ASC
`

func (q *Queries) ListBoards(ctx context.Context, workspaceID string) ([]Board, error) {
	rows, err := q.db.QueryContext(ctx, listBoards, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Board, 0)
	for rows.Next() {
		var i Board
		if err := rows.Scan(&i.ID, &i.WorkspaceID, &i.RemoteID, &i.Name, &i.ViewDefault); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const createColumn = `-- name: CreateColumn :exec
INSERT INTO columns (id, board_id, remote_id, name, position, wip_limit)
VALUES (?, ?, ?, ?, ?, ?)
`

type CreateColumnParams struct {
	ID       string
	BoardID  string
	RemoteID sql.NullString
	Name     string
	Position int64
	WipLimit sql.NullInt64
}

func (q *Queries) CreateColumn(ctx context.Context, arg CreateColumnParams) error {
	_, err := q.db.ExecContext(ctx, createColumn, arg.ID, arg.BoardID, arg.RemoteID, arg.Name, arg.Position, arg.WipLimit)
	return err
}

const listColumns = `-- name: ListColumns :many
SELECT id, board_id, remote_id, name, position, wip_limit
FROM columns
WHERE board_id = ?
ORDER BY position ASC
`

func (q *Queries) ListColumns(ctx context.Context, boardID string) ([]Column, error) {
	rows, err := q.db.QueryContext(ctx, listColumns, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Column, 0)
	for rows.Next() {
		var i Column
		if err := rows.Scan(&i.ID, &i.BoardID, &i.RemoteID, &i.Name, &i.Position, &i.WipLimit); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const createTask = `-- name: CreateTask :exec
INSERT INTO tasks (
  id,
  provider_id,
  workspace_id,
  board_id,
  column_id,
  remote_id,
  title,
  description_md,
  status,
  priority,
  due_at,
  estimate_minutes,
  assignee,
  labels_json,
  position,
  created_at,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`

type CreateTaskParams struct {
	ID              string
	ProviderID      string
	WorkspaceID     string
	BoardID         sql.NullString
	ColumnID        sql.NullString
	RemoteID        sql.NullString
	Title           string
	DescriptionMd   string
	Status          sql.NullString
	Priority        int64
	DueAt           sql.NullString
	EstimateMinutes sql.NullInt64
	Assignee        sql.NullString
	LabelsJSON      string
	Position        float64
	CreatedAt       string
	UpdatedAt       string
}

func (q *Queries) CreateTask(ctx context.Context, arg CreateTaskParams) error {
	_, err := q.db.ExecContext(ctx, createTask,
		arg.ID,
		arg.ProviderID,
		arg.WorkspaceID,
		arg.BoardID,
		arg.ColumnID,
		arg.RemoteID,
		arg.Title,
		arg.DescriptionMd,
		arg.Status,
		arg.Priority,
		arg.DueAt,
		arg.EstimateMinutes,
		arg.Assignee,
		arg.LabelsJSON,
		arg.Position,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}

const updateTask = `-- name: UpdateTask :exec
UPDATE tasks
SET
  title = COALESCE(?, title),
  description_md = COALESCE(?, description_md),
  status = COALESCE(?, status),
  priority = COALESCE(?, priority),
  due_at = COALESCE(?, due_at),
  column_id = COALESCE(?, column_id),
  labels_json = COALESCE(?, labels_json),
  updated_at = ?
WHERE id = ?
`

type UpdateTaskParams struct {
	Title         sql.NullString
	DescriptionMd sql.NullString
	Status        sql.NullString
	Priority      sql.NullInt64
	DueAt         sql.NullString
	ColumnID      sql.NullString
	LabelsJSON    sql.NullString
	UpdatedAt     string
	ID            string
}

func (q *Queries) UpdateTask(ctx context.Context, arg UpdateTaskParams) error {
	_, err := q.db.ExecContext(ctx, updateTask,
		arg.Title,
		arg.DescriptionMd,
		arg.Status,
		arg.Priority,
		arg.DueAt,
		arg.ColumnID,
		arg.LabelsJSON,
		arg.UpdatedAt,
		arg.ID,
	)
	return err
}

const getTask = `-- name: GetTask :one
SELECT
  id,
  provider_id,
  workspace_id,
  board_id,
  column_id,
  remote_id,
  title,
  description_md,
  status,
  priority,
  due_at,
  estimate_minutes,
  assignee,
  labels_json,
  position,
  created_at,
  updated_at
FROM tasks
WHERE id = ?
`

func (q *Queries) GetTask(ctx context.Context, id string) (Task, error) {
	row := q.db.QueryRowContext(ctx, getTask, id)
	var i Task
	err := row.Scan(
		&i.ID,
		&i.ProviderID,
		&i.WorkspaceID,
		&i.BoardID,
		&i.ColumnID,
		&i.RemoteID,
		&i.Title,
		&i.DescriptionMd,
		&i.Status,
		&i.Priority,
		&i.DueAt,
		&i.EstimateMinutes,
		&i.Assignee,
		&i.LabelsJSON,
		&i.Position,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listTasks = `-- name: ListTasks :many
SELECT
  id,
  provider_id,
  workspace_id,
  board_id,
  column_id,
  remote_id,
  title,
  description_md,
  status,
  priority,
  due_at,
  estimate_minutes,
  assignee,
  labels_json,
  position,
  created_at,
  updated_at
FROM tasks
WHERE workspace_id = ?
  AND (? = '' OR LOWER(title) LIKE '%' || LOWER(?) || '%')
  AND (? = '' OR column_id = ?)
  AND (? = '' OR status = ?)
  AND (? = 0 OR (due_at IS NOT NULL AND due_at <= ?))
ORDER BY updated_at DESC
`

type ListTasksParams struct {
	WorkspaceID   string
	TitleQuery    string
	ColumnID      string
	Status        string
	DueSoonActive int64
	DueSoonBefore string
}

func (q *Queries) ListTasks(ctx context.Context, arg ListTasksParams) ([]Task, error) {
	rows, err := q.db.QueryContext(ctx, listTasks,
		arg.WorkspaceID,
		arg.TitleQuery,
		arg.TitleQuery,
		arg.ColumnID,
		arg.ColumnID,
		arg.Status,
		arg.Status,
		arg.DueSoonActive,
		arg.DueSoonBefore,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Task, 0)
	for rows.Next() {
		var i Task
		if err := rows.Scan(
			&i.ID,
			&i.ProviderID,
			&i.WorkspaceID,
			&i.BoardID,
			&i.ColumnID,
			&i.RemoteID,
			&i.Title,
			&i.DescriptionMd,
			&i.Status,
			&i.Priority,
			&i.DueAt,
			&i.EstimateMinutes,
			&i.Assignee,
			&i.LabelsJSON,
			&i.Position,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const moveTask = `-- name: MoveTask :exec
UPDATE tasks
SET column_id = ?, status = ?, position = ?, updated_at = ?
WHERE id = ?
`

type MoveTaskParams struct {
	ColumnID  sql.NullString
	Status    sql.NullString
	Position  float64
	UpdatedAt string
	ID        string
}

func (q *Queries) MoveTask(ctx context.Context, arg MoveTaskParams) error {
	_, err := q.db.ExecContext(ctx, moveTask, arg.ColumnID, arg.Status, arg.Position, arg.UpdatedAt, arg.ID)
	return err
}

const createComment = `-- name: CreateComment :exec
INSERT INTO comments (id, task_id, provider_id, remote_id, body_md, author, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
`

type CreateCommentParams struct {
	ID         string
	TaskID     string
	ProviderID string
	RemoteID   sql.NullString
	BodyMd     string
	Author     sql.NullString
	CreatedAt  string
}

func (q *Queries) CreateComment(ctx context.Context, arg CreateCommentParams) error {
	_, err := q.db.ExecContext(ctx, createComment,
		arg.ID,
		arg.TaskID,
		arg.ProviderID,
		arg.RemoteID,
		arg.BodyMd,
		arg.Author,
		arg.CreatedAt,
	)
	return err
}

const listComments = `-- name: ListComments :many
SELECT id, task_id, provider_id, remote_id, body_md, author, created_at
FROM comments
WHERE task_id = ?
ORDER BY created_at ASC
`

func (q *Queries) ListComments(ctx context.Context, taskID string) ([]Comment, error) {
	rows, err := q.db.QueryContext(ctx, listComments, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Comment, 0)
	for rows.Next() {
		var i Comment
		if err := rows.Scan(&i.ID, &i.TaskID, &i.ProviderID, &i.RemoteID, &i.BodyMd, &i.Author, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
