# Pipeline Orchestrator

**Domain**: fab-workflow

## Overview

The pipeline orchestrator automates multi-change execution in dependency order. It reads a YAML manifest (`fab/pipelines/*.yaml`), resolves the dependency chain, and dispatches each change into an isolated worktree where `fab-ff` runs the full pipeline (tasks → apply → review → hydrate). The manifest is a live contract — the human adds entries while the orchestrator processes earlier ones.

V1 is serial (one change at a time). The orchestrator runs indefinitely until killed with Ctrl+C.

## Requirements

### Manifest Format

Pipeline manifests are YAML files in `fab/pipelines/` with:
- `base` — branch that root nodes branch from (typically `main`)
- `changes[]` — list of entries, each with `id`, `depends_on`, and optional `stage`

The `stage` field is written by the orchestrator. Valid values: `intake`, `spec`, `tasks`, `apply`, `review`, `hydrate` (intermediate), `done`, `failed`, `invalid` (terminal).

`depends_on` supports at most one entry in v1. Multi-parent dependencies require manual merge of parent branches first.

### Orchestrator Core (run.sh)

`run.sh` accepts a manifest path and runs an infinite dispatch loop:

1. Re-reads the manifest from disk (live editing support)
2. Validates: schema, circular deps, single-dep constraint, reference integrity
3. Identifies dispatchable changes (deps all `done`, self not terminal)
4. Dispatches first dispatchable in list order (serial, deterministic)
5. After dispatch returns pane ID, enters a **unified polling loop** per change (5-second interval)
6. If nothing dispatchable: sleeps 30 seconds (configurable via `PIPELINE_POLL_INTERVAL`), re-reads. Uses `\r` in-place line update (no scrolling).

**Unified polling loop** (`poll_change()`): Monitors the interactive Claude session via `.status.yaml` polling and pane-alive checks. State machine:
- `polling_fab_ff` → detects `hydrate:done` (triggers ship) or `*:failed` (marks failed)
- `shipping` → pushes `/changes:ship pr` via `tmux send-keys`, polls `gh pr view` for PR creation
- Terminal states: `done` (PR detected) or `failed` (timeout, pane death, pipeline failure)

**Progress rendering**: Each poll iteration calls `stageman.sh progress-line` and renders in-place: `[pipeline] <id>: <progress> (<elapsed>)`.

**Configurable timeouts**: `PIPELINE_FF_TIMEOUT` (default 1800s/30min), `PIPELINE_SHIP_TIMEOUT` (default 300s/5min).

**Stage classification for resumability**:
- Terminal (`done`, `failed`, `invalid`) — skip permanently
- Intermediate (pipeline stage names) — re-dispatch into fresh worktree
- Absent — dispatch normally

On SIGINT: kills all tracked interactive panes, prints structured summary (Completed/Failed/Blocked/Skipped/Pending with worktree paths), exits 130.

Output: `[pipeline]`-prefixed status lines, in-place progress updates per change.

### Change Dispatch (dispatch.sh)

`dispatch.sh` handles a single change and returns immediately (no polling):

1. **Worktree creation** — branch name is `{branch_prefix}{change-id}`. Root nodes: `wt-create --non-interactive --worktree-open skip <change-branch>`. Dependent nodes: `git branch <change-branch> origin/<parent-branch>` first, then `wt-create`.
2. **Artifact provisioning** — copies `fab/changes/<id>/` from source repo to worktree if not present
3. **Prerequisite validation** — intake.md, spec.md, confidence gate. Writes `invalid` on failure.
4. **fab-switch** — `claude -p --dangerously-skip-permissions` (print mode, no context needed)
5. **Interactive pane creation** — launches `claude --dangerously-skip-permissions '/fab-ff'` in a tmux split pane. First dispatch: horizontal split (`-h`); subsequent: vertical split stacked below previous (`-v -t $LAST_PANE_ID`)
6. **Output** — two stdout lines: worktree path + pane ID. `run.sh` captures both.

`dispatch.sh` does NOT poll, wait, or ship. All completion detection and shipping is handled by `run.sh`'s polling loop.

Infrastructure failures (wt-create, claude, git) abort the orchestrator entirely.

Worktrees and interactive panes are left in place after dispatch for manual inspection.

### Example Scaffold

`fab/pipelines/example.yaml` — fully commented-out annotated example covering manifest format, dependency syntax, stage values, prerequisites, and the live-editing contract.

## Scripts

| Script | Location | Purpose |
|--------|----------|---------|
| `batch-fab-pipeline.sh` | `fab/.kit/scripts/batch-fab-pipeline.sh` | User-facing entry point (listing, matching, delegation) |
| `run.sh` | `fab/.kit/scripts/pipeline/run.sh` | Main orchestrator loop |
| `dispatch.sh` | `fab/.kit/scripts/pipeline/dispatch.sh` | Per-change dispatch |

### batch-fab-pipeline.sh

User-facing entry point on PATH. Owns all UX: no-args/`--list` lists available pipelines from `fab/pipelines/*.yaml` (excluding `example.yaml`), `-h`/`--help` prints usage, positional arguments use case-insensitive substring matching against manifest basenames. Arguments with `/` or ending `.yaml` bypass matching. Delegates to `pipeline/run.sh` via `exec` with arg passthrough.

### Change ID Resolution

