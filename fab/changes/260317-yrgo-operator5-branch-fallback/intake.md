# Intake: Operator5 — Use Case Registry, Branch Fallback, and Proactive Monitoring

**Change**: 260317-yrgo-operator5-branch-fallback
**Created**: 2026-03-17
**Status**: Draft

## Origin

> Create operator5 as a successor to operator4 — adding a use case registry model, branch fallback resolution, and proactive monitoring use cases (change monitoring, Linear inbox, PR staleness).

Conversational `/fab-discuss` session that evolved through three phases:

1. **Branch fallback** — user wanted the operator to find changes that exist only in git branches. Discussion concluded that `fab` is orthogonal to git, so this belongs in the operator skill (not the CLI). Branch name scanning is trivially achievable by the agent.

2. **Use case registry** — user wanted the operator to handle multiple concurrent monitoring concerns via a single `/loop`, with each concern as a toggleable "use case" controlled by a persistent config file and conversational commands. Each tick displays a status roster of active use cases.

3. **PR staleness** — user wanted automatic detection and rebase of stale PRs, routed to agents in tabs (not executed by the operator directly).

Key rejected alternatives:
- **CLI-level branch resolution** (`fab resolve --search-branches`, `--branch` output mode, automatic fallback) — rejected because fab operates on change folders (filesystem/YAML), not git branches.

## Why

1. **Branch blindness**: When a change folder exists only in a git branch (not checked out in any worktree), the operator can't find it. Users must manually provide branch names, breaking the "coordinate, don't execute" model.

2. **Single-concern monitoring**: Operator4's monitoring is limited to enrolled changes. Real workflows have multiple concurrent concerns — change progress, incoming work from Linear, PR staleness — that all need periodic attention. Currently the user must manually check each.

3. **PR drift**: Open PRs go stale as main advances. Without proactive rebasing, PRs accumulate merge conflicts that are harder to resolve later. The operator already has a "Rebase all" mode but it's manual — this makes it automatic.

## What Changes

**Important**: Do NOT modify `fab/.kit/skills/fab-operator4.md`. This change creates a new **`fab-operator5`** skill that is operator4's full content plus the additions below. Operator4 remains untouched as the previous version.

### 1. Use Case Registry

Operator5 replaces operator4's single-purpose monitoring with a **use case registry** — a set of named, toggleable concerns that the operator checks on each `/loop` tick.

#### Config file: `.fab-operator.yaml` (repo root, hidden)

```yaml
use_cases:
  monitor-changes:
    enabled: true
  linear-inbox:
    enabled: false
    config:
      assignee: "@me"
  pr-freshness:
    enabled: true
```

#### Conversational toggling

Users enable/disable use cases via natural language:
- "Turn on Linear monitoring"
- "Stop watching for stale PRs"
- "What's active?"

The operator writes to `.fab-operator.yaml` on the user's behalf. No manual file editing.

#### Tick-start status roster

Every `/loop` tick begins by displaying active use cases:

```
── Operator tick (17:32) ──────────────
🟢 Monitor changes (3 agents: 2 active, 1 idle 4m)
🟢 PR freshness (2 PRs clean, 1 stale)
⚪ Linear inbox (disabled)
────────────────────────────────────────
```

Green dot (🟢) = enabled and running. White dot (⚪) = disabled. Each use case reports a one-line summary after its check.

#### Loop lifecycle change

Operator4: `/loop` starts when a change is enrolled, stops when monitored set is empty.
Operator5: `/loop` runs as long as **any use case is enabled**. The loop is the operator's heartbeat. It stops only when all use cases are disabled or the user says "stop."

### 2. Branch Fallback Resolution

New subsection in **Section 3: Safety Model**, after "Pre-Send Validation" and before "Bounded Retries & Escalation".
<!-- clarified: placement — Section 3 Safety Model, after Pre-Send Validation -->

