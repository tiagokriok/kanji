# Kanji CLI v1 Phase 2 Atomic Tasks and Sequenced Issue List

Status: Draft
Parent blueprint: `docs/prd/cli-v1.md`
Prerequisite: Phase 1 complete
Scope: Phase 2 only
Goal: Break Phase 2 into atomic one-session tasks and a sequenced issue list with titles and acceptance criteria

---

## 1. Phase 2 Goal

Deliver write commands built primarily on existing application capabilities, while staying strictly within the bounded Phase 2 scope.

Phase 2 includes:

- `workspace create`
- `workspace update`
- `board create`
- `board update`
- `column create`
- `column update`
- `column reorder`
- `task create`
- `task update`
- `task move`
- `task delete`
- `comment create`

Phase 2 may add:
- small CLI-runtime helpers
- small application-layer helpers/use-case adapters
- shared input parsing/validation helpers
- shared output refinements for write flows

Phase 2 must not pull in Phase 3 scope:
- workspace delete
- board delete
- column delete
- comment update/delete
- schema hardening migrations
- uniqueness constraints
- expanded doctor checks tied to new constraints

---

## 2. Phase 2 Assumptions

Assume Phase 1 is complete and available:

- CLI root and command tree exist
- global flags/env/config exist
- namespaced `cli_context` / `tui_state` store exists
- selector validation layer exists
- output rendering helpers exist
- bootstrap guard exists
- read commands exist for all first-class resources
- `kanji tui` works via CLI runtime
- topic help and generated docs framework exist

---

## 3. Atomic Task Checklist

Each task is intended to fit in one coding session.

Conventions:
- TDD first
- prefer minimal vertical slices
- each task should end with:
  - `gofmt`
  - relevant targeted tests
  - ideally `go test ./...`

### Shared write foundations

- [ ] **P2-01** Add write-command shared validation helpers
- [ ] **P2-02** Add rich text input source parsing helpers
- [ ] **P2-03** Add priority parsing helper
- [ ] **P2-04** Add due-date parsing helper
- [ ] **P2-05** Add write-scope resolution helpers
- [ ] **P2-06** Add standardized write success/result render helpers

### Workspace / board writes

- [ ] **P2-07** Add `workspace create`
- [ ] **P2-08** Add `workspace update`
- [ ] **P2-09** Add board column spec parser for `board create`
- [ ] **P2-10** Add `board create`
- [ ] **P2-11** Add `board update`

### Column writes

- [ ] **P2-12** Add deterministic default color helper for standalone column create
- [ ] **P2-13** Add `column create`
- [ ] **P2-14** Add `column update`
- [ ] **P2-15** Add `column reorder`

### Task writes

- [ ] **P2-16** Add task-create input assembly helpers
- [ ] **P2-17** Add `task create`
- [ ] **P2-18** Add task-update patch assembly helpers
- [ ] **P2-19** Add `task update`
- [ ] **P2-20** Add task destination-column resolution helpers for moves
- [ ] **P2-21** Add `task move`
- [ ] **P2-22** Add `task delete`

### Comment writes

- [ ] **P2-23** Add `comment create`

### Help/docs hardening for phase 2 commands

- [ ] **P2-24** Add curated help/examples for all phase 2 write commands
- [ ] **P2-25** Update generated CLI docs and README examples for phase 2 write flows

---

## 4. Recommended Execution Order

1. P2-01 — write-command shared validation helpers
2. P2-02 — rich text input source parsing helpers
3. P2-03 — priority parsing helper
4. P2-04 — due-date parsing helper
5. P2-05 — write-scope resolution helpers
6. P2-06 — write result render helpers
7. P2-07 — `workspace create`
8. P2-08 — `workspace update`
9. P2-09 — board column spec parser
10. P2-10 — `board create`
11. P2-11 — `board update`
12. P2-12 — deterministic column default color helper
13. P2-13 — `column create`
14. P2-14 — `column update`
15. P2-15 — `column reorder`
16. P2-16 — task create input assembly helpers
17. P2-17 — `task create`
18. P2-18 — task update patch assembly helpers
19. P2-19 — `task update`
20. P2-20 — task move destination resolution helpers
21. P2-21 — `task move`
22. P2-22 — `task delete`
23. P2-23 — `comment create`
24. P2-24 — curated help/examples hardening
25. P2-25 — docs/README update for phase 2

---

## 5. Sequenced Issue List

---

