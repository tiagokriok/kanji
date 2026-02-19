CREATE TABLE providers (
  id TEXT PRIMARY KEY,
  type TEXT NOT NULL,
  name TEXT NOT NULL,
  auth_json TEXT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE workspaces (
  id TEXT PRIMARY KEY,
  provider_id TEXT NOT NULL,
  remote_id TEXT NULL,
  name TEXT NOT NULL,
  FOREIGN KEY (provider_id) REFERENCES providers(id)
);

CREATE TABLE boards (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL,
  remote_id TEXT NULL,
  name TEXT NOT NULL,
  view_default TEXT NOT NULL,
  FOREIGN KEY (workspace_id) REFERENCES workspaces(id)
);

CREATE TABLE columns (
  id TEXT PRIMARY KEY,
  board_id TEXT NOT NULL,
  remote_id TEXT NULL,
  name TEXT NOT NULL,
  color TEXT NOT NULL DEFAULT '#6B7280'
    CHECK (color GLOB '#[0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f]'),
  position INTEGER NOT NULL,
  wip_limit INTEGER NULL,
  FOREIGN KEY (board_id) REFERENCES boards(id)
);

CREATE TABLE tasks (
  id TEXT PRIMARY KEY,
  provider_id TEXT NOT NULL,
  workspace_id TEXT NOT NULL,
  board_id TEXT NULL,
  column_id TEXT NULL,
  remote_id TEXT NULL,
  title TEXT NOT NULL,
  description_md TEXT NOT NULL DEFAULT '',
  status TEXT NULL,
  priority INTEGER NOT NULL DEFAULT 0,
  due_at TEXT NULL,
  estimate_minutes INTEGER NULL,
  assignee TEXT NULL,
  labels_json TEXT NOT NULL DEFAULT '[]',
  position REAL NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (provider_id) REFERENCES providers(id),
  FOREIGN KEY (workspace_id) REFERENCES workspaces(id),
  FOREIGN KEY (board_id) REFERENCES boards(id),
  FOREIGN KEY (column_id) REFERENCES columns(id)
);

CREATE TABLE comments (
  id TEXT PRIMARY KEY,
  task_id TEXT NOT NULL,
  provider_id TEXT NOT NULL,
  remote_id TEXT NULL,
  body_md TEXT NOT NULL,
  author TEXT NULL,
  created_at TEXT NOT NULL,
  FOREIGN KEY (task_id) REFERENCES tasks(id),
  FOREIGN KEY (provider_id) REFERENCES providers(id)
);

CREATE TABLE sync_queue (
  id TEXT PRIMARY KEY,
  provider_id TEXT NOT NULL,
  entity TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  action TEXT NOT NULL,
  payload_json TEXT NOT NULL,
  attempts INTEGER NOT NULL DEFAULT 0,
  last_error TEXT NULL,
  created_at TEXT NOT NULL,
  FOREIGN KEY (provider_id) REFERENCES providers(id)
);

CREATE INDEX idx_tasks_workspace_id ON tasks(workspace_id);
CREATE INDEX idx_tasks_column_id ON tasks(column_id);
CREATE INDEX idx_tasks_updated_at ON tasks(updated_at);
CREATE INDEX idx_tasks_due_at ON tasks(due_at);
CREATE INDEX idx_comments_task_created ON comments(task_id, created_at);
CREATE INDEX idx_columns_board_position ON columns(board_id, position);
