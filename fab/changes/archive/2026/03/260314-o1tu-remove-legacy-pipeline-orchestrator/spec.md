# Spec: Remove Legacy Pipeline Orchestrator

**Change**: 260314-o1tu-remove-legacy-pipeline-orchestrator
**Created**: 2026-03-14
**Affected memory**: `docs/memory/fab-workflow/pipeline-orchestrator.md` (remove), `docs/memory/fab-workflow/kit-architecture.md` (modify), `docs/memory/fab-workflow/index.md` (modify)

## Non-Goals

- Modifying any operator skill (`/fab-operator1`, `/fab-operator2`, `/fab-operator3`) — they have no dependency on the pipeline system
- Touching archived changes that reference pipeline manifests — archives are historical records
- Removing `fab/.kit/hooks/on-stop.sh` or `on-session-start.sh` — these serve the operator/runtime system, not the pipeline

## Script Removal

### Requirement: Delete Pipeline Scripts

All pipeline orchestrator scripts SHALL be removed from `fab/.kit/scripts/`:

- `batch-pipeline.sh` (user-facing entry point)
- `batch-pipeline-series.sh` (sequential chain shorthand)
- `pipeline/run.sh` (main orchestrator loop)
- `pipeline/dispatch.sh` (per-change dispatch)

The `fab/.kit/scripts/pipeline/` directory SHALL be removed entirely after its contents are deleted.

#### Scenario: Pipeline scripts no longer exist after removal
- **GIVEN** the change has been applied
- **WHEN** listing files in `fab/.kit/scripts/`
- **THEN** `batch-pipeline.sh` and `batch-pipeline-series.sh` do not exist
- **AND** the `pipeline/` subdirectory does not exist

#### Scenario: No dangling references to pipeline scripts
- **GIVEN** the pipeline scripts have been removed
- **WHEN** searching the codebase for `batch-pipeline` or `pipeline/run.sh` or `pipeline/dispatch.sh`
- **THEN** no references remain in active code (README, fab-help.sh, kit-architecture memory)

## Scaffold Removal

### Requirement: Remove Pipeline Scaffold Artifacts

The scaffold directory `fab/.kit/scaffold/fab/pipelines/` SHALL be removed entirely, including `example.yaml`. New projects initialized via `/fab-setup` SHALL NOT receive a `fab/pipelines/` directory.

#### Scenario: New project setup produces no pipelines directory
- **GIVEN** the scaffold `fab/pipelines/` directory has been removed
- **WHEN** a new project runs `/fab-setup`
- **THEN** no `fab/pipelines/` directory is created in the project

## User Data Removal

### Requirement: Remove Pipeline Manifests from This Repo

The `fab/pipelines/` directory SHALL be removed from this repository, including:

- `example.yaml` (documentation-as-code example)
- `pipeline1.yaml` (completed historical manifest)

#### Scenario: Pipeline user data directory removed
- **GIVEN** the change has been applied
- **WHEN** checking for `fab/pipelines/` in the repo
- **THEN** the directory does not exist

## Test Removal

### Requirement: Delete Orphaned BATS Test Suite

The file `src/sh/pipeline/test.bats` SHALL be removed. With the pipeline scripts deleted, the test suite has no code to test.

#### Scenario: Pipeline tests removed
- **GIVEN** the pipeline scripts have been deleted
- **WHEN** checking for `src/sh/pipeline/test.bats`
- **THEN** the file does not exist
- **AND** the `src/sh/pipeline/` directory is removed if empty

## Documentation Updates

### Requirement: Remove Pipeline Entries from README

The Shell Utilities table in `README.md` SHALL have the `batch-pipeline.sh` and `batch-pipeline-series.sh` rows removed.

#### Scenario: README no longer references pipeline scripts
- **GIVEN** the change has been applied
- **WHEN** reading the Shell Utilities table in `README.md`
- **THEN** no `batch-pipeline` entries appear

### Requirement: Remove Pipeline Entries from fab-help.sh

The `batch_to_group` mapping in `fab/.kit/scripts/fab-help.sh` SHALL have the `batch-pipeline` and `batch-pipeline-series` entries removed.

#### Scenario: fab-help no longer lists pipeline commands
- **GIVEN** the change has been applied
- **WHEN** running `fab-help.sh`
- **THEN** no pipeline-related entries appear in any category

## Gitignore Cleanup

### Requirement: Remove Pipeline Gitignore Entries

The line `fab/pipelines/.series-*.yaml` SHALL be removed from:
1. The repo root `.gitignore`
2. The scaffold `fab/.kit/scaffold/fragment-.gitignore`

#### Scenario: Gitignore entries cleaned up
- **GIVEN** the change has been applied
- **WHEN** reading `.gitignore` and `fab/.kit/scaffold/fragment-.gitignore`
- **THEN** neither file contains `fab/pipelines/.series-*.yaml`

## Memory Cleanup

### Requirement: Delete Pipeline Orchestrator Memory File

The file `docs/memory/fab-workflow/pipeline-orchestrator.md` SHALL be deleted entirely. The system it documents no longer exists.

#### Scenario: Memory file removed
- **GIVEN** the change has been applied
- **WHEN** checking `docs/memory/fab-workflow/`
- **THEN** `pipeline-orchestrator.md` does not exist

