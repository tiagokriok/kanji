# Kanji CLI v1 Phase 1 Atomic Tasks and Sequenced Issue List

Status: Draft
Parent blueprint: `docs/prd/cli-v1.md`
Scope: Phase 1 only
Goal: Break Phase 1 into atomic one-session tasks and a sequenced issue list with titles and acceptance criteria

---

## 1. Phase 1 Goal

Deliver the CLI foundation and first usable read/operational surface for `kanji`:

- `kanji` no-arg shows didactic help
- `kanji tui` launches TUI
- shared runtime/config/context plumbing exists
- operational commands exist:
  - `db info`
  - `db doctor`
  - `db migrate up`
  - `db migrate status`
  - `data bootstrap`
  - `data seed`
  - `context show/set/clear`
- read commands exist for all first-class resources:
  - `workspace list/get`
  - `board list/get`
  - `column list/get`
  - `task list/get --include-comments`
  - `comment list/get`
- human and JSON output contracts are both implemented
- explicit CLI context inference exists via namespaced shared state
- TUI behavior is preserved during and after cutover

---

## 2. Atomic Task Checklist

Each task is intended to fit in one coding session.

Conventions:
- TDD first
- prefer minimal vertical slices
- each task should end with:
  - `gofmt`
  - relevant targeted tests
  - ideally `go test ./...`

### Foundation spine

- [ ] **P1-01** Create new CLI entrypoint skeleton
- [ ] **P1-02** Add global runtime/config plumbing
- [ ] **P1-03** Add namespace resolution helper
- [ ] **P1-04** Add shared namespaced state store
- [ ] **P1-07** Add bootstrap/runtime app wiring helper
- [ ] **P1-08** Add `data bootstrap`
- [ ] **P1-09** Add bootstrap guard helper for domain commands
- [ ] **P1-14** Add root topic help commands
- [ ] **P1-15** Add shared selector validation layer skeleton
- [ ] **P1-17** Add shared output rendering helpers

### Context + ops

- [ ] **P1-05** Add `context show`
- [ ] **P1-06** Add `context clear`
- [ ] **P1-16** Add `context set`
- [ ] **P1-10** Add `db info`
- [ ] **P1-11** Add `db migrate up`
- [ ] **P1-12** Add `db migrate status`
- [ ] **P1-13** Add `data seed`
- [ ] **P1-28** Add `db doctor`

### Read surface

- [ ] **P1-18** Add `workspace list`
- [ ] **P1-19** Add `workspace get`
- [ ] **P1-20** Add `board list`
- [ ] **P1-21** Add `board get`
- [ ] **P1-22** Add `column list`
- [ ] **P1-23** Add `column get`
- [ ] **P1-24** Add `task list`
- [ ] **P1-25** Add `task get`
- [ ] **P1-26** Add `comment list`
- [ ] **P1-27** Add `comment get`

### TUI cutover + docs

- [ ] **P1-29** Add `kanji tui` command and rewire TUI startup
- [ ] **P1-30** Cut over from `cmd/app` to `cmd/kanji`
- [ ] **P1-31** Generate initial CLI docs under `docs/cli/`

---

## 3. Recommended Execution Order

1. P1-01 — CLI entrypoint skeleton
2. P1-02 — global config/env plumbing
3. P1-03 — namespace resolution
4. P1-04 — shared namespaced state store
5. P1-07 — runtime/service wiring helper
6. P1-08 — `data bootstrap`
7. P1-09 — bootstrap guard helper
8. P1-14 — topic help commands
9. P1-15 — selector validation layer skeleton
10. P1-17 — output rendering helpers
11. P1-05 — `context show`
12. P1-06 — `context clear`
13. P1-16 — `context set`
14. P1-10 — `db info`
15. P1-11 — `db migrate up`
16. P1-12 — `db migrate status`
17. P1-13 — `data seed`
18. P1-28 — `db doctor`
19. P1-18 — `workspace list`
20. P1-19 — `workspace get`
21. P1-20 — `board list`
22. P1-21 — `board get`
23. P1-22 — `column list`
24. P1-23 — `column get`
25. P1-24 — `task list`
26. P1-25 — `task get`
27. P1-26 — `comment list`
28. P1-27 — `comment get`
29. P1-29 — `kanji tui`
30. P1-30 — cutover to `cmd/kanji`
31. P1-31 — docs generation + README update

