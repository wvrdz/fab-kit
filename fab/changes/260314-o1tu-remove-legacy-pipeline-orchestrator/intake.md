# Intake: Remove Legacy Pipeline Orchestrator

**Change**: 260314-o1tu-remove-legacy-pipeline-orchestrator
**Created**: 2026-03-14
**Status**: Draft

## Origin

> Now that we are going the operator route, we can remove the older "pipeline" commands... I mean "batch-pipeline-*" commands, and the fab/pipelines folder (& any other dependencies for the automated pipelines). Also: fab/.kit/scaffold/fab/pipelines. If needed, also add a migration step (to cleanup .gitignore entries etc).

Discussion-mode conversation confirmed the full inventory and agreed on scope before `/fab-new`.

## Why

The batch-pipeline system (`batch-pipeline.sh`, `batch-pipeline-series.sh`, `pipeline/run.sh`, `pipeline/dispatch.sh`) was the v1 approach to multi-change coordination — a serial bash orchestrator that read YAML manifests and dispatched `fab-ff` into worktree panes. The operator skills (`/fab-operator1`, `/fab-operator2`) supersede this entirely with a more flexible, parallel, interactive coordination model that operates through tmux pane observation and `tmux send-keys`.

Keeping the dead code adds maintenance burden (7 files, ~400 lines of bash, a BATS test suite, a memory file, scaffold artifacts, gitignore entries) and confuses users who see `fab/pipelines/` in their project without understanding it's deprecated.

## What Changes

### 1. Delete Scripts

Remove the following files from `fab/.kit/scripts/`:

- `batch-pipeline.sh` — manifest-based entry point (listing, partial name matching, delegation)
- `batch-pipeline-series.sh` — sequential chain shorthand (generates temp manifest, delegates)
- `pipeline/run.sh` — main orchestrator loop (dispatch, poll, ship, SIGINT handling)
- `pipeline/dispatch.sh` — per-change dispatch (worktree creation, artifact provisioning, pane creation, send-keys)

After deletion, the `fab/.kit/scripts/pipeline/` directory should be removed entirely (it will be empty).

### 2. Delete Scaffold Artifacts

Remove `fab/.kit/scaffold/fab/pipelines/` directory:

- `fab/.kit/scaffold/fab/pipelines/example.yaml` — the scaffold copy of the example manifest

This prevents new projects from getting a `fab/pipelines/` directory during setup.

### 3. Delete User Data Directory

Remove `fab/pipelines/` from this repo:

- `fab/pipelines/example.yaml` — documentation-as-code example manifest
- `fab/pipelines/pipeline1.yaml` — a completed pipeline manifest (historical)

### 4. Delete BATS Test Suite

Remove `src/sh/pipeline/test.bats` — the test suite for `run.sh` and `dispatch.sh` pure-logic functions (38+ tests). With the scripts gone, the tests are orphaned.

### 5. Delete Memory File

Remove `docs/memory/fab-workflow/pipeline-orchestrator.md` — the post-implementation memory documenting the orchestrator's manifest format, dispatch loop, polling, shipping, design decisions, and changelog.

### 6. Update Memory Index

Edit `docs/memory/fab-workflow/index.md`:
- Remove the `pipeline-orchestrator` row from the file table

### 7. Update Memory Domain Index

Edit `docs/memory/index.md`:
- Remove `pipeline-orchestrator` from the `fab-workflow` domain's Memory Files list

### 8. Update Kit Architecture Memory

Edit `docs/memory/fab-workflow/kit-architecture.md`:
- Remove `batch-pipeline.sh` and `batch-pipeline-series.sh` from the scripts directory tree
- Remove the `pipeline/` subdirectory from the scripts tree
- Remove the `#### batch-pipeline.sh` and `#### batch-pipeline-series.sh` sections
- Remove references to pipeline scripts in the "Batch Scripts" naming pattern note
- Remove the `dispatch.sh` reference in the Non-Interactive Porcelain Output Contract
- Remove the changelog entry for `260222-bcfy-batch-pipeline-series-rename`

### 9. Update fab-help.sh

Edit `fab/.kit/scripts/fab-help.sh`:
- Remove the `batch-pipeline` and `batch-pipeline-series` entries from the `batch_to_group` mapping (or equivalent category mapping)

### 10. Clean Up .gitignore Files

**Repo `.gitignore`**: Remove the line `fab/pipelines/.series-*.yaml`

**Scaffold `fragment-.gitignore`**: Remove the line `fab/pipelines/.series-*.yaml`

### 11. Add Migration

Create `fab/.kit/migrations/0.37.0-to-0.38.0.md` with instructions for deployed projects to:

1. Remove `fab/pipelines/` directory if it exists
2. Remove the `fab/pipelines/.series-*.yaml` line from `.gitignore` if present
3. No `.status.yaml` or `config.yaml` changes needed

## Affected Memory

- `fab-workflow/pipeline-orchestrator`: (remove) Entire memory file deleted — system no longer exists
- `fab-workflow/kit-architecture`: (modify) Remove pipeline script entries from directory tree, script descriptions, batch naming pattern, and changelog
- `fab-workflow/index`: (modify) Remove pipeline-orchestrator from file table

## Impact

- **Scripts**: 4 files deleted, 1 directory removed (`pipeline/`)
- **Scaffold**: 1 directory removed (`fab/pipelines/`)
- **User data**: 1 directory removed (`fab/pipelines/`)
- **Tests**: 1 BATS file deleted
- **Memory**: 1 file deleted, 2 files edited
- **Memory index**: 1 entry removed from domain index, 1 from top-level index
- **Help**: 2 entries removed from fab-help.sh
- **Gitignore**: 1 line removed from each of 2 files
- **Migration**: 1 new migration file

No runtime behavior changes — the pipeline orchestrator was not invoked by any skill or hook. The operator skills have no dependency on these files.

## Open Questions

- None — scope is well-defined from the discussion session.

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Delete all 4 pipeline scripts and the pipeline/ directory | Discussed — user explicitly listed these | S:95 R:90 A:95 D:95 |
| 2 | Certain | Delete fab/pipelines/ (user data) and scaffold copy | Discussed — user explicitly requested | S:95 R:85 A:90 D:95 |
| 3 | Certain | Delete pipeline-orchestrator.md memory file | Discussed — system no longer exists, memory is orphaned | S:90 R:90 A:95 D:95 |
| 4 | Certain | Remove fab-help.sh entries | Directly follows from script deletion | S:85 R:95 A:95 D:95 |
| 5 | Certain | Clean .gitignore entries in both repo and scaffold | Discussed — user explicitly mentioned .gitignore cleanup | S:95 R:95 A:95 D:95 |
| 6 | Certain | Add migration for deployed projects | Discussed — user explicitly requested migration | S:95 R:85 A:90 D:95 |
| 7 | Certain | Delete src/sh/pipeline/test.bats | Tests are orphaned when scripts are deleted | S:85 R:85 A:95 D:95 |
| 8 | Confident | Migration targets 0.37.0-to-0.38.0 range | Latest migration is 0.34.0-to-0.37.0, current VERSION is 0.36.3, next release will be 0.37.0+ | S:70 R:90 A:80 D:75 |
| 9 | Certain | Leave archived changes untouched | Archives are historical records — discussed and agreed | S:90 R:95 A:95 D:95 |
| 10 | Confident | Update kit-architecture.md memory references | Follows from deletion — references become stale without cleanup | S:80 R:85 A:90 D:90 |

10 assumptions (9 certain, 1 confident, 0 tentative, 0 unresolved).
