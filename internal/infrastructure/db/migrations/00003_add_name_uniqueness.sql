-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS idx_workspaces_name_unique ON workspaces(LOWER(TRIM(name)));
CREATE UNIQUE INDEX IF NOT EXISTS idx_boards_name_unique ON boards(workspace_id, LOWER(TRIM(name)));
CREATE UNIQUE INDEX IF NOT EXISTS idx_columns_name_unique ON columns(board_id, LOWER(TRIM(name)));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_workspaces_name_unique;
DROP INDEX IF EXISTS idx_boards_name_unique;
DROP INDEX IF EXISTS idx_columns_name_unique;
-- +goose StatementEnd