### Issue 1 — Add shared write-command validation helpers

**ID**: P2-01  
**Title**: `feat(cli): add shared validation helpers for write command conflicts and requirements`

**Goal**
- Centralize recurring write-command validation patterns:
  - required mutation presence
  - mutually exclusive clear/value flags
  - write confirmation preconditions where applicable later
  - repeated-label normalization helpers where useful

**Dependencies**
- Phase 1 complete

**Acceptance criteria**
- shared helper layer exists for write-command validation
- conflict cases can be reused across resource commands
- tests cover representative invalid combinations
- command handlers no longer hand-roll identical conflict logic

---

### Issue 2 — Add rich text input source parsing helpers

**ID**: P2-02  
**Title**: `feat(cli): add shared rich text source parsing for inline file and stdin`

**Goal**
- Parse rich text field inputs from:
  - inline flag
  - `--*-file <path>`
  - `--*-file -` for stdin
- enforce strict mutual exclusivity and fail-fast behavior

**Dependencies**
- Phase 1 complete

**Acceptance criteria**
- shared helper exists for task descriptions and comment bodies
- invalid combinations fail clearly
- stdin path is testable
- tests cover inline/file/stdin and conflict cases

---

### Issue 3 — Add priority parsing helper

**ID**: P2-03  
**Title**: `feat(cli): add shared priority parsing for numeric and label forms`

**Goal**
- Support both numeric and label-based priority inputs
- normalize to canonical internal numeric values

**Dependencies**
- Phase 1 complete

**Acceptance criteria**
- helper accepts numeric and known label forms
- exact mapping is tested
- invalid priorities fail with actionable errors

---

### Issue 4 — Add due-date parsing helper

**ID**: P2-04  
**Title**: `feat(cli): add shared due-date parsing for date and RFC3339 inputs`

**Goal**
- Support narrow, documented due-date inputs:
  - `YYYY-MM-DD`
  - RFC3339
- date-only values resolve to end-of-day semantics

**Dependencies**
- Phase 1 complete

**Acceptance criteria**
- helper parses date-only and RFC3339 inputs
- date-only resolves to end-of-day behavior consistently
- invalid inputs fail clearly
- tests cover timezone-safe expectations where practical

---

### Issue 5 — Add write-scope resolution helpers

**ID**: P2-05  
**Title**: `feat(cli): add write scope resolution from explicit selectors or explicit cli_context`

**Goal**
- Centralize write-scope resolution rules:
  - explicit flags win
  - explicit CLI context may satisfy missing write scope
  - missing required write scope fails clearly
- implement the refined rule that writes may rely on explicit `kanji context set`

**Dependencies**
- Phase 1 context store and selector layer

**Acceptance criteria**
- helpers exist for workspace/board write scope resolution
- helpers distinguish explicit CLI context from passive TUI state
- tests cover explicit flags, context-backed scope, and missing-scope failures

---

### Issue 6 — Add write success/result render helpers

**ID**: P2-06  
**Title**: `feat(cli): add shared render helpers for write command summaries and resolved metadata`

**Goal**
- Standardize success outputs for create/update/move/delete commands
- support optional `resolved` metadata blocks in JSON
- support concise human summaries with contextual notes when relevant

**Dependencies**
- Phase 1 output helpers

**Acceptance criteria**
- shared helpers exist for singular write results
- JSON output supports canonical object + optional `resolved`
- human output can mention context resolution when relevant
- representative tests exist

---

### Issue 7 — Add workspace create

**ID**: P2-07  
**Title**: `feat(cli): add workspace create command with smart default board`

**Goal**
- Add `kanji workspace create`
- create workspace and auto-create default board `Main`
- optionally `--set-context`

**Dependencies**
- P2-05
- P2-06
- bootstrap guard from phase 1

**Acceptance criteria**
- command requires bootstrap first
- `--name` required
- default board `Main` is created
- output treats workspace as primary resource and includes created board
- `--set-context` sets workspace + created `Main` board
- human and JSON outputs are tested

---

### Issue 8 — Add workspace update

**ID**: P2-08  
**Title**: `feat(cli): add workspace update command`

**Goal**
- Add `kanji workspace update`
- patch-style, effectively rename in v1

**Dependencies**
- P2-06
- selector layer from phase 1

**Acceptance criteria**
- supports `--workspace-id` and `--workspace`
- requires `--name`
- uses patch/update semantics in command contract
- human and JSON outputs are tested

---

