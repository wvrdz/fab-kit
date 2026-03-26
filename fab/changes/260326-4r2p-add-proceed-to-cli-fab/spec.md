# Spec: Add fab-proceed to _cli-fab.md

**Change**: 260326-4r2p-add-proceed-to-cli-fab
**Created**: 2026-03-26
**Affected memory**: `docs/memory/fab-workflow/execution-skills.md`

## Non-Goals

- Adding new CLI commands — PR 278 introduced only a skill, no Go binary changes
- Restructuring `_cli-fab.md` layout — additions follow the existing document structure
- Documenting `/fab-proceed`'s full behavior — that belongs in the skill file and specs, not the CLI guide

## CLI Guide: fab-proceed Consumer Documentation

### Requirement: fab resolve SHALL document fab-proceed as a caller

The `fab resolve` section in `_cli-fab.md` SHALL include a **Notable callers** note documenting that `/fab-proceed` invokes `fab resolve --folder` (with stderr suppressed) in its Step 1 state detection to check whether an active change exists.

#### Scenario: Reader looks up fab resolve callers
- **GIVEN** a skill author reading the `fab resolve` section of `_cli-fab.md`
- **WHEN** they look for which skills use `fab resolve --folder`
- **THEN** `/fab-proceed` SHALL be listed as a caller with its usage pattern (`fab resolve --folder 2>/dev/null` for active change detection)

### Requirement: fab change switch SHALL document fab-proceed as a caller

The `fab change` section's `switch` subcommand row in `_cli-fab.md` SHALL include a note that `/fab-proceed` dispatches `fab change switch` via a subagent when an unactivated intake is detected.

#### Scenario: Reader looks up fab change switch callers
- **GIVEN** a skill author reading the `fab change` section of `_cli-fab.md`
- **WHEN** they check who invokes the `switch` subcommand
- **THEN** `/fab-proceed` SHALL be documented as dispatching `fab change switch "<change-name>"` via subagent

### Requirement: fab log Callers table SHALL NOT add a fab-proceed row

The `fab log` Callers table SHALL NOT add a dedicated `/fab-proceed` row because `/fab-proceed` does not call `fab log command` directly. Its subagents (`/fab-new`, `/fab-switch`) use `fab log` through their own documented paths, which are already captured in the existing Callers entries.

#### Scenario: fab-proceed's indirect logging is already covered
- **GIVEN** `/fab-proceed` dispatches `/fab-new` as a subagent
- **WHEN** `/fab-new` calls `fab change new --log-args`
- **THEN** the existing `fab change new` auto-log Callers entry covers this path
- **AND** no duplicate `/fab-proceed` row is needed in the Callers table

### Requirement: Documentation additions SHALL follow existing _cli-fab.md patterns

All additions to `_cli-fab.md` MUST use the same formatting patterns already present in the file: brief inline notes or small tables, not new top-level sections. The additions SHOULD be minimal — a sentence or note per command section, not paragraphs.

#### Scenario: Addition to fab resolve section
- **GIVEN** the `fab resolve` section currently has no callers documentation
- **WHEN** `/fab-proceed` is added as a caller
- **THEN** the addition SHALL be a brief `**Notable callers**:` line after the flags table, consistent with the `**Callers**:` pattern used in the `fab log` section

#### Scenario: Addition to fab change section
- **GIVEN** the `fab change` subcommand table documents `switch` on one row
- **WHEN** `/fab-proceed` is added as a consumer
- **THEN** the addition SHALL be a brief note after the subcommand table (not inline in the table cell)

### Requirement: Affected memory update SHALL add CLI invocation detail

The `docs/memory/fab-workflow/execution-skills.md` memory file's `/fab-proceed` section already describes the skill's dispatch table and behavior. If the CLI invocation patterns (`fab resolve --folder`, `fab change switch`) are not already present in the memory, they SHOULD be confirmed as present or added.

#### Scenario: Memory already covers CLI invocations
- **GIVEN** `execution-skills.md` already mentions `fab resolve --folder` in the `/fab-proceed` section
- **WHEN** the hydrate step checks for completeness
- **THEN** no memory modification is needed for CLI invocation patterns

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | No new CLI commands to document | Confirmed from intake #1 — PR 278 added a skill only | S:95 R:95 A:95 D:95 |
| 2 | Certain | Document fab-proceed as consumer of fab resolve and fab change switch | Confirmed from intake #2 — verified by reading fab-proceed.md Step 1 and dispatch table | S:90 R:90 A:90 D:90 |
| 3 | Certain | No new fab log Callers row needed | Upgraded from intake #3 Confident → Certain — verified fab-proceed.md confirms no direct fab log calls | S:90 R:90 A:90 D:90 |
| 4 | Certain | Use Notable callers pattern for fab resolve | _cli-fab.md has Callers table under fab log; fab resolve has no callers section yet — add one using the same pattern | S:85 R:95 A:90 D:85 |
| 5 | Certain | Memory execution-skills already covers CLI patterns | Verified — execution-skills.md line 15 mentions `fab resolve --folder` explicitly | S:95 R:95 A:95 D:95 |
| 6 | Certain | This is a docs change type | Confirmed from intake #5 — only markdown files modified | S:95 R:95 A:95 D:95 |

6 assumptions (6 certain, 0 confident, 0 tentative, 0 unresolved).