**Trigger**: User-initiated resolution only. Does not apply during monitoring ticks.
<!-- clarified: trigger scope — user-initiated only, not monitoring ticks -->

**Resolution flow** (after `fab resolve` returns non-zero):

1. Scan local and remote branch names:
   ```bash
   git for-each-ref --format='%(refname:short)' refs/heads/ refs/remotes/ | grep -i "<query>"
   ```
   <!-- clarified: searches both local and remote branches -->
2. If exactly one match: choose response based on user intent:
   - **Read-only query** (status check, "what stage is X at?"): read `.status.yaml` directly from the branch via `git show <branch>:fab/changes/<folder>/.status.yaml` — no worktree needed
   - **Action query** (send command, resume work): offer to create a worktree
   ```
   Can't find {query} in any worktree. Found branch `{branch}`. Create a worktree for it?
   ```
   <!-- clarified: read-only queries use git show, worktree creation only for action queries -->
3. If multiple matches: present disambiguation list
4. If no match: report "not found locally or in any branch"

**After user confirms worktree creation**, the operator follows the existing "Known change" spawning rule from `_naming.md`:
```
wt create --non-interactive --worktree-name <name> <branch-name>
```

### 3. Use Case: `monitor-changes`

Operator4's existing monitoring system (sections 4-5: monitored set, monitoring tick, auto-nudge, stuck detection) reframed as a use case. Behavior is identical — enrollment triggers, removal triggers, 6-step tick, auto-nudge answer model — just housed under the use case registry.

### 4. Use Case: `linear-inbox`

Watches Linear for new issues assigned to the user and offers to spawn agents.

**Detection**:
1. Query Linear via MCP (`mcp__claude_ai_Linear__list_issues`) for issues assigned to the configured user, filtered by status (e.g., `Backlog`, `Todo`)
2. Compare against known changes — match by Linear issue ID stored in `.status.yaml` (`issues` array) across all active changes and archived changes
3. New = issue exists in Linear but has no matching change

**Action on new issue**:
- Report: `"New issue: DEV-123 — {title}. Spawn an agent?"`
- On confirmation: spawn worktree + agent with `/fab-new DEV-123` (the existing Linear ticket flow in `/fab-new`)
- Deduplication: once spawned, the change's `.status.yaml` records the issue ID, preventing re-detection

**Config**:
```yaml
linear-inbox:
  enabled: false
  config:
    assignee: "@me"    # Linear user filter
```

### 5. Use Case: `pr-freshness`

Monitors open PRs for staleness and routes rebase commands to agents in tabs.

**Detection**:
```bash
gh pr list --author @me --state open --json number,headRefName,mergeStateStatus
```

- `mergeStateStatus == "BEHIND"` → PR is stale (main has advanced), needs rebase
- `mergeStateStatus == "DIRTY"` → PR has conflicts, needs rebase + conflict resolution
- `mergeStateStatus == "CLEAN"` → up to date, no action

**Action on stale PR**:
1. Match PR's `headRefName` to an agent tab via pane-map (branch name == change folder name)
2. If agent tab found and agent is idle: send `"rebase on main, resolve any conflicts, and push"` to that pane
3. If agent tab found but agent is busy: skip, report: `"PR #{n}: stale but agent is busy"`
4. If no agent tab: report: `"PR #{n} ({branch}): stale but no agent running. Spawn one?"` — ties into branch fallback (the branch exists, no worktree)
5. If PR is `DIRTY`: still route to agent, but flag: `"PR #{n}: has conflicts — agent may need help"`

**The operator does NOT run `git rebase` itself.** It sends instructions to the agent in the tab. Consistent with "coordinate, don't execute."

### 6. Tab Preparation Procedure

Before dispatching work to an agent in a tab, the operator MUST ensure the tab is ready. This is a shared procedure used by multiple playbooks (spawn, unstick, broadcast, autopilot, pr-freshness rebase routing).

