# fab-operator5

## Summary

Multi-agent coordination layer with use case registry, branch fallback resolution, and proactive monitoring. Successor to operator4 — contains all operator4 behavior plus: use case registry (toggleable monitoring concerns), branch name fallback when `fab resolve` fails, three built-in use cases (`monitor-changes`, `linear-inbox`, `pr-freshness`), tab preparation procedure, and "playbooks" (renamed from "modes of operation").

Self-contained — does not inherit from any other operator skill. All behavior is defined in `fab/.kit/skills/fab-operator5.md` plus the standard `_` files loaded via `_preamble.md`. External tool reference (`_cli-external.md`) is loaded in the operator's own startup section.

Not a lifecycle enforcer — the operator coordinates across agents and proxies routine user input, not advancing stages or making pipeline decisions.

---

## Section Structure

The skill is organized into 12 sections:

1. **Principles** — identity (coordinate don't execute), routing discipline, context discipline (never loads change artifacts), state re-derivation
2. **Startup** — always-load context layer, `_cli-external.md` load, `.fab-operator.yaml` read, orientation (pane map + use case roster + ready signal), outside-tmux degradation
3. **Safety Model** — confirmation tiers, pre-send validation, **branch fallback resolution** (new: scan `refs/heads/` + `refs/remotes/` on user-initiated resolve failure), bounded retries & escalation
4. **Use Case Registry** — `.fab-operator.yaml` config schema, conversational toggling, tick-start status roster (🟢/⚪), loop lifecycle (heartbeat model), `monitor-changes` use case (operator4's monitoring system reframed)
5. **Auto-Nudge** — question detection, answer model, re-capture before send, logging (unchanged from operator4)
6. **Use Case: `linear-inbox`** — MCP-based Linear issue detection, deduplication, spawn-on-confirm
7. **Use Case: `pr-freshness`** — `gh pr list` staleness detection, action matrix (idle/busy/no-agent/dirty), routes rebase to agents
8. **Tab Preparation Procedure** — 5-step shared procedure (pane → idle → switch → branch → dispatch) used by playbooks and use cases
9. **Playbooks** — renamed from "Modes of Operation"; 9 on-demand coordination patterns (broadcast, sequenced rebase, merge PRs, spawn agent, status dashboard, unstick agent, notification, rebase all, autopilot)
10. **Autopilot** — queue ordering, per-change loop, failure matrix, interruptibility, resumability (unchanged from operator4)
11. **Configuration** — intervals + `.fab-operator.yaml` reference
12. **Key Properties** — standard properties table + `.fab-operator.yaml` usage

---

## Primitives

All tool references are in shared `_` files — operator5 does not duplicate tool tables.

| Primitive | Reference |
|-----------|-----------|
| `fab pane-map`, `fab resolve`, `fab change list`, `fab status`, `fab score` | `_cli-fab.md` |
| `wt list`, `wt create`, `wt delete`, `tmux` commands, `/loop` | `_cli-external.md` |
| Change folder, branch, worktree naming | `_naming.md` |
| `gh pr list` | GitHub CLI (system) |
| `mcp__claude_ai_Linear__list_issues` | Linear MCP integration |

---

## What's New vs Operator4

| Feature | Operator4 | Operator5 |
|---------|-----------|-----------|
| Branch fallback | Not available | Scans `refs/heads/` + `refs/remotes/` on resolve failure |
| Monitoring model | Single-purpose monitored set | Use case registry with 3 toggleable concerns |
| `.fab-operator.yaml` | Not used | Persistent use case config |
| Status roster | Not available | Tick-start 🟢/⚪ display |
| Loop lifecycle | Starts on enrollment, stops when set empty | Runs while any use case enabled (heartbeat) |
| Linear integration | Not available | `linear-inbox` use case |
| PR freshness | Manual "Rebase all" mode only | Automatic `pr-freshness` use case |
| Tab preparation | Implicit (pre-send validation only) | Explicit 5-step procedure (switch + branch alignment) |
| Section 6 naming | "Modes of Operation" | "Playbooks" |

---

## Branch Fallback

- Placement: Section 3 (Safety Model), after Pre-Send Validation
- Trigger: user-initiated resolution only (not monitoring ticks)
- Scope: local (`refs/heads/`) + remote (`refs/remotes/`) branches
- Read-only queries: `git show` for `.status.yaml` without worktree
- Action queries: offer worktree creation via `wt create`
- Matching: same case-insensitive substring/ID matching as `fab resolve`

---

## Use Case Registry

- Config: `.fab-operator.yaml` at repo root (persistent, hidden)
- Toggling: conversational natural language → operator writes config
- Roster: tick-start display with 🟢 (enabled + summary) / ⚪ (disabled)
- Loop: heartbeat model — runs while any use case enabled
- Built-in use cases: `monitor-changes`, `linear-inbox`, `pr-freshness` (fixed set)

---

## Tab Preparation Procedure

Shared 5-step procedure before dispatching work to a tab:

1. Verify pane exists
2. Check agent is idle
3. Check change is active → `/fab-switch` if needed
4. Check branch alignment → `/git-branch` if needed
5. Dispatch command

Referenced by: all playbooks that send commands, `pr-freshness` rebase routing.

---

## Key Properties

| Property | Value |
|----------|-------|
| Requires active change? | No |
| Runs preflight? | No |
| Read-only? | No — sends commands to other agents, auto-answers questions, writes `.fab-operator.yaml` |
| Idempotent? | Yes — state is re-derived before every action |
| Advances stage? | No |
| Outputs `Next:` line? | No — ends with ready signal |
| Loads change artifacts? | No — coordination context only |
| Requires tmux? | Yes for pane-map, resolve --pane, monitoring, auto-nudge; status-only mode without |
| Uses `/loop`? | Yes — heartbeat for use case registry |
| Uses `.fab-operator.yaml`? | Yes — persistent use case config |

---

## Resolved Design Decisions

1. **Standalone operator5 over modifying operator4.** Operator4 remains as the previous stable version. Consistent with the versioning pattern — new capabilities go into new operator versions.

2. **Branch name matching over content inspection.** Scans `refs/heads/` and `refs/remotes/` ref names, not `git ls-tree`. Branch names follow `YYMMDD-{id}-{slug}` convention — content inspection is unnecessary for identification.

3. **Fixed use case set over plugin model.** Three built-in use cases, fully specified in the skill file. New use cases require a new operator version. Keeps the skill self-contained.

4. **Playbooks over modes.** "Modes of Operation" renamed to "Playbooks" — these are coordination recipes/guidebook, not state the operator switches between.

5. **Tab preparation as shared procedure.** Factored out of individual playbooks into a reusable 5-step procedure. Ensures consistent tab state (active change + branch) before dispatching.

6. **All inherited operator4 decisions apply.** All-auto-answer, re-capture before send, `/fab-fff` for autopilot, `/git-branch` after new change.
