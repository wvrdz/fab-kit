# Tasks: Improve wt list and delete commands

**Change**: 260316-mvcv-improve-wt-list-delete
**Spec**: `spec.md`
**Intake**: `intake.md`

## Phase 1: Core Implementation

- [x] T001 [P] Refactor `handleFormattedOutput` in `src/go/wt/cmd/list.go` to compute dynamic column widths: iterate entries to find max width per column (Name, Branch, Status, Path), accounting for ANSI color codes in display width. Replace fixed `%-14s %-22s` format with dynamically computed padding.

- [x] T002 [P] Add column headers and separator row to `handleFormattedOutput` in `src/go/wt/cmd/list.go`: print `Name`, `Branch`, `Status`, `Path` header row with same padding as data rows, followed by a dash separator row where each column's dashes match its header label length.

- [x] T003 [P] Add relative path computation to `handleFormattedOutput` in `src/go/wt/cmd/list.go`: compute the parent directory of `ctx.WorktreesDir`, then display main worktree as `{repoName}/` and other worktrees as `{repoName}.worktrees/{wtName}/` by making paths relative to that parent.

- [x] T004 [P] Modify `handleBranchCleanup` in `src/go/wt/cmd/delete.go` to implement tri-state logic: when `deleteBranch == ""` (auto mode), only delete if `branch == wtName`; when `"true"`, force delete; when `"false"`, skip. Print note when auto-mode skips: `Skipped branch deletion: {branch} ≠ worktree name ({wtName}). Use --delete-branch true to force.`

- [x] T005 [P] Remove the `deleteBranch = "true"` default assignment in `deleteCmd().RunE` in `src/go/wt/cmd/delete.go` (line 38). The empty string `""` must flow through to `handleBranchCleanup` as the auto-mode sentinel.

## Phase 2: Integration & Edge Cases

- [x] T006 Verify orphan `wt/{wtName}` branch cleanup still works in auto-mode: the cleanup block in `handleBranchCleanup` must execute regardless of the primary branch deletion decision. Ensure the orphan cleanup code path is outside the new conditional.

- [x] T007 Test `wt list` output with edge cases: single worktree (main only), worktree with detached HEAD, worktree with very long branch name, dirty worktree with unpushed commits. Verify column alignment is correct in all cases.

- [x] T008 Build and verify: `cd src/go/wt && go build -o ../../../fab/.kit/bin/wt .` — confirm both `wt list` and `wt delete` work correctly with the changes.

---

## Execution Order

- T001, T002, T003 are interdependent (all modify `handleFormattedOutput`) — execute T001 first (column width computation), then T002 (headers/separator), then T003 (relative paths)
- T004 and T005 are tightly coupled — execute T005 first (remove default), then T004 (add tri-state logic)
- T006 depends on T004/T005
- T007 depends on T001-T003
- T008 depends on all previous tasks