**Steps**:
1. **Verify pane exists** — refresh pane map (existing pre-send validation)
2. **Check agent is idle** — if busy, warn and require confirmation (existing pre-send validation)
3. **Check change is active** — read the tab's active change via pane-map. If the target change is not the active change in that tab, send `/fab-switch <change>` first
4. **Check branch alignment** — if the tab's git branch doesn't match the change folder name, send `/git-branch` to align it
5. **Dispatch the command** — e.g., `/fab-fff`, `/fab-continue`, rebase instruction

This ensures tabs are in a consistent state before receiving work. Without this, commands sent to a tab with the wrong active change or wrong branch will fail or act on the wrong change.

### 7. Playbooks (carried over from operator4)

Operator4's "Modes of Operation" (Section 6) are renamed to **playbooks** — documented coordination patterns for common user requests. They are on-demand (user-triggered), not loop-driven. The operator recognizes the user's intent and follows the relevant playbook.

Carried over from operator4:
- **Broadcast** — send a command to all idle agents
- **Sequenced rebase** — "when X finishes, rebase Y on main"
- **Merge PRs** — merge completed PRs at ship/review-pr stage
- **Spawn agent** — new worktree + agent from backlog idea
- **Status dashboard** — concise summary of all agents
- **Unstick agent** — nudge a stuck agent
- **Notification** — "tell me when X finishes"
- **Rebase all** — make all PRs mergeable
- **Autopilot** — drive a queue of changes through the full pipeline

All playbooks that dispatch work to a tab use the **Tab Preparation Procedure** (section 6 above) before sending commands.

### 8. Cleanup: delete legacy operator scripts

Delete the following obsolete launcher scripts:

- `fab/.kit/scripts/fab-operator1.sh`
- `fab/.kit/scripts/fab-operator2.sh`
- `fab/.kit/scripts/fab-operator3.sh`

These are leftovers from the operator inheritance chain (operator1→2→3→4) that was replaced by the standalone operator4 rewrite. `fab/.kit/scripts/fab-operator4.sh` remains as the active launcher.

## Affected Memory

- `fab-workflow/execution-skills`: (modify) Document operator5 use case registry, branch fallback, and three built-in use cases

## Impact

- **New skill file**: `fab/.kit/skills/fab-operator5.md` — operator4's full content + use case registry + branch fallback + 3 use cases
- **New spec file**: `docs/specs/skills/SPEC-fab-operator5.md` — corresponding spec
- **New config file**: `.fab-operator.yaml` — use case registry config (repo root, hidden)
- **Operator4 untouched**: `fab/.kit/skills/fab-operator4.md` and `docs/specs/skills/SPEC-fab-operator4.md` remain as-is
- **Deleted scripts**: `fab/.kit/scripts/fab-operator{1,2,3}.sh` — legacy launchers from the inheritance chain
- **No CLI changes**: `fab resolve`, `fab change`, and all other `fab` subcommands remain unchanged
- **No template changes**: No new templates or status fields
- **No migration needed**: Pure skill addition + cleanup

## Open Questions

- None remaining — all clarified in discussion.

## Clarifications

### Session 2026-03-17

