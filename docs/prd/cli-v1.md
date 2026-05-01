# Kanji CLI v1 Implementation Blueprint

Status: Draft
Owner: kanji
Scope: Transform kanji into a CLI-first app with TUI as a subcommand

## 1. Purpose

Transform `kanji` from a TUI-first app into a CLI-first, automation-friendly tool that:

- shows didactic help when run with no args
- launches the TUI only via `kanji tui`
- supports full first-class resource operations for:
  - workspaces
  - boards
  - columns
  - tasks
  - comments
  - context
- is understandable by humans, coding agents, and LLMs
- provides stable human and JSON output modes
- preserves TUI behavior during the transition

## 2. Goals

- CLI-first UX with noun-first command grammar
- Very detailed built-in help as a product requirement
- Strong selector semantics with exact, normalized matching
- Safe destructive operations with explicit confirmation and dry-run support
- Per-directory namespaced context for parallel agent workflows
- Shared application layer for both CLI and TUI
- Generated Markdown CLI docs in-repo

## 3. Non-goals

- Shell completions in initial v1
- Manpages in initial v1
- CLI editor-launch integration in initial v1
- Provider as a first-class public resource in v1
- `--quiet` mode in v1
- `--cwd` or `--no-context` in v1
- Migration `down` / `reset` in v1
- Column delete task cascade in v1
- Label filtering in `task list` v1

## 4. Breaking Behavior Changes

- `kanji` with no args will show help instead of launching the TUI
- `kanji tui` becomes the only TUI entrypoint
- entrypoint moves from `cmd/app` to `cmd/kanji`
- `--migrate` and `--seed` top-level flags are retired in favor of subcommands

## 5. Final UX Rules

### 5.1 Invocation model

- `kanji` => help
- `kanji tui` => TUI
- `kanji <noun> <verb>` => CLI command

### 5.2 Command grammar

Use noun-first grammar only.

Examples:

- `kanji task create ...`
- `kanji board get ...`
- `kanji context set ...`
- `kanji db migrate up`

No aliases in v1.

### 5.3 Output modes

- Default: human-readable
- `--json`: structured machine output
- `--verbose`: adds operational resolution detail in human mode only

### 5.4 Destructive commands

- Non-interactive by default
- Require explicit confirmation flags
- Parent cascades require `--cascade`
- High-level destructive commands support `--dry-run`

## 6. Top-level Command Tree

```text
kanji
  help
    concepts
    context
    selectors
    output

  tui

  workspace
    list
    get
    create
    update
    delete

  board
    list
    get
    create
    update
    delete

  column
    list
    get
    create
    update
    delete
    reorder

  task
    list
    get
    create
    update
    move
    delete

  comment
    list
    get
    create
    update
    delete

  context
    show
    set
    clear

  db
    info
    doctor
    migrate
      up
      status

  data
    bootstrap
    seed
```

## 7. Global Flags and Env

### 7.1 Persistent flags

Available on all commands:

- `--db-path <path>`
- `--json`
- `--verbose`

### 7.2 Env vars

Operational settings only:

- `KANJI_DB_PATH`
- `KANJI_CONTEXT`

### 7.3 Resolution precedence

1. explicit flags
2. env vars
3. persisted namespaced CLI context
4. built-in defaults

## 8. Context Model

## 8.1 Namespace

- default namespace key = exact normalized cwd
- if `KANJI_CONTEXT` is set, it fully replaces cwd-derived namespace key
- applies to CLI and TUI equally

## 8.2 Shared state store

Store is shared between CLI and TUI, but with separate semantics:

```json
{
  "namespaces": {
    "<namespace-key>": {
      "cli_context": {
        "workspace_id": "...",
        "board_id": "..."
      },
      "tui_state": {
        "last_workspace_id": "...",
        "last_board_id": "..."
      }
    }
  }
}
```

Rules:

- CLI inference uses only `cli_context`
- TUI startup precedence:
  1. `cli_context`
  2. `tui_state`
  3. bootstrap fallback
- `context clear` clears only `cli_context`
- destructive deletes sanitize both `cli_context` and `tui_state`

## 8.3 Context command semantics

- `context show` shows:
  - namespace key
  - namespace source (`cwd` or `KANJI_CONTEXT`)
  - effective CLI context
  - TUI state secondarily
- `context set` validates immediately
- `context set --workspace ...` clears board if board omitted
- `context set --board ...` by name requires workspace scope

