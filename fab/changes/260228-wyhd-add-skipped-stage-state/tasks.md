# Tasks: Add "skipped" Stage State

**Change**: 260228-wyhd-add-skipped-stage-state
**Spec**: `spec.md`
**Intake**: `intake.md`

## Phase 1: Schema Changes

- [x] T001 [P] Add `skipped` state definition to `fab/.kit/schemas/workflow.yaml` — new entry in `states` list with `id: skipped`, `symbol: "⏭"`, `description: "Stage intentionally bypassed"`, `terminal: true`
- [x] T002 [P] Update `allowed_states` in `fab/.kit/schemas/workflow.yaml` — add `skipped` to `spec`, `tasks`, `apply`, `review`, `hydrate` stages (not `intake`). Review already has `failed`; append `skipped` after it
- [x] T003 [P] Add `skip` transition to `transitions.default` in `fab/.kit/schemas/workflow.yaml` — `event: skip`, `from: [pending]`, `to: skipped`
- [x] T004 [P] Update `reset` transitions in `fab/.kit/schemas/workflow.yaml` — add `skipped` to `from` array in both `transitions.default` reset and `transitions.review` reset: `from: [done, ready, skipped]`
- [x] T005 [P] Update progression rules in `fab/.kit/schemas/workflow.yaml` — update `current_stage.rule`, `next_stage.rule`, `completion.rule`, and `validation.prerequisites.rule` to include `skipped` alongside `done`

## Phase 2: Core Implementation

- [x] T006 Add `event_skip` function to `fab/.kit/scripts/lib/statusman.sh` — implements `pending → skipped` with forward cascade (all downstream `pending` stages → `skipped`). No auto-activate, no metrics side-effect. Atomic write (tmpfile + mv). Signature: `event_skip <status_file> <stage> [driver]`
- [x] T007 Add `skip` CLI subcommand to `fab/.kit/scripts/lib/statusman.sh` — dispatch section and help text. Accepts `skip <change> <stage> [driver]`, resolves via `resolve_to_status`, dispatches to `event_skip`
- [x] T008 Update `get_current_stage` in `fab/.kit/scripts/lib/statusman.sh` — fallback logic tracks last `done` or `skipped` stage (currently only tracks `done`)
- [x] T009 [P] Update `get_display_stage` in `fab/.kit/scripts/lib/statusman.sh` — Tier 3 fallback tracks `skipped` alongside `done` for "last resolved" detection
- [x] T010 [P] Update `get_progress_line` in `fab/.kit/scripts/lib/statusman.sh` — add `skipped) parts+=("$stage ⏭") ;;` case in the rendering loop
- [x] T011 [P] Update `_apply_metrics_side_effect` in `fab/.kit/scripts/lib/statusman.sh` — add `skipped` case that deletes metrics (same as `pending`)

## Phase 3: Tests

- [x] T012 Add tests for `skip` event to `src/lib/statusman/test.bats` — pending→skipped, forward cascade, rejects non-pending, skipped in progress-line, skipped in current-stage fallback, skipped in display-stage, reset from skipped, validate-status-file accepts skipped for non-intake and rejects for intake, _apply_metrics deletes metrics for skipped

---

## Execution Order

- T001–T005 are independent schema changes, all parallelizable
- T006 depends on T003 (needs `skip` transition in schema for `lookup_transition`)
- T007 depends on T006 (dispatches to `event_skip`)
- T008 depends on T001 (needs `skipped` state to be valid)
- T009–T011 depend on T001 (need `skipped` state to be valid)
- T012 depends on T006–T011 (tests the complete implementation)
