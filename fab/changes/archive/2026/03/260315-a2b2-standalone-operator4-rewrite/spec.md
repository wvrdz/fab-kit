# Spec: Standalone Operator4 Rewrite

**Change**: 260315-a2b2-standalone-operator4-rewrite
**Created**: 2026-03-15
**Affected memory**: `docs/memory/fab-workflow/execution-skills.md` (modify), `docs/memory/fab-workflow/kit-architecture.md` (modify)

## Non-Goals

- Changing operator4's runtime behavior — auto-nudge, monitoring, autopilot all preserved exactly
- Splitting `_preamble.md` into framework + context files
- Modifying `_generation.md`
- Adding selective-load mechanism to `_preamble.md`

## Skills: Standalone Operator4

### Requirement: Self-Contained Skill File

`fab/.kit/skills/fab-operator4.md` SHALL be a fully self-contained skill that does not reference or inherit from `fab-operator1.md`, `fab-operator2.md`, or `fab-operator3.md`. An agent reading operator4 SHALL understand all operator behavior from this single file plus the standard `_` files loaded via `_preamble.md`.

#### Scenario: Agent loads operator4
- **GIVEN** an agent invoked with `/fab-operator4`
- **WHEN** it follows the preamble instructions and reads operator4
- **THEN** it has complete knowledge of all operator behavior without reading any other operator skill file

#### Scenario: No inheritance directive
- **GIVEN** the operator4 skill file
- **WHEN** an agent reads the file header
- **THEN** there is no instruction to read operator1, operator2, or operator3

### Requirement: Section Structure

The standalone operator4 SHALL be organized into these sections in order:

1. **Principles** — identity, routing discipline, context discipline, state re-derivation (with why for each)
2. **Startup** — context loading, orientation, outside-tmux degradation
3. **Safety Model** — confirmation tiers, pre-send validation, bounded retries, context discipline
4. **Monitoring System** — monitored set, `/loop` lifecycle, monitoring tick (6 steps)
5. **Auto-Nudge** — question detection, answer model, re-capture guard, logging
6. **Modes of Operation** — shared rhythm + compact table for all 8 modes
7. **Autopilot** — queue ordering, per-change loop, failure matrix, interruptibility
8. **Configuration** — session-scoped settings
9. **Key Properties** — standard properties table

#### Scenario: Complete monitoring tick
- **GIVEN** the monitoring system section
- **WHEN** an agent reads the monitoring tick definition
- **THEN** all 6 steps are fully specified (stage advance, pipeline completion, review failure, pane death, auto-nudge, stuck detection) without reference to external operator files

#### Scenario: Auto-nudge self-contained
- **GIVEN** the auto-nudge section
- **WHEN** an agent reads the question detection and answer model
- **THEN** the simplified decision list (items 1-6), all guards (Claude turn boundary, blank capture, idle-only), capture window (`-l 20`), pattern matching rules, re-capture guard, and logging format are all specified inline

### Requirement: Writing Style

The skill SHALL explain the "why" behind constraints instead of heavy-handed imperative directives. Imperatives (MUST/SHALL) are reserved for true safety constraints (pre-send validation, bounded retries, confirmation for destructive actions). All other guidance explains reasoning so the agent can apply judgment in edge cases.

#### Scenario: Principle has rationale
- **GIVEN** the principles section states "the operator never loads change artifacts"
- **WHEN** an agent reads this constraint
- **THEN** the rationale is provided (e.g., "context window is reserved for coordination state")

### Requirement: Size Target

The operator4 main file SHOULD be approximately 300 lines. Tool references, naming conventions, and fab CLI documentation are provided by `_` files and SHALL NOT be duplicated in operator4.

#### Scenario: No tool table duplication
- **GIVEN** operator4 needs to reference `fab pane-map`
- **WHEN** the agent reads operator4
- **THEN** operator4 does not contain a tool reference table for fab CLI commands (these are in `_cli-fab.md`)

### Requirement: Operator-Only Context Loading

Operator4 SHALL load `fab/.kit/skills/_cli-external.md` in its own startup section. This file is NOT part of the always-load layer in `_preamble.md`.

#### Scenario: Operator loads external tools reference
- **GIVEN** an agent invoked with `/fab-operator4`
- **WHEN** it follows operator4's startup instructions
- **THEN** it reads `_cli-external.md` for wt, tmux, and `/loop` reference
- **AND** pipeline skills (e.g., `/fab-continue`) do NOT load `_cli-external.md`

### Requirement: Autopilot Pipeline

Operator4 SHALL use `/fab-fff` (not `/fab-ff`) for autopilot gate checks and pipeline invocations.

#### Scenario: Autopilot spawns pipeline
- **GIVEN** operator4 is running autopilot for a change
- **WHEN** the confidence gate passes
- **THEN** the operator sends `/fab-fff` to the agent (not `/fab-ff`)

### Requirement: Operator Sends `/git-branch` After New Change

When the operator spawns an agent for a new change (from backlog), it SHALL send `/git-branch` to the agent after detecting the intake stage has advanced (indicating `/fab-new` completed and the change folder exists). This aligns the branch name with the change folder name per `_naming.md` conventions.

