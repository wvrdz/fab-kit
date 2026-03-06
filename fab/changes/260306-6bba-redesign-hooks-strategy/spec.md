# Spec: Redesign Hooks Strategy

**Change**: 260306-6bba-redesign-hooks-strategy
**Created**: 2026-03-06
**Affected memory**:
- `docs/memory/fab-workflow/kit-architecture.md`
- `docs/memory/fab-workflow/schemas.md`
- `docs/memory/fab-workflow/planning-skills.md`
- `docs/memory/fab-workflow/execution-skills.md`
- `docs/memory/fab-workflow/setup.md`

## Non-Goals

- User-configured stage hooks — explicitly rejected as "too unreliable"; user enforcement stays in `project/*` files
- Language-specific templates — fab-kit stays language-neutral (4vj0, qg80, rwt1 rejected)
- Absorbing pipeline orchestrator's yq usage (~40 uses) — separate concern, pipeline already lists yq as prerequisite
- Hooking `fab status advance` — must happen after SRAD questions in fab-new, not on artifact write
- Hooking stage transitions (`finish`, `start`, `fail`, `reset`) — these are intentional agent decisions
- Hooking `fab log command` — skill name cannot be detected from hook events
- PreCompact, SessionEnd, SessionStart enhancement hooks — assessed and rejected

## Hooks: PostToolUse Artifact Bookkeeping

### Requirement: Hook Script for Artifact Write Events

A new hook script `fab/.kit/hooks/on-artifact-write.sh` SHALL be registered for both `PostToolUse` `Write` and `PostToolUse` `Edit` Claude Code hook events. The script SHALL detect fab artifact writes and execute bookkeeping commands automatically.

#### Scenario: Intake Written
- **GIVEN** the agent writes or edits `fab/changes/{name}/intake.md`
- **WHEN** the PostToolUse hook fires
- **THEN** the hook SHALL infer the change type from intake content via keyword matching and call `fab status set-change-type <change> <type>`
- **AND** the hook SHALL compute indicative confidence via `fab score --stage intake <change>`

#### Scenario: Spec Written
- **GIVEN** the agent writes or edits `fab/changes/{name}/spec.md`
- **WHEN** the PostToolUse hook fires
- **THEN** the hook SHALL compute confidence via `fab score <change>`

#### Scenario: Tasks Written
- **GIVEN** the agent writes or edits `fab/changes/{name}/tasks.md`
- **WHEN** the PostToolUse hook fires
- **THEN** the hook SHALL count `- [ ]` and `- [x]` lines in the file content and call `fab status set-checklist <change> total <N>` where N is the total count

#### Scenario: Checklist Written
- **GIVEN** the agent writes or edits `fab/changes/{name}/checklist.md`
- **WHEN** the PostToolUse hook fires
- **THEN** the hook SHALL call `fab status set-checklist <change> generated true`
- **AND** the hook SHALL count all task items (`- [ ]` + `- [x]`) and call `fab status set-checklist <change> total <N>`
- **AND** the hook SHALL count completed items (`- [x]`) and call `fab status set-checklist <change> completed <M>`

#### Scenario: Non-Artifact File
- **GIVEN** the agent writes or edits a file that does NOT match `fab/changes/*/intake.md`, `fab/changes/*/spec.md`, `fab/changes/*/tasks.md`, or `fab/changes/*/checklist.md`
- **WHEN** the PostToolUse hook fires
- **THEN** the hook SHALL exit 0 immediately (fast-path no-op)

#### Scenario: Bookkeeping Command Failure
- **GIVEN** any `fab` CLI command invoked by the hook fails
- **WHEN** the hook processes the failure
- **THEN** the hook SHALL ignore the error and exit 0 — bookkeeping failures MUST NOT interrupt the agent

### Requirement: Hook Input Processing

The hook script SHALL read JSON from stdin containing `tool_input.file_path` (and optionally `tool_input.content` for Write events). For Edit events where content is not provided on stdin, the hook SHALL read the file from disk if needed.

#### Scenario: Write Event with Content
- **GIVEN** a PostToolUse Write event fires
- **WHEN** the hook reads stdin JSON
- **THEN** `tool_input.file_path` SHALL be used for path matching and `tool_input.content` SHALL be used for content analysis (keyword matching, item counting)