## 9. Selector Model

## 9.1 Canonical selector flags

- `--workspace-id`
- `--board-id`
- `--column-id`
- `--task-id`
- `--comment-id`

## 9.2 Name selector flags

- `--workspace`
- `--board`
- `--column`
- `--task`

## 9.3 Matching rules

Name matching is:

- trim surrounding whitespace
- case-insensitive exact match
- no fuzzy matching
- no substring matching
- no prefix matching

## 9.4 ID + name together

If both canonical ID and name selector are supplied:

- both are resolved
- both must match the same resource
- mismatch is an error

## 9.5 Scope rules

- reads may infer from `cli_context`
- writes may rely on previously explicit `kanji context set`
- missing required scope => hard error, never silent broadening
- `task list` supports workspace scope, board narrows it

## 10. Uniqueness Invariants

Enforce in CLI validation, application layer, and DB constraints.

- workspace name unique globally, case-insensitive
- board name unique within workspace, case-insensitive
- column name unique within board, case-insensitive
- task titles remain non-unique

Rationale:

- makes human selectors deterministic
- keeps task selectors convenience-only

## 11. Human and JSON Output Contracts

## 11.1 Human output

- list commands => tables
- get commands => key/value detail blocks
- create/update/delete => concise summaries
- IDs always included in list output
- when relevant, mention context resolution in human output
- `--verbose` may include:
  - namespace key
  - namespace source
  - DB path
  - resolution details
  - operation counts

## 11.2 JSON output

Always wrapped.

### Lists

```json
{
  "tasks": [],
  "count": 0,
  "scope": {}
}
```

### Singular

```json
{
  "task": {}
}
```

### Delete

```json
{
  "deleted": {
    "resource": "task",
    "id": "..."
  }
}
```

### Errors

```json
{
  "error": {
    "code": "validation_error",
    "message": "...",
    "details": {}
  }
}
```

### Resolved metadata

Optional standardized block:

```json
{
  "task": {},
  "resolved": {
    "workspace": {},
    "board": {},
    "column": {}
  }
}
```

Derived fields are allowed where stable and useful, but canonical fields remain primary.

## 11.3 Exit codes

- `0` success
- `1` operational/domain/validation failure
- `2` usage/parsing/help misuse

## 12. Help and Docs Contract

## 12.1 Top-level help must teach

Top-level help should include:

1. what kanji is
2. command grammar
3. bootstrap-first workflow
4. context model
5. selector model
6. output modes
7. destructive command policy
8. common workflows
9. command list

## 12.2 Topic help required

- `kanji help concepts`
- `kanji help context`
- `kanji help selectors`
- `kanji help output`

## 12.3 Per-command help

Each major command should include:

- short summary
- long explanation
- required and optional flags
- selector rules relevant to that command
- JSON/output notes where relevant
- one human-oriented example
- one agent/script-oriented example
- references to topic help where appropriate

## 12.4 Generated docs

Generate Markdown docs under:

- `docs/cli/`

Top-level README becomes CLI-first and TUI-secondary.

## 13. Exact Command and Flag Matrix

## 13.1 `kanji tui`

Purpose:
- Launch TUI using same runtime/config/context rules as CLI

Flags:
- global flags only

Example:
```bash
kanji tui
KANJI_CONTEXT=agent-a kanji tui --db-path /tmp/kanji.db
```

## 13.2 `kanji workspace`

### `workspace list`

Purpose:
- List all workspaces

Flags:
- global only

Human columns:
- ID
- Name

Examples:
```bash
kanji workspace list
kanji workspace list --json
```

### `workspace get`

Flags:
- one of:
  - `--workspace-id`
  - `--workspace`
- optional:
  - `--include-boards`

Examples:
```bash
kanji workspace get --workspace "Product"
kanji workspace get --workspace-id ws_123 --include-boards --json
```

### `workspace create`

Behavior:
- requires bootstrap first
- creates workspace
- auto-creates default board `Main`
- primary created resource is workspace
- output includes default board too
- `--set-context` sets workspace + created `Main` board

Flags:
- required:
  - `--name`
- optional:
  - `--set-context`

Examples:
```bash
kanji workspace create --name Product
kanji workspace create --name Product --set-context --json
```

### `workspace update`

Behavior:
- patch-style
- in v1, effectively rename

Flags:
- selector:
  - `--workspace-id` or `--workspace`
- required mutation:
  - `--name`

