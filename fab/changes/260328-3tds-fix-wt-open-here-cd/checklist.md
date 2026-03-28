# Quality Checklist: Fix wt open "Open Here" cd Mechanism

**Change**: 260328-3tds-fix-wt-open-here-cd
**Generated**: 2026-03-28
**Spec**: `spec.md`

## Functional Completeness
- [ ] CHK-001 shell-setup subcommand: `wt shell-setup` outputs a shell wrapper function to stdout
- [ ] CHK-002 Shell detection: Subcommand detects bash/zsh from `$SHELL` and outputs appropriate wrapper
- [ ] CHK-003 Unsupported shell: Exits 1 with stderr error for non-bash/zsh shells
- [ ] CHK-004 WT_WRAPPER env var: Wrapper sets `WT_WRAPPER=1` before invoking `command wt`
- [ ] CHK-005 Stderr hint: Hint printed when `WT_WRAPPER` unset, suppressed when set
- [ ] CHK-006 Help text updated: `wt --help` shows `eval "$(wt shell-setup)"` instead of raw function

## Behavioral Correctness
- [ ] CHK-007 cd protocol unchanged: `open_here` still prints `cd -- "/path"` to stdout (existing behavior preserved)
- [ ] CHK-008 Wrapper passes through normal output: Non-cd output from `wt list` etc. displayed unchanged
- [ ] CHK-009 Wrapper preserves exit code: Binary exit code propagated through wrapper function
- [ ] CHK-010 Create open_here still works: `wt create` with "Open here" suppresses path line and prints cd

## Scenario Coverage
- [ ] CHK-011 Bash setup: `eval "$(wt shell-setup)"` in bash defines working `wt` function
- [ ] CHK-012 Zsh setup: Same for zsh
- [ ] CHK-013 Special characters in path: cd with quoted paths containing spaces works through wrapper
- [ ] CHK-014 Open here without wrapper: cd command printed + hint on stderr
- [ ] CHK-015 Open here with wrapper: cd command eval'd, no hint

## Edge Cases & Error Handling
- [ ] CHK-016 Empty SHELL variable: Handled gracefully (error or reasonable default)
- [ ] CHK-017 Hint on create path: `wt create` open_here also shows hint when no wrapper

## Code Quality
- [ ] CHK-018 Pattern consistency: New code follows naming and structural patterns of surrounding wt commands
- [ ] CHK-019 No unnecessary duplication: Existing utilities reused where applicable

## Documentation Accuracy
- [ ] CHK-020 Help text: Updated to reference shell-setup instead of inline wrapper code

## Cross References
- [ ] CHK-021 create.go consistency: Both open.go and create.go open_here paths show the same hint behavior

## Notes

- Check items as you review: `- [x]`
- All items must pass before `/fab-continue` (hydrate)
- If an item is not applicable, mark checked and prefix with **N/A**: `- [x] CHK-008 **N/A**: {reason}`