#### Scenario: Edit Event without Content
- **GIVEN** a PostToolUse Edit event fires
- **WHEN** the hook reads stdin JSON
- **THEN** `tool_input.file_path` SHALL be used for path matching and the hook SHALL read the file from disk for content analysis

### Requirement: Change Name Derivation

The hook SHALL derive the change name from the file path by extracting the folder name between `fab/changes/` and the artifact filename.

#### Scenario: Path Parsing
- **GIVEN** file path `fab/changes/260306-6bba-redesign-hooks-strategy/intake.md`
- **WHEN** the hook parses the path
- **THEN** the change name SHALL be `260306-6bba-redesign-hooks-strategy`

### Requirement: Hook Output

The hook SHOULD return JSON on stdout with an `additionalContext` field to inform the agent what was auto-handled.

#### Scenario: Successful Bookkeeping
- **GIVEN** the hook successfully processes an artifact write
- **WHEN** it completes bookkeeping
- **THEN** stdout SHALL contain JSON like `{"additionalContext": "Bookkeeping: score 4.2/5.0, type: refactor"}` (content varies by artifact)

### Requirement: Change Type Inference in Hook

The hook SHALL infer change type from intake content using keyword matching (case-insensitive, first match wins):

1. fix, bug, broken, regression → `fix`
2. refactor, restructure, consolidate, split, rename → `refactor`
3. docs, document, readme, guide → `docs`
4. test, spec, coverage → `test`
5. ci, pipeline, deploy, build → `ci`
6. chore, cleanup, maintenance, housekeeping → `chore`
7. Otherwise → `feat`

#### Scenario: Keyword Match
- **GIVEN** intake content contains "Refactor the hooks system"
- **WHEN** the hook scans for change type keywords
- **THEN** the inferred type SHALL be `refactor`

## Runtime CLI: fab runtime Subcommands

### Requirement: fab runtime set-idle

A new `fab runtime set-idle <change>` subcommand SHALL write the `agent.idle_since` timestamp to `.fab-runtime.yaml` at the repo root, keyed by the change's full folder name.

#### Scenario: Set Idle Timestamp
- **GIVEN** a valid change reference `6bba`
- **WHEN** `fab runtime set-idle 6bba` is executed
- **THEN** `.fab-runtime.yaml` SHALL contain `{change_folder}.agent.idle_since` set to the current Unix timestamp
- **AND** if the file does not exist, it SHALL be created

#### Scenario: Invalid Change
- **GIVEN** an invalid change reference
- **WHEN** `fab runtime set-idle invalid` is executed
- **THEN** the command SHALL exit non-zero with an error message

### Requirement: fab runtime clear-idle

A new `fab runtime clear-idle <change>` subcommand SHALL delete the `agent` block from `.fab-runtime.yaml` for the specified change.

#### Scenario: Clear Agent Block
- **GIVEN** `.fab-runtime.yaml` contains an `agent` block for the change
- **WHEN** `fab runtime clear-idle <change>` is executed
- **THEN** the `agent` block for that change SHALL be removed
- **AND** other change entries in the file SHALL be preserved

#### Scenario: No Existing Entry
- **GIVEN** `.fab-runtime.yaml` does not contain an entry for the change
- **WHEN** `fab runtime clear-idle <change>` is executed
- **THEN** the command SHALL exit 0 (no-op)

#### Scenario: Missing Runtime File
- **GIVEN** `.fab-runtime.yaml` does not exist
- **WHEN** `fab runtime clear-idle <change>` is executed
- **THEN** the command SHALL exit 0 (no-op)

### Requirement: Standard Change Resolution

Both `fab runtime` subcommands SHALL use the standard `<change>` argument convention (4-char ID, substring, full folder name) via the existing `resolve` package.

#### Scenario: Resolution by ID
- **GIVEN** change ID `6bba`
- **WHEN** passed to `fab runtime set-idle`
- **THEN** the full folder name SHALL be resolved and used as the YAML key

## Hook Migration: yq to fab runtime

### Requirement: Migrate on-stop.sh

