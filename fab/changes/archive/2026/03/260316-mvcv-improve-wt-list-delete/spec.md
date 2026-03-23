# Spec: Improve wt list and delete commands

**Change**: 260316-mvcv-improve-wt-list-delete
**Created**: 2026-03-16
**Affected memory**: N/A — no memory updates needed

## wt list: Formatted Output

### Requirement: Column Headers

The `handleFormattedOutput` function SHALL print a header row with columns `Name`, `Branch`, `Status`, `Path` after the title/location lines and before any data rows. The header row SHALL use the same column widths as data rows.

#### Scenario: Headers printed above data rows
- **GIVEN** a repository with at least one worktree
- **WHEN** `wt list` is run (no `--json` or `--path` flags)
- **THEN** the output SHALL contain a header row: `Name`, `Branch`, `Status`, `Path`
- **AND** the header row SHALL appear after the `Location:` line and blank line
- **AND** the header row SHALL be followed by a separator row of dashes

#### Scenario: No worktrees
- **GIVEN** a repository with only the main worktree
- **WHEN** `wt list` is run
- **THEN** headers SHALL still be printed (main worktree is always listed)

### Requirement: Separator Line

A separator row SHALL be printed immediately below the header row. Each column's separator SHALL consist of dashes (`-`) matching the length of that column's header text.

#### Scenario: Separator matches header widths
- **GIVEN** the header row has been printed
- **WHEN** the separator row is rendered
- **THEN** each column SHALL have dashes equal to the header label length (e.g., `Name` → `----`, `Branch` → `------`, `Status` → `------`, `Path` → `----`)

### Requirement: Dynamic Column Widths

Column widths SHALL be computed dynamically from the maximum display width of each column's data (including the header label). The fixed-width format strings (`%-14s %-22s`) SHALL be replaced with dynamically computed padding.

#### Scenario: Short names and branches
- **GIVEN** all worktree names are <= 6 chars and all branches <= 10 chars
- **WHEN** `wt list` is run
- **THEN** columns SHALL be padded to the maximum width of data in that column (minimum: header label width)

#### Scenario: Long branch name
- **GIVEN** a worktree with branch `260313-9jyv-main-process-token-refresh` (39 chars)
- **WHEN** `wt list` is run
- **THEN** the `Branch` column width SHALL accommodate the longest branch name
- **AND** all other rows SHALL be padded to the same width

### Requirement: Relative Paths

The `Path` column SHALL display paths relative to the parent directory of the worktrees directory, rather than full absolute paths. The main worktree path SHALL be shown as `{repo-name}/` and other worktrees as `{repo-name}.worktrees/{wt-name}/`.

#### Scenario: Main worktree path
- **GIVEN** the main worktree is at `/home/user/code/loom`
- **WHEN** `wt list` renders the main worktree's path
- **THEN** the path SHALL be displayed as `loom/`

#### Scenario: Non-main worktree path
- **GIVEN** a worktree at `/home/user/code/loom.worktrees/arctic-eagle`
- **WHEN** `wt list` renders that worktree's path
- **THEN** the path SHALL be displayed as `loom.worktrees/arctic-eagle/`

### Requirement: Current Worktree Marker Alignment

The current worktree marker (`*`) SHALL appear in a 2-character prefix column (position 0-1), with the `Name` column starting at position 2. The header row SHALL have 2 spaces in the prefix column to align with data rows.

#### Scenario: Current worktree marked
- **GIVEN** the user is currently in worktree `arctic-eagle`
- **WHEN** `wt list` is run
- **THEN** the `arctic-eagle` row SHALL have `* ` as prefix (asterisk + space)
- **AND** all other rows SHALL have `  ` as prefix (two spaces)
- **AND** the header row SHALL have `  ` as prefix

### Requirement: JSON and Path Modes Unchanged

The `--json` output and `--path` lookup modes SHALL remain unchanged. Only the human-readable formatted output is affected.

#### Scenario: JSON output unchanged
- **GIVEN** any repository state
- **WHEN** `wt list --json` is run
- **THEN** the JSON output SHALL be identical to the current format (including absolute paths in the `path` field)

#### Scenario: Path lookup unchanged
- **GIVEN** a worktree named `arctic-eagle`
- **WHEN** `wt list --path arctic-eagle` is run
- **THEN** the output SHALL be the absolute path (unchanged behavior)

## wt delete: Safe Branch Deletion

### Requirement: Auto-Mode Branch Safety

When `--delete-branch` is not explicitly passed (value is `""`), `handleBranchCleanup` SHALL only delete the branch if `branch == wtName`. If `branch != wtName`, it SHALL skip the primary branch deletion and print a note.

#### Scenario: Branch matches worktree name (auto-delete)
- **GIVEN** worktree `arctic-eagle` is on branch `arctic-eagle`
- **WHEN** `wt delete arctic-eagle` is run without `--delete-branch`
- **THEN** branch `arctic-eagle` SHALL be deleted (local and remote)
- **AND** the `wt/arctic-eagle` orphan branch SHALL be cleaned up as before

