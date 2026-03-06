# Intake: Add `/fab-conductor` Skill

**Change**: 260306-qkov-conductor-skill
**Created**: 2026-03-06
**Status**: Draft

## Origin

> Extended `/fab-discuss` session exploring multi-agent coordination in tmux. The user described wanting an "orchestrator" that can observe agents running in tmux panes and interact with them as the user — sending commands, coordinating rebases, spawning worktrees. The name evolved from "orchestrator" to "connector" to "conductor." A detailed spec was drafted at `docs/specs/skills/SPEC-fab-conductor.md` during the discussion.

Conversational — multiple rounds of design refinement before this intake.

## Why

1. **Tab-hopping doesn't scale** — the assembly-line workflow (`batch-fab-switch-change`, `batch-fab-new-backlog`) creates N tmux tabs with N agents. Today, coordinating them requires the user to manually hop between tabs: checking status, sending commands, sequencing rebases. With 5+ agents this becomes the bottleneck.

2. **Cross-agent actions have no home** — rebasing agent B after agent A finishes, broadcasting `/fab-continue` to all idle agents, merging completed PRs in order — these require awareness of multiple agents simultaneously. No existing skill or script handles this. The user does it manually.

3. **Natural language over scripts** — `batch-fab-switch-change --all` works but is rigid. "Start working on the retry logic idea" is more natural than remembering script flags and manually creating worktrees. The conductor translates intent into multi-agent actions.

## What Changes

### New skill: `/fab-conductor`

A Claude Code skill that runs in a dedicated tmux pane. It loads the always-load layer (like `/fab-discuss`) but NOT change-specific artifacts — it only needs coordination primitives.

### Primitives consumed

The conductor relies on:

| Primitive | Source | Purpose |
|-----------|--------|---------|
| `fab pane-map` | New CLI subcommand (separate change bh45) | Unified view: pane → worktree → change → stage → agent state |
| `fab status show --all` | Existing CLI | Change stage and progress |
| `fab runtime` | Existing CLI | Agent idle/active detection |
| `fab send-keys <change> "<text>"` | **New** CLI subcommand | Send text to a target agent's tmux pane |

### New subcommand: `fab send-keys`

```
fab send-keys <change> "<text>"
```

Resolves `<change>` to a tmux pane (via pane-map data), then runs `tmux send-keys -t <pane> "<text>" Enter`. Uses standard change resolution (4-char ID, substring, full name).

Safety requirements:
- Validate target pane exists before sending
- Return error if pane is gone (agent exited, tab closed)
- The conductor checks idle state before calling `send-keys`; the subcommand itself does NOT enforce idle checks (separation of concerns — policy in the skill, mechanism in the CLI)

### Use cases

**1. Broadcast a command to all agents**
User: "run /fab-continue in all idle agents"
- Refreshes pane map, filters for idle agents, sends `/fab-continue` to each via `fab send-keys`

**2. Sequenced rebase after completion**
User: "when r3m7 finishes, rebase k8ds on main"
- Polls `fab status show --all`, waits for r3m7 to reach target stage, then sends rebase command to k8ds

**3. Merge completed PRs**
User: "merge all PRs that are ready"
- Checks for changes at `ship` or `review-pr (pass)` stage, runs `gh pr merge` directly from conductor's shell

**4. Spawn new worktree + agent from idea**
User: "start working on the retry logic idea"
- Looks up idea via `idea show`, creates worktree via `wt-create --non-interactive`, opens tmux tab, sends `/fab-new <description>`

**5. Status dashboard**
User: "what's everyone doing?"
- Runs `fab pane-map`, presents human-readable summary

**6. Unstick a stuck agent**
User: "r3m7 looks stuck, nudge it"
- Confirms agent is idle at a non-terminal stage, reads current stage, sends `/fab-continue` or a targeted prompt
- Can diagnose via `tmux capture-pane -t <pane> -p` to read recent terminal output and reason about what went wrong

**7. Sequential pipeline execution (autopilot)**
User: "take all these changes to completion: bh45, qkov, ab12" — or "run all my changes, highest confidence first."
- Autonomous pipeline operator: for each change in order, creates worktree, opens agent, checks confidence, runs `/fab-ff` or `/fab-fff`, monitors until done, merges PR, rebases remaining changes on updated main, moves to next
- Ordering: user-provided, confidence-based (highest first), or hybrid (user constraints + confidence optimization)
- Failure handling: flags low confidence to user, lets rework loops run up to budget, detects rebase conflicts, reports agent crashes
- Resumable: reconstructs state from `fab status show --all` if interrupted
- Requires deep lifecycle knowledge: confidence gates, stage progression, `/fab-ff` vs `/fab-fff` differences, rework budget, PR merge mechanics

**8. Notification surface**
The conductor pane is the single notification surface for cross-agent events. "Tell me when r3m7 finishes" or "notify me if anything fails" — the conductor holds the instruction in conversation context and checks on the next user interaction. User-driven, not event-driven.

### Interaction model

- Long-running Claude session in its own tmux pane
- NOT a daemon or polling loop — waits for user input, acts, reports back
- User talks to it like any other Claude session

### Confirmation model

