# lazytask

Local-first TUI task manager MVP in Go.

## Stack

- Go
- Bubble Tea + Bubbles + Lip Gloss
- SQLite (file-based, `modernc.org/sqlite`)
- Goose migrations
- SQLC-compatible schema and queries (`sqlc.yaml` + SQL files)

## Architecture

Hexagonal (Ports & Adapters):

- Domain: entities + repository/provider ports (`internal/domain`)
- Application: use cases (`internal/application`)
- Infrastructure: SQLite adapter, migrations, SQL query layer, repository adapters (`internal/infrastructure`)
- UI: Bubble Tea TUI, depends only on application services (`internal/ui`)

## Database

Default DB path:

- `~/.config/lazytask/app.db`

Override with:

- `--db-path /path/to/app.db`

### Migrate

Run app migrations only:

```bash
go run ./cmd/app --migrate
```

Or with goose directly:

```bash
go run github.com/pressly/goose/v3/cmd/goose -dir internal/infrastructure/db/migrations sqlite3 ~/.config/lazytask/app.db up
```

### Seed

- First-run bootstrap is automatic (provider/workspace/board/default columns).
- Optional sample tasks:

```bash
go run ./cmd/app --seed
```

## SQLC

Config and SQL definitions:

- `sqlc.yaml`
- `internal/infrastructure/db/sqlc/schema.sql`
- `internal/infrastructure/db/sqlc/queries.sql`

Generate/re-generate code:

```bash
sqlc generate
```

The project includes a typed query package under `internal/infrastructure/db/sqlc` integrated with repositories.

## Run

```bash
go run ./cmd/app
```

## Keybindings

- `q` / `Ctrl+C`: quit
- `Tab`: switch List/Kanban
- `d`: toggle details pane
- `j`/`k` or arrows: move selection
- `Enter`:
  - List: open details
  - Kanban: move selected task to next column
- `m`: move selected task to next column
- `n`: new task
- `e`: edit title
- `E`: edit description (`Ctrl+S` to save, `Esc` cancel)
- `c`: add comment
- `/`: search by title
- `s`: cycle column filter
- `z`: toggle due-soon filter (next 7 days)

## Project Structure

```text
cmd/app/main.go

internal/
  domain/
    task.go
    comment.go
    board.go
    column.go
    provider.go
    repository_interfaces.go

  application/
    task_service.go
    comment_service.go
    filters.go
    bootstrap_service.go

  ui/
    app_model.go
    list_view.go
    kanban_view.go
    detail_view.go
    keymap.go

  infrastructure/
    db/
      sqlite.go
      sqlc/
      migrations/
    repositories/
      task_repository.go
      comment_repository.go
      setup_repository.go
    providers/
      local_provider.go
```