---

## 4. Sequenced Issue List

Each issue below is designed as a small, implementation-ready slice. Use the sequence order unless an explicit dependency chain is intentionally changed.

---

### Issue 1 — Create CLI root entrypoint skeleton

**ID**: P1-01  
**Title**: `feat(cli): add cmd/kanji root command skeleton`

**Goal**
- Introduce `cmd/kanji/main.go`
- Add Cobra root command
- Make `kanji` with no args show didactic help

**Dependencies**
- none

**Acceptance criteria**
- `go run ./cmd/kanji` prints root help
- root command exists with placeholder structure for future subcommands
- tests cover no-arg help behavior
- existing TUI code remains untouched

---

### Issue 2 — Add global config and env resolution

**ID**: P1-02  
**Title**: `feat(cli): add global runtime config flags and env precedence`

**Goal**
- Add persistent flags:
  - `--db-path`
  - `--json`
  - `--verbose`
- Add env support:
  - `KANJI_DB_PATH`
  - `KANJI_CONTEXT`
- Implement precedence rules

**Dependencies**
- P1-01

**Acceptance criteria**
- root help shows all persistent flags
- runtime config object exists under `cmd/kanji/internal/...`
- flag > env > default precedence is tested
- global flags are available throughout the command tree scaffold

---

### Issue 3 — Add namespace resolution helper

**ID**: P1-03  
**Title**: `feat(cli): add namespace resolution from cwd and KANJI_CONTEXT`

**Goal**
- Resolve active namespace key from exact cwd by default
- Fully override namespace key via `KANJI_CONTEXT`
- Record namespace source (`cwd` or env override)

**Dependencies**
- P1-02

**Acceptance criteria**
- helper returns namespace key + namespace source
- cwd-derived and env-override cases are unit-tested
- normalization behavior is deterministic and documented in tests

---

### Issue 4 — Add shared namespaced state store

**ID**: P1-04  
**Title**: `feat(cli): add shared namespace state store for cli_context and tui_state`

**Goal**
- Implement shared persisted state store keyed by namespace
- Store separate state for:
  - `cli_context`
  - `tui_state`

**Dependencies**
- P1-03

**Acceptance criteria**
- state store can load/save multiple namespaces
- missing file is handled safely
- state round-trip is tested
- namespace isolation is tested

---

### Issue 5 — Add CLI runtime/service wiring helper

**ID**: P1-07  
**Title**: `refactor(cli): centralize db and service runtime wiring`

**Goal**
- Create reusable CLI runtime helper to open DB, construct store, repos, and services
- Avoid per-command wiring duplication

**Dependencies**
- P1-02

**Acceptance criteria**
- command handlers can reuse one runtime/service builder
- runtime helper supports DB path from resolved config
- helper is isolated from Cobra-specific parsing concerns
- targeted tests exist where practical

---

### Issue 6 — Add explicit bootstrap command

**ID**: P1-08  
**Title**: `feat(cli): add data bootstrap command`

**Goal**
- Add `kanji data bootstrap`
- Make it explicit and idempotent
- Return created-vs-existing summary in human and JSON modes

**Dependencies**
- P1-07

**Acceptance criteria**
- `kanji data bootstrap` succeeds on empty DB
- rerunning bootstrap is safe
- human output summarizes created/existing state
- JSON output is wrapped and structured
- integration tests cover first run and rerun

---

### Issue 7 — Add bootstrap guard for domain commands

**ID**: P1-09  
**Title**: `feat(cli): add bootstrap guard with actionable errors`

**Goal**
- Add reusable guard for non-bootstrap domain commands
- Fail with actionable guidance instead of silently bootstrapping

**Dependencies**
- P1-08

