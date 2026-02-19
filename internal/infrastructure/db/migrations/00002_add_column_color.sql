-- +goose Up
-- +goose StatementBegin
ALTER TABLE columns
ADD COLUMN color TEXT NOT NULL DEFAULT '#6B7280'
CHECK (color GLOB '#[0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f]');

UPDATE columns
SET color = CASE LOWER(name)
  WHEN 'todo' THEN '#60A5FA'
  WHEN 'doing' THEN '#F59E0B'
  WHEN 'done' THEN '#22C55E'
  ELSE color
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Intentionally no-op. SQLite/libSQL/D1 compatibility makes dropping columns unsafe.
SELECT 1;
-- +goose StatementEnd
