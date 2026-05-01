# Kanji CLI v1 Phase 3 Atomic Tasks and Sequenced Issue List

Status: Draft
Parent blueprint: `docs/prd/cli-v1.md`
Prerequisite: Phases 1 and 2 complete
Scope: Phase 3 only
Goal: Break Phase 3 into atomic one-session tasks and a sequenced issue list with titles and acceptance criteria

---

## 1. Phase 3 Goal

Deliver full CRUD completion, integrity hardening, migration support, and destructive workflow safety.

Phase 3 includes:

- missing CRUD completion:
  - `workspace delete`
  - `board delete`
  - `column delete`
  - `comment update`
  - `comment delete`
- application/repository/query expansion needed to support those commands cleanly
- DB integrity hardening for selector-safe uniqueness invariants
- migration upgrade handling
- `db doctor` expansion for integrity diagnostics tied to the new invariants and context hygiene
- automatic cleanup of invalid `cli_context` and `tui_state` after destructive deletes

Phase 3 must not drift into Phase 4 polish work except where a task requires minimal help/docs updates for correctness.

---

## 2. Phase 3 Assumptions

Assume previous phases are complete and available:

- CLI root/help/topics exist
- runtime/config/context store exists
- selector validation/output rendering layers exist
- read commands exist for all first-class resources
- write commands from Phase 2 exist for:
  - workspace create/update
  - board create/update
  - column create/update/reorder
  - task create/update/move/delete
  - comment create
- generated docs pipeline exists
- `kanji tui` works through the CLI runtime

---

## 3. Atomic Task Checklist

Each task is intended to fit in one coding session.

Conventions:
- TDD first
- prefer vertical slices
- each task should end with:
  - `gofmt`
  - relevant targeted tests
  - ideally `go test ./...`

### Integrity and migration foundations

- [ ] **P3-01** Add duplicate-name diagnostic queries and helpers
- [ ] **P3-02** Add migration preflight/upgrade handling helpers for selector-safe constraints
- [ ] **P3-03** Add migrations for case-insensitive workspace, board, and column uniqueness
- [ ] **P3-04** Expand `db doctor` for selector-hostile data and migration blockers

### Comment CRUD completion

- [ ] **P3-05** Expand application and repository support for comment update/delete
- [ ] **P3-06** Add `comment update`
- [ ] **P3-07** Add `comment delete`

### Workspace delete

- [ ] **P3-08** Add workspace cascade impact summary use case and queries
- [ ] **P3-09** Add namespace state sanitization helper for workspace deletion
- [ ] **P3-10** Add `workspace delete`

### Board delete

- [ ] **P3-11** Add board cascade impact summary use case and queries
- [ ] **P3-12** Add namespace state sanitization helper for board deletion
- [ ] **P3-13** Add `board delete`

### Column delete

- [ ] **P3-14** Add column occupancy/count helpers and reassignment use case
- [ ] **P3-15** Add `column delete`

### Hardening / contract completion

- [ ] **P3-16** Add destructive command dry-run rendering and JSON result helpers
- [ ] **P3-17** Add Phase 3 help/docs updates for destructive command semantics
- [ ] **P3-18** Update generated CLI docs and README references for full CRUD completion

---

## 4. Recommended Execution Order

1. P3-01 — duplicate-name diagnostic queries and helpers
2. P3-02 — migration preflight/upgrade handling helpers
3. P3-03 — uniqueness migrations
4. P3-04 — doctor expansion
5. P3-05 — comment update/delete app+repo support
6. P3-06 — `comment update`
7. P3-07 — `comment delete`
8. P3-08 — workspace cascade impact summary use case and queries
9. P3-09 — workspace deletion namespace sanitization helper
10. P3-10 — `workspace delete`
11. P3-11 — board cascade impact summary use case and queries
12. P3-12 — board deletion namespace sanitization helper
13. P3-13 — `board delete`
14. P3-14 — column occupancy/reassignment use case
15. P3-15 — `column delete`
16. P3-16 — destructive dry-run rendering + JSON result helpers
17. P3-17 — help/docs hardening for destructive semantics
18. P3-18 — generated docs/README update

---

## 5. Sequenced Issue List

---

### Issue 1 — Add duplicate-name diagnostic queries and helpers

**ID**: P3-01  
**Title**: `feat(cli): add diagnostic queries for selector-hostile duplicate names`

**Goal**
- Detect duplicate resource names that break selector assumptions:
  - workspace names globally
  - board names within workspace
  - column names within board
- Use normalized matching rules:
  - trim surrounding whitespace
  - case-insensitive exact comparison

**Dependencies**
- Phases 1 and 2 complete

**Acceptance criteria**
- diagnostic query helpers exist for all three uniqueness scopes
- results surface enough metadata to explain conflicts clearly
- tests cover duplicate and non-duplicate cases
- helpers are reusable by both migrations and `db doctor`

---

### Issue 2 — Add migration preflight / upgrade handling helpers

**ID**: P3-02  
**Title**: `feat(cli): add migration preflight and upgrade handling for integrity hardening`

