# Intake: fab-proceed Orchestrator

**Change**: 260325-kxw7-fab-proceed-orchestrator
**Created**: 2026-03-26
**Status**: Draft

## Origin

> User observed a recurring workflow pattern across sessions: `/fab-discuss` → `/fab-new` → `/fab-switch` → `/git-branch` → `/fab-fff`. The same suffix sequence (`fab-switch` → `git-branch` → `fab-fff`) also appears after standalone `/fab-new` invocations. User proposed a single command — `/fab-proceed` — that detects the current state and runs the appropriate prefix steps before delegating to `/fab-fff`. Discussion mode (conversational, multiple rounds).

Key decisions from discussion:
- **No arguments, no flags** — the skill infers everything from context
- **Always `fab-fff`** — never `fab-ff`, no `--force` passthrough
- **Synthesize from conversation** — when no intake exists, the skill mines conversation context to generate a description for `/fab-new` (same as Step 4: Conversation Context Mining in `/fab-new`)
- **Error on empty context** — if called with no prior discussion and no existing intake, error out rather than prompting
- **Stay in current worktree** — does not create or switch worktrees
- **Branch via `git-branch` only** — does not do any direct branch switching; delegates to `/git-branch` which handles create/checkout/rename logic

## Why

The `fab-switch` → `git-branch` → `fab-fff` sequence is pure ceremony — three commands with no user decisions between them. The user runs this sequence after every `/fab-new` and after many `/fab-discuss` sessions. Each invocation requires the user to remember the sequence and wait for each step to complete before invoking the next. A single command that detects state and runs the right prefix steps eliminates this friction without adding complexity — it's a thin orchestrator that delegates to existing skills.

If we don't build this, users continue manually typing 3-4 commands in sequence every time they want to go from "planning done" to "pipeline running." The time cost is modest per invocation but compounds across sessions.

## What Changes

### New skill: `fab/.kit/skills/fab-proceed.md`

A context-aware orchestrator skill that detects the current pipeline state and runs the minimum prefix steps needed before delegating to `/fab-fff`.

#### State detection and dispatch table

The skill inspects the current state and determines which steps to run:

| Detected state | Steps composed |
|---|---|
| No conversation context, no intake anywhere | **Error**: "Nothing to proceed with — start a discussion or run /fab-new first." |
| Conversation context exists, no intake | Synthesize description → `/fab-new` → `/fab-switch` → `/git-branch` → `/fab-fff` |
| Intake exists, change not active (no `.fab-status.yaml` or points elsewhere) | `/fab-switch` → `/git-branch` → `/fab-fff` |
| Active change, no matching git branch | `/git-branch` → `/fab-fff` |
| Active change + matching branch | `/fab-fff` |

#### State detection logic

1. **Check for active change**: Run `fab resolve --folder 2>/dev/null`. If exits 0, an active change exists — read its `.status.yaml` for stage.
2. **Check for matching branch**: Run `git branch --show-current` and compare with the resolved change folder name. If they match, branch exists.
3. **Check for unactivated intake**: If no active change, scan `fab/changes/` for folders with `intake` stage at `ready` state that match the conversation context (most recent creation).
4. **Check for conversation context**: If no intake exists anywhere, check whether prior conversation contains substantive discussion that can be synthesized into an intake description. If empty/greeting-only conversation, error out.

#### Subagent dispatch

Each prefix step is dispatched as a subagent (same pattern as `/fab-fff` per `_preamble.md` § Subagent Dispatch):

- **`/fab-new`**: Dispatched with synthesized description from conversation context mining. The subagent reads the `/fab-new` skill file and follows its full behavior including SRAD scoring and indicative confidence.
- **`/fab-switch`**: Dispatched with the change name (from `/fab-new` output or resolved from `fab/changes/`). Mechanical — creates the `.fab-status.yaml` symlink.
- **`/git-branch`**: Dispatched after switch. Creates or checks out the branch matching the active change.
- **`/fab-fff`**: Final delegation. The skill hands off completely — `/fab-fff` handles its own preflight, confidence gates, and full pipeline.

