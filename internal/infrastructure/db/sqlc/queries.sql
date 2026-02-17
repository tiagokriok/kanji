-- name: CreateProvider :exec
INSERT INTO providers (id, type, name, auth_json, created_at)
VALUES (?, ?, ?, ?, ?);

-- name: ListProviders :many
SELECT id, type, name, auth_json, created_at
FROM providers
ORDER BY created_at ASC;

-- name: CreateWorkspace :exec
INSERT INTO workspaces (id, provider_id, remote_id, name)
VALUES (?, ?, ?, ?);

-- name: ListWorkspaces :many
SELECT id, provider_id, remote_id, name
FROM workspaces
ORDER BY name ASC;

-- name: CreateBoard :exec
INSERT INTO boards (id, workspace_id, remote_id, name, view_default)
VALUES (?, ?, ?, ?, ?);

-- name: ListBoards :many
SELECT id, workspace_id, remote_id, name, view_default
FROM boards
WHERE workspace_id = ?
ORDER BY name ASC;

-- name: CreateColumn :exec
INSERT INTO columns (id, board_id, remote_id, name, position, wip_limit)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListColumns :many
SELECT id, board_id, remote_id, name, position, wip_limit
FROM columns
WHERE board_id = ?
ORDER BY position ASC;

-- name: CreateTask :exec
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
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateTask :exec
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
WHERE id = ?;

-- name: GetTask :one
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
WHERE id = ?;

-- name: ListTasks :many
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
ORDER BY updated_at DESC;

-- name: MoveTask :exec
UPDATE tasks
SET column_id = ?, status = ?, position = ?, updated_at = ?
WHERE id = ?;

-- name: CreateComment :exec
INSERT INTO comments (id, task_id, provider_id, remote_id, body_md, author, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListComments :many
SELECT id, task_id, provider_id, remote_id, body_md, author, created_at
FROM comments
WHERE task_id = ?
ORDER BY created_at ASC;
