# Intake: Improve wt list and delete commands

**Change**: 260316-mvcv-improve-wt-list-delete
**Created**: 2026-03-16
**Status**: Draft

## Origin

> Better wt list output. Add headers, indent better (check in ~/code/wvrdz/loom for example). Also, in "wt delete", by default only delete the branch with the same name as the worktree (the worktree branch) — don't delete branch if the user switched to another one — by default.

One-shot natural language input. Two independent improvements to the `wt` CLI tool (Go, Cobra).

## Why

1. **`wt list` readability**: The current output has no column headers — users must infer which column is the name, branch, status, or path. Fixed-width formatting (`%-14s %-22s`) causes misalignment when names or branches exceed the expected length, and full absolute paths add noise. As the number of worktrees grows, this becomes harder to scan.

2. **`wt delete` branch safety**: The current default deletes whatever branch the worktree is currently on (the `RunE` function defaults `deleteBranch = ""` to `"true"`, then `handleBranchCleanup` unconditionally deletes). This is dangerous when the user has switched to a different branch inside the worktree. For example: a worktree named `arctic-eagle` was created on its default branch (`arctic-eagle`), then the user ran `git checkout 260316-feat-important-work`. Running `wt delete arctic-eagle` would delete branch `260316-feat-important-work` — potentially destroying in-progress feature work. The fix: only auto-delete the branch when it matches the worktree name (i.e., the branch was auto-created by `wt create` and has no independent value).

## What Changes

### `wt list` — Headers and dynamic column widths

Current output (from `handleFormattedOutput` in `src/go/wt/cmd/list.go:117`):
```
Worktrees for: loom
Location: /home/parallels/code/wvrdz/loom.worktrees

* (main)        main                          /home/parallels/code/wvrdz/loom
  9jyv           260313-9jyv-main-process-token-refresh        /path/...
  arctic-eagle   arctic-eagle           * /path/...
  lunar-sparrow  lunar-sparrow          * /path/...

Total: 6 worktree(s)
```

Target improvements:
1. **Add column headers**: `Name`, `Branch`, `Status`, `Path` — printed as a header row below the title/location lines
2. **Dynamic column widths**: Compute max width per column from data, then pad accordingly. This eliminates misalignment when names or branches are long.
3. **Relative paths**: Show paths relative to the worktrees parent directory (e.g., `loom/` for main, `loom.worktrees/arctic-eagle/` for worktrees) instead of full absolute paths. Keeps output compact.
4. **Separator line**: A thin separator (`----` dashes) under headers for scannability.

Target output sketch:
```
Worktrees for: loom
Location: /home/parallels/code/wvrdz/loom.worktrees

  Name           Branch                                    Status  Path
  ----           ------                                    ------  ----
* (main)         main                                              loom/
  9jyv           260313-9jyv-main-process-token-refresh            loom.worktrees/9jyv/
  arctic-eagle   arctic-eagle                              *       loom.worktrees/arctic-eagle/
  lunar-sparrow  lunar-sparrow                             *       loom.worktrees/lunar-sparrow/

Total: 6 worktree(s)
```

The JSON output (`--json`) and path lookup (`--path`) modes are unchanged.

### `wt delete` — Safe branch deletion default

Current behavior in `handleBranchCleanup` (`src/go/wt/cmd/delete.go:599`):
- The `RunE` function defaults `deleteBranch = ""` to `"true"` (line 38)
- `handleBranchCleanup` then always deletes the worktree's current branch (local + remote) when `deleteBranch == "true"`
- Also cleans up orphaned `wt/{wtName}` branches

New default behavior:
- **Only delete the branch if `branch == wtName`** (the branch name matches the worktree directory name). This means the branch was auto-created by `wt create` and has no independent value.
- **Skip branch deletion if `branch != wtName`** — the user switched to a different branch, which likely has independent value. Print a note: `Skipped branch deletion: {branch} ≠ worktree name ({wtName}). Use --delete-branch true to force.`
- The `wt/{wtName}` cleanup continues unchanged regardless.
- **Explicit `--delete-branch true`** overrides this safety check and always deletes the branch (existing behavior preserved for automation/scripts).

Implementation approach: stop defaulting `""` to `"true"` in the `RunE` function. Instead, pass the raw value through to `handleBranchCleanup` and add tri-state logic:
- `""` (no flag passed) → "auto" mode: delete only if `branch == wtName`
- `"true"` (explicit flag) → force delete (current behavior)
- `"false"` (explicit flag) → never delete (current behavior)

The `handleBranchCleanup` signature already receives both `branch` and `wtName` — the comparison is straightforward.

## Affected Memory

No memory updates needed — `wt` is a fab-kit tool, not a domain component with spec-level behavior documented in memory.

## Impact

- **Source files**: `src/go/wt/cmd/list.go`, `src/go/wt/cmd/delete.go`
- **Binary**: `fab/.kit/bin/wt` (recompile after changes)
- **Downstream consumers**: `fab pane-map` uses `wt list --json` — JSON output is unchanged. Scripts using `wt delete --non-interactive` with explicit `--delete-branch true` are unaffected (explicit flag bypasses safety).
- **Breaking change**: `wt delete` will stop deleting switched-to branches by default. This is the intended behavior change and is safer than the current default.

## Open Questions

None — both changes are well-scoped with clear requirements.

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Column headers use plain text, no box-drawing characters | Consistent with existing `wt` and `fab pane-map` output style | S:80 R:95 A:90 D:95 |
| 2 | Certain | JSON output (`--json`) is unchanged | User only mentioned human-readable output | S:85 R:90 A:95 D:95 |
| 3 | Certain | Branch safety check compares `branch == wtName` | User explicitly stated "same name as the worktree" — observable in real data (arctic-eagle on arctic-eagle vs 9jyv on different branch) | S:90 R:85 A:90 D:90 |
| 4 | Certain | Dynamic column widths computed from max entry width | Consistent with table-output patterns in fab CLI tools | S:95 R:90 A:85 D:80 |
| 5 | Certain | Paths shown as relative to worktrees parent directory | Reduces noise; parent dir is already shown in the Location line | S:95 R:90 A:75 D:70 |
| 6 | Certain | Explicit `--delete-branch true` bypasses safety check | Needed for automation/scripts; preserves backwards compat for explicit callers | S:95 R:80 A:80 D:85 |
| 7 | Certain | Separator line uses `----` dashes under headers | Consistent with existing CLI output conventions | S:95 R:95 A:80 D:75 |

7 assumptions (7 certain, 0 confident, 0 tentative, 0 unresolved).