`run.sh` resolves each manifest entry's `id` field through `changeman resolve` before dispatching. This allows manifests to use short IDs (e.g., `a7k2`) or partial names. The manifest's internal consistency is preserved — `id` and `depends_on` values match each other as written. Resolution maps to the actual `fab/changes/` folder only at dispatch time. Resolution failure marks the change as `invalid` in the manifest.

### Worktree Naming

`dispatch.sh` uses `--worktree-name "$CHANGE_ID"` in its `wt-create` invocation, producing readable worktree directory names matching the change ID. This follows the pattern established by `batch-fab-switch-change.sh`.

### Progress Rendering

`stageman.sh progress-line` produces a single-line visual pipeline progress string for left-pane rendering. Done stages joined by ` → `, active stage + ` ⏳`, failed + ` ✗`, pending omitted. All-done appends ` ✓`. Examples: `spec → tasks → apply ⏳`, `intake → spec → tasks → apply → review ✗`.

### Stage Detection

`run.sh`'s polling loop uses `stageman progress-map` to scan for terminal states. `hydrate:done` triggers shipping; any `*:failed` marks the change failed. This replaced the previous `display-stage` approach (which does not surface failed states).

### Shipping

After `hydrate:done` is detected, `run.sh` pushes `/changes:ship pr` to the interactive Claude session via `tmux send-keys`. The session retains full fab-ff conversation context, producing contextual commit messages and PR descriptions. Ship completion is detected by polling `gh pr view` from the worktree.

### Pane Lifecycle

Each dispatched change gets its own tmux split pane (stacked vertically in the right panel). Sessions remain open for inspection — they are NOT auto-killed on completion. SIGINT kills all tracked panes.

## Design Decisions

### Serial Execution in V1
**Decision**: Process one change at a time in topological order.
**Why**: Avoids concurrent worktree management, manifest race conditions, and interleaved output. Parallel is a documented stretch goal.
**Rejected**: Background processes with PID tracking.

### Infinite Loop with SIGINT Exit
**Decision**: run.sh runs indefinitely, polling for new manifest entries.
**Why**: Supports the live-contract model — the human stays ahead, adding entries at their pace. The orchestrator is a daemon-like process, not a batch job.
**Rejected**: Exit when all changes done (prevents the human from adding more entries).

### Hybrid Model: claude -p for Switch, Interactive for fab-ff + Ship
**Decision**: fab-switch uses `claude -p` (print mode); fab-ff runs in an interactive Claude session; shipping is pushed to the same session via `tmux send-keys`.
**Why**: fab-switch is trivial (no context benefit from interactivity). fab-ff produces rich conversation context that ship needs for contextual commit messages and PR descriptions. The interactive session preserves this context across the full pipeline.
**Rejected**: All-`claude -p` (3 separate sessions, no shared context — produced generic shipping output), all-interactive (wasteful for fab-switch).

### Polling in run.sh, Not dispatch.sh
**Decision**: `dispatch.sh` spawns the interactive pane and exits immediately. `run.sh` owns all polling, progress rendering, and completion detection.
**Why**: Keeps dispatch.sh simple and stateless. run.sh already owns the main loop and SIGINT handling. Centralizing all waiting avoids duplicated SIGINT handling.
**Rejected**: dispatch.sh waits for completion (blocks serial dispatch, duplicates SIGINT handling).

### Stacked Vertical Splits
**Decision**: Each dispatched change gets its own vertical split within the right panel. Sessions never auto-killed.
**Why**: Users need to inspect failed or completed sessions. Killing panes loses diagnostic context. Vertical stacking provides natural visual grouping.
**Rejected**: Single reusable right pane (can't inspect previous sessions), horizontal splits (poor screen width use).

### Single-Dependency Restriction
**Decision**: `depends_on` limited to one entry in v1.
**Why**: Multi-parent dependencies create a branch topology problem — D depending on B and C can't branch from both. Resolution requires manual merge.
**Rejected**: Branching from last-completed parent (loses other parent's code).

### Infrastructure Failures Abort
**Decision**: wt-create/claude/git failures abort the orchestrator entirely.
**Why**: Infrastructure failures indicate a broken environment. Continuing would likely fail on subsequent changes.
**Rejected**: Mark individual change failed and continue (masks environment issues).

## Changelog

| Change | Date | Summary |
|--------|------|---------|
| 260221-ay66-interactive-pipeline-pane | 2026-02-21 | Replaced passive `tail -f` log pane with interactive Claude sessions per dispatch. fab-ff now runs in interactive mode (not `claude -p`), shipping via `tmux send-keys` to the same session. Added unified polling loop in `run.sh` with progress-line rendering and state machine (polling_fab_ff → shipping → done/failed). Added `stageman.sh progress-line` command. Stacked vertical pane layout. Removed `ship()` from dispatch.sh. |
| 260221-i0z6-move-env-packages-add-fab-pipeline | 2026-02-21 | Added `batch-fab-pipeline.sh` user-facing entry point with listing, partial name matching, help, and `exec` delegation. Added changeman resolve for manifest change IDs in `run.sh`. Added `--worktree-name` to `wt-create` call in `dispatch.sh`. Replaced raw yq hydrate check with stageman `display-stage` in `dispatch.sh`. |
| 260221-wy0e-pipeline-orchestrator | 2026-02-21 | Initial implementation — serial orchestrator with run.sh + dispatch.sh, YAML manifest format, wt-create integration, Claude CLI execution, stacked PRs, SIGINT handling, example scaffold |