**Goal**
- Introduce explicit upgrade handling around selector-safe constraints
- fail clearly when legacy data blocks constraint application

**Dependencies**
- P3-01

**Acceptance criteria**
- migration preflight helper exists or migration failure path is explicitly wrapped
- duplicate-name conflicts produce actionable errors
- tests cover blocked-upgrade scenarios on temp DB fixtures
- error messages suggest `kanji db doctor` or concrete next steps

---

### Issue 3 — Add uniqueness migrations

**ID**: P3-03  
**Title**: `feat(db): add case-insensitive uniqueness constraints for workspace board and column names`

**Goal**
- Enforce selector-safe invariants in the DB
- add case-insensitive normalized uniqueness constraints/indexes for:
  - workspaces globally
  - boards within workspace
  - columns within board

**Dependencies**
- P3-02

**Acceptance criteria**
- forward migrations add uniqueness enforcement using normalized comparisons
- migrations succeed on clean/valid datasets
- migrations fail clearly on invalid legacy datasets
- tests cover valid inserts and duplicate rejection cases

---

### Issue 4 — Expand db doctor for integrity and migration blockers

**ID**: P3-04  
**Title**: `feat(cli): expand db doctor for uniqueness conflicts and migration blockers`

**Goal**
- Make `kanji db doctor` report:
  - selector-hostile duplicate names
  - missing/incomplete bootstrap state
  - migration blocker conditions related to new constraints
  - dangling namespace context references if present

**Dependencies**
- P3-01
- P3-03

**Acceptance criteria**
- `db doctor` returns findings for duplicate-name conflicts
- wrapped JSON findings include stable codes and details
- human output is concise and actionable
- exit code remains `1` when findings exist
- tests cover success and multiple-finding scenarios

---

### Issue 5 — Expand application and repository support for comment update/delete

**ID**: P3-05  
**Title**: `feat(app): add comment update and delete use cases with repository support`

**Goal**
- Expand application layer and repositories to support comment body mutation and deletion
- keep CLI business logic out of repositories directly

**Dependencies**
- previous phases complete

**Acceptance criteria**
- application-layer use cases/methods exist for comment update and delete
- repository interface and infrastructure support required operations
- sqlc/query additions exist as needed
- tests cover happy path and not-found/error behavior

---

### Issue 6 — Add comment update command

**ID**: P3-06  
**Title**: `feat(cli): add comment update command`

**Goal**
- Add `kanji comment update`
- patch-style but practically whole-body replacement
- support inline/file/stdin body sources

**Dependencies**
- P3-05
- Phase 2 rich text parsing helpers

**Acceptance criteria**
- `comment update --comment-id ...` exists
- body source mutual exclusivity enforced
- human and JSON outputs are tested
- output uses wrapped singular result
- help/examples reflect body replacement semantics

---

### Issue 7 — Add comment delete command

**ID**: P3-07  
**Title**: `feat(cli): add comment delete command`

**Goal**
- Add `kanji comment delete`
- require explicit confirmation

**Dependencies**
- P3-05

**Acceptance criteria**
- `comment delete --comment-id ... --yes` works
- missing `--yes` fails with actionable guidance
- human and JSON outputs are tested
- delete result shape matches CLI contract

---

### Issue 8 — Add workspace cascade impact summary use case and queries

**ID**: P3-08  
**Title**: `feat(app): add workspace delete impact summary and cascade execution support`

**Goal**
- Add use case support for workspace deletion with:
  - dry-run impact summary
  - cascade execution
- impact must include counts for:
  - workspace
  - boards
  - columns
  - tasks
  - comments

**Dependencies**
- P3-03

**Acceptance criteria**
- application-layer use case exists for workspace delete preview and execution
- query support exists for impact counting
- tests cover preview and execution paths on temp DB
- behavior is explicit and non-interactive

---

### Issue 9 — Add workspace deletion namespace sanitization helper

**ID**: P3-09  
**Title**: `feat(cli): sanitize cli_context and tui_state after workspace deletion`

**Goal**
- After deleting a workspace, clear invalid namespace state in the shared store:
  - clear workspace
  - clear board
- affect both `cli_context` and `tui_state`

**Dependencies**
- P3-08
- Phase 1 shared context store

**Acceptance criteria**
- shared helper exists for post-workspace-delete state cleanup
- cleanup affects both context semantics correctly
- tests cover active and inactive namespace cases

---

### Issue 10 — Add workspace delete command

**ID**: P3-10  
**Title**: `feat(cli): add workspace delete command with cascade and dry-run`

**Goal**
- Add `kanji workspace delete`
- require `--yes --cascade`
- support `--dry-run`
- clear invalid namespace state automatically after real delete

**Dependencies**
- P3-08
- P3-09
- destructive render helpers or existing output layer

**Acceptance criteria**
- dry-run reports impact summary in human and JSON modes
- real delete requires `--yes --cascade`
- real delete sanitizes `cli_context` and `tui_state`
- human output mentions context/state adjustments when relevant
- tests cover dry-run, real delete, and missing-flag failure cases

---

