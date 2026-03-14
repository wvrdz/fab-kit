# Tasks: Remove Legacy Pipeline Orchestrator

**Change**: 260314-o1tu-remove-legacy-pipeline-orchestrator
**Spec**: `spec.md`
**Intake**: `intake.md`

## Phase 1: Delete Pipeline Files

- [x] T001 [P] Delete `fab/.kit/scripts/batch-pipeline.sh`
- [x] T002 [P] Delete `fab/.kit/scripts/batch-pipeline-series.sh`
- [x] T003 [P] Delete `fab/.kit/scripts/pipeline/run.sh` and `fab/.kit/scripts/pipeline/dispatch.sh`, then remove the empty `fab/.kit/scripts/pipeline/` directory
- [x] T004 [P] Delete `fab/.kit/scaffold/fab/pipelines/example.yaml`, then remove the empty `fab/.kit/scaffold/fab/pipelines/` directory
- [x] T005 [P] Delete `fab/pipelines/example.yaml` and `fab/pipelines/pipeline1.yaml`, then remove the empty `fab/pipelines/` directory
- [x] T006 [P] Delete `src/sh/pipeline/test.bats`, then remove the empty `src/sh/pipeline/` directory
- [x] T007 [P] Delete `docs/memory/fab-workflow/pipeline-orchestrator.md`

## Phase 2: Update References

- [x] T008 [P] Remove `batch-pipeline.sh` and `batch-pipeline-series.sh` rows from the Shell Utilities table in `README.md`
- [x] T009 [P] Remove `batch-pipeline` and `batch-pipeline-series` entries from the `batch_to_group` mapping in `fab/.kit/scripts/fab-help.sh`
- [x] T010 [P] Remove the `fab/pipelines/.series-*.yaml` line from `.gitignore`
- [x] T011 [P] Remove the `fab/pipelines/.series-*.yaml` line from `fab/.kit/scaffold/fragment-.gitignore`
- [x] T012 [P] Remove the `pipeline-orchestrator` row from `docs/memory/fab-workflow/index.md`
- [x] T013 [P] Remove `pipeline-orchestrator` from the `fab-workflow` domain's Memory Files list in `docs/memory/index.md`
- [x] T014 Update `docs/memory/fab-workflow/kit-architecture.md`: remove `batch-pipeline.sh`, `batch-pipeline-series.sh`, and `pipeline/` directory entries from the scripts directory tree; remove script description sections; remove pipeline changelog entries; remove `dispatch.sh` from Non-Interactive Porcelain Output Contract if referenced

## Phase 3: Migration

- [x] T015 Create `fab/.kit/migrations/0.37.0-to-0.38.0.md` with instructions to remove `fab/pipelines/` and clean `.gitignore`

---

## Execution Order

- Phase 1 tasks (T001–T007) are all independent and parallelizable
- Phase 2 tasks (T008–T013) are all independent and parallelizable; T014 is larger but independent
- Phase 3 (T015) is independent of all other tasks
- No cross-phase dependencies — all phases can technically run in parallel, but logical ordering is cleaner
