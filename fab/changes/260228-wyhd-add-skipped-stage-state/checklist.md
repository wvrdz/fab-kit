# Quality Checklist: Add "skipped" Stage State

**Change**: 260228-wyhd-add-skipped-stage-state
**Generated**: 2026-02-28
**Spec**: `spec.md`
**Reviewed**: 2026-02-28

## Functional Completeness

- [x] CHK-001 Skipped state definition: `workflow.yaml` contains `skipped` state with correct `id`, `symbol`, `description`, `terminal` fields
- [x] CHK-002 Allowed states: `skipped` present in `allowed_states` for spec, tasks, apply, review, hydrate — absent for intake
- [x] CHK-003 Skip transition: `transitions.default` has `skip` event with `from: [pending]`, `to: skipped`
- [x] CHK-004 Reset transition: both `transitions.default` and `transitions.review` reset entries include `skipped` in `from` array
- [x] CHK-005 Progression rules: all four rules updated to reference `skipped` alongside `done`
- [x] CHK-006 event_skip function: exists in `statusman.sh` with correct signature and behavior
- [x] CHK-007 CLI skip subcommand: `statusman.sh skip <change> <stage> [driver]` dispatches to `event_skip`
- [x] CHK-008 get_current_stage: treats `skipped` like `done` in fallback logic
- [x] CHK-009 get_display_stage: treats `skipped` like `done` in Tier 3 fallback
- [x] CHK-010 get_progress_line: renders `skipped` stages with `⏭` symbol
- [x] CHK-011 _apply_metrics_side_effect: deletes metrics for `skipped` state

## Behavioral Correctness

- [x] CHK-012 Skip forward cascade: skipping a stage cascades all downstream `pending` → `skipped`
- [x] CHK-013 Skip cascade conservative: cascade only affects `pending` stages, not `done`/`active`/`ready`/`failed`
- [x] CHK-014 No auto-activate after skip: skipping does not activate the next stage
- [x] CHK-015 Reset from skipped: resetting a `skipped` stage → `active`, downstream → `pending`

## Scenario Coverage

- [x] CHK-016 Skip pending stage: `pending → skipped` succeeds
- [x] CHK-017 Skip rejects non-pending: `active`, `done`, `ready` all rejected
- [x] CHK-018 Skip mid-pipeline: only downstream pending stages affected
- [x] CHK-019 Current stage skips over skipped: fallback finds first pending after last done/skipped
- [x] CHK-020 Display stage with skipped: returns `skipped` stage as Tier 3 when it's the last resolved
- [x] CHK-021 Progress line with skipped: renders `stage ⏭` in output
- [x] CHK-022 All-skipped completion: intake done + all else skipped → completion marker `✓`
- [x] CHK-023 Validate accepts skipped for non-intake stages
- [x] CHK-024 Validate rejects skipped for intake stage

## Edge Cases & Error Handling

- [x] CHK-025 Skip with invalid stage name: exits non-zero with error
- [x] CHK-026 Skip with nonexistent file: exits non-zero with error
- [x] CHK-027 Skip with insufficient args: prints usage to stderr, exits 1
- [x] CHK-028 Intake skip attempt blocked by schema: `validate_stage_state` rejects, `lookup_transition` fails (no skip in intake override or default for intake's allowed states)

## Code Quality

- [x] CHK-029 Pattern consistency: `event_skip` follows the same structure as `event_start`, `event_finish`, `event_reset` (validation, lookup, tmpfile, atomic write)
- [x] CHK-030 No unnecessary duplication: forward cascade reuses the stage iteration pattern from `event_reset`
- [x] CHK-031 Readability: function and case additions follow existing code style (indentation, comments, variable naming)

## Documentation Accuracy

- [x] CHK-032 Help text: `--help` output includes `skip` subcommand with correct usage and description
- [x] CHK-033 Progression rules: YAML comments in `workflow.yaml` accurately describe the updated behavior

## Cross References

- [x] CHK-034 Schema/script consistency: `event_skip` uses `lookup_transition` which reads the schema — transitions match
- [x] CHK-035 Test coverage: all new behaviors have corresponding test cases in `src/lib/statusman/test.bats`

## Notes

- Check items as you review: `- [x]`
- All items must pass before `/fab-continue` (hydrate)
- If an item is not applicable, mark checked and prefix with **N/A**: `- [x] CHK-008 **N/A**: {reason}`
