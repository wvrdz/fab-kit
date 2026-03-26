# Spec: fab-proceed Orchestrator

**Change**: 260325-kxw7-fab-proceed-orchestrator
**Created**: 2026-03-26
**Affected memory**: `docs/memory/fab-workflow/execution-skills.md`

## Non-Goals

- Accepting arguments or flags — the skill infers everything from context
- Creating or switching worktrees — stays in the current worktree
- Implementing pipeline logic — delegates entirely to `/fab-fff`
- Providing `--force` passthrough — if the user needs `--force`, they invoke `/fab-fff` directly

## Skill: State Detection

### Requirement: State Detection Pipeline

`/fab-proceed` SHALL detect the current pipeline state by executing the following checks in order:

1. **Active change check**: Run `fab resolve --folder 2>/dev/null`. If exits 0, an active change exists.
2. **Branch check**: Run `git branch --show-current` and compare with the resolved change folder name. Match means the branch is already set up.
3. **Unactivated intake check**: If no active change, scan `fab/changes/` (excluding `archive/`) for folders. If exactly one non-archived change exists, use it. If multiple exist, select the most recently created by folder date prefix (`YYMMDD`).
4. **Conversation context check**: If no intake exists anywhere, evaluate whether prior conversation contains substantive discussion. An empty conversation, a greeting-only conversation, or a conversation with no technical content SHALL be treated as "no context."

The skill MUST NOT prompt the user for input at any detection step — it either resolves automatically or errors.

#### Scenario: Active change with matching branch

- **GIVEN** an active change `260325-kxw7-fab-proceed-orchestrator` pointed to by `.fab-status.yaml`
- **AND** the current git branch is `260325-kxw7-fab-proceed-orchestrator`
- **WHEN** `/fab-proceed` is invoked
- **THEN** the skill skips switch and branch steps
- **AND** dispatches directly to `/fab-fff`

#### Scenario: Active change without matching branch

- **GIVEN** an active change exists via `.fab-status.yaml`
- **AND** the current git branch is `main`
- **WHEN** `/fab-proceed` is invoked
- **THEN** the skill dispatches `/git-branch` (which creates/checks out the matching branch)
- **AND** then dispatches `/fab-fff`

#### Scenario: Unactivated intake exists

- **GIVEN** no active change (`.fab-status.yaml` absent or dangling)
- **AND** `fab/changes/260325-kxw7-fab-proceed-orchestrator/` exists with `intake.md`
- **WHEN** `/fab-proceed` is invoked
- **THEN** the skill dispatches `/fab-switch` with the change name
- **AND** then dispatches `/git-branch`
- **AND** then dispatches `/fab-fff`

#### Scenario: Multiple unactivated intakes

- **GIVEN** no active change
- **AND** two change folders exist: `260325-ab12-feature-a/` and `260324-cd34-feature-b/`
- **WHEN** `/fab-proceed` is invoked
- **THEN** the skill selects `260325-ab12-feature-a/` (most recent by date prefix)
- **AND** proceeds with switch → branch → fff

#### Scenario: Conversation context, no intake

- **GIVEN** no active change and no change folders in `fab/changes/`
- **AND** the conversation contains substantive technical discussion about a feature
- **WHEN** `/fab-proceed` is invoked
- **THEN** the skill synthesizes a description from the conversation context
- **AND** dispatches `/fab-new` with the synthesized description
- **AND** then dispatches `/fab-switch` with the newly created change name
- **AND** then dispatches `/git-branch`
- **AND** then dispatches `/fab-fff`

#### Scenario: No context, no intake

- **GIVEN** no active change, no change folders, and no substantive conversation history
- **WHEN** `/fab-proceed` is invoked
- **THEN** the skill outputs: `Nothing to proceed with — start a discussion or run /fab-new first.`
- **AND** stops without invoking any sub-skills

### Requirement: Conversation Context Synthesis

When `/fab-proceed` needs to create an intake (no existing intake found), it SHALL synthesize a description from the conversation by extracting:

- Decisions made with rationale
- Alternatives rejected and why
- Constraints identified
- Specific values agreed upon

The synthesized description MUST be substantive enough for `/fab-new` to generate a complete intake without prompting. The synthesis follows the same pattern as `/fab-new` Step 4 (Conversation Context Mining).

#### Scenario: Rich discussion context