#### Scenario: Branch differs from worktree name (skip deletion)
- **GIVEN** worktree `9jyv` is on branch `260313-9jyv-main-process-token-refresh`
- **WHEN** `wt delete 9jyv` is run without `--delete-branch`
- **THEN** the branch `260313-9jyv-main-process-token-refresh` SHALL NOT be deleted
- **AND** a note SHALL be printed: `Skipped branch deletion: 260313-9jyv-main-process-token-refresh ≠ worktree name (9jyv). Use --delete-branch true to force.`
- **AND** the `wt/9jyv` orphan branch SHALL still be cleaned up

### Requirement: Explicit Override

When `--delete-branch true` is explicitly passed, `handleBranchCleanup` SHALL unconditionally delete the branch regardless of whether it matches the worktree name. This preserves backwards compatibility for scripts and automation.

#### Scenario: Explicit flag forces deletion of mismatched branch
- **GIVEN** worktree `9jyv` is on branch `260313-9jyv-main-process-token-refresh`
- **WHEN** `wt delete 9jyv --delete-branch true` is run
- **THEN** branch `260313-9jyv-main-process-token-refresh` SHALL be deleted (local and remote)

#### Scenario: Explicit false prevents deletion
- **GIVEN** worktree `arctic-eagle` is on branch `arctic-eagle`
- **WHEN** `wt delete arctic-eagle --delete-branch false` is run
- **THEN** the branch SHALL NOT be deleted

### Requirement: Tri-State Default Removal

The `RunE` function SHALL NOT default `deleteBranch = ""` to `"true"`. The empty string SHALL be passed through to `handleBranchCleanup` as-is, representing "auto" mode.

#### Scenario: Default value is empty string
- **GIVEN** the `deleteCmd()` function defines `deleteBranch` with default `""`
- **WHEN** the user runs `wt delete` without `--delete-branch`
- **THEN** the `RunE` function SHALL NOT modify `deleteBranch` from `""`
- **AND** `handleBranchCleanup` SHALL receive `""` as the `deleteBranch` parameter

### Requirement: Orphan Branch Cleanup Unchanged

The `wt/{wtName}` orphan branch cleanup in `handleBranchCleanup` SHALL continue to execute regardless of the `deleteBranch` value or the branch safety check result.

#### Scenario: Orphan cleanup runs even when primary branch is skipped
- **GIVEN** worktree `9jyv` is on branch `260313-9jyv-main-process-token-refresh`
- **AND** a local branch `wt/9jyv` exists
- **WHEN** `wt delete 9jyv` is run without `--delete-branch`
- **THEN** the primary branch deletion SHALL be skipped (mismatch)
- **AND** the `wt/9jyv` branch SHALL still be deleted (local and remote)

### Requirement: Multi-Delete and Delete-All Consistency

The branch safety logic SHALL apply uniformly across all delete paths: `handleDeleteCurrent`, `handleDeleteByName`, `handleDeleteMultiple`, and `handleDeleteAll`. All paths pass `deleteBranch` to `handleBranchCleanup`, which is the single point of enforcement.

#### Scenario: Multi-delete with mixed branches
- **GIVEN** worktree `arctic-eagle` on branch `arctic-eagle` and worktree `9jyv` on branch `260313-9jyv-main-process-token-refresh`
- **WHEN** `wt delete arctic-eagle 9jyv` is run without `--delete-branch`
- **THEN** `arctic-eagle`'s branch SHALL be deleted (match)
- **AND** `9jyv`'s branch SHALL be skipped with a note (mismatch)

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Column headers use plain text, no box-drawing characters | Confirmed from intake #1 — consistent with existing wt and fab pane-map output style | S:80 R:95 A:90 D:95 |
| 2 | Certain | JSON output (`--json`) is unchanged | Confirmed from intake #2 — user only mentioned human-readable output | S:85 R:90 A:95 D:95 |
| 3 | Certain | Branch safety check compares `branch == wtName` | Confirmed from intake #3 — user explicitly stated "same name as the worktree" | S:90 R:85 A:90 D:90 |
| 4 | Certain | Dynamic column widths computed from max entry width per column | Confirmed from intake #4 — standard table layout approach | S:95 R:90 A:85 D:80 |
| 5 | Certain | Paths shown as relative to worktrees parent directory | Confirmed from intake #5 — parent dir already in Location line | S:95 R:90 A:75 D:70 |
| 6 | Certain | Explicit `--delete-branch true` bypasses safety check | Confirmed from intake #6 — needed for automation/scripts | S:95 R:80 A:80 D:85 |
| 7 | Certain | Separator line uses dashes matching header label length | Confirmed from intake #7 — consistent with CLI conventions | S:95 R:95 A:80 D:75 |
| 8 | Certain | Relative path format uses `{repo-name}/` for main and `{repo-name}.worktrees/{wt-name}/` for others | Derived from intake target output sketch — consistent with worktree directory structure | S:85 R:90 A:90 D:85 |
| 9 | Certain | Safety note message format: `Skipped branch deletion: {branch} ≠ worktree name ({wtName}). Use --delete-branch true to force.` | Directly from intake specification | S:90 R:95 A:85 D:90 |
| 10 | Certain | Orphan wt/{wtName} branch cleanup is independent of the safety check | Confirmed from intake — cleanup continues unchanged | S:90 R:85 A:90 D:95 |

10 assumptions (10 certain, 0 confident, 0 tentative, 0 unresolved).
