# Intake: Create Operator 6

**Change**: 260323-1arx-create-operator-six
**Created**: 2026-03-23
**Status**: Draft

## Origin

> creating operator 6 - clean rewrite

Conversational design session. User and agent collaboratively designed the operator6 skill section by section, making explicit decisions on what to keep, cut, and restructure from operator5. The skill file was written during the discussion and exists at `fab/.kit/skills/fab-operator6.md` (~275 lines).

## Why

The operator skill has evolved through two major revisions (operator4 → operator5), accreting features at each step: use case registries, `.fab-operator.yaml` config files, playbook catalogs, autopilot queues. The result is a ~430-line skill file (operator5) that attempts to cover every coordination scenario exhaustively. This creates several problems:

1. **Prompt bloat** — the operator runs in a long-lived session. A 430-line system prompt consumes context budget that should be reserved for coordination state (pane maps, stage snapshots).
2. **Rigid playbook model** — hard-coded playbooks (broadcast, sequenced rebase, merge PRs, etc.) are brittle. New coordination patterns require editing the skill file.
3. **Redundant specification** — many "playbooks" are just common-sense applications of the core principles (coordinate don't execute, state re-derivation, pre-send validation). Spelling them out adds length without adding capability.
4. **Configuration overhead** — `.fab-operator.yaml` with use case toggling adds a persistence layer the operator must manage, creating another source of state to track.

If left unaddressed, each new coordination pattern will further inflate the skill, and the operator's effectiveness will degrade as its system prompt competes with working context for token budget.

## What Changes

### New skill: `fab/.kit/skills/fab-operator6.md`

A ground-up rewrite of the operator skill (~275 lines vs operator5's ~430). Designed in a collaborative discussion session. Seven sections:

**§1 Principles** — 6 principles (up from 4):
- Carried forward: coordinate don't execute, not a lifecycle enforcer, context discipline, state re-derivation
- New: **automate the routine** — operator exists to take work off the user's hands, not just relay commands
- New: **self-manage context** — run `/clear` when context approaches capacity; continuity via `.fab-operator.yaml`

**§2 Startup** — heavier context loading than op5. Operator reads `_cli-fab.md`, `_cli-external.md`, and `_naming.md` in addition to the always-load layer, giving it full command vocabulary for routing decisions. Tmux is a **hard stop** (no status-only fallback). Orientation moved into the loop as the first tick output.

**§3 Safety** — confirmation tiers (autopilot removed from destructive tier — individual actions within autopilot have their own tiers), pre-send validation (expanded to 4 steps: pane exists, agent idle, change active, branch aligned), branch fallback, bounded retries.

**§4 The Loop** — 3-minute tick interval. `.fab-operator.yaml` simplified to just monitored set + autopilot queue (no use case registry, no toggling). Loop stops when monitored set is empty and no autopilot. User prompt restarts it. Tick behavior: refresh → stage advance → completion → review failure → pane death → auto-nudge → stuck detection → autopilot step → persist → loop lifecycle check.

**§5 Auto-Nudge** — carried forward largely unchanged from op5. Question detection patterns, guards, answer model, re-capture before send, logging. Too precise to compress significantly.

**§6 Coordination Patterns** — replaces op5's playbook catalog. Includes pipeline reference, command vocabulary, spawn command pattern (reads `agent.spawn_command` from config.yaml, defaults to `claude --dangerously-skip-permissions`), "working a change" example with concrete tmux spawn command, and autopilot as a coordination pattern (not a standalone section).

**§7 Key Properties** — standard table, tmux requirement changed to hard stop.

### `.fab-operator.yaml` redesign

Simplified from use-case registry to pure state persistence:

```yaml
monitored:
  r3m7:
    pane: "%3"
    stage: apply
    agent: active
    enrolled_at: "2026-03-23T17:30:00Z"
    last_transition: "2026-03-23T17:32:00Z"
autopilot:
  queue: [ab12, cd34, ef56]
  current: cd34
  completed: [ab12]
  state: running  # running | paused | null
```

Purpose: survive `/clear` cycles. Operator re-reads this file after clearing context to restore monitored set and autopilot queue.

### Retire operator 4

- Deleted `fab/.kit/skills/fab-operator4.md`
- Deleted `fab/.kit/scripts/fab-operator4.sh`

### Operator 5 retained temporarily

`fab-operator5.md` and `fab-operator5.sh` kept until operator6 is proven out. Will be removed in a follow-up.

### New launcher: `fab/.kit/scripts/fab-operator6.sh`

Created. Identical structure to op5 launcher, references `/fab-operator6`.

### Update deployed copies

After skill source changes, run `fab-sync.sh` to update `.claude/skills/` deployed copies.

## Affected Memory

- `fab-workflow/execution-skills`: (modify) Update operator section to reflect operator6 replacing 4 (and eventually 5)
- `fab-workflow/kit-architecture`: (modify) Update references to operator skill and script files

## Impact

- **Skills**: `fab/.kit/skills/fab-operator6.md` (new), `fab-operator4.md` (deleted)
- **Scripts**: `fab/.kit/scripts/fab-operator6.sh` (new), `fab-operator4.sh` (deleted)
- **Deployed copies**: `.claude/skills/fab-operator*.md` files updated via sync
- **Spec files**: `docs/specs/skills/SPEC-fab-operator*.md` may need updating per constitution constraint
- **No config.yaml changes** — operator is not referenced in project config
- **No template changes** — operator doesn't use artifact templates

## Open Questions

None — all design decisions resolved during collaborative discussion session.

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Core principles (coordinate don't execute, state re-derivation, context discipline, not a lifecycle enforcer) carry forward | Discussed — foundational to operator concept | S:95 R:90 A:95 D:95 |
| 2 | Certain | Two new principles: automate the routine, self-manage context | Discussed — user explicitly requested both | S:95 R:90 A:90 D:95 |
| 3 | Certain | Monitoring tick, question detection, and auto-nudge carry forward unchanged | Discussed — operator's primary value, too precise to compress | S:95 R:85 A:90 D:95 |
| 4 | Certain | Safety model carries forward with autopilot removed from destructive tier | Discussed — user explicitly requested removal; individual actions have own tiers | S:95 R:95 A:95 D:95 |
| 5 | Certain | Tmux is a hard stop — no status-only fallback | Discussed — user said "refuse to start" | S:95 R:85 A:90 D:95 |
| 6 | Certain | Drop exhaustive playbook catalog in favor of principles + examples | Discussed — operator infers patterns from pipeline knowledge | S:95 R:80 A:85 D:90 |
| 7 | Certain | `.fab-operator.yaml` kept for monitored set + autopilot queue persistence | Discussed — user confirmed persistence needed for `/clear` recovery | S:95 R:85 A:90 D:95 |
| 8 | Certain | Drop use case registry and toggling from `.fab-operator.yaml` | Discussed — monitoring is always-on; Linear/PR concerns are conversational | S:90 R:80 A:85 D:90 |
| 9 | Certain | 3-minute loop interval | Discussed — user chose this explicitly | S:95 R:95 A:80 D:95 |
| 10 | Certain | Operator loads full command vocabulary (_cli-fab, _cli-external, _naming) at startup | Discussed — operator needs pipeline understanding to route correctly | S:95 R:90 A:90 D:95 |
| 11 | Certain | Orientation output moved into the loop (first tick) | Discussed — keeps startup minimal | S:90 R:90 A:85 D:90 |
| 12 | Certain | Operator4 deleted immediately, operator5 retained until proven out | Discussed — user confirmed this approach | S:95 R:80 A:90 D:95 |
| 13 | Certain | Autopilot is a coordination pattern, not a standalone section | Discussed — user confirmed autopilot is "a coordination pattern" | S:95 R:80 A:85 D:90 |
| 14 | Confident | Spawn command reads `agent.spawn_command` from config.yaml | Derived from `lib/spawn.sh` pattern — consistent with existing launcher scripts | S:80 R:90 A:90 D:85 |
| 15 | Confident | ~275 lines (38% reduction from op5) is sufficient | Discussed — user accepted the scaffold, may compress further during review | S:75 R:95 A:75 D:80 |

15 assumptions (13 certain, 2 confident, 0 tentative, 0 unresolved).