### Requirement: Update Memory Domain Index

The file `docs/memory/fab-workflow/index.md` SHALL have the `pipeline-orchestrator` row removed from its file table.

#### Scenario: Domain index no longer references pipeline-orchestrator
- **GIVEN** the change has been applied
- **WHEN** reading `docs/memory/fab-workflow/index.md`
- **THEN** no row mentions `pipeline-orchestrator`

### Requirement: Update Top-Level Memory Index

The file `docs/memory/index.md` SHALL have `pipeline-orchestrator` removed from the `fab-workflow` domain's Memory Files list.

#### Scenario: Top-level index no longer lists pipeline-orchestrator
- **GIVEN** the change has been applied
- **WHEN** reading `docs/memory/index.md`
- **THEN** the `fab-workflow` row's Memory Files column does not include `pipeline-orchestrator`

### Requirement: Update Kit Architecture Memory

The file `docs/memory/fab-workflow/kit-architecture.md` SHALL be updated to remove all pipeline orchestrator references:

- Remove `batch-pipeline.sh` and `batch-pipeline-series.sh` from the scripts directory tree
- Remove the `pipeline/` subdirectory (containing `run.sh` and `dispatch.sh`) from the scripts tree
- Remove the `#### batch-pipeline.sh` and `#### batch-pipeline-series.sh` sections
- Remove `dispatch.sh` from the Non-Interactive Porcelain Output Contract if referenced
- Remove the `260222-bcfy-batch-pipeline-series-rename` changelog entry and any other pipeline-specific changelog entries

#### Scenario: Kit architecture memory has no pipeline references
- **GIVEN** the change has been applied
- **WHEN** searching `kit-architecture.md` for `pipeline`, `batch-pipeline`, or `dispatch.sh`
- **THEN** no matches are found (except in the changelog entry for this change itself, if added)

## Migration

### Requirement: Create Migration for Deployed Projects

A migration file `fab/.kit/migrations/0.37.0-to-0.38.0.md` SHALL be created with instructions for deployed projects to:

1. Remove the `fab/pipelines/` directory if it exists
2. Remove the `fab/pipelines/.series-*.yaml` line from `.gitignore` if present

No `.status.yaml` or `config.yaml` schema changes are needed.

#### Scenario: Migration file exists with correct instructions
- **GIVEN** the change has been applied
- **WHEN** reading `fab/.kit/migrations/0.37.0-to-0.38.0.md`
- **THEN** the file contains removal instructions for `fab/pipelines/` and the `.gitignore` entry
- **AND** no status or config schema changes are mentioned

## Deprecated Requirements

### Pipeline Orchestrator (entire subsystem)

**Reason**: Superseded by the operator skills (`/fab-operator1`, `/fab-operator2`, `/fab-operator3`) which provide a more flexible, parallel, interactive coordination model via tmux pane observation and `tmux send-keys`.

**Migration**: Use `/fab-operator1` (or `/fab-operator2`, `/fab-operator3`) for multi-change coordination. No data migration needed — pipeline manifests are not consumed by operators.

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Delete all 4 pipeline scripts and the pipeline/ directory | Confirmed from intake #1 — user explicitly listed these | S:95 R:90 A:95 D:95 |
| 2 | Certain | Delete fab/pipelines/ (user data) and scaffold copy | Confirmed from intake #2 — user explicitly requested | S:95 R:85 A:90 D:95 |
| 3 | Certain | Delete pipeline-orchestrator.md memory file | Confirmed from intake #3 — system no longer exists | S:90 R:90 A:95 D:95 |
| 4 | Certain | Remove fab-help.sh entries | Confirmed from intake #4 — directly follows from script deletion | S:85 R:95 A:95 D:95 |
| 5 | Certain | Clean .gitignore entries in both repo and scaffold | Confirmed from intake #5 — user explicitly mentioned | S:95 R:95 A:95 D:95 |
| 6 | Certain | Add migration for deployed projects | Confirmed from intake #6 — user explicitly requested | S:95 R:85 A:90 D:95 |
| 7 | Certain | Delete src/sh/pipeline/test.bats | Confirmed from intake #7 — tests are orphaned | S:85 R:85 A:95 D:95 |
| 8 | Confident | Migration targets 0.37.0-to-0.38.0 range | Confirmed from intake #8 — latest migration is 0.34.0-to-0.37.0, VERSION is 0.36.3 | S:70 R:90 A:80 D:75 |
| 9 | Certain | Leave archived changes untouched | Confirmed from intake #9 — archives are historical | S:90 R:95 A:95 D:95 |
| 10 | Certain | Update kit-architecture.md memory references | Upgraded from intake #10 Confident — confirmed by reading the file, clear what to remove | S:85 R:85 A:95 D:90 |
| 11 | Certain | Remove entire pipeline/ subdirectory under scripts/ | Directory will be empty after run.sh and dispatch.sh deleted | S:90 R:90 A:95 D:95 |
| 12 | Certain | Remove src/sh/pipeline/ directory if empty after test deletion | Only file in directory is test.bats | S:85 R:90 A:95 D:95 |

12 assumptions (11 certain, 1 confident, 0 tentative, 0 unresolved).
