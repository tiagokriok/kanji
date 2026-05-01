# Kanji CLI v1 Phase 4 Atomic Tasks and Sequenced Issue List

Status: Draft
Parent blueprint: `docs/prd/cli-v1.md`
Prerequisite: Phases 1, 2, and 3 complete
Scope: Phase 4 only
Goal: Break Phase 4 into atomic one-session tasks and a sequenced issue list with titles and acceptance criteria

---

## 1. Phase 4 Goal

Deliver final hardening and polish for the CLI contract, with emphasis on:

- generated CLI docs completeness
- top-level didactic help quality
- examples for humans and agents
- consistency audit across commands
- JSON contract review and stabilization
- README finalization
- command/help/output coherence across the full v1 surface

Phase 4 should avoid introducing new product scope unless a polish task exposes a correctness bug that must be fixed to preserve the agreed contract.

---

## 2. Phase 4 Assumptions

Assume previous phases are complete and available:

- full Phase 1 command/runtime/context surface exists
- full Phase 2 write surface exists
- full Phase 3 CRUD completion/integrity/delete semantics exist
- generated docs pipeline exists
- README has already been partially updated during earlier phases
- `kanji tui` works and remains behavior-preserving

---

## 3. Atomic Task Checklist

Each task is intended to fit in one coding session.

Conventions:
- TDD where code changes are involved
- doc verification where doc generation/help behavior is involved
- each task should end with:
  - `gofmt` if Go files changed
  - relevant tests
  - ideally `go test ./...`

### Help hardening

- [ ] **P4-01** Audit and harden root didactic help content
- [ ] **P4-02** Audit and harden topic help content
- [ ] **P4-03** Audit per-command long help for required sections
- [ ] **P4-04** Add/normalize human and agent examples across command families

### Output contract hardening

- [ ] **P4-05** Audit wrapped JSON success payloads across all commands
- [ ] **P4-06** Audit wrapped JSON error payloads across all commands
- [ ] **P4-07** Audit human table outputs for stable columns and ID visibility
- [ ] **P4-08** Audit human key/value detail outputs for consistency
- [ ] **P4-09** Audit verbose output behavior across representative commands

### Docs generation and repo docs

- [ ] **P4-10** Generate and verify full CLI docs under `docs/cli/`
- [ ] **P4-11** Rewrite README as final CLI-first quickstart and product entrypoint doc
- [ ] **P4-12** Add docs consistency checks for examples and command references

### Final consistency and release hardening

- [ ] **P4-13** Audit selector terminology and flag naming consistency across all help/docs
- [ ] **P4-14** Audit destructive command messaging and dry-run guidance consistency
- [ ] **P4-15** Add representative integration snapshot/golden coverage for help and JSON contracts
- [ ] **P4-16** Final CLI v1 acceptance sweep against the blueprint

---

## 4. Recommended Execution Order

1. P4-01 — root didactic help audit
2. P4-02 — topic help audit
3. P4-03 — per-command long help audit
4. P4-04 — example normalization across command families
5. P4-05 — JSON success payload audit
6. P4-06 — JSON error payload audit
7. P4-07 — human table output audit
8. P4-08 — human key/value output audit
9. P4-09 — verbose output audit
10. P4-10 — full generated CLI docs under `docs/cli/`
11. P4-11 — final README rewrite
12. P4-12 — docs consistency checks
13. P4-13 — selector/flag terminology audit
14. P4-14 — destructive messaging audit
15. P4-15 — help and JSON contract golden coverage
16. P4-16 — final blueprint acceptance sweep

---

## 5. Sequenced Issue List

---

### Issue 1 — Audit and harden root didactic help content

**ID**: P4-01  
**Title**: `docs(cli): harden root help into final didactic entrypoint experience`

**Goal**
- Make `kanji` no-arg help a strong first-run experience for humans and agents
- ensure it teaches:
  - what kanji is
  - command grammar
  - bootstrap-first workflow
  - context model
  - selector model
  - output modes
  - destructive command policy
  - common workflows

**Dependencies**
- full CLI command tree complete

**Acceptance criteria**
- no-arg help is curated, not generic-only
- root help references topic help and major command families clearly
- examples include bootstrap-first setup and TUI launch
- tests or golden assertions cover key root-help anchors

---

### Issue 2 — Audit and harden topic help content