**Acceptance criteria**
- guard exists as shared helper
- missing-bootstrap error includes actionable next step
- commands can reuse the guard consistently
- tests cover error code/message shape

---

### Issue 8 — Add topic help commands

**ID**: P1-14  
**Title**: `feat(cli): add topic help commands for concepts context selectors output`

**Goal**
- Add:
  - `kanji help concepts`
  - `kanji help context`
  - `kanji help selectors`
  - `kanji help output`

**Dependencies**
- P1-01

**Acceptance criteria**
- all topic commands are routed and rendered
- top-level help references topic help
- tests cover command presence and key content anchors
- help content is curated, not placeholder text

---

### Issue 9 — Add shared selector validation skeleton

**ID**: P1-15  
**Title**: `feat(cli): add centralized selector validation and error model`

**Goal**
- Create shared selector/runtime layer for resource resolution
- Establish reusable error types:
  - mismatch
  - ambiguity
  - not found
  - validation

**Dependencies**
- P1-07
- P1-09

**Acceptance criteria**
- shared selector package exists under `cmd/kanji/internal/...`
- exact normalized matching rules are implemented and tested
- mismatch/ambiguity/not-found error shapes are reusable
- command handlers can consume selector helpers instead of hand-rolling logic

---

### Issue 10 — Add shared output rendering helpers

**ID**: P1-17  
**Title**: `feat(cli): add shared human and json output renderers`

**Goal**
- Standardize output rendering across commands
- Support:
  - human tables
  - human key/value blocks
  - wrapped JSON success payloads
  - wrapped JSON errors

**Dependencies**
- P1-02

**Acceptance criteria**
- renderer helpers exist and are reusable
- wrapped list/singular/error payloads are tested
- table rendering supports ID columns by default
- JSON contract matches `docs/prd/cli-v1.md`

---

### Issue 11 — Add context show command

**ID**: P1-05  
**Title**: `feat(cli): add context show command`

**Goal**
- Display namespace metadata and effective context
- Show TUI state secondarily

**Dependencies**
- P1-04
- P1-17

**Acceptance criteria**
- `kanji context show` supports human output
- `kanji context show --json` supports wrapped JSON output
- output includes namespace key and namespace source
- tests cover empty and populated context state

---

### Issue 12 — Add context clear command

**ID**: P1-06  
**Title**: `feat(cli): add context clear command`

**Goal**
- Clear only `cli_context`
- Preserve `tui_state`

**Dependencies**
- P1-05

**Acceptance criteria**
- `kanji context clear` clears only explicit CLI context
- human and JSON outputs are supported
- tests prove `tui_state` is preserved unchanged

---

### Issue 13 — Add context set command

**ID**: P1-16  
**Title**: `feat(cli): add context set command with immediate validation`

**Goal**
- Add `kanji context set`
- Support workspace and board selectors by ID or exact normalized name
- Workspace-only set clears board

**Dependencies**
- P1-05
- P1-09
- P1-15

**Acceptance criteria**
- `context set` validates selectors immediately
- board-name resolution requires workspace scope
- workspace-only set clears board context
- human and JSON outputs are supported
- tests cover valid, invalid, mismatch, and board-without-workspace-name cases

---

### Issue 14 — Add db info command

**ID**: P1-10  
**Title**: `feat(cli): add db info command`

**Goal**
- Add operational introspection for:
  - effective DB path
  - DB existence
  - namespace
  - bootstrap status

**Dependencies**
- P1-07
- P1-08
- P1-17

**Acceptance criteria**
- `kanji db info` supports human and JSON outputs
- reports DB path, existence, namespace, and bootstrap status
- tests cover missing DB and bootstrapped DB cases

---

### Issue 15 — Add db migrate up command

**ID**: P1-11  
**Title**: `feat(cli): add db migrate up command`

**Goal**
- Replace old `--migrate` flag behavior with `kanji db migrate up`

**Dependencies**
- P1-07

**Acceptance criteria**
- `kanji db migrate up` runs forward migrations only
- integration test covers success path on temp DB
- help documents purpose and non-goals clearly