Examples:
```bash
kanji workspace update --workspace "Product" --name "Platform"
```

### `workspace delete`

Behavior:
- requires `--yes --cascade`
- supports `--dry-run`
- clears invalid CLI/TUI context automatically if needed

Flags:
- selector:
  - `--workspace-id` or `--workspace`
- required for real delete:
  - `--yes`
  - `--cascade`
- optional:
  - `--dry-run`

Examples:
```bash
kanji workspace delete --workspace "Platform" --cascade --dry-run
kanji workspace delete --workspace-id ws_123 --yes --cascade
```

## 13.3 `kanji board`

### `board list`

Scope:
- requires workspace via flags or explicit CLI context

Flags:
- workspace selector:
  - `--workspace-id` or `--workspace`

Human columns:
- ID
- Name

Examples:
```bash
kanji board list --workspace Product
kanji board list --json
```

### `board get`

Flags:
- workspace selector required for board-name resolution
- board selector:
  - `--board-id` or `--board`
- optional:
  - `--include-columns`

Examples:
```bash
kanji board get --workspace Product --board Roadmap
kanji board get --board-id bd_123 --include-columns --json
```

### `board create`

Behavior:
- preserves smart defaults
- if no `--column` provided, default columns are created
- if any `--column` provided, use only provided columns
- `--set-context` updates current namespace context to new board

Flags:
- workspace selector required unless current explicit context already provides board scope parent
- required:
  - `--name`
- repeated optional custom columns:
  - `--column "Todo:#60A5FA"`
- optional:
  - `--set-context`

Examples:
```bash
kanji board create --workspace Product --name Roadmap
kanji board create --workspace Product --name Triage \
  --column "Backlog:#60A5FA" \
  --column "Doing:#F59E0B" \
  --column "Done:#22C55E"
```

### `board update`

Behavior:
- patch-style
- v1 only exposes name mutation

Flags:
- board selector
- optional workspace selector for name resolution
- required mutation:
  - `--name`

### `board delete`

Behavior:
- requires `--yes --cascade`
- supports `--dry-run`
- clears invalid context/state automatically

Flags:
- board selector
- optional workspace selector for name resolution
- required for real delete:
  - `--yes`
  - `--cascade`
- optional:
  - `--dry-run`

## 13.4 `kanji column`

### `column list`

Scope:
- requires board via flags or explicit CLI context

Flags:
- board selector
- optional workspace selector for board-name resolution

Human columns:
- ID
- Name
- Color
- Position
- WIP Limit

### `column get`

Flags:
- column selector
- board selector required for name-based column resolution
- optional workspace selector for board-name resolution

No expansion in v1.

### `column create`

Behavior:
- requires board scope
- `--color` optional
- omitted color uses next default palette color

Flags:
- board selector
- required:
  - `--name`
- optional:
  - `--color`
  - `--wip-limit`

Examples:
```bash
kanji column create --workspace Product --board Roadmap --name Blocked
kanji column create --board Roadmap --workspace Product --name Review --color '#A78BFA' --wip-limit 2
```

### `column update`

Behavior:
- patch-style metadata only
- no ordering fields here

Flags:
- column selector
- board selector required for name-based resolution
- optional:
  - `--name`
  - `--color`
  - `--wip-limit`
  - `--clear-wip-limit`

### `column delete`

Behavior:
- default refuse if tasks exist
- allow reassignment with `--move-tasks-to`
- no cascade delete tasks in v1

Flags:
- column selector
- board selector required for name-based resolution
- optional:
  - `--move-tasks-to`
  - `--move-tasks-to-id`
- required for actual delete:
  - `--yes`

Examples:
```bash
kanji column delete --workspace Product --board Roadmap --column Blocked --yes
kanji column delete --workspace Product --board Roadmap --column Blocked --move-tasks-to Doing --yes
```

### `column reorder`

Behavior:
- full ordered list required
- repeated ordered selector flags
- all columns in board must appear exactly once

Flags:
- board selector
- repeated ordered selectors:
  - `--column`
  - or `--column-id`

Examples:
```bash
kanji column reorder --workspace Product --board Roadmap \
  --column Backlog \
  --column Doing \
  --column Done
```

## 13.5 `kanji task`

### `task list`

Scope:
- requires at least workspace scope
- board is optional narrowing
- may infer from explicit CLI context only

Flags:
- scope:
  - `--workspace-id` or `--workspace`
  - optional `--board-id` or `--board`
