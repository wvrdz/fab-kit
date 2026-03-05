# Intake: Orchestrator Idle Hooks

**Change**: 260305-bs5x-orchestrator-idle-hooks
**Created**: 2026-03-05
**Status**: Draft

## Origin

> Add SessionStart and Stop hooks (shell scripts in fab/.kit/) that write/clear `agent.idle_since` in the active change's `.status.yaml`. These hooks enable the existing pipeline orchestrator (and future multi-change coordination) to detect when an agent window is idle and ready for the next pipeline command — replacing fixed-delay heuristics with an explicit filesystem signal.

Discussion context: Explored the `agent-orchestrator` project (a TypeScript monorepo for coordinating 30+ AI agents via tmux). Concluded that fab-kit needs a much simpler approach — filesystem-as-bus, shell-as-glue — consistent with the constitution. The key missing piece: an explicit "agent is idle" signal. Claude Code's `SessionStart` and `Stop` hook events provide exactly this.

## Why

The existing pipeline orchestrator (`run.sh` + `dispatch.sh`) uses fixed delays to guess when an agent is ready:
- `CLAUDE_STARTUP_DELAY=3s` before sending fab-switch
- `POST_SWITCH_DELAY=5s` before sending fab-ff
- `PIPELINE_SHIP_DELAY=8s` after hydrate:done before sending /git-pr

These are timing heuristics — too short and the command is swallowed, too long and time is wasted. An explicit `idle_since` timestamp in `.status.yaml` lets the orchestrator (or any future coordination tool) know definitively when the agent stopped and is ready for input.

Without this, multi-change coordination remains dependent on fragile timing assumptions that vary by machine speed, model latency, and change complexity.

## What Changes

### New `agent` block in `.status.yaml`

Add an `agent` key to `.status.yaml` with a single field:

```yaml
agent:
  idle_since: 1741193400    # unix timestamp, set by Stop hook
```

When the agent is active (working), the `agent` block is absent or `idle_since` is cleared. When the agent stops (finishes its turn), the `Stop` hook writes the current unix timestamp.

The `agent` block is NOT part of the status template — it's ephemeral runtime state managed exclusively by hooks. It should not be initialized in `fab/.kit/templates/status.yaml`.

### SessionStart hook script

A shell script at `fab/.kit/hooks/on-session-start.sh` that:

1. Reads `fab/current` to get the active change ID (exit silently if no active change)
2. Resolves the change directory via `fab/.kit/bin/fab resolve --dir`
3. Clears `agent.idle_since` from `.status.yaml` using `yq`:
   ```bash
   yq -i 'del(.agent)' "$status_file"
   ```
4. Exits 0 always (hooks must not block the agent)

### Stop hook script

A shell script at `fab/.kit/hooks/on-stop.sh` that:

1. Reads `fab/current` to get the active change ID (exit silently if no active change)
2. Resolves the change directory via `fab/.kit/bin/fab resolve --dir`
3. Writes `agent.idle_since` to `.status.yaml` using `yq`:
   ```bash
   yq -i '.agent.idle_since = '$(date +%s)'' "$status_file"
   ```
4. Exits 0 always

### Hook registration

The hooks need to be registered in Claude Code's settings. This is a user-side configuration — the hooks reference scripts inside `fab/.kit/hooks/`. Registration format (in `.claude/settings.json` or equivalent):

```json
{
  "hooks": {
    "SessionStart": [{ "command": "bash fab/.kit/hooks/on-session-start.sh" }],
    "Stop": [{ "command": "bash fab/.kit/hooks/on-stop.sh" }]
  }
}
```

Registration is out of scope for this change — it's a user/project configuration concern. The change only delivers the scripts.

### Path resolution

Both scripts must work from any working directory (worktrees, repo root). They should use `git rev-parse --show-toplevel` to find the repo root, then resolve paths relative to it.

## Affected Memory

- `fab-workflow/kit-architecture`: (modify) Document the `fab/.kit/hooks/` directory and hook scripts
- `fab-workflow/schemas`: (modify) Document the `agent` block in `.status.yaml` schema
- `fab-workflow/pipeline-orchestrator`: (modify) Note that `idle_since` hooks exist as an explicit idle signal, complementing the existing polling approach

## Impact

- **`.status.yaml` schema**: New optional `agent` block — purely additive, no existing fields affected
- **Pipeline orchestrator**: Future changes can replace fixed delays with `idle_since` polling. This change does NOT modify `run.sh`/`dispatch.sh` — it only provides the signal.
- **New directory**: `fab/.kit/hooks/` — new convention for Claude Code hook scripts shipped with the kit
- **Dependencies**: Requires `yq` (already a fab-kit dependency) and `git` (always present)

## Open Questions

- Should the hooks also write to a shared location (e.g., `fab/.agent-status`) for multi-worktree visibility, or is per-change `.status.yaml` sufficient for now?

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Use unix timestamp (not ISO 8601) for `idle_since` | Shell `date +%s` is simpler and sufficient; yq handles integers natively; orchestrator does arithmetic on timestamps | S:90 R:95 A:90 D:95 |
| 2 | Certain | Clear entire `agent` block on SessionStart (not just null the field) | Cleaner — absence means "active", presence means "idle". No stale partial state. | S:85 R:90 A:85 D:90 |
| 3 | Certain | Hook scripts live at `fab/.kit/hooks/` | Follows constitution (kit is self-contained, portable via cp -r). New directory is analogous to `fab/.kit/scripts/`. | S:80 R:90 A:90 D:90 |
| 4 | Certain | Don't modify status.yaml template | `agent` block is ephemeral runtime state, not part of change lifecycle. Template should only contain lifecycle fields. | S:85 R:95 A:90 D:90 |
| 5 | Confident | Hook registration is out of scope | Discussed — hooks are user/project config. Kit delivers scripts; user wires them. Consistent with constitution's portability principle. | S:75 R:85 A:80 D:75 |
| 6 | Confident | Don't modify run.sh/dispatch.sh in this change | Discussed — this change provides the signal only. Consuming the signal is a separate change. Keeps scope minimal. | S:80 R:90 A:75 D:80 |
| 7 | Confident | Use `git rev-parse --show-toplevel` for path resolution | Hooks run from arbitrary CWDs (worktrees, subdirectories). git rev-parse is the canonical way to find repo root. | S:75 R:85 A:85 D:80 |
| 8 | Tentative | Per-change `.status.yaml` is sufficient (no shared file) | For single-session multi-window, each window has its own worktree with its own `.status.yaml`. The orchestrator already knows which worktree to poll. Multi-session coordination might need a shared file, but that's a future concern. | S:60 R:70 A:60 D:55 |
<!-- assumed: per-change status file sufficient — orchestrator already polls per-worktree .status.yaml, shared file deferred to multi-session use case -->

8 assumptions (4 certain, 3 confident, 1 tentative, 0 unresolved).