---

### Issue 16 — Add db migrate status command

**ID**: P1-12  
**Title**: `feat(cli): add db migrate status command`

**Goal**
- Expose migration state under the new CLI tree

**Dependencies**
- P1-11

**Acceptance criteria**
- `kanji db migrate status` supports human and JSON outputs
- integration tests cover migrated/unmigrated temp DB states

---

### Issue 17 — Add data seed command

**ID**: P1-13  
**Title**: `feat(cli): add data seed command for sample demo data`

**Goal**
- Add `kanji data seed`
- Position it explicitly as sample/demo-only and non-production
- Return best-effort idempotent summary

**Dependencies**
- P1-07
- P1-08
- P1-17

**Acceptance criteria**
- `kanji data seed` works after bootstrap
- help explicitly calls seed non-production sample data
- summary output works in human and JSON modes
- rerun behavior is tested and documented as best-effort idempotent

---

### Issue 18 — Add db doctor command

**ID**: P1-28  
**Title**: `feat(cli): add read-only db doctor command`

**Goal**
- Add `kanji db doctor`
- Return findings with standard human/JSON output
- Exit `1` when findings exist

**Dependencies**
- P1-10
- P1-04
- P1-17

**Acceptance criteria**
- `db doctor` is read-only
- supports wrapped JSON findings
- human output summarizes findings clearly
- exits `1` when issues are found
- tests cover ok and failing cases

---

### Issue 19 — Add workspace list command

**ID**: P1-18  
**Title**: `feat(cli): add workspace list command`

**Goal**
- First domain read command
- Global list, no context required

**Dependencies**
- P1-07
- P1-09
- P1-17

**Acceptance criteria**
- `kanji workspace list` supports table output
- `kanji workspace list --json` returns wrapped list payload
- empty and populated DB cases are tested

---

### Issue 20 — Add workspace get command

**ID**: P1-19  
**Title**: `feat(cli): add workspace get command`

**Goal**
- Support get by ID or exact normalized name
- Optional `--include-boards`

**Dependencies**
- P1-15
- P1-18
- P1-17

**Acceptance criteria**
- `workspace get` supports ID and name selectors
- `--include-boards` works in human and JSON modes
- key/value output format is used for human mode
- tests cover not-found and selector behavior

---

### Issue 21 — Add board list command

**ID**: P1-20  
**Title**: `feat(cli): add board list command with explicit context inference`

**Goal**
- Add scoped board list command
- Require workspace via flags or `cli_context`

**Dependencies**
- P1-16
- P1-15
- P1-17

**Acceptance criteria**
- board list refuses missing scope with actionable error
- board list can infer workspace from `cli_context`
- human and JSON outputs are supported
- tests cover explicit scope and inferred scope

---

### Issue 22 — Add board get command

**ID**: P1-21  
**Title**: `feat(cli): add board get command`

**Goal**
- Support board lookup by ID or exact normalized name
- Optional `--include-columns`

**Dependencies**
- P1-20
- P1-17

**Acceptance criteria**
- board-name resolution requires workspace scope
- `--include-columns` is supported
- human output uses key/value detail block
- JSON output is wrapped and includes expansions when requested

---

### Issue 23 — Add column list command

**ID**: P1-22  
**Title**: `feat(cli): add column list command`

**Goal**
- Add scoped column list command
- Require board via flags or `cli_context`

**Dependencies**
- P1-16
- P1-15
- P1-17

**Acceptance criteria**
- column list refuses missing board scope
- human table includes:
  - ID
  - Name
  - Color
  - Position
  - WIP Limit
- explicit and inferred scope paths are tested

---

### Issue 24 — Add column get command

**ID**: P1-23  
**Title**: `feat(cli): add column get command`

**Goal**
- Support column get by ID or exact normalized name
- No expansion flags in v1

**Dependencies**
- P1-22
- P1-17

**Acceptance criteria**
- board scope is required for name-based column resolution
- human output uses key/value detail block
- JSON output is wrapped
- tests cover exact normalized name resolution and not-found cases

---