| Risk | Examples | Behavior |
|------|----------|----------|
| Read-only | Status check, pane map | No confirmation |
| Recoverable | Send `/fab-continue`, rebase | Announce before sending |
| Destructive | Merge PR, archive, delete worktree | Confirm before executing |

### Pre-send validation

Before sending keys to any pane, the conductor MUST:
1. Verify target pane still exists
2. Check agent is idle via `fab runtime` — sending to a busy agent risks corrupting its work
3. If agent is not idle, warn user and ask for confirmation

### What the conductor is not

- Not a daemon — it's a Claude session, not a background process
- Not a lifecycle enforcer — agents manage their own pipeline
- Not a scheduler — no queuing or priority management
- Not a replacement for batch scripts — those remain for non-interactive use

## Affected Memory

- `fab-workflow/execution-skills`: (modify) Document the conductor skill alongside existing execution skills
- `fab-workflow/kit-architecture`: (modify) Document `fab send-keys` subcommand
- `fab-workflow/distribution`: (modify) Note that the conductor skill ships with the kit

## Impact

- **New skill file**: `fab/.kit/skills/fab-conductor.md` (source) → `.claude/skills/fab-conductor` (deployed)
- **New spec file**: `docs/specs/skills/SPEC-fab-conductor.md` (already drafted)
- **Go binary**: New `send-keys` subcommand
- **Dependencies**: `fab pane-map` (change bh45) must land first — conductor consumes its output
- **Existing skills**: No modifications — conductor is purely additive
- **External deps**: tmux (required at runtime), `gh` (for PR merging use case)

## Open Questions

None — all four open questions resolved during discussion (see Assumptions #11–14).

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Name is "conductor" | Discussed — user explicitly chose "conductor" over "orchestrator" and "connector" | S:95 R:95 A:95 D:95 |
| 2 | Certain | This is a skill (`/fab-conductor`), not a package | Discussed — requires AI agent for NL interpretation, not a shell script | S:90 R:85 A:95 D:95 |
| 3 | Certain | Conductor is NOT a lifecycle enforcer | Discussed — agents self-govern; conductor handles cross-agent coordination only | S:95 R:90 A:90 D:95 |
| 4 | Certain | Observation via `fab pane-map` + `fab status show --all` + `fab runtime` | Discussed — three primitives for unified observation | S:90 R:85 A:90 D:90 |
| 5 | Certain | Interaction via `fab send-keys <change> "<text>"` | Discussed — indirection through fab CLI for standard change resolution | S:90 R:80 A:90 D:90 |
| 6 | Certain | Pre-send idle check is mandatory | Discussed — sending to busy agent corrupts its work | S:90 R:70 A:90 D:95 |
| 7 | Certain | Not a daemon — user-driven Claude session | Discussed — waits for input, acts, reports back | S:95 R:85 A:90 D:95 |
| 8 | Confident | Confirmation model: read-only (none), recoverable (announce), destructive (confirm) | Discussed — three tiers. Exact boundary of "destructive" vs "recoverable" may shift | S:80 R:80 A:80 D:75 |
| 9 | Certain | Eight use cases: broadcast, sequenced rebase, merge PRs, spawn worktree, status dashboard, unstick agent, sequential pipeline execution (autopilot), notification surface | Discussed — all eight confirmed through conversation | S:90 R:85 A:85 D:90 |
| 10 | Confident | Depends on `fab pane-map` (change bh45) landing first | Discussed — conductor consumes pane-map output. Could degrade gracefully without it but design assumes it | S:80 R:75 A:85 D:80 |
| 11 | Certain | No standing orders in v1 — session context suffices | Resolved — user agreed. Persistence deferred until real friction. Future home: `fab/project/conductor.yaml` | S:90 R:85 A:85 D:90 |
| 12 | Certain | User-driven, not event-driven — conductor pane is single notification surface | Resolved — user agreed. Refreshes state before each action, holds "notify me" instructions in conversation context | S:90 R:80 A:85 D:90 |
| 13 | Certain | Single-repo only, by design | Resolved — user confirmed. Fab-kit's worktree model is same-repo checkouts | S:95 R:90 A:95 D:95 |
| 14 | Certain | `fab send-keys` targets fab-managed worktrees only | Resolved — user confirmed. Change resolution requires a change argument; raw `tmux send-keys` available for edge cases | S:90 R:85 A:90 D:90 |
| 15 | Certain | Autopilot mode supports user-provided, confidence-based, or hybrid ordering | Discussed — user wants both explicit ordering and conductor-determined ordering | S:85 R:80 A:80 D:85 |
| 16 | Confident | Autopilot flags failures to user rather than making autonomous recovery decisions | Discussed — low confidence, rebase conflicts, and exhausted rework budgets all escalate to user. Conservative: skip and continue rather than guess | S:75 R:70 A:75 D:70 |
| 17 | Confident | Conductor must understand change lifecycle deeply: confidence gates, stage progression, /fab-ff vs /fab-fff, rework budget | Discussed — autopilot requires lifecycle knowledge to make gating and routing decisions | S:85 R:75 A:80 D:80 |

17 assumptions (13 certain, 4 confident, 0 tentative, 0 unresolved).