**ID**: P4-02  
**Title**: `docs(cli): finalize topic help for concepts context selectors and output`

**Goal**
- Make topic help complete, coherent, and aligned with final implementation

**Dependencies**
- P4-01

**Acceptance criteria**
- `help concepts`, `help context`, `help selectors`, and `help output` all reflect final semantics
- selector theory is centralized and consistent
- context docs clearly distinguish `cli_context` from `tui_state`
- tests or golden assertions cover key content anchors

---

### Issue 3 — Audit per-command long help sections

**ID**: P4-03  
**Title**: `docs(cli): audit all command help for required long-form sections`

**Goal**
- Ensure each major command help includes required sections:
  - purpose
  - selector rules relevant to that command
  - required/optional flags
  - output notes where needed
  - examples
  - related commands/topics where useful

**Dependencies**
- P4-02

**Acceptance criteria**
- all first-class and operational commands meet the curated help standard
- missing or inconsistent sections are fixed
- tests or docs checks validate presence of important help sections where practical

---

### Issue 4 — Normalize human and agent examples across command families

**ID**: P4-04  
**Title**: `docs(cli): normalize human and agent-oriented examples across all command families`

**Goal**
- Ensure examples exist and follow consistent patterns for:
  - humans
  - agents/scripts (`--json`, explicit selectors, or context-driven flows where appropriate)

**Dependencies**
- P4-03

**Acceptance criteria**
- each major command family has both human and agent-style examples
- examples use final command grammar and current runtime rules
- examples are consistent about selector styles and output flags

---

### Issue 5 — Audit wrapped JSON success payloads

**ID**: P4-05  
**Title**: `feat(cli): audit and harden wrapped JSON success payloads across all commands`

**Goal**
- Verify the JSON success contract is consistent across:
  - list commands
  - get commands
  - create/update/delete commands
  - move/reorder/doctor/bootstrap commands

**Dependencies**
- previous phases complete

**Acceptance criteria**
- all JSON success payloads are wrapped, not raw
- list outputs use resource-named wrappers and counts/scope where applicable
- singular outputs use wrapped named objects
- optional `resolved` blocks are present only where meaningful and implemented consistently
- representative integration/golden tests cover major families

---

### Issue 6 — Audit wrapped JSON error payloads

**ID**: P4-06  
**Title**: `feat(cli): audit and harden wrapped JSON error payloads`

**Goal**
- Ensure JSON errors are consistent across commands and failure modes

**Dependencies**
- P4-05

**Acceptance criteria**
- errors use wrapped `{ error: { code, message, details? } }` shape
- common error families are applied consistently:
  - usage
  - validation
  - not found
  - ambiguous selector
  - conflict
  - internal
- representative integration tests cover common failure types

---

### Issue 7 — Audit human table outputs

**ID**: P4-07  
**Title**: `feat(cli): audit human-readable list tables for stable columns and ID visibility`

**Goal**
- Verify all list commands use stable, readable tables
- ensure IDs are always present

**Dependencies**
- previous phases complete

**Acceptance criteria**
- each list command uses the agreed default column set
- IDs always appear in tables
- scope-aware omission of redundant parent columns is applied consistently
- representative output tests/goldens exist

---

### Issue 8 — Audit human key/value detail outputs

**ID**: P4-08  
**Title**: `feat(cli): audit human key-value detail outputs for singular commands`

**Goal**
- Make `get` and relevant singular outputs consistent and readable
- ensure multiline fields render clearly

**Dependencies**
- previous phases complete

**Acceptance criteria**
- singular human output uses key/value block style consistently
- descriptions/comments render clearly
- related scope metadata is shown consistently where relevant
- representative tests/goldens exist

---

### Issue 9 — Audit verbose output behavior

**ID**: P4-09  
**Title**: `feat(cli): audit verbose output behavior across representative commands`

**Goal**
- Verify `--verbose` remains narrow, useful, and consistent
- ensure it can reveal:
  - namespace key
  - namespace source
  - DB path
  - resolution details
  - operation counts where relevant

**Dependencies**
- previous phases complete

**Acceptance criteria**
- `--verbose` behaves consistently on representative read, write, and operational commands
- it does not alter JSON contract correctness
- representative tests verify verbose-only fields/content where practical

---

### Issue 10 — Generate and verify full CLI docs under docs/cli