### Issue 25 — Add task list command

**ID**: P1-24  
**Title**: `feat(cli): add task list command with scoped filters`

**Goal**
- Add task list with workspace-required scope and board narrowing
- Expose selected v1 filter subset

**Dependencies**
- P1-16
- P1-15
- P1-17

**Acceptance criteria**
- task list requires at least workspace scope
- board scope is optional narrowing
- filters implemented:
  - `--query`
  - `--column`
  - `--priority`
  - `--due-soon`
  - `--overdue`
  - `--no-due-date`
  - `--sort`
- due flags are validated for conflicts
- human table and wrapped JSON outputs are tested

---

### Issue 26 — Add task get command

**ID**: P1-25  
**Title**: `feat(cli): add task get command with optional comment expansion`

**Goal**
- Add task get by ID or strongly scoped title
- Support optional `--include-comments`

**Dependencies**
- P1-24
- P1-17

**Acceptance criteria**
- default output returns task only
- `--include-comments` appends/embeds comments appropriately
- task-title resolution is exact and strongly scoped
- ambiguity and not-found behavior are tested

---

### Issue 27 — Add comment list command

**ID**: P1-26  
**Title**: `feat(cli): add comment list command`

**Goal**
- Add list-by-task comment read command

**Dependencies**
- P1-15
- P1-17
- P1-25

**Acceptance criteria**
- comment list requires task selector
- human table includes:
  - ID
  - Task ID
  - Author
  - Created
  - Body preview
- wrapped JSON output is supported
- tests cover task-id path and selector resolution behavior

---

### Issue 28 — Add comment get command

**ID**: P1-27  
**Title**: `feat(cli): add comment get command`

**Goal**
- Add `comment get --comment-id ...`

**Dependencies**
- P1-26
- may require minimal application/repository expansion if not already available

**Acceptance criteria**
- comment get supports ID-only selector
- human output uses key/value detail block
- wrapped JSON output is supported
- tests cover not-found case

---

### Issue 29 — Add tui command and runtime integration

**ID**: P1-29  
**Title**: `feat(cli): add tui command and wire shared runtime rules`

**Goal**
- Make `kanji tui` the official TUI launch path
- Reuse same runtime/config/context rules as CLI

**Dependencies**
- P1-07
- P1-16
- P1-04

**Acceptance criteria**
- `kanji tui` launches TUI
- TUI startup prefers `cli_context`, then `tui_state`
- existing TUI behavior remains supported
- smoke coverage exists where practical

---

### Issue 30 — Cut over from cmd/app to cmd/kanji

**ID**: P1-30  
**Title**: `refactor(cli): cut over entrypoint from cmd/app to cmd/kanji`

**Goal**
- Remove old generic entrypoint
- Make CLI-first binary canonical
- Update repo tooling/docs accordingly

**Dependencies**
- P1-29

**Acceptance criteria**
- `cmd/app` is removed
- Makefile points to `cmd/kanji`
- README examples stop referencing `cmd/app`
- no-arg help and `kanji tui` behavior are documented

---

### Issue 31 — Generate initial CLI docs and rewrite README orientation

**ID**: P1-31  
**Title**: `docs(cli): generate phase-1 CLI docs and make README CLI-first`

**Goal**
- Generate CLI reference docs under `docs/cli/`
- Rewrite README to reflect CLI-first model

**Dependencies**
- P1-14
- P1-18 through P1-29
- P1-30

**Acceptance criteria**
- `docs/cli/` exists with phase-1 command coverage
- README is CLI-first and TUI-secondary
- bootstrap-first workflow is prominent
- README links to generated CLI docs

---

## 5. Suggested First Five Sessions

1. **P1-01** — CLI root skeleton  
2. **P1-02** — global config/env plumbing  
3. **P1-03 + P1-04** — namespace + shared state store  
4. **P1-07 + P1-08 + P1-09** — runtime wiring + bootstrap + guard  
5. **P1-05 + P1-06 + P1-16** — context show/clear/set  

This sequence yields a real operational base before resource read commands expand.