`fab/.kit/hooks/on-stop.sh` SHALL be rewritten to use `fab runtime set-idle` instead of `yq`. The `command -v yq` guard and all `yq` invocations SHALL be removed.

#### Scenario: Stop Hook with fab CLI
- **GIVEN** the Stop hook fires and an active change exists
- **WHEN** the hook executes
- **THEN** it SHALL call `"$fab_cmd" runtime set-idle "$change_folder" 2>/dev/null || true`
- **AND** the script SHALL NOT reference `yq`

#### Scenario: Missing fab CLI
- **GIVEN** the fab binary is not found at `$repo_root/fab/.kit/bin/fab`
- **WHEN** the hook checks for the binary
- **THEN** it SHALL exit 0 (silent no-op)

### Requirement: Migrate on-session-start.sh

`fab/.kit/hooks/on-session-start.sh` SHALL be rewritten to use `fab runtime clear-idle` instead of `yq`.

#### Scenario: Session Start Hook with fab CLI
- **GIVEN** the SessionStart hook fires and an active change exists
- **WHEN** the hook executes
- **THEN** it SHALL call `"$fab_cmd" runtime clear-idle "$change_folder" 2>/dev/null || true`
- **AND** the script SHALL NOT reference `yq`

## Hook Sync: Matcher Support

### Requirement: Extend 5-sync-hooks.sh for PostToolUse Matchers

`fab/.kit/sync/5-sync-hooks.sh` SHALL be extended to register PostToolUse hooks with tool-specific matchers (`Write`, `Edit`). The `map_event()` function SHALL return both event name and matcher.

#### Scenario: PostToolUse Write Registration
- **GIVEN** hook file `on-artifact-write.sh` exists in `fab/.kit/hooks/`
- **WHEN** the sync script processes it
- **THEN** `.claude/settings.local.json` SHALL contain a `PostToolUse` entry with `"matcher": "Write"` pointing to the hook

#### Scenario: PostToolUse Edit Registration
- **GIVEN** hook file `on-artifact-write.sh` exists
- **WHEN** the sync script processes it
- **THEN** `.claude/settings.local.json` SHALL also contain a `PostToolUse` entry with `"matcher": "Edit"` pointing to the same hook

#### Scenario: Existing Matcher-less Hooks Unaffected
- **GIVEN** existing hooks `on-stop.sh` and `on-session-start.sh`
- **WHEN** the sync script processes them
- **THEN** they SHALL continue to register with `"matcher": ""` as before

### Requirement: Filename-to-Event Mapping

The `map_event()` function (or its replacement) SHALL support a mapping table that returns event+matcher pairs. The mapping for `on-artifact-write.sh` SHALL produce two entries: `PostToolUse/Write` and `PostToolUse/Edit`.

#### Scenario: Mapping Table
- **GIVEN** the mapping table in the sync script
- **WHEN** `on-artifact-write.sh` is looked up
- **THEN** it SHALL return: `PostToolUse Write` and `PostToolUse Edit`
- **AND** `on-stop.sh` SHALL return: `Stop ""`
- **AND** `on-session-start.sh` SHALL return: `SessionStart ""`

## Skills: Remove Bookkeeping Instructions

### Requirement: Remove Bookkeeping from fab-new.md

`fab-new.md` Step 6 (infer change type) and Step 7 (indicative confidence) SHALL be removed. The hook now handles both operations when `intake.md` is written.

#### Scenario: fab-new After Change
- **GIVEN** the agent generates `intake.md` via `/fab-new`
- **WHEN** the Write tool completes
- **THEN** the PostToolUse hook SHALL automatically infer change type and compute indicative confidence
- **AND** `fab-new.md` SHALL NOT contain instructions to call `fab status set-change-type` or `fab score --stage intake`

Step 9 (`fab status advance`) SHALL remain — it must happen after SRAD questions, not on write.

### Requirement: Remove Bookkeeping from fab-continue.md

The explicit `fab score <change>` call after spec generation SHALL be removed. The hook handles this when `spec.md` is written or edited.