- filters:
  - `--query`
  - `--column-id` or `--column`
  - `--priority`
  - `--due-soon`
  - `--overdue`
  - `--no-due-date`
  - `--sort <priority|due|title|updated|created>`

Human columns:
- ID
- Title
- Column
- Priority
- Due
- Labels
- Updated

Examples:
```bash
kanji task list
kanji task list --workspace Product --priority high --sort due
kanji task list --json
```

### `task get`

Flags:
- selector:
  - `--task-id`
  - or `--task`
- if task name used:
  - require strong scope, at least workspace, ideally board for deterministic usage
- optional:
  - `--include-comments`

Examples:
```bash
kanji task get --task-id task_123
kanji task get --workspace Product --board Roadmap --task "Fix login" --include-comments --json
```

### `task create`

Behavior:
- requires bootstrap first
- board required unless current explicit CLI context already provides it
- column may be passed by name within board
- if column omitted, use first board column by position
- status is derived from column, not exposed directly

Flags:
- scope:
  - workspace selector
  - board selector unless explicit context already provides it
- required:
  - `--title`
- optional:
  - `--column-id` or `--column`
  - `--description`
  - `--description-file <path|- >`
  - `--priority <0..5|critical|urgent|high|medium|low|none>`
  - `--due <YYYY-MM-DD|RFC3339>`
  - repeated `--label`
  - `--assignee`
  - `--estimate-minutes`

Examples:
```bash
kanji task create --workspace Product --board Roadmap --title "Fix login"
kanji task create --title "Fix login" --column Doing --priority high --description-file spec.md
cat spec.md | kanji task create --title "Fix login" --description-file - --json
```

### `task update`

Behavior:
- patch-style metadata only
- no workflow/column flags exposed
- explicit clear flags for nullable/list fields

Flags:
- selector:
  - `--task-id`
  - or `--task` with strong scope
- optional mutations:
  - `--title`
  - `--description`
  - `--description-file`
  - `--priority`
  - `--due`
  - repeated `--label`
  - `--assignee`
  - `--estimate-minutes`
  - `--clear-due-date`
  - `--clear-labels`
  - `--clear-assignee`
  - `--clear-estimate-minutes`

### `task move`

Behavior:
- canonical workflow transition command
- destination is column-driven
- raw status not exposed

Flags:
- task selector
- destination:
  - `--to-column-id`
  - or `--to-column`
- board scope required for name-based destination

Examples:
```bash
kanji task move --task-id task_123 --to-column-id col_doing
kanji task move --workspace Product --board Roadmap --task "Fix login" --to-column Doing
```

### `task delete`

Behavior:
- requires `--yes`
- cascades comments automatically

Flags:
- task selector
- required:
  - `--yes`

## 13.6 `kanji comment`

### `comment list`

Scope:
- requires task selector

Flags:
- `--task-id`
- or scoped task title selector

Human columns:
- ID
- Task ID
- Author
- Created
- Body preview

### `comment get`

Flags:
- `--comment-id`

### `comment create`

Flags:
- task selector
- required one body source:
  - `--body`
  - `--body-file <path|- >`
- optional:
  - `--author`

Examples:
```bash
kanji comment create --task-id task_123 --body "Looks good"
cat note.md | kanji comment create --task-id task_123 --body-file - --json
```

### `comment update`

Behavior:
- patch-style, practically full body replacement

Flags:
- `--comment-id`
- required one body source:
  - `--body`
  - `--body-file <path|- >`

### `comment delete`

Flags:
- `--comment-id`
- required:
  - `--yes`

## 13.7 `kanji context`

### `context show`

Shows:
- namespace key
- namespace source
- CLI context
- TUI state secondary

### `context set`

Flags:
- workspace selector optional if board ID uniquely resolves its parent
- board name resolution requires workspace scope

Examples:
```bash
kanji context set --workspace Product
kanji context set --workspace Product --board Roadmap
kanji context set --workspace-id ws_123 --board-id bd_456
```

### `context clear`

Behavior:
- clears only `cli_context`

## 13.8 `kanji db`

### `db info`

Shows:
- effective DB path
- existence
- active namespace
- bootstrap status

### `db doctor`

Behavior:
- read-only
- reports findings
- exit `1` if findings exist

Checks should include:
- bootstrap missing/incomplete
- selector-hostile duplicate names
- dangling context references
- migration-state problems

### `db migrate up`

Behavior:
- run forward migrations only

