# Tasks: Add fab-proceed to _cli-fab.md

**Change**: 260326-4r2p-add-proceed-to-cli-fab
**Spec**: `spec.md`
**Intake**: `intake.md`

## Phase 1: Core Implementation

- [x] T001 [P] Add **Notable callers** note to `fab resolve` section in `fab/.kit/skills/_cli-fab.md` — document `/fab-proceed` invoking `fab resolve --folder 2>/dev/null` for active change detection
- [x] T002 [P] Add **Notable callers** note after the `fab change` subcommand table in `fab/.kit/skills/_cli-fab.md` — document `/fab-proceed` dispatching `fab change switch` via subagent for unactivated intakes

## Phase 2: Verification

- [x] T003 Verify `docs/memory/fab-workflow/execution-skills.md` already covers `/fab-proceed`'s CLI invocation patterns (`fab resolve --folder`, `fab change switch`) — confirm no memory update needed

---

## Execution Order

- T001 and T002 are independent ([P])
- T003 runs after T001/T002 to verify completeness