### Issue 9 — Add board column spec parser

**ID**: P2-09  
**Title**: `feat(cli): add repeated board column spec parser for board create`

**Goal**
- Parse repeated board column specs in the form:
  - `--column "Name:#RRGGBB"`
- validate color and non-empty names

**Dependencies**
- Phase 1 foundation

**Acceptance criteria**
- parser exists as shared helper
- repeated specs preserve order
- invalid shapes fail clearly
- tests cover valid and invalid inputs

---

### Issue 10 — Add board create

**ID**: P2-10  
**Title**: `feat(cli): add board create command with smart defaults and custom columns`

**Goal**
- Add `kanji board create`
- preserve smart defaults when no custom columns are provided
- use only provided columns when any custom columns are supplied
- support `--set-context`

**Dependencies**
- P2-05
- P2-06
- P2-09

**Acceptance criteria**
- workspace scope required via flags or explicit context
- `--name` required
- repeated custom `--column` supported
- no custom columns => default columns created
- any custom columns => only provided columns created
- `--set-context` sets current board context
- tests cover default and custom column flows

---

### Issue 11 — Add board update

**ID**: P2-11  
**Title**: `feat(cli): add board update command`

**Goal**
- Add `kanji board update`
- expose name mutation only in v1

**Dependencies**
- P2-06
- selector layer from phase 1

**Acceptance criteria**
- supports board selector by ID or exact normalized name
- board-name resolution uses workspace scope where required
- requires `--name`
- human and JSON outputs are tested

---

### Issue 12 — Add deterministic default color helper for standalone column create

**ID**: P2-12  
**Title**: `feat(cli): add deterministic palette-based default color helper for column create`

**Goal**
- Reuse default board palette logic for standalone column creation
- choose next palette color deterministically

**Dependencies**
- existing app/domain behavior knowledge from current codebase

**Acceptance criteria**
- helper returns deterministic next color based on board column order/position
- helper is covered by tests
- no fixed-gray fallback is used for the common standalone path when palette logic applies

---

### Issue 13 — Add column create

**ID**: P2-13  
**Title**: `feat(cli): add column create command`

**Goal**
- Add `kanji column create`
- support optional `--color` and `--wip-limit`
- apply deterministic default color if `--color` omitted

**Dependencies**
- P2-05
- P2-06
- P2-12

**Acceptance criteria**
- board scope required via flags or explicit context
- `--name` required
- `--color` optional
- `--wip-limit` optional
- omitted color uses deterministic palette helper
- human and JSON outputs are tested

---

### Issue 14 — Add column update

**ID**: P2-14  
**Title**: `feat(cli): add column update command`

**Goal**
- Add `kanji column update`
- patch-style metadata update only

**Dependencies**
- P2-06
- selector layer from phase 1

**Acceptance criteria**
- supports selectors by ID or exact normalized name
- supports:
  - `--name`
  - `--color`
  - `--wip-limit`
  - `--clear-wip-limit`
- no ordering fields are exposed
- conflicting flag combinations fail clearly
- tests cover patch and clear behavior

---

### Issue 15 — Add column reorder

**ID**: P2-15  
**Title**: `feat(cli): add column reorder command with full ordered selector set`

**Goal**
- Add `kanji column reorder`
- require full ordered list using repeated ordered selectors

**Dependencies**
- selector layer from phase 1
- output helpers

**Acceptance criteria**
- board scope required
- repeated `--column` or `--column-id` preserves order
- all current board columns must appear exactly once
- duplicates and omissions fail clearly
- human and JSON outputs are tested

---

### Issue 16 — Add task create input assembly helpers

**ID**: P2-16  
**Title**: `feat(cli): add task create input assembly helpers`

**Goal**
- Assemble `task create` application inputs from flags/context
- support title, description, labels, priority, due date, assignee, estimate, and column-driven status resolution

**Dependencies**
- P2-02
- P2-03
- P2-04
- P2-05

**Acceptance criteria**
- helper supports all approved create fields
- repeated `--label` becomes canonical label slice
- board required unless explicit context provides it
- column omission defaults to first board column
- tests cover explicit and context-backed create assembly

---

### Issue 17 — Add task create

**ID**: P2-17  
**Title**: `feat(cli): add task create command`

**Goal**
- Add `kanji task create`
- preserve column-driven workflow semantics

**Dependencies**
- P2-06
- P2-16