### `db migrate status`

Behavior:
- show migration state

## 13.9 `kanji data`

### `data bootstrap`

Behavior:
- explicit initialization command
- required before normal domain commands
- idempotent
- reports created vs existing

### `data seed`

Behavior:
- sample/demo data only
- non-production
- best-effort idempotent
- reports created counts

## 14. Internal Package Layout Proposal

```text
cmd/kanji/
  main.go
  root.go

  internal/
    commands/
      workspace/
      board/
      column/
      task/
      comment/
      context/
      db/
      data/
      tui/
      topics/

    runtime/
      app.go            # service/repo/bootstrap wiring
      config.go         # flags/env resolution
      db.go             # db open/migrate helpers
      bootstrap.go      # bootstrap checks

    contextstate/
      store.go          # shared namespaced store
      namespace.go      # cwd/env namespace resolution

    selectors/
      workspace.go
      board.go
      column.go
      task.go
      comment.go
      errors.go

    output/
      human.go
      json.go
      tables.go
      detail.go
      errors.go

    help/
      topics.go
      examples.go
```

Notes:
- CLI-only logic stays under `cmd/kanji/internal/...`
- only promote helpers into `internal/application` or broader shared packages if truly interface-agnostic

## 15. Missing / Expanded Application Use Cases

All CLI business operations must go through the application layer.

## 15.1 New or expanded services/use cases

### Workspace use cases
- get workspace by ID/name
- list workspaces
- create workspace with default board
- update workspace name
- delete workspace with cascade dry-run/summary support

### Board use cases
- get board by ID/name within workspace
- list boards in workspace
- create board with smart default or custom columns
- update board name
- delete board with cascade dry-run/summary support

### Column use cases
- get column by ID/name within board
- list columns in board
- create column with deterministic default color
- update column name/color/wip-limit
- delete column with:
  - empty-column fast path
  - reassignment path via `--move-tasks-to`
- reorder columns with full-order validation

### Task use cases
- list tasks by scoped filters
- get task by ID and optionally include comments
- resolve task by exact title within scope
- create task with column-driven defaulting
- update metadata patch only
- move task by destination column
- delete task with comment cascade

### Comment use cases
- list comments by task
- get comment by ID
- create comment
- update comment body
- delete comment

### Context state use cases
- set CLI context for namespace
- clear CLI context for namespace
- show CLI context + TUI state for namespace
- sanitize invalid namespace state after deletes

### Doctor use cases
- inspect bootstrap state
- inspect uniqueness invariants
- inspect dangling namespaced context references
- inspect migration health

## 16. Required Domain / Repository / SQL Work

## 16.1 Domain interface expansions

Likely additions needed:

- `SetupRepository`
  - get workspace by id/name helpers or equivalent list+resolve support
  - delete workspace
  - delete board
  - delete column
  - update column metadata
  - get board/column detail helpers if application layer should not re-scan collections repeatedly
- `CommentRepository`
  - get by id
  - update
  - delete
- `TaskRepository`
  - get already exists
  - may need richer list filters if future selector resolution benefits

Exact interface names can be refined during implementation, but the use cases above must be supported without CLI bypassing business logic.

## 16.2 SQL/query additions

Add SQLC queries for at least:

- comment get by id
- comment update
- comment delete
- workspace delete
- board delete
- column update metadata
- column delete
- counts for dry-run summaries:
  - board cascade impact
  - workspace cascade impact
- doctor diagnostics queries for duplicate-name detection and dangling references

## 16.3 Required migrations

### Integrity and selector safety

Add case-insensitive normalized uniqueness constraints/indexes for:

- workspaces global name uniqueness
- boards uniqueness within workspace
- columns uniqueness within board

Recommended SQLite strategy:
- unique indexes on normalized name expressions, e.g. `lower(trim(name))`

### Delete semantics support

Ensure DB and app support these semantics cleanly:

- task delete cascades comments
- board/workspace deletes can be summarized and executed safely
- column delete supports reassignment path

### Helpful indexes

Add/confirm indexes for common selector and list paths:

- boards by workspace
- columns by board
- tasks by workspace
- tasks by board
- comments by task
- normalized name lookup indexes where relevant

## 16.4 Upgrade handling

Required:
- explicit migration-time handling for duplicate names or invalid legacy data
- clear errors if uniqueness constraints cannot be applied cleanly
- `db doctor` should detect selector-hostile legacy states

