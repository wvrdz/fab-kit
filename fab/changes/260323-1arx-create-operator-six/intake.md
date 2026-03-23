# Intake: Create Operator 6

**Change**: 260323-1arx-create-operator-six
**Created**: 2026-03-23
**Status**: Draft

## Origin

> creating operator 6 - clean rewrite

Extended conversational design session. User and agent collaboratively designed the operator6 skill section by section, making explicit decisions on what to keep, cut, and restructure from operator5. Two rounds of parallel sub-agent review (5 agents each) identified and fixed gaps. The skill file exists at `fab/.kit/skills/fab-operator6.md` (~376 lines).

## Why

The operator skill has evolved through two major revisions (operator4 → operator5), accreting features at each step: use case registries, `.fab-operator.yaml` config files, playbook catalogs, autopilot queues. The result is a ~430-line skill file (operator5) that attempts to cover every coordination scenario exhaustively. This creates several problems:

1. **Prompt bloat** — the operator runs in a long-lived session. A 430-line system prompt consumes context budget that should be reserved for coordination state (pane maps, stage snapshots).
2. **Rigid playbook model** — hard-coded playbooks (broadcast, sequenced rebase, merge PRs, etc.) are brittle. New coordination patterns require editing the skill file.
3. **Redundant specification** — many "playbooks" are just common-sense applications of the core principles (coordinate don't execute, state re-derivation, pre-send validation). Spelling them out adds length without adding capability.
4. **Configuration overhead** — `.fab-operator.yaml` with use case toggling adds a persistence layer the operator must manage, creating another source of state to track.

If left unaddressed, each new coordination pattern will further inflate the skill, and the operator's effectiveness will degrade as its system prompt competes with working context for token budget.

## What Changes

### New skill: `fab/.kit/skills/fab-operator6.md` (9 sections, ~376 lines)

A ground-up rewrite of the operator skill. Designed in a collaborative discussion session with two rounds of parallel sub-agent review.

**§1 Principles** — 6 principles (up from 4):
- Carried forward: coordinate don't execute (with exception for maintenance actions like merge/archive), not a lifecycle enforcer, context discipline, state re-derivation
- New: **automate the routine** — operator exists to take work off the user's hands
- New: **self-manage context** — run `/clear` when context approaches capacity; continuity via `.fab-operator.yaml`

**§2 Startup** — heavier context loading than op5. Operator reads `_cli-fab.md`, `_cli-external.md`, and `_naming.md` in addition to the always-load layer, giving it full command vocabulary for routing decisions. Tmux is a **hard stop** (no status-only fallback). Loop starts only when monitored set is non-empty, autopilot active, or watches exist.

**§3 Safety** — confirmation tiers (autopilot removed from destructive tier), pre-send validation (4 steps: pane exists, agent idle, change active, branch aligned), branch fallback (proceeds without confirmation on single match), bounded retries.

**§4 The Loop** — 3-minute tick interval. Persistent `tick_count`. Framed status output with color indicators:
```
── Operator ── 17:32 ── tick #47 ── 3 monitored · autopilot 1/3 · 1 watch ──
  r3m7  🟢 apply → review
  k8ds  🟡 review · idle 18m ⚠
  👁 linear-bugs  2 known · 1 completed · last check 17:29
```
Tick steps: snapshot → auto-nudge → watches → autopilot → removals → persist → lifecycle check.

**§5 Auto-Nudge** — carried forward from op5 with pre-send validation reference. Question detection patterns, guards, answer model, re-capture before send, logging.

**§6 Coordination Patterns** — replaces op5's playbook catalog. Pipeline reference, spawn command pattern (reads `agent.spawn_command` from config.yaml), "working a change" example, autopilot with queue ordering strategies (user-provided with `--base`, confidence-based, hybrid), `--reuse` for respawns, cleanup step, rebase targets `origin/main` explicitly.

**§7 Watches** — new capability replacing op5's hard-coded `linear-inbox` and `pr-freshness` use cases. Generic external source monitoring:
- 4 structured fields: `source`, `query`, `stop_stage`, `known` (capped at 200)
- `instructions` field for free-form natural language (trigger conditions, concurrency limits, label filters)
- `spawned_by` linkage on monitored entries for deterministic concurrency counting
- `enabled` for pause/resume, `last_checked`/`last_error` for observability
- `completed` list tracks items that reached `stop_stage`
- Error handling: 3-failure auto-disable with backoff
- "Test watch" dry-run command for validating instructions before going live

**§8 Configuration** — loop interval (3m) and stuck threshold (15m), overridable via natural language.

**§9 Key Properties** — standard table.

### `.fab-operator.yaml` redesign

Persistent state with three concerns:

```yaml
tick_count: 47
monitored:
  r3m7:
    pane: "%3"
    stage: apply
    agent: active
    stop_stage: null
    spawned_by: null
    enrolled_at: "2026-03-23T17:30:00Z"
    last_transition: "2026-03-23T17:32:00Z"
autopilot:
  queue: [ab12, cd34, ef56]
  current: cd34
  completed: [ab12]
  state: running
watches:
  linear-bugs:
    enabled: true
    source: linear
    query: { project: "DEV", status: [Backlog, Todo], assignee: "@me" }
    stop_stage: intake
    known: [DEV-988, DEV-992]
    completed: [DEV-985]
    last_checked: "2026-03-23T17:29:00Z"
    last_error: null
    instructions: >
      Spawn agents for issues older than 1 hour with label 'bug'.
      Max 2 concurrent agents from this watch.
```

### Retire operator 4

- Deleted `fab/.kit/skills/fab-operator4.md`
- Deleted `fab/.kit/scripts/fab-operator4.sh`

### Operator 5 retained temporarily

`fab-operator5.md` and `fab-operator5.sh` kept until operator6 is proven out. Will be removed in a follow-up.

### New launcher: `fab/.kit/scripts/fab-operator6.sh`

Created. Identical structure to op5 launcher, references `/fab-operator6`. Byobu references removed.

### README updated

Multi-Agent Coordination table updated with full operator lineage (v1-v6), showing status (retired/available/current) and incremental capability additions.

### Update deployed copies

After skill source changes, run `fab-sync.sh` to update `.claude/skills/` deployed copies.

## Affected Memory

- `fab-workflow/execution-skills`: (modify) Update operator section to reflect operator6 replacing 4 (and eventually 5)
- `fab-workflow/kit-architecture`: (modify) Update references to operator skill and script files

## Impact

- **Skills**: `fab/.kit/skills/fab-operator6.md` (new), `fab-operator4.md` (deleted)
- **Scripts**: `fab/.kit/scripts/fab-operator6.sh` (new), `fab-operator4.sh` (deleted)
- **README**: Multi-Agent Coordination table updated
- **Deployed copies**: `.claude/skills/fab-operator*.md` files updated via sync
- **No config.yaml changes** — operator is not referenced in project config
- **No template changes** — operator doesn't use artifact templates

## Open Questions

None — all design decisions resolved during collaborative discussion session.

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Core principles carry forward with "coordinate don't execute" clarified (maintenance actions exempt) | Discussed — foundational, with explicit exception for merge/archive/delete | S:95 R:90 A:95 D:95 |
| 2 | Certain | Two new principles: automate the routine, self-manage context | Discussed — user explicitly requested both | S:95 R:90 A:90 D:95 |
| 3 | Certain | Monitoring tick, question detection, and auto-nudge carry forward | Discussed — operator's primary value, too precise to compress | S:95 R:85 A:90 D:95 |
| 4 | Certain | Safety model carries forward with autopilot removed from destructive tier | Discussed — user explicitly requested removal | S:95 R:95 A:95 D:95 |
| 5 | Certain | Tmux is a hard stop — no status-only fallback | Discussed — user said "refuse to start" | S:95 R:85 A:90 D:95 |
| 6 | Certain | Drop exhaustive playbook catalog in favor of principles + examples | Discussed — operator infers patterns from pipeline knowledge | S:95 R:80 A:85 D:90 |
| 7 | Certain | `.fab-operator.yaml` kept for monitored set + autopilot queue + watches persistence | Discussed — user confirmed persistence needed for `/clear` recovery | S:95 R:85 A:90 D:95 |
| 8 | Certain | Drop use case registry; replace with generic watches system | Discussed — user designed watches conversationally | S:95 R:80 A:85 D:90 |
| 9 | Certain | 3-minute loop interval | Discussed — user chose this explicitly | S:95 R:95 A:80 D:95 |
| 10 | Certain | Operator loads full command vocabulary at startup | Discussed — needs pipeline understanding to route correctly | S:95 R:90 A:90 D:95 |
| 11 | Certain | Framed tick output with color indicators and tick count | Discussed — user chose option A (grouped) then refined to box-drawn frame | S:95 R:90 A:85 D:90 |
| 12 | Certain | Operator4 deleted immediately, operator5 retained until proven out | Discussed — user confirmed this approach | S:95 R:80 A:90 D:95 |
| 13 | Certain | Autopilot is a coordination pattern with queue ordering strategies | Discussed — user confirmed; confidence-based/hybrid/user-provided restored after review | S:95 R:80 A:85 D:90 |
| 14 | Certain | Watches: 4 structured fields + instructions for everything else | Discussed — user chose to collapse trigger/action into instructions, keep stop_stage structured | S:95 R:85 A:85 D:90 |
| 15 | Certain | Watches: spawned_by linkage, known cap (200), enabled, last_checked/last_error, completed list | Discussed — user approved all 7 review findings | S:95 R:85 A:90 D:95 |
| 16 | Certain | "Test watch" dry-run command | Discussed — user asked how it would work, approved the design | S:95 R:90 A:85 D:90 |
| 17 | Certain | Rebase targets origin/main explicitly, not vague "rebase on main" | Backlog item [djkp] — addressed per user request | S:95 R:90 A:90 D:95 |
| 18 | Certain | Branch fallback proceeds without confirmation | Backlog item [02eh] — user said "just proceed" | S:95 R:85 A:90 D:95 |
| 19 | Certain | README updated with operator evolution table (v1-v6) | Discussed — user requested table entries tracking operator evolution | S:95 R:90 A:90 D:95 |
| 20 | Confident | Spawn command reads `agent.spawn_command` from config.yaml | Derived from `lib/spawn.sh` pattern — consistent with existing launchers | S:80 R:90 A:90 D:85 |

20 assumptions (19 certain, 1 confident, 0 tentative, 0 unresolved).
