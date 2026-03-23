# Quality Checklist: Remove Legacy Pipeline Orchestrator

**Change**: 260314-o1tu-remove-legacy-pipeline-orchestrator
**Generated**: 2026-03-14
**Spec**: `spec.md`

## Functional Completeness

- [x] CHK-001 Script Removal: All 4 pipeline scripts deleted (`batch-pipeline.sh`, `batch-pipeline-series.sh`, `pipeline/run.sh`, `pipeline/dispatch.sh`)
- [x] CHK-002 Directory Removal: `fab/.kit/scripts/pipeline/` directory removed
- [x] CHK-003 Scaffold Removal: `fab/.kit/scaffold/fab/pipelines/` directory removed
- [x] CHK-004 User Data Removal: `fab/pipelines/` directory removed
- [x] CHK-005 Test Removal: `src/sh/pipeline/test.bats` and `src/sh/pipeline/` directory removed
- [x] CHK-006 Memory File Removal: `docs/memory/fab-workflow/pipeline-orchestrator.md` deleted
- [x] CHK-007 Migration Created: `fab/.kit/migrations/0.37.0-to-0.38.0.md` exists with correct instructions

## Removal Verification

- [x] CHK-008 README: No `batch-pipeline` entries in Shell Utilities table
- [x] CHK-009 fab-help.sh: No `batch-pipeline` or `batch-pipeline-series` entries in `batch_to_group`
- [x] CHK-010 .gitignore: No `fab/pipelines/.series-*.yaml` line in repo `.gitignore`
- [x] CHK-011 scaffold .gitignore: No `fab/pipelines/.series-*.yaml` line in `fragment-.gitignore`
- [x] CHK-012 Domain index: No `pipeline-orchestrator` row in `docs/memory/fab-workflow/index.md`
- [x] CHK-013 Top-level index: No `pipeline-orchestrator` in `docs/memory/index.md` fab-workflow entry
- [x] CHK-014 Kit architecture: No `batch-pipeline`, `pipeline/`, or pipeline changelog entries in `kit-architecture.md`

## Scenario Coverage

- [x] CHK-015 No dangling references: Grep for `batch-pipeline`, `pipeline/run.sh`, `pipeline/dispatch.sh`, `pipeline-orchestrator` returns no hits in active code/docs
- [x] CHK-016 Migration content: Migration file has removal instructions for `fab/pipelines/` and `.gitignore` entry, no schema changes mentioned

## Code Quality

- [x] CHK-017 Pattern consistency: Remaining scripts in `fab/.kit/scripts/` unmodified except where entries removed
- [x] CHK-018 No unnecessary duplication: No leftover pipeline references in any file

## Documentation Accuracy

- [x] CHK-019 Memory indexes: Both `index.md` files accurately reflect the remaining memory files
- [x] CHK-020 Kit architecture tree: Directory tree in `kit-architecture.md` matches actual `fab/.kit/` structure after deletions

## Cross References

- [x] CHK-021 No broken cross-references: No remaining docs/memory files link to `pipeline-orchestrator.md`

## Notes

- Check items as you review: `- [x]`
- All items must pass before `/fab-continue` (hydrate)
- If an item is not applicable, mark checked and prefix with **N/A**: `- [x] CHK-008 **N/A**: {reason}`
