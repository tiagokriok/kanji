# kanji

CLI-first task manager with an optional TUI. Local-first, SQLite-backed.

## Stack

- Go
- Cobra (CLI framework)
- Bubble Tea + Bubbles + Lip Gloss (TUI)
- SQLite (file-based, `modernc.org/sqlite`)
- Goose migrations
- SQLC-compatible schema and queries

## Quickstart

```bash
# Bootstrap the system
go run ./cmd/kanji data bootstrap

# Launch the TUI
go run ./cmd/kanji tui

# Or use CLI commands
go run ./cmd/kanji workspace list
go run ./cmd/kanji task list --workspace-id <id>
```

## CLI

kanji is a CLI-first tool. The TUI is launched explicitly via a subcommand.

### Global flags

- `--db-path` - path to SQLite database (env: `KANJI_DB_PATH`)
- `--json` - output in JSON format
- `--verbose` - enable verbose output

### First run

```bash
# Bootstrap creates default provider, workspace, board, and columns
kanji data bootstrap

# Seed sample/demo data (optional)
kanji data seed
```

### Context

kanji uses a namespace model based on the current working directory.

```bash
# Show current namespace and context
kanji context show

# Set explicit workspace and/or board
kanji context set --workspace "My Workspace" --board "My Board"

# Clear explicit context
kanji context clear
```

### Operational commands

```bash
# Database info
kanji db info

# Run migrations
kanji db migrate up

# Migration status
kanji db migrate status

# Database diagnostics
kanji db doctor
```

### Resource commands

```bash
# Workspaces
kanji workspace list
kanji workspace get --workspace-id <id>
kanji workspace get --workspace "My Workspace"

# Boards
kanji board list --workspace-id <id>
kanji board get --board-id <id>

# Columns
kanji column list --board-id <id>
kanji column get --column-id <id>

# Tasks
kanji task list --workspace-id <id>
kanji task list --workspace-id <id> --board-id <id> --query "search"
kanji task get --task-id <id>

# Comments
kanji comment list --task-id <id>
kanji comment get --comment-id <id>
```

### TUI

```bash
kanji tui
```

## Database

Default DB path: `~/.config/kanji/app.db`

Override with `--db-path` or `KANJI_DB_PATH` env var.

## Architecture

Hexagonal (Ports & Adapters):

- Domain: entities + repository/provider ports (`internal/domain`)
- Application: use cases (`internal/application`)
- Infrastructure: SQLite adapter, migrations, SQL query layer, repository adapters (`internal/infrastructure`)
- UI: Bubble Tea TUI (`internal/ui`)
- CLI: Cobra-based command tree (`cmd/kanji`)

## Project Structure

```text
cmd/kanji/main.go
cmd/kanji/internal/cli/

internal/
  domain/
  application/
  ui/
  infrastructure/
    db/
    repositories/
    store/
  state/
```