- **GIVEN** a conversation that discussed "adding retry logic to the queue with exponential backoff, max 3 retries, 1s base delay"
- **WHEN** `/fab-proceed` synthesizes the description
- **THEN** the description includes the specific values (3 retries, 1s base, exponential backoff)
- **AND** `/fab-new` receives enough detail to generate a complete intake

#### Scenario: Minimal discussion context

- **GIVEN** a conversation that only said "we should add retry logic"
- **WHEN** `/fab-proceed` synthesizes the description
- **THEN** the description captures what was said without fabricating details
- **AND** `/fab-new` generates the intake with appropriate Tentative/Unresolved assumptions for the gaps

## Skill: Dispatch Behavior

### Requirement: Subagent Dispatch Pattern

Each prefix step (fab-new, fab-switch, git-branch) SHALL be dispatched as a subagent using the Agent tool (`general-purpose` subagent) per `_preamble.md` § Subagent Dispatch. Each subagent prompt MUST include standard subagent context files.

The final `/fab-fff` delegation is NOT dispatched as a subagent — it is invoked via the Skill tool in the current context, since it is the terminal operation and its output should be visible to the user.

#### Scenario: fab-new subagent dispatch

- **GIVEN** conversation context exists but no intake
- **WHEN** `/fab-proceed` dispatches `/fab-new`
- **THEN** the subagent prompt includes the synthesized description
- **AND** the subagent reads `fab/.kit/skills/fab-new.md` and follows its full behavior
- **AND** the subagent returns the created change folder name

#### Scenario: fab-switch subagent dispatch

- **GIVEN** an unactivated intake exists
- **WHEN** `/fab-proceed` dispatches `/fab-switch`
- **THEN** the subagent runs `fab/.kit/bin/fab change switch "<change-name>"`
- **AND** returns the switch confirmation

#### Scenario: git-branch subagent dispatch

- **GIVEN** an active change exists but no matching branch
- **WHEN** `/fab-proceed` dispatches `/git-branch`
- **THEN** the subagent reads `fab/.kit/skills/git-branch.md` and follows its behavior
- **AND** returns the branch creation/checkout result

### Requirement: fab-fff Terminal Delegation

`/fab-proceed` SHALL invoke `/fab-fff` via the Skill tool (not as a subagent) as its final step. This ensures `/fab-fff` runs in the main context with full user visibility of its output, confidence gates, and pipeline progress.

The skill SHALL NOT pass `--force` or any other flags to `/fab-fff`. If `/fab-fff` fails a confidence gate, it stops normally and the user intervenes.

#### Scenario: fab-fff invocation

- **GIVEN** all prefix steps (switch, branch) have completed
- **WHEN** `/fab-proceed` invokes `/fab-fff`
- **THEN** `/fab-fff` runs its own preflight, checks confidence gates, and executes the pipeline
- **AND** its output is directly visible to the user

#### Scenario: fab-fff gate failure

- **GIVEN** `/fab-proceed` invoked `/fab-fff`
- **AND** the indicative confidence is 2.0 (below 3.0 threshold)
- **WHEN** `/fab-fff` checks the intake gate
- **THEN** `/fab-fff` stops with its standard gate failure message
- **AND** `/fab-proceed` does not retry or bypass the gate

### Requirement: No Arguments or Flags

`/fab-proceed` SHALL NOT accept any arguments or flags. The skill header is:

```
# /fab-proceed
```

If the user passes arguments (e.g., `/fab-proceed --force`), the skill SHALL ignore them — they are not parsed or passed through.

#### Scenario: Invocation with spurious arguments

- **GIVEN** a user invokes `/fab-proceed some-arg`
- **WHEN** the skill starts
- **THEN** arguments are ignored
- **AND** the skill proceeds with normal state detection

### Requirement: Idempotent Re-run

`/fab-proceed` SHALL be safe to re-run. On re-invocation, state detection identifies which steps are already complete and skips them. If the active change already has a matching branch, only `/fab-fff` is invoked. If `/fab-fff` itself is resumable (stages already `done`), it skips those stages.

#### Scenario: Re-run after partial completion

- **GIVEN** `/fab-proceed` previously ran `/fab-new` and `/fab-switch` but was interrupted before `/git-branch`
- **WHEN** `/fab-proceed` is re-invoked
- **THEN** state detection finds an active change without a matching branch
- **AND** the skill runs `/git-branch` → `/fab-fff` (skipping switch)

