# Quality Checklist: Improve wt list and delete commands

**Change**: 260316-mvcv-improve-wt-list-delete
**Generated**: 2026-03-16
**Spec**: `spec.md`

## Functional Completeness
- [x] CHK-001 Column Headers: `wt list` prints Name, Branch, Status, Path header row
- [x] CHK-002 Separator Line: Dash separator row printed below headers with correct dash counts
- [x] CHK-003 Dynamic Column Widths: Column widths computed from max data width, no fixed-width format strings
- [x] CHK-004 Relative Paths: Main worktree shows `{repo}/`, others show `{repo}.worktrees/{name}/`
- [x] CHK-005 Auto-Mode Branch Safety: `handleBranchCleanup` only deletes when `branch == wtName` in auto mode
- [x] CHK-006 Explicit Override: `--delete-branch true` forces deletion regardless of name match
- [x] CHK-007 Tri-State Default: `deleteBranch` stays `""` when flag not passed (no default to `"true"`)

## Behavioral Correctness
- [x] CHK-008 JSON Output Unchanged: `wt list --json` produces identical output to before
- [x] CHK-009 Path Lookup Unchanged: `wt list --path <name>` returns absolute path as before
- [x] CHK-010 Orphan Cleanup Independent: `wt/{wtName}` branch cleanup runs regardless of safety check result
- [x] CHK-011 Multi-Delete Consistency: Safety logic applies per-worktree in `handleDeleteMultiple` and `handleDeleteAll`

## Scenario Coverage
- [x] CHK-012 Branch matches worktree name: branch auto-deleted in auto mode
- [x] CHK-013 Branch differs from worktree name: branch skipped with note in auto mode
- [x] CHK-014 Explicit `--delete-branch true` with mismatched branch: branch deleted
- [x] CHK-015 Explicit `--delete-branch false` with matching branch: branch not deleted
- [x] CHK-016 Current worktree marker alignment: `*` prefix aligns with header spacing

## Edge Cases & Error Handling
- [x] CHK-017 Single worktree (main only): headers still print, no crash
- [x] CHK-018 Detached HEAD: list displays correctly, delete handles gracefully
- [x] CHK-019 Long branch names: columns expand, no truncation or overflow

## Code Quality
- [x] CHK-020 Pattern consistency: New code follows naming and structural patterns of surrounding code
- [x] CHK-021 No unnecessary duplication: Existing utilities reused where applicable

## Notes

- Check items as you review: `- [x]`
- All items must pass before `/fab-continue` (hydrate)
- If an item is not applicable, mark checked and prefix with **N/A**: `- [x] CHK-008 **N/A**: {reason}`