**ID**: P4-10  
**Title**: `docs(cli): generate and verify full command reference under docs/cli`

**Goal**
- Generate complete Markdown CLI docs under `docs/cli/`
- ensure all first-class resources, operational commands, and topics are represented

**Dependencies**
- P4-01 through P4-09

**Acceptance criteria**
- docs tree exists and is complete for v1 command surface
- topic docs and command docs align with final implementation
- command references use final naming and examples
- generation process is documented and repeatable

---

### Issue 11 — Final README rewrite

**ID**: P4-11  
**Title**: `docs(cli): rewrite README as final CLI-first product entrypoint`

**Goal**
- Make README the final CLI-first entrypoint doc
- TUI becomes secondary and explicitly launched via `kanji tui`

**Dependencies**
- P4-10

**Acceptance criteria**
- README quickstart starts with CLI and bootstrap-first flow
- no-arg help behavior is documented
- `kanji tui` is documented as explicit TUI entrypoint
- README links to generated CLI docs and core help topics where appropriate

---

### Issue 12 — Add docs consistency checks

**ID**: P4-12  
**Title**: `docs(cli): add consistency checks for command refs and examples`

**Goal**
- Add lightweight checks ensuring docs/examples do not drift badly from final command tree and naming

**Dependencies**
- P4-10
- P4-11

**Acceptance criteria**
- there is a repeatable way to detect obvious stale command names/examples in docs
- checks cover core command tree and major examples where practical
- docs drift is reduced for future changes

---

### Issue 13 — Audit selector terminology and flag naming consistency

**ID**: P4-13  
**Title**: `docs(cli): audit selector terminology and flag naming across help docs and output`

**Goal**
- Verify aggressive selector naming consistency:
  - `--workspace-id`, `--workspace`
  - `--board-id`, `--board`
  - `--column-id`, `--column`
  - `--task-id`, `--task`
  - `--comment-id`
- ensure wording stays aligned everywhere

**Dependencies**
- P4-10

**Acceptance criteria**
- no inconsistent selector flag variants remain in help/docs/output examples
- selector wording aligns with `help selectors`
- audit findings are fixed, not just documented

---

### Issue 14 — Audit destructive messaging and dry-run guidance consistency

**ID**: P4-14  
**Title**: `docs(cli): audit destructive command messaging and dry-run guidance`

**Goal**
- Ensure destructive workflows are explained consistently across commands, docs, and examples
- especially:
  - `--yes`
  - `--cascade`
  - `--dry-run`
  - column reassignment semantics

**Dependencies**
- Phase 3 delete commands complete

**Acceptance criteria**
- all destructive commands use consistent confirmation guidance
- dry-run examples exist where relevant
- human failure messages for missing destructive flags are consistent and actionable
- docs/help match final behavior

---

### Issue 15 — Add representative golden coverage for help and JSON contracts

**ID**: P4-15  
**Title**: `test(cli): add golden coverage for root help topic help and representative json contracts`

**Goal**
- Create durable regression coverage for the final CLI contract
- focus on:
  - root help
  - topic help
  - representative command help
  - representative JSON success payloads
  - representative JSON error payloads

**Dependencies**
- P4-05 through P4-14

**Acceptance criteria**
- representative help and JSON contract snapshots/goldens exist
- tests fail meaningfully when command/help/output contracts drift
- coverage spans at least one command from each major command family

---

### Issue 16 — Final CLI v1 acceptance sweep against blueprint

**ID**: P4-16  
**Title**: `chore(cli): run final acceptance sweep against cli-v1 blueprint`

**Goal**
- Validate implemented CLI surface against `docs/prd/cli-v1.md`
- close the gap between plan and delivery

**Dependencies**
- all Phase 4 tasks complete

**Acceptance criteria**
- explicit checklist exists comparing implementation to blueprint sections
- discrepancies are either fixed or documented as intentional deviations
- all phase acceptance criteria from Phases 1–4 are reviewed and confirmed
- CLI v1 is ready for issue closure / release milestone

---

## 6. Suggested First Five Sessions

1. **P4-01** — root help audit  
2. **P4-02** — topic help audit  
3. **P4-03** — per-command long help audit  
4. **P4-04** — example normalization  
5. **P4-05 + P4-06** — JSON success/error contract audit  

This sequence hardens the most visible contract surfaces first, before generated docs and final acceptance work.