**Acceptance criteria**
- requires bootstrap first
- `--title` required
- board required unless explicit context already provides it
- `--column` / `--column-id` optional and resolved within board
- no raw `--status` flag exposed
- human and JSON outputs are tested, including `resolved` metadata

---

### Issue 18 — Add task update patch assembly helpers

**ID**: P2-18  
**Title**: `feat(cli): add task update patch assembly helpers`

**Goal**
- Build task metadata patch inputs from approved update flags
- handle clear flags and conflict validation

**Dependencies**
- P2-01
- P2-02
- P2-03
- P2-04

**Acceptance criteria**
- helper supports approved patch fields only
- workflow/column changes are not exposed here
- clear/value conflicts fail clearly
- tests cover patch and clear assembly behavior

---

### Issue 19 — Add task update

**ID**: P2-19  
**Title**: `feat(cli): add task update command`

**Goal**
- Add `kanji task update`
- patch-style metadata updates only

**Dependencies**
- P2-06
- P2-18
- selector layer from phase 1

**Acceptance criteria**
- supports task selectors by ID or strongly scoped title
- exposes approved patch fields only
- does not expose workflow/column move flags
- human and JSON outputs are tested

---

### Issue 20 — Add task move destination resolution helpers

**ID**: P2-20  
**Title**: `feat(cli): add destination column resolution helpers for task move`

**Goal**
- Resolve move destinations by ID or board-scoped exact normalized column name
- enforce column-driven move semantics

**Dependencies**
- P2-05
- selector layer from phase 1

**Acceptance criteria**
- helper resolves `--to-column-id` and `--to-column`
- board scope is required for name-based destination
- no raw status destination is supported
- tests cover destination resolution and invalid scope

---

### Issue 21 — Add task move

**ID**: P2-21  
**Title**: `feat(cli): add task move command`

**Goal**
- Add `kanji task move`
- make it the canonical workflow transition command

**Dependencies**
- P2-06
- P2-20

**Acceptance criteria**
- task selector supported by ID or strongly scoped title
- destination column supported by ID or board-scoped name
- output reflects move result and resolved metadata where useful
- tests cover explicit and context-backed flows

---

### Issue 22 — Add task delete

**ID**: P2-22  
**Title**: `feat(cli): add task delete command`

**Goal**
- Add `kanji task delete`
- require explicit confirmation
- rely on existing task delete semantics, including comment cascade behavior already supported by business logic/DB path

**Dependencies**
- P2-06
- selector layer from phase 1

**Acceptance criteria**
- requires `--yes`
- supports task selector by ID or strongly scoped title
- human and JSON outputs are tested
- error if confirmation flag missing is actionable

---

### Issue 23 — Add comment create

**ID**: P2-23  
**Title**: `feat(cli): add comment create command`

**Goal**
- Add `kanji comment create`
- support body from inline, file, or stdin
- optional author field

**Dependencies**
- P2-02
- P2-06
- selector layer from phase 1

**Acceptance criteria**
- task selector required
- body source required and mutually exclusive
- optional `--author` supported
- human and JSON outputs are tested
- tests cover inline/file/stdin sources

---

### Issue 24 — Add curated help/examples for phase 2 write commands

**ID**: P2-24  
**Title**: `docs(cli): add curated help and examples for phase 2 write commands`

**Goal**
- Harden command help for all newly added write commands
- include both human and agent/script-oriented examples

**Dependencies**
- all phase 2 write commands landed

**Acceptance criteria**
- each phase 2 write command has:
  - detailed long help
  - selector reminders
  - human example
  - agent/script example
- topic help references remain consistent
- representative help content anchors are tested where practical

---

### Issue 25 — Update generated CLI docs and README examples for phase 2

**ID**: P2-25  
**Title**: `docs(cli): update generated docs and README for phase 2 write workflows`

**Goal**
- Extend generated docs under `docs/cli/`
- add representative phase 2 write examples to README or linked docs

**Dependencies**
- P2-24

**Acceptance criteria**
- generated docs include phase 2 write commands
- README or linked docs demonstrate representative write workflows
- docs remain aligned with the blueprint contract

---

## 6. Suggested First Five Sessions

1. **P2-01** — write validation helpers  
2. **P2-02** — rich text source parsing  
3. **P2-03 + P2-04** — priority and due-date parsing  
4. **P2-05 + P2-06** — write scope + write result rendering  
5. **P2-07 + P2-08** — workspace create/update  

This keeps Phase 2 strongly vertical and lets the first write family land early.