### Issue 11 — Add board cascade impact summary use case and queries

**ID**: P3-11  
**Title**: `feat(app): add board delete impact summary and cascade execution support`

**Goal**
- Add use case support for board deletion with:
  - dry-run impact summary
  - cascade execution
- impact must include counts for:
  - board
  - columns
  - tasks
  - comments

**Dependencies**
- P3-03

**Acceptance criteria**
- application-layer use case exists for board delete preview and execution
- query support exists for impact counting
- tests cover preview and execution behavior on temp DB

---

### Issue 12 — Add board deletion namespace sanitization helper

**ID**: P3-12  
**Title**: `feat(cli): sanitize cli_context and tui_state after board deletion`

**Goal**
- After deleting a board, clear invalid board selections while preserving valid workspace selection
- affect both `cli_context` and `tui_state`

**Dependencies**
- P3-11
- Phase 1 shared context store

**Acceptance criteria**
- helper clears board only and preserves valid workspace when appropriate
- affects both CLI and TUI state representations
- tests cover selected-board and unrelated-board cases

---

### Issue 13 — Add board delete command

**ID**: P3-13  
**Title**: `feat(cli): add board delete command with cascade and dry-run`

**Goal**
- Add `kanji board delete`
- require `--yes --cascade`
- support `--dry-run`
- sanitize invalid namespace state after real delete

**Dependencies**
- P3-11
- P3-12

**Acceptance criteria**
- dry-run returns impact summary in human and JSON outputs
- real delete requires `--yes --cascade`
- post-delete state cleanup occurs automatically
- tests cover dry-run, real delete, and missing-flag error paths

---

### Issue 14 — Add column occupancy / reassignment use case

**ID**: P3-14  
**Title**: `feat(app): add column delete occupancy checks and reassignment path`

**Goal**
- Add application support for column deletion semantics:
  - refuse delete if tasks exist by default
  - allow reassignment via `--move-tasks-to`
  - no task cascade in v1
- support both previewable validation and execution

**Dependencies**
- P3-03

**Acceptance criteria**
- use case exists to inspect whether column contains tasks
- use case supports reassignment destination validation within board scope
- repository/query support exists as needed
- tests cover empty column, occupied column, and reassignment flow

---

### Issue 15 — Add column delete command

**ID**: P3-15  
**Title**: `feat(cli): add column delete command with reassignment-driven semantics`

**Goal**
- Add `kanji column delete`
- require `--yes`
- refuse occupied-column delete unless reassignment target is supplied

**Dependencies**
- P3-14

**Acceptance criteria**
- `column delete` supports selectors by ID or exact normalized name
- missing `--yes` fails clearly
- occupied-column delete without reassignment fails clearly
- reassignment path succeeds and preserves tasks
- human and JSON outputs are tested

---

### Issue 16 — Add destructive dry-run rendering and JSON result helpers

**ID**: P3-16  
**Title**: `feat(cli): add shared rendering helpers for dry-run impacts and delete results`

**Goal**
- Standardize destructive command output for:
  - dry-run impact summaries
  - real delete summaries
  - JSON result shapes for deleted resources and affected counts

**Dependencies**
- P3-08
- P3-11
- existing output layer

**Acceptance criteria**
- shared render helpers exist for dry-run and destructive success paths
- JSON shapes match the blueprint contract
- human output includes impact summaries where relevant
- representative tests exist for workspace and board deletes

---

### Issue 17 — Add Phase 3 help/docs hardening for destructive semantics

**ID**: P3-17  
**Title**: `docs(cli): harden help and examples for full CRUD and destructive semantics`

**Goal**
- Update command help for newly added destructive and comment CRUD commands
- emphasize:
  - `--yes`
  - `--cascade`
  - `--dry-run`
  - reassignment semantics for column delete

**Dependencies**
- P3-06
- P3-07
- P3-10
- P3-13
- P3-15

**Acceptance criteria**
- each new Phase 3 command has curated long help
- examples include both human and agent/script forms
- selector and destructive reminders are correct and consistent
- representative help content anchors are tested where practical

---

### Issue 18 — Update generated CLI docs and README references for full CRUD completion

**ID**: P3-18  
**Title**: `docs(cli): update generated docs and README for full CRUD completion`

**Goal**
- Extend generated docs under `docs/cli/`
- reflect full CRUD completion for v1 command surface
- update README references where phase-3 command surface materially changes user guidance

**Dependencies**
- P3-17

**Acceptance criteria**
- generated docs include Phase 3 command surface and examples
- README or linked docs explain destructive workflows and doctor usage
- docs remain aligned with `docs/prd/cli-v1.md`

---

## 6. Suggested First Five Sessions

1. **P3-01** — duplicate-name diagnostic helpers  
2. **P3-02** — migration preflight / upgrade handling  
3. **P3-03** — uniqueness migrations  
4. **P3-04** — `db doctor` expansion  
5. **P3-05 + P3-06** — comment update/delete app support + command  

This order lands the schema/integrity backbone first, then starts the smallest missing CRUD family before higher-risk deletes.
