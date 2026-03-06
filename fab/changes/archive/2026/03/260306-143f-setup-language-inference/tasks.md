# Tasks: Replace Template-Driven Language Detection with Agent-Inferred Conventions

**Change**: 260306-143f-setup-language-inference
**Spec**: `spec.md`
**Intake**: `intake.md`

## Phase 1: Cleanup

- [x] T001 [P] Delete template directories: `fab/.kit/templates/constitutions/` (3 files) and `fab/.kit/templates/configs/` (3 files)
- [x] T002 [P] Remove the "Language template advisory" block (section 2b, lines 219–247) from `fab/.kit/sync/2-sync-workspace.sh`

## Phase 2: Core Implementation

- [x] T003 Rewrite step 1b-lang in `fab/.kit/skills/fab-setup.md`: replace the template-lookup detection table and application logic (lines 83–115 of the Bootstrap Behavior section) with the three-phase agent-inference flow (Detection → Inference → Write) as specified in `spec.md`

## Phase 3: Integration & Edge Cases

- [x] T004 Verify idempotency: ensure the new step 1b-lang text describes merge/update behavior — read existing `fab/project/*` content before writing, skip already-present conventions, preserve user edits
- [x] T005 Verify the `fab/.kit/templates/` directory still contains artifact templates (`intake.md`, `spec.md`, `tasks.md`, `checklist.md`, `status.yaml`) after template directory deletion — no accidental removal

---

## Execution Order

- T001 and T002 are independent — can run in parallel
- T003 depends on T001 (template references must be gone before rewriting the step that referenced them)
- T004 depends on T003 (verifies the rewritten content)
- T005 depends on T001 (verifies artifact templates survive deletion)