#### Scenario: fab-continue Spec Stage
- **GIVEN** the agent generates `spec.md` via `/fab-continue`
- **WHEN** the Write tool completes
- **THEN** the PostToolUse hook SHALL automatically compute confidence
- **AND** `fab-continue.md` SHALL NOT contain instructions to call `fab score` after spec generation

All stage transitions (`finish`, `start`, `reset`, `fail`) SHALL remain in `fab-continue.md`.

### Requirement: Remove Bookkeeping from fab-ff.md

Step 4's three `set-checklist` calls SHALL be removed. The score computation after spec (not the gate check) SHALL be removed.

#### Scenario: fab-ff Planning Complete
- **GIVEN** the agent generates `checklist.md` via `/fab-ff`
- **WHEN** the Write tool completes
- **THEN** the PostToolUse hook SHALL automatically set checklist metadata
- **AND** `fab-ff.md` Step 4 SHALL NOT contain `set-checklist` calls

The intake gate (`fab score --check-gate --stage intake`) and spec gate (`fab score --check-gate`) SHALL remain — these are read-only gate validations, not bookkeeping writes.

### Requirement: Remove Bookkeeping from fab-fff.md

Same as fab-ff: Step 4's three `set-checklist` calls SHALL be removed.

#### Scenario: fab-fff Planning Complete
- **GIVEN** the agent generates `checklist.md` via `/fab-fff`
- **WHEN** the Write tool completes
- **THEN** the PostToolUse hook SHALL automatically set checklist metadata

### Requirement: Remove Bookkeeping from fab-clarify.md

Step 7 (recompute confidence via `fab score` in suggest mode) SHALL be removed. The hook fires when `spec.md` is edited by `/fab-clarify`.

#### Scenario: fab-clarify Suggest Mode
- **GIVEN** the agent edits `spec.md` via `/fab-clarify`
- **WHEN** the Edit tool completes
- **THEN** the PostToolUse hook SHALL automatically recompute confidence

### Requirement: Remove Bookkeeping from _generation.md

Checklist Generation Procedure step 6 (three `set-checklist` CLI commands) SHALL be removed. The hook handles this when `checklist.md` is written.

#### Scenario: Checklist Generation
- **GIVEN** the agent writes `checklist.md` via the Checklist Generation Procedure
- **WHEN** the Write tool completes
- **THEN** the PostToolUse hook SHALL automatically set `generated true`, `total <N>`, and `completed <M>`

## Constitution: Update §I Wording

### Requirement: Broaden Script Language Reference

Constitution §I SHALL change "shell scripts" to "scripts" to accommodate non-shell hook scripts and future scripting.

#### Scenario: Constitution Text
- **GIVEN** the current text: "All workflow logic MUST live in markdown skill files and shell scripts."
- **WHEN** this change is applied
- **THEN** the text SHALL read: "All workflow logic MUST live in markdown skill files and scripts."
- **AND** the constitution version SHALL be bumped and `Last Amended` date updated

## Design Decisions

### Single Hook Script for Both Write and Edit Matchers
**Decision**: One script (`on-artifact-write.sh`) handles both PostToolUse Write and PostToolUse Edit events.
- *Why*: Same detection logic (path matching) and same bookkeeping commands. No reason for separate scripts.
- *Rejected*: Separate `on-artifact-write.sh` and `on-artifact-edit.sh` — duplicated logic, maintenance burden.

### Hook Returns additionalContext
**Decision**: The hook returns JSON with `additionalContext` to inform the agent what was auto-handled.
- *Why*: PostToolUse hooks can return structured JSON. This keeps the agent informed without requiring explicit instructions.
- *Rejected*: Silent execution — the agent would have no visibility into what happened.

### Hook Reads Content from Stdin or Disk
**Decision**: The hook reads `tool_input.content` from stdin JSON for Write events, and reads from disk for Edit events.
- *Why*: Write events include full content in stdin; Edit events may only include the diff, not the full file content.
- *Rejected*: Always reading from disk — slower and unnecessary for Write events.

### Checklist Hook Counts Both Total and Completed
**Decision**: When `checklist.md` is written or edited, the hook counts all items for total and `[x]` items for completed.
- *Why*: This handles both initial creation (all unchecked) and review updates (items progressively checked) with one mechanism. Eliminates the need for separate `set-checklist completed` calls in review behavior.
- *Rejected*: Only handling initial creation — would still require manual completed count updates during review.

