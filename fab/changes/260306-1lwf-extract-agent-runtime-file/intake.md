# Intake: Extract Agent Runtime to Gitignored File

**Change**: 260306-1lwf-extract-agent-runtime-file
**Created**: 2026-03-06
**Status**: Draft

## Origin

> Move agent ephemeral runtime data (agent.idle_since) from per-change .status.yaml to a repo-root gitignored singleton file (.fab-runtime.yaml). Hooks (on-stop.sh, on-session-start.sh) write/clear agent block there instead. File is keyed by change folder name. Update .gitignore, both hook scripts, and schemas.md memory. Pipeline orchestrator docs should note the new location. No Go code changes needed — hooks use yq directly. This is a refactor to reduce noisy writes to tracked .status.yaml on every agent turn.

Interaction mode: conversational. Preceded by a `/fab-discuss` session with extensive blast radius analysis. Three options were evaluated (per-change sidecar, repo-root singleton, Go Save split). User chose Option B (repo-root singleton) after reviewing tradeoffs.

Key decisions from conversation:
- Option B chosen over A (per-change `.runtime.yaml` sidecar) and C (split Go `Save()` path)
- Minimal scope: only `agent` block moves — `stage_metrics` and `last_updated` stay in `.status.yaml`
- No Go binary changes needed since hooks use `yq` directly, not the Go status API
- Full blast radius analysis confirmed zero Go consumers of the `agent` block

## Why

The `on-stop.sh` hook fires on **every agent response turn** and writes `agent.idle_since` to the active change's `.status.yaml` via `yq -i`. Similarly, `on-session-start.sh` clears the `agent` block on every new session. This causes:

1. **Constant git noise** — `.status.yaml` shows up in `git diff` / `git status` after every agent interaction, even one-liner responses
2. **Merge friction** — parallel sessions or worktrees touching the same change see unnecessary conflicts in a tracked file
3. **Semantic mismatch** — `agent.idle_since` is ephemeral runtime state (seconds-granularity timestamps), not workflow state. It doesn't belong in a committed artifact

The original `260305-bs5x-orchestrator-idle-hooks` change explicitly deferred the separate-file approach as a "future concern." The write frequency in practice has proven this concern is current, not future.

## What Changes

### 1. New file: `.fab-runtime.yaml` at repo root

A gitignored YAML file at the repository root, keyed by change folder name. Created on first write (by whichever hook fires first). Hooks create it if absent.

```yaml
# .fab-runtime.yaml — ephemeral agent runtime state (gitignored)
# Keyed by change folder name. Managed by fab/.kit/hooks/.

260306-1lwf-extract-agent-runtime-file:
  agent:
    idle_since: 1741193400
```

Structure:
- Top-level keys are change folder names (full `YYMMDD-XXXX-slug` format)
- Each change entry contains an `agent` block with `idle_since` (unix timestamp)
- File is **append-friendly** — hooks only touch their change's key, never the full file
- No schema enforcement — this is a loose runtime sidecar, not a workflow artifact

### 2. Update `on-stop.sh`

Change target from `.status.yaml` to `.fab-runtime.yaml`:

```bash
# Before (in .status.yaml):
yq -i ".agent.idle_since = $(date +%s)" "$status_file"

# After (in .fab-runtime.yaml):
runtime_file="$repo_root/.fab-runtime.yaml"
[ -f "$runtime_file" ] || echo "{}" > "$runtime_file"
yq -i ".\"$change_name\".agent.idle_since = $(date +%s)" "$runtime_file"
```

The hook still resolves the active change via `fab/current` + `fab resolve` — it just writes to a different file. The `change_name` used as the key is the full folder name (from `fab resolve --folder`).

### 3. Update `on-session-start.sh`

Change target from `.status.yaml` to `.fab-runtime.yaml`:

```bash
# Before (in .status.yaml):
yq -i 'del(.agent)' "$status_file"

# After (in .fab-runtime.yaml):
runtime_file="$repo_root/.fab-runtime.yaml"
[ -f "$runtime_file" ] || exit 0
yq -i "del(.\"$change_name\".agent)" "$runtime_file"
```

Note: if `.fab-runtime.yaml` doesn't exist, the session-start hook exits cleanly (nothing to clear).

### 4. Update `.gitignore`

Add `.fab-runtime.yaml` to the gitignore:

```
# Agent runtime state (ephemeral, per-session)
.fab-runtime.yaml
```

### 5. Update `docs/memory/fab-workflow/schemas.md`

Update the "Ephemeral Runtime State" section (lines 65-79) to document the new file location. The `agent` block description stays the same — only the file path and keying structure change.

### 6. Update `docs/memory/fab-workflow/pipeline-orchestrator.md`

Add a note that the agent idle state is now in `.fab-runtime.yaml` (repo root, keyed by change folder name) instead of `.status.yaml`. The pipeline orchestrator (`run.sh`) doesn't currently read the agent block — it polls `.status.yaml` for progress state — but the memory should reflect the new location for future consumers.

## Affected Memory

- `fab-workflow/schemas`: (modify) Update Ephemeral Runtime State section — new file location and keyed structure
- `fab-workflow/pipeline-orchestrator`: (modify) Note agent idle state location change

## Impact

- **Hook scripts**: `fab/.kit/hooks/on-stop.sh` and `fab/.kit/hooks/on-session-start.sh` — primary change targets
- **`.gitignore`**: New entry for `.fab-runtime.yaml`
- **Memory docs**: Two files updated for accuracy
- **No Go code changes**: The Go binary (`statusfile.go`, `status.go`) never reads or writes the `agent` block — hooks use `yq` directly
- **No skill changes**: No skill references or consumes `agent.idle_since`
- **No template changes**: `fab/.kit/templates/status.yaml` never included the `agent` block
- **Pipeline orchestrator**: `run.sh` polls `.status.yaml` progress, not the agent block — no code change needed, only memory doc update
- **Worktree behavior**: Each worktree gets its own `.fab-runtime.yaml` at its repo root — no cross-worktree contention
- **Distribution**: `.fab-runtime.yaml` is gitignored and ephemeral — no impact on `fab/.kit/` distribution

## Open Questions

None — all design decisions were resolved in the preceding discussion.

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Only `agent` block moves — `stage_metrics`, `last_updated`, `confidence` stay in `.status.yaml` | Discussed — user explicitly chose minimal scope. Stage metrics are low-frequency (stage transitions only) and consumed by git-pr | S:95 R:90 A:95 D:95 |
| 2 | Certain | Repo-root singleton (Option B) over per-change sidecar (Option A) | Discussed — user explicitly chose Option B | S:95 R:85 A:90 D:95 |
| 3 | Certain | No Go binary changes needed | Confirmed by blast radius analysis — Go code never reads/writes `agent` block, hooks use yq | S:95 R:95 A:95 D:95 |
| 4 | Certain | File keyed by full change folder name (`YYMMDD-XXXX-slug`) | Discussed — consistent with how hooks already resolve the change name via `fab resolve --folder` | S:90 R:85 A:90 D:90 |
| 5 | Confident | Hooks create `.fab-runtime.yaml` on first write if absent | Standard pattern — `on-stop.sh` seeds with `{}` before writing. Consistent with hooks' must-never-fail contract | S:75 R:90 A:85 D:85 |
| 6 | Confident | No cleanup/GC for stale entries in `.fab-runtime.yaml` | File is gitignored and ephemeral. Stale entries for archived changes are harmless. User can delete the file at any time | S:70 R:95 A:80 D:80 |
| 7 | Certain | `.gitignore` entry is at repo root level (`.fab-runtime.yaml`) | Follows existing pattern — `fab/current` and other runtime files are gitignored at root | S:90 R:95 A:90 D:95 |

7 assumptions (5 certain, 2 confident, 0 tentative, 0 unresolved).