## 17. Test Strategy by Phase

## 17.1 Test categories

### Unit tests
- namespace resolution (`cwd` vs `KANJI_CONTEXT`)
- context-store semantics
- selector parsing/validation/mismatch/ambiguity
- flag conflict validation
- output renderers
- JSON shapes
- topic/help content presence where practical

### Integration tests
- command execution against temp DB
- migrations/bootstrap/seed/doctor
- resource list/get/create/update/delete flows
- JSON and human output verification
- destructive dry-run summaries
- context mutation and cleanup after deletes

### Selective E2E
- bootstrap -> context set -> create/list/get/update/delete workflow
- board/column/task/comment representative flows
- `kanji tui` smoke test only

## 17.2 Phase acceptance criteria

### Phase 1: foundations + ops + read surface

Deliver:
- `cmd/kanji` entrypoint
- no-arg didactic help
- topic help commands
- shared namespaced store with `cli_context` + `tui_state`
- global flags/env wiring
- `db info`
- `db doctor`
- `db migrate up/status`
- `data bootstrap`
- `data seed`
- `context show/set/clear`
- read commands:
  - `workspace list/get`
  - `board list/get`
  - `column list/get`
  - `task list/get --include-comments`
  - `comment list/get`
- `kanji tui`

Acceptance criteria:
- `kanji` shows didactic help
- `kanji tui` launches existing TUI behavior
- context inference works only from `cli_context`
- JSON and human output work for all phase-1 commands
- doctor returns structured findings and exit `1` on problems
- generated docs pipeline can produce phase-1 CLI docs

### Phase 2: writes on existing capabilities

Deliver:
- `workspace create/update`
- `board create/update`
- `column create/update/reorder`
- `task create/update/move/delete`
- `comment create`

Acceptance criteria:
- writes may use explicit selectors or explicit CLI context
- smart defaults behave as specified
- dry-run not yet required except where already in phase scope
- all phase-2 write commands have curated help and examples
- integration tests verify human and JSON outputs

### Phase 3: full CRUD completion + invariants + migrations

Deliver:
- `workspace delete`
- `board delete`
- `column delete`
- `comment update/delete`
- missing app/repo/query support
- uniqueness migrations
- doctor diagnostics for legacy issues

Acceptance criteria:
- uniqueness invariants enforced in DB and app layer
- parent deletes require `--yes --cascade`
- `--dry-run` impact summaries implemented for high-level deletes
- deleting selected resources sanitizes `cli_context` and `tui_state`
- migration failure paths are explicit and documented

### Phase 4: docs/help hardening

Deliver:
- generated docs under `docs/cli/`
- README rewritten CLI-first
- help consistency audit
- JSON contract review
- example audit for humans + agents

Acceptance criteria:
- all major commands documented in generated docs
- README quickstart uses bootstrap-first flow
- topic help and command help are internally consistent
- representative examples validated in tests where practical

## 18. Rollout and Breaking Change Checklist

- add `cmd/kanji/main.go`
- remove `cmd/app`
- update Makefile targets from `cmd/app` to `cmd/kanji`
- update README examples
- update any scripts/tests/docs referring to `cmd/app`
- communicate no-arg behavior change:
  - old: launch TUI
  - new: show help
- communicate new explicit TUI path:
  - `kanji tui`

## 19. Representative Examples

### Bootstrap-first flow

```bash
kanji data bootstrap
kanji context set --workspace Main --board Main
kanji task list
kanji tui
```

### Agent-friendly flow

```bash
KANJI_CONTEXT=agent-a kanji data bootstrap --json
KANJI_CONTEXT=agent-a kanji context set --workspace Product --board Roadmap --json
KANJI_CONTEXT=agent-a kanji task create --title "Fix login" --column Doing --priority high --json
KANJI_CONTEXT=agent-a kanji task list --json
```

### Safe destructive flow

```bash
kanji board delete --workspace Product --board Roadmap --cascade --dry-run
kanji board delete --workspace Product --board Roadmap --yes --cascade
```

### Rich text flow

```bash
cat task.md | kanji task create --title "Write spec" --description-file -
kanji comment update --comment-id c_123 --body-file note.md
```

## 20. Immediate Follow-up After This Blueprint

After this document is accepted:

1. break phase 1 into concrete issues/tasks
2. implement command/runtime foundation first
3. land phase 1 before expanding deep CRUD/migrations
4. derive generated docs after command contracts stabilize