### fab runtime Uses Go YAML Library
**Decision**: The `fab runtime` subcommands use the Go `gopkg.in/yaml.v3` library (already in the project) to read/write `.fab-runtime.yaml`, replacing yq.
- *Why*: The Go binary is always present. yq is an external dependency that may not be installed. Consistent with the "all through fab CLI" pattern.
- *Rejected*: Keeping yq — silent failure when not installed, additional dependency.

### Stale Changes Already Deleted
**Decision**: The stale changes (4vj0, qg80, rwt1, shk2) referenced in the intake are already deleted from all branches. No deletion task needed.
- *Why*: Verified via git — none exist in this worktree or on main.
- *Rejected*: Adding deletion tasks for non-existent folders.

### Language Detection Already Absent
**Decision**: No language detection exists in `fab-setup.md`. No removal task needed.
- *Why*: Searched the skill file — no Phase 1b-lang or language detection logic present.
- *Rejected*: Adding a removal task for non-existent code.

## Deprecated Requirements

### Manual Bookkeeping in Planning Skills
**Reason**: Automated via PostToolUse hook. Skills no longer need explicit `fab score`, `set-change-type`, or `set-checklist` instructions after artifact generation.
**Migration**: The hook handles these operations automatically when artifacts are written or edited.

### yq Dependency in Hook Scripts
**Reason**: Replaced by `fab runtime` Go subcommands. The Go binary is always present.
**Migration**: `on-stop.sh` and `on-session-start.sh` use `fab runtime set-idle` and `fab runtime clear-idle` respectively.

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Hooks are for kit-internal mechanics only — user enforcement stays in project/* files | Confirmed from intake #1 — user explicitly rejected user-configured hooks | S:95 R:85 A:95 D:95 |
| 2 | Certain | PostToolUse (Write/Edit) is the only new hook event | Confirmed from intake #2 — all 18 events assessed | S:95 R:85 A:95 D:95 |
| 3 | Certain | Constitution §I changes "shell scripts" to "scripts" | Confirmed from intake #4 — user explicitly requested | S:95 R:90 A:95 D:95 |
| 4 | Certain | Stage transitions remain agent-directed (not hookable) | Confirmed from intake #5 — advance/finish/fail/reset are intentional decisions | S:90 R:85 A:90 D:90 |
| 5 | Certain | `fab status advance` stays in fab-new (not hookable) | Confirmed from intake #6 — must happen after SRAD questions | S:90 R:80 A:90 D:90 |
| 6 | Certain | Existing hooks migrate from yq to `fab runtime` subcommands | Confirmed from intake #7 — eliminates silent-fail yq dependency | S:90 R:85 A:90 D:90 |
| 7 | Certain | Single script handles both Write and Edit matchers | Confirmed from intake #8 — same logic for both events | S:85 R:90 A:85 D:90 |
| 8 | Certain | Stale changes already deleted — no task needed | Verified via git — 4vj0, qg80, rwt1, shk2 not in worktree or main | S:95 R:95 A:95 D:95 |
| 9 | Certain | Language detection absent from fab-setup.md — no task needed | Verified by searching fab-setup.md — no Phase 1b-lang or detection logic found | S:95 R:95 A:95 D:95 |
| 10 | Confident | Hook returns additionalContext to inform agent | Confirmed from intake #9 — PostToolUse hooks can return JSON with additionalContext field | S:80 R:90 A:80 D:85 |
| 11 | Confident | Pipeline orchestrator keeps yq — separate concern | Confirmed from intake #10 — absorbing manifest parsing is a much larger change | S:80 R:85 A:85 D:80 |
| 12 | Confident | Checklist hook counts both total and completed on every write/edit | Natural extension — eliminates separate completed count updates in review. No intake disagreement | S:75 R:85 A:80 D:80 |
| 13 | Confident | fab runtime uses existing Go YAML library (yaml.v3) | Project already depends on yaml.v3 for statusfile operations. Consistent pattern | S:80 R:90 A:85 D:85 |

13 assumptions (9 certain, 4 confident, 0 tentative, 0 unresolved).
