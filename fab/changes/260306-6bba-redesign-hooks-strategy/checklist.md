# Quality Checklist: Redesign Hooks Strategy

**Change**: 260306-6bba-redesign-hooks-strategy
**Generated**: 2026-03-06
**Spec**: `spec.md`

## Functional Completeness

- [ ] CHK-001 on-artifact-write.sh: Hook script exists and handles intake.md, spec.md, tasks.md, checklist.md artifact writes
- [ ] CHK-002 Change type inference: Hook infers correct change type from intake content using keyword matching
- [ ] CHK-003 Indicative confidence: Hook calls `fab score --stage intake` when intake.md is written
- [ ] CHK-004 Spec confidence: Hook calls `fab score` when spec.md is written or edited
- [ ] CHK-005 Tasks counting: Hook counts task items and calls `set-checklist total` when tasks.md is written
- [ ] CHK-006 Checklist metadata: Hook sets generated/total/completed when checklist.md is written or edited
- [ ] CHK-007 fab runtime set-idle: Command writes idle_since timestamp to .fab-runtime.yaml
- [ ] CHK-008 fab runtime clear-idle: Command removes agent block from .fab-runtime.yaml
- [ ] CHK-009 on-stop.sh migration: Uses `fab runtime set-idle` instead of yq
- [ ] CHK-010 on-session-start.sh migration: Uses `fab runtime clear-idle` instead of yq
- [ ] CHK-011 5-sync-hooks.sh matchers: Registers on-artifact-write.sh with PostToolUse Write and Edit matchers
- [ ] CHK-012 Skill bookkeeping removed: fab-new, fab-continue, fab-ff, fab-fff, fab-clarify, _generation.md no longer contain removed bookkeeping instructions

## Behavioral Correctness

- [ ] CHK-013 Non-artifact files: Hook exits 0 immediately for paths not matching fab artifact patterns
- [ ] CHK-014 Hook failure tolerance: All bookkeeping command failures in hook are silently ignored (exit 0)
- [ ] CHK-015 Missing fab CLI: on-stop.sh and on-session-start.sh exit 0 when fab binary not found
- [ ] CHK-016 Missing runtime file: `fab runtime clear-idle` exits 0 when .fab-runtime.yaml doesn't exist
- [ ] CHK-017 Gate checks preserved: fab-ff.md still contains intake gate and spec gate checks (read-only)
- [ ] CHK-018 Stage transitions preserved: fab-continue.md, fab-ff.md, fab-fff.md still contain all finish/start/reset/fail calls

## Removal Verification

- [ ] CHK-019 No yq in hooks: on-stop.sh and on-session-start.sh contain no yq references
- [ ] CHK-020 No manual bookkeeping: fab-new.md has no set-change-type or fab score --stage intake instructions
- [ ] CHK-021 No manual scoring: fab-continue.md has no fab score call after spec generation
- [ ] CHK-022 No manual checklist: fab-ff.md and fab-fff.md Step 4 has no set-checklist calls
- [ ] CHK-023 No manual clarify scoring: fab-clarify.md has no fab score call in suggest mode
- [ ] CHK-024 No generation checklist commands: _generation.md has no set-checklist CLI commands in checklist procedure

## Scenario Coverage

- [ ] CHK-025 Intake write scenario: Writing intake.md triggers type inference + indicative confidence
- [ ] CHK-026 Spec write scenario: Writing spec.md triggers confidence computation
- [ ] CHK-027 Tasks write scenario: Writing tasks.md triggers total count update
- [ ] CHK-028 Checklist write scenario: Writing checklist.md triggers generated/total/completed update
- [ ] CHK-029 Edit event scenario: Editing spec.md or checklist.md triggers same bookkeeping as write
- [ ] CHK-030 Stop hook scenario: Agent stop sets idle timestamp via fab runtime
- [ ] CHK-031 Session start scenario: Session start clears idle via fab runtime

## Edge Cases & Error Handling

- [ ] CHK-032 Invalid change reference: fab runtime commands exit non-zero for unresolvable changes
- [ ] CHK-033 Runtime file creation: fab runtime set-idle creates .fab-runtime.yaml if it doesn't exist
- [ ] CHK-034 No active change: Hook scripts exit 0 when fab/current is missing or empty
- [ ] CHK-035 Concurrent entries: fab runtime preserves other change entries when modifying one

## Code Quality

- [ ] CHK-036 Pattern consistency: New Go code follows existing cmd/fab and internal package patterns
- [ ] CHK-037 No unnecessary duplication: Hook script reuses fab CLI commands, doesn't duplicate logic
- [ ] CHK-038 Readability: Following existing project conventions for error handling and code structure

## Documentation Accuracy

- [ ] CHK-039 _scripts.md updated: fab runtime command reference added with set-idle and clear-idle docs
- [ ] CHK-040 Spec cross-reference: SPEC-hooks.md referenced in docs/specs/skills/ (if applicable)

## Cross References

- [ ] CHK-041 Constitution §I: Text already reads "scripts" (not "shell scripts") — verify no regression
- [ ] CHK-042 Memory affected files: All 5 affected memory domains identified for hydration

## Notes

- Check items as you review: `- [x]`
- All items must pass before `/fab-continue` (hydrate)
- If an item is not applicable, mark checked and prefix with **N/A**: `- [x] CHK-008 **N/A**: {reason}`