#### Skill properties

| Property | Value |
|----------|-------|
| Arguments | None |
| Flags | None |
| Requires active change? | No — creates one if needed |
| Runs preflight? | No — delegates to `/fab-fff` which runs its own |
| Read-only? | No — creates intake, switches active pointer, creates branch, runs pipeline |
| Idempotent? | Yes — re-running detects completed steps and skips them |
| Advances stage? | No directly — delegates to skills that do |
| Outputs `Next:` line? | Inherits from `/fab-fff` output |

### Updates to existing files

#### `_preamble.md` — State Table

Add `fab-proceed` as an available command in relevant states:

| State | Current commands | Addition |
|---|---|---|
| intake | `/fab-continue`, `/fab-ff`, `/fab-fff`, `/fab-clarify` | Add `/fab-proceed` |
| initialized | `/fab-new`, `/docs-hydrate-memory` | Add `/fab-proceed` (when conversation context exists) |

#### Spec and memory updates

- Update `docs/specs/skills.md` with `/fab-proceed` entry
- Create per-skill flow diagram at `docs/specs/skills/SPEC-fab-proceed.md`
- Update `docs/memory/fab-workflow/execution-skills.md` with the new skill

## Affected Memory

- `fab-workflow/execution-skills`: (modify) Add `fab-proceed` as a new orchestrator skill — state detection, dispatch table, relationship to `fab-fff`

## Impact

- **Skill files**: New `fab/.kit/skills/fab-proceed.md` — new skill source
- **Deployed copies**: `.claude/skills/fab-proceed/` — generated by `fab-sync.sh` on next sync
- **Preamble**: `_preamble.md` state table gets new entries for `/fab-proceed`
- **Specs**: `docs/specs/skills.md` and new `docs/specs/skills/SPEC-fab-proceed.md`
- **No Go changes**: This is a pure markdown skill — no changes to the `fab` CLI binary
- **No template changes**: Uses existing skills' templates via delegation
- **No migration needed**: New skill, no existing user data to restructure

## Open Questions

- When multiple unactivated intakes exist in `fab/changes/`, how should the skill pick which one to switch to? Most recent by folder date prefix? Or should it error with ambiguity?

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Always delegates to `/fab-fff`, never `/fab-ff` | Discussed — user explicitly chose fff | S:95 R:90 A:95 D:95 |
| 2 | Certain | No arguments, no flags | Discussed — user explicitly said "no arguments needed, also no --force" | S:95 R:85 A:95 D:95 |
| 3 | Certain | Error on empty context (no discussion, no intake) | Discussed — user explicitly confirmed "yes, error out" | S:95 R:90 A:95 D:95 |
| 4 | Certain | Stay in current worktree | Discussed — user confirmed | S:95 R:90 A:95 D:95 |
| 5 | Certain | Branch management only via `/git-branch` delegation | Discussed — user said "should not switch to a branch unless via git-branch command" | S:95 R:85 A:95 D:95 |
| 6 | Confident | Prefix steps dispatched as subagents per `_preamble.md` § Subagent Dispatch | Follows established orchestrator pattern (fab-fff). User confirmed "orchestrator over orchestrators" | S:80 R:80 A:85 D:85 |
| 7 | Confident | State detection uses `fab resolve`, `git branch --show-current`, and `fab/changes/` scan | Standard tooling — mirrors how existing skills detect state | S:70 R:85 A:85 D:80 |
| 8 | Confident | Skill lives at `fab/.kit/skills/fab-proceed.md` following standard conventions | Constitution mandates skills in `fab/.kit/skills/`. No reason to deviate | S:75 R:90 A:90 D:90 |
| 9 | Tentative | When multiple unactivated intakes exist, pick the most recently created (by folder date prefix) | Reasonable default but could cause confusion in parallel workflows. Open question raised. <!-- assumed: most-recent-intake heuristic — simplest default, may need refinement --> | S:40 R:60 A:50 D:45 |

9 assumptions (5 certain, 3 confident, 1 tentative, 0 unresolved).
