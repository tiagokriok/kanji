# Kanji CLI Commands

## Global Flags

All commands support these persistent flags:

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `--db-path` | `KANJI_DB_PATH` | `~/.config/kanji/app.db` | Path to SQLite database |
| `--json` | - | false | Output in JSON format |
| `--verbose` | - | false | Enable verbose output |

---

## Root Command

### `kanji`

Shows didactic help when run with no arguments.

```bash
kanji
```

---

## Data Operations

### `kanji data bootstrap`

Bootstrap the system with default provider, workspace, board, and columns.
Idempotent -- safe to run multiple times.

```bash
kanji data bootstrap
kanji data bootstrap --json
```

### `kanji data seed`

Seed sample/demo data. Non-production use only.

```bash
kanji data seed
kanji data seed --json
```

---

## Database Operations

### `kanji db info`

Show database status and metadata.

```bash
kanji db info
kanji db info --json
```

### `kanji db migrate up`

Run forward migrations.

```bash
kanji db migrate up
```

### `kanji db migrate status`

Show migration status.

```bash
kanji db migrate status
kanji db migrate status --json
```

### `kanji db doctor`

Run read-only database diagnostics. Exits with code 1 if issues found.

```bash
kanji db doctor
kanji db doctor --json
```

---

## Context Operations

### `kanji context show`

Show current namespace and context.

```bash
kanji context show
kanji context show --json
```

### `kanji context set`

Set explicit CLI context for the current namespace.

```bash
kanji context set --workspace-id <id>
kanji context set --workspace "My Workspace"
kanji context set --workspace-id <id> --board-id <id>
kanji context set --workspace "My Workspace" --board "My Board"
```

### `kanji context clear`

Clear explicit CLI context (preserves TUI state).

```bash
kanji context clear
kanji context clear --json
```

---

## Workspace Operations

### `kanji workspace list`

List all workspaces. Global resource, no context required.

```bash
kanji workspace list
kanji workspace list --json
```

### `kanji workspace create`

Create a new workspace. Auto-creates a default `Main` board.

```bash
kanji workspace create --name "My Workspace"
kanji workspace create --name "My Workspace" --set-context
kanji workspace create --name "My Workspace" --json
```

### `kanji workspace update`

Update a workspace name.

```bash
kanji workspace update --workspace-id <id> --name "New Name"
kanji workspace update --workspace "Old Name" --name "New Name"
```

### `kanji workspace get`

Get a workspace by ID or name.

```bash
kanji workspace get --workspace-id <id>
kanji workspace get --workspace "My Workspace"
kanji workspace get --workspace-id <id> --json
```

---

## Board Operations

### `kanji board list`

List boards for a workspace. Requires workspace scope.

```bash
kanji board list --workspace-id <id>
kanji board list --workspace "My Workspace"
kanji board list  # infers from cli_context if set
```

### `kanji board create`

Create a new board. Supports custom columns or smart defaults.

```bash
kanji board create --name "My Board" --workspace-id <id>
kanji board create --name "My Board" --workspace-id <id> --column "Todo:#FFFFFF" --column "Done:#000000"
kanji board create --name "My Board" --workspace-id <id> --set-context
```

### `kanji board update`

Update a board name.

```bash
kanji board update --board-id <id> --name "New Name"
kanji board update --board "Old Name" --workspace-id <id> --name "New Name"
```

### `kanji board get`

Get a board by ID or name.

```bash
kanji board get --board-id <id>
kanji board get --board "My Board" --workspace-id <id>
```

---

## Column Operations

### `kanji column list`

List columns for a board. Requires board scope.

```bash
kanji column list --board-id <id>
kanji column list --board "My Board"  # requires workspace context
```

### `kanji column create`

Create a new column. Uses deterministic palette color if omitted.

```bash
kanji column create --name "Review" --board-id <id>
kanji column create --name "Review" --board-id <id> --color "#FF0000"
kanji column create --name "Review" --board-id <id> --wip-limit 5
```

### `kanji column update`

Update column metadata.

```bash
kanji column update --column-id <id> --name "New Name"
kanji column update --column-id <id> --color "#00FF00"
kanji column update --column-id <id> --wip-limit 3
kanji column update --column-id <id> --clear-wip-limit
```

### `kanji column reorder`

Reorder columns for a board.

```bash
kanji column reorder --board-id <id> --column-id <id1> --column-id <id2> --column-id <id3>
```

### `kanji column get`

Get a column by ID or name.

```bash
kanji column get --column-id <id>
kanji column get --column "Todo" --board-id <id>
```

---

## Task Operations

### `kanji task list`

List tasks for a workspace. Supports optional board narrowing and filters.

```bash
kanji task list --workspace-id <id>
kanji task list --workspace-id <id> --board-id <id>
kanji task list --workspace-id <id> --query "search term"
kanji task list --workspace-id <id> --column <column-id>
kanji task list --workspace-id <id> --due-soon 7
```

### `kanji task create`

Create a new task.

```bash
kanji task create --title "My Task" --workspace-id <id>
kanji task create --title "My Task" --workspace-id <id> --board-id <id>
kanji task create --title "My Task" --workspace-id <id> --column-id <id>
kanji task create --title "My Task" --workspace-id <id> --priority high
kanji task create --title "My Task" --workspace-id <id> --due-date 2026-05-01
kanji task create --title "My Task" --workspace-id <id> --description-file task.md
```

### `kanji task update`

Update task metadata.

```bash
kanji task update --task-id <id> --title "New Title"
kanji task update --task-id <id> --priority low
kanji task update --task-id <id> --due-date 2026-05-01
kanji task update --task-id <id> --description-file new_desc.md
```

### `kanji task move`

Move a task to another column.

```bash
kanji task move --task-id <id> --to-column-id <id>
kanji task move --task "My Task" --workspace-id <id> --to-column "Done"
```

### `kanji task delete`

Delete a task. Requires explicit confirmation.

```bash
kanji task delete --task-id <id> --yes
kanji task delete --task "My Task" --workspace-id <id> --yes
```

### `kanji task get`

Get a task by ID or title.

```bash
kanji task get --task-id <id>
kanji task get --task "My Task" --workspace-id <id>
kanji task get --task-id <id> --include-comments
```

---

## Comment Operations

### `kanji comment list`

List comments for a task.

```bash
kanji comment list --task-id <id>
kanji comment list --task-id <id> --json
```

### `kanji comment create`

Create a comment on a task.

```bash
kanji comment create --task-id <id> --body "Great work!"
kanji comment create --task-id <id> --body-file comment.md
kanji comment create --task-id <id> --body-file -  # stdin
kanji comment create --task-id <id> --body "Note" --author "Alice"
```

### `kanji comment get`

Get a comment by ID.

```bash
kanji comment get --comment-id <id>
kanji comment get --comment-id <id> --json
```

---

## TUI

### `kanji tui`

Launch the interactive TUI.

```bash
kanji tui
```

---

## Help Topics

### `kanji help concepts`

Core kanji concepts and hierarchy.

### `kanji help context`

Namespace and context model.

### `kanji help selectors`

Resource selection rules.

### `kanji help output`

Output formats and options.