## Skill: Integration

### Requirement: State Table Update

`_preamble.md` state table SHALL include `/fab-proceed` in the following states:

| State | Addition |
|-------|----------|
| initialized | `/fab-proceed` (available when conversation context exists) |
| intake | `/fab-proceed` |

#### Scenario: State table at initialized state

- **GIVEN** the project is initialized but no changes exist
- **WHEN** a skill outputs a `Next:` line for the initialized state
- **THEN** `/fab-proceed` appears as an available command alongside `/fab-new`

### Requirement: Skill Registration

The skill file SHALL be created at `fab/.kit/skills/fab-proceed.md` with standard frontmatter:

```yaml
---
name: fab-proceed
description: "Context-aware orchestrator — detects state, runs prefix steps (fab-new, fab-switch, git-branch), then delegates to fab-fff."
---
```

The skill SHALL be deployed to `.claude/skills/fab-proceed/` via `fab-sync.sh` on next sync.

#### Scenario: Skill file location

- **GIVEN** the skill is implemented
- **WHEN** a user runs `fab-sync.sh`
- **THEN** `.claude/skills/fab-proceed/` is created with the deployed copy

## Skill: Output

### Requirement: Progress Reporting

`/fab-proceed` SHALL report which prefix steps it runs before delegating to `/fab-fff`:

```
/fab-proceed — detecting state...

{Step reports, one per line}

Handing off to /fab-fff...
```

Step report format (only for steps actually executed):
- `Created intake: {change-name}` (when fab-new ran)
- `Activated: {change-name}` (when fab-switch ran)
- `Branch: {branch-name} ({action})` (when git-branch ran, action = created/checked out/already active)

After the prefix report, `/fab-fff` takes over and produces its own output.

#### Scenario: Full prefix execution

- **GIVEN** conversation context exists, no intake, no active change, no branch
- **WHEN** `/fab-proceed` completes all prefix steps
- **THEN** output shows all three step reports before `/fab-fff` output

#### Scenario: Only fab-fff needed

- **GIVEN** active change exists with matching branch
- **WHEN** `/fab-proceed` runs
- **THEN** output shows only the detecting state line and the handoff line
- **AND** `/fab-fff` output follows immediately

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Always delegates to `/fab-fff`, never `/fab-ff` | Confirmed from intake #1 — user explicitly chose fff | S:95 R:90 A:95 D:95 |
| 2 | Certain | No arguments, no flags | Confirmed from intake #2 — user explicit | S:95 R:85 A:95 D:95 |
| 3 | Certain | Error on empty context | Confirmed from intake #3 — user explicit | S:95 R:90 A:95 D:95 |
| 4 | Certain | Stay in current worktree | Confirmed from intake #4 — user explicit | S:95 R:90 A:95 D:95 |
| 5 | Certain | Branch management via `/git-branch` delegation only | Confirmed from intake #5 — user explicit | S:95 R:85 A:95 D:95 |
| 6 | Certain | `/fab-fff` invoked via Skill tool (not subagent) for user visibility | Spec-level decision: terminal operation output must be visible to user. Highly reversible if wrong. | S:85 R:90 A:90 D:85 |
| 7 | Confident | Prefix steps dispatched as subagents | Confirmed from intake #6 — follows established orchestrator pattern | S:80 R:80 A:85 D:85 |
| 8 | Confident | State detection order: active change → branch → unactivated intake → conversation context | Upgraded from intake #7 — standard pipeline state inspection, mirrors existing skills | S:75 R:85 A:85 D:80 |
| 9 | Confident | Multiple unactivated intakes resolved by most-recent date prefix | Upgraded from intake #9 (was Tentative) — simplest heuristic, parallel workflows are operator-driven which would have active changes | S:65 R:65 A:70 D:60 |
| 10 | Confident | Skill file at `fab/.kit/skills/fab-proceed.md` | Confirmed from intake #8 — constitution mandates skills in `fab/.kit/skills/` | S:75 R:90 A:90 D:90 |
| 11 | Confident | Spurious arguments are silently ignored | No argument parsing = nothing to reject. Consistent with how shell scripts handle extra args to no-arg commands | S:70 R:90 A:80 D:75 |

11 assumptions (6 certain, 5 confident, 0 tentative, 0 unresolved).