#### Scenario: New change from backlog
- **GIVEN** the operator spawned an agent for a backlog idea
- **WHEN** the monitoring tick detects the agent's change has advanced past intake
- **THEN** the operator sends `/git-branch` to the agent's pane

#### Scenario: Existing change
- **GIVEN** the operator spawns an agent for an already-existing change
- **WHEN** it creates the worktree
- **THEN** it passes the change folder name as the branch argument to `wt create` and does NOT need to send `/git-branch`

## Skills: `_cli-external.md`

### Requirement: External Tool Reference

A new file `fab/.kit/skills/_cli-external.md` SHALL document external (non-fab) CLI tools used by the operator: `wt` (worktree manager), `tmux`, and `/loop`.

#### Scenario: wt commands documented
- **GIVEN** an agent reads `_cli-external.md`
- **WHEN** it needs to create a worktree
- **THEN** it finds the `wt create` command with all flags (`--non-interactive`, `--worktree-name`, `--reuse`, `--base`, branch argument)

#### Scenario: tmux commands documented
- **GIVEN** an agent reads `_cli-external.md`
- **WHEN** it needs to capture terminal output
- **THEN** it finds `tmux capture-pane -t <pane> -p [-l N]` with usage notes

### Requirement: Operator-Only Loading

`_cli-external.md` SHALL NOT be added to the always-load list in `_preamble.md`. It is loaded only by skills that explicitly reference it (currently only operator4).

#### Scenario: Pipeline skill does not load external tools
- **GIVEN** an agent invoked with `/fab-continue`
- **WHEN** it follows `_preamble.md` context loading
- **THEN** it does NOT read `_cli-external.md`

## Skills: `_naming.md`

### Requirement: Naming Conventions Reference

A new file `fab/.kit/skills/_naming.md` SHALL document naming conventions for change folders, git branches, worktree directories, and operator spawning rules.

#### Scenario: Change folder pattern
- **GIVEN** an agent reads `_naming.md`
- **WHEN** it needs the change folder naming pattern
- **THEN** it finds `{YYMMDD}-{XXXX}-{slug}` with the note that `fab change new` generates this

#### Scenario: Operator spawning rules for known change
- **GIVEN** an agent reads the operator spawning rules in `_naming.md`
- **WHEN** the change already exists
- **THEN** the convention says to use the change folder name as the branch argument to `wt create`

#### Scenario: Operator spawning rules for new change
- **GIVEN** an agent reads the operator spawning rules in `_naming.md`
- **WHEN** spawning for a backlog idea (no change exists yet)
- **THEN** the convention says `wt create` auto-generates the worktree name, and the operator sends `/git-branch` after `/fab-new` completes

### Requirement: Always-Load

`_naming.md` SHALL be added to the always-load list in `_preamble.md` alongside `_cli-fab.md`.

#### Scenario: Any skill loads naming conventions
- **GIVEN** an agent invoked with any skill that follows `_preamble.md`
- **WHEN** it reads the always-load files
- **THEN** `_naming.md` is included in the list

## Skills: Rename `_scripts.md` to `_cli-fab.md`

### Requirement: File Rename

`fab/.kit/skills/_scripts.md` SHALL be renamed to `fab/.kit/skills/_cli-fab.md`. Content is unchanged.

#### Scenario: Preamble references updated
- **GIVEN** `_preamble.md` previously referenced `_scripts.md`
- **WHEN** the rename is applied
- **THEN** all references in `_preamble.md` point to `_cli-fab.md`

## Skills: Cross-References in `git-branch.md` and `git-pr.md`

### Requirement: Naming Cross-Reference

`git-branch.md` and `git-pr.md` SHALL include a cross-reference to `_naming.md` for naming conventions. All procedural logic in these skills is preserved unchanged.

#### Scenario: git-branch cross-reference
- **GIVEN** `git-branch.md`
- **WHEN** an agent reads the file
- **THEN** it contains a note: "Branch naming conventions are defined in `_naming.md`."
- **AND** the branch creation/checkout/rename procedures are unchanged

## Skills: Delete Operator1/2/3

### Requirement: Remove Legacy Operator Files

`fab/.kit/skills/fab-operator1.md`, `fab/.kit/skills/fab-operator2.md`, and `fab/.kit/skills/fab-operator3.md` SHALL be deleted from the source directory.

#### Scenario: No ghost skills
- **GIVEN** operator1/2/3 source files are deleted
- **WHEN** `fab-sync.sh` runs
- **THEN** the corresponding deployed copies in `.claude/skills/` are also removed

## Scripts: `fab-sync.sh`

### Requirement: Exact-Match Invariant

After `fab-sync.sh` runs, the set of skill files in `.claude/skills/` SHALL exactly match the set of skill files in `fab/.kit/skills/`. No stale files remain; no source files are missing from the deployed set.

#### Scenario: Rename handled
- **GIVEN** `_scripts.md` was renamed to `_cli-fab.md` in source
- **WHEN** `fab-sync.sh` runs
- **THEN** `.claude/skills/_cli-fab.md` exists AND `.claude/skills/_scripts.md` does NOT exist