| # | Action | Detail |
|---|--------|--------|
| 7 | Clarified | Trigger scope: user-initiated resolution only, not monitoring ticks |
| 8 | Clarified | Search both local and remote branches |
| 9 | Clarified | Read-only queries use `git show`; worktree creation only for action queries |
| 6 | Clarified | Placement: Section 3 (Safety Model), after Pre-Send Validation |

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Create new operator5, do not modify operator4. No CLI changes | Discussed — user explicitly rejected CLI-level approaches; later clarified to create operator5 | S:95 R:90 A:95 D:95 |
| 2 | Certain | Use `git for-each-ref` + grep for branch name scanning | Discussed — user confirmed branch name matching, not branch content inspection | S:90 R:95 A:90 D:90 |
| 3 | Certain | Same substring/ID matching as existing resolution | Branch names follow `YYMMDD-{id}-{slug}` convention per `_naming.md` | S:85 R:90 A:95 D:95 |
| 4 | Confident | Offer worktree creation when a branch match is found | Natural next step per existing operator spawning rules in `_naming.md` | S:75 R:80 A:85 D:80 |
| 5 | Confident | Single-match auto-report, multi-match disambiguation | Standard resolution pattern consistent with `fab resolve` behavior | S:70 R:85 A:85 D:80 |
| 6 | Certain | Branch fallback in Section 3 (Safety Model), after Pre-Send Validation | Clarified — user chose option 1 | S:95 R:90 A:60 D:50 |
| 7 | Certain | Branch fallback triggers on user-initiated resolution only, not monitoring ticks | Clarified — user agreed | S:95 R:90 A:90 D:95 |
| 8 | Certain | Search both local (`refs/heads/`) and remote (`refs/remotes/`) branches | Clarified — user requested remote branch search | S:95 R:85 A:90 D:95 |
| 9 | Certain | Read-only queries use `git show` for status; worktree creation only for action queries | Clarified — user agreed | S:95 R:90 A:90 D:90 |
| 10 | Certain | Use case registry model — toggleable named concerns with persistent config file | Discussed — user designed this model: config in hidden file, conversational toggling, tick-start roster | S:95 R:85 A:90 D:95 |
| 11 | Certain | Config file is `.fab-operator.yaml` at repo root | Discussed — user specified hidden file, persistent across sessions | S:90 R:90 A:85 D:85 |
| 12 | Certain | Tick-start roster with green/white dots showing active/disabled use cases | Discussed — user specified the display format | S:95 R:95 A:95 D:95 |
| 13 | Certain | Loop runs as long as any use case is enabled, not tied to monitored set | Follows from use case registry model — loop is the operator's heartbeat | S:85 R:85 A:90 D:90 |
| 14 | Certain | `monitor-changes` use case = operator4's existing monitoring reframed | Discussed — user's first example use case | S:90 R:90 A:95 D:95 |
| 15 | Certain | `linear-inbox` use case watches for assigned issues, offers to spawn agents | Discussed — user's second example use case | S:90 R:80 A:80 D:85 |
| 16 | Certain | `pr-freshness` use case detects stale PRs via `gh pr list` mergeStateStatus | Discussed — user requested this; `gh` API provides staleness status directly | S:90 R:85 A:90 D:90 |
| 17 | Certain | PR rebase routed to agents in tabs, not executed by operator directly | Discussed — user explicitly stated "handled by agents running in the other tabs via a message sent from operator5" | S:95 R:85 A:95 D:95 |
| 18 | Confident | Three built-in use cases shipped with operator5, fixed set (not user-extensible) | Not explicitly discussed — but consistent with operator versioning pattern (new use cases → operator6) | S:60 R:85 A:80 D:70 |
| 19 | Confident | All use cases run every tick (same cadence), no per-use-case intervals | Not explicitly discussed — simpler model, `gh pr list` is lightweight enough for every tick | S:55 R:90 A:80 D:75 |
| 20 | Certain | Tab preparation procedure: check active change → fab-switch → git-branch → dispatch | Discussed — user specified this sequence for ensuring tabs are ready before work | S:90 R:85 A:90 D:90 |
| 21 | Certain | Operator4's "Modes of Operation" renamed to "Playbooks" — on-demand coordination patterns | Discussed — user noted these are a guidebook/cheat sheet, not modes. "Playbooks" agreed | S:95 R:95 A:90 D:90 |
| 22 | Certain | Operator5 has 3 loop-driven use cases + 9 playbooks carried from operator4 | Discussed — user asked for explicit accounting of total capabilities | S:90 R:90 A:95 D:95 |

22 assumptions (18 certain, 4 confident, 0 tentative).