#### Scenario: Deletions handled
- **GIVEN** `fab-operator1.md`, `fab-operator2.md`, `fab-operator3.md` were deleted from source
- **WHEN** `fab-sync.sh` runs
- **THEN** the corresponding files in `.claude/skills/` are also removed

#### Scenario: New files deployed
- **GIVEN** `_cli-external.md` and `_naming.md` are added to source
- **WHEN** `fab-sync.sh` runs
- **THEN** they appear in `.claude/skills/`

## Specs: Update `docs/specs/skills/`
<!-- clarified: constitution §Additional Constraints requires SPEC-*.md updates when skill files change; intake Impact section already identified this -->

### Requirement: Spec File Alignment

The `docs/specs/skills/` directory SHALL be updated to reflect all skill file changes:

1. `SPEC-fab-operator4.md` — rewritten to match standalone operator4
2. `SPEC-fab-operator1.md`, `SPEC-fab-operator2.md`, `SPEC-fab-operator3.md` — deleted (matching source file deletions)

New `_` files (`_cli-external.md`, `_naming.md`) and the renamed `_cli-fab.md` do not require individual SPEC files — they are internal partials, not user-invocable skills.

#### Scenario: Operator4 spec updated
- **GIVEN** `fab-operator4.md` is rewritten as standalone
- **WHEN** the change is applied
- **THEN** `docs/specs/skills/SPEC-fab-operator4.md` reflects the new standalone structure

#### Scenario: Legacy operator specs removed
- **GIVEN** `fab-operator1.md`, `fab-operator2.md`, `fab-operator3.md` are deleted
- **WHEN** the change is applied
- **THEN** `SPEC-fab-operator1.md`, `SPEC-fab-operator2.md`, `SPEC-fab-operator3.md` are also deleted

## Deprecated Requirements

### Operator Inheritance Chain

**Reason**: Operator4 is now standalone. The inheritance pattern (operator1→2→3→4) is replaced by a single self-contained file.
**Migration**: All behavior from operator1/2/3 is inlined into operator4. The source files are deleted.

## Design Decisions

1. **`_cli-external.md` is operator-only, not always-load**
   - *Why*: wt, tmux, and `/loop` are only relevant to the operator. Loading them for every pipeline skill wastes ~60 lines of context for no benefit.
   - *Rejected*: Always-load for simplicity — the cost (context waste in 20+ skills) outweighs the benefit (one fewer line in operator4's startup).

2. **Delete operator1/2/3 rather than archive**
   - *Why*: Git history preserves all versions. Dead skill files in the directory risk ghost triggers (agents matching on "operator" keyword in descriptions). Sync script would deploy them as stale skills.
   - *Rejected*: Move to `fab/.kit/skills/archive/` — requires new sync exclusion logic and the directory doesn't exist.

3. **Cross-reference `_naming.md` in git-branch/git-pr rather than extract procedures**
   - *Why*: The procedures in git-branch (rename/checkout/create logic) and git-pr (branch matching/nudge) *apply* naming conventions but aren't naming conventions themselves. Extracting them would fragment procedural logic across files.
   - *Rejected*: Aggressive extraction of naming-related lines — creates indirection for procedures that are only used in one place.

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Operator4 fully standalone — no inheritance | Confirmed from intake #1 — user explicitly requested | S:95 R:80 A:90 D:95 |
| 2 | Certain | `_preamble.md` stays intact | Confirmed from intake #2 — not worth splitting | S:90 R:85 A:85 D:90 |
| 3 | Certain | `_scripts.md` → `_cli-fab.md` rename | Confirmed from intake #3 | S:95 R:90 A:90 D:95 |
| 4 | Certain | `_cli-external.md` operator-only load | Clarified from intake #4 + #11 — user specified operator-only | S:95 R:85 A:90 D:95 |
| 5 | Certain | `_naming.md` always-load | Confirmed from intake #5 + clarification | S:95 R:85 A:90 D:95 |
| 6 | Certain | `_generation.md` unchanged | Confirmed from intake #6 | S:90 R:95 A:90 D:95 |
| 7 | Certain | Operator sends `/git-branch` after new change | Confirmed from intake #7 — user confirmed operator responsible | S:90 R:75 A:85 D:85 |
| 8 | Certain | Explain-the-why writing style | Confirmed from intake #8 — skill-creator guidance | S:85 R:85 A:80 D:85 |
| 9 | Confident | ~300 line target for operator4 | Carried from intake #9 — reasonable estimate, may vary | S:75 R:90 A:80 D:80 |
| 10 | Certain | Delete operator1/2/3 outright | Clarified from intake #10 — user confirmed deletion | S:95 R:70 A:90 D:95 |
| 11 | Certain | Sync exact-match invariant | Clarified from intake #12 — deployed = source | S:95 R:80 A:90 D:95 |
| 12 | Certain | git-branch/git-pr cross-reference only | Clarified from intake #13 — procedures stay, add reference line | S:90 R:90 A:85 D:90 |

12 assumptions (11 certain, 1 confident, 0 tentative, 0 unresolved).
