# Spec: Standardize archiveman.sh Dispatcher Integration

**Change**: 260305-jv4y-standardize-archiveman-dispatcher
**Created**: 2026-03-05
**Affected memory**: `docs/memory/fab-workflow/kit-architecture.md`

## Shell Dispatcher: Archive Command Pass-Through

### Requirement: Dispatcher SHALL pass arguments to archiveman.sh without injection

The `fab` shell dispatcher's `archive` case SHALL forward all positional arguments to `archiveman.sh` without prepending a hardcoded `"archive"` subcommand. This standardizes the `archive` entry to match all other dispatcher entries (plain pass-through).

#### Scenario: Archive via dispatcher

- **GIVEN** the shell backend is active (no `fab-go` or `fab-rust` binary)
- **WHEN** the user runs `fab archive <change> --description "..."`
- **THEN** the dispatcher executes `archiveman.sh <change> --description "..."` (no injected `archive` arg)
- **AND** archiveman.sh treats `<change>` as a change name (default-to-archive fallback)

#### Scenario: Restore via dispatcher

- **GIVEN** the shell backend is active
- **WHEN** the user runs `fab archive restore <change>`
- **THEN** the dispatcher executes `archiveman.sh restore <change>`
- **AND** archiveman.sh recognizes `restore` as a subcommand and restores the archived change

#### Scenario: List via dispatcher

- **GIVEN** the shell backend is active
- **WHEN** the user runs `fab archive list`
- **THEN** the dispatcher executes `archiveman.sh list`
- **AND** archiveman.sh recognizes `list` as a subcommand and lists archived changes

## archiveman.sh: Default-to-Archive Fallback

### Requirement: Unknown first argument SHALL default to archive subcommand

When `$1` is not a recognized subcommand (`archive`, `restore`, `list`, `--help`, `-h`), archiveman.sh SHALL treat all arguments as `cmd_archive "$@"` (no shift — `$1` is the change name). The empty-argument case SHALL remain an error.

#### Scenario: Change name as first argument (new default path)

- **GIVEN** archiveman.sh is invoked with a change name as `$1` (e.g., `archiveman.sh t3st --description "..."`)
- **WHEN** `$1` does not match any known subcommand
- **THEN** archiveman.sh delegates to `cmd_archive` with all arguments (`$@`)
- **AND** the archive operation succeeds identically to `archiveman.sh archive t3st --description "..."`

#### Scenario: Explicit archive subcommand still works

- **GIVEN** archiveman.sh is invoked with `archive` as `$1`
- **WHEN** the user or script runs `archiveman.sh archive <change> --description "..."`
- **THEN** archiveman.sh recognizes `archive`, shifts, and delegates to `cmd_archive`
- **AND** backwards compatibility is preserved

#### Scenario: Empty arguments remain an error

- **GIVEN** archiveman.sh is invoked with no arguments
- **WHEN** `$1` is empty
- **THEN** archiveman.sh exits non-zero with an error message

### Requirement: Recognized subcommands SHALL continue to work unchanged

The `restore`, `list`, `--help`, and `-h` subcommands SHALL continue to function identically. No behavioral changes to these paths.

#### Scenario: Restore subcommand

- **GIVEN** archiveman.sh is invoked as `archiveman.sh restore <change>`
- **WHEN** `$1` is `restore`
- **THEN** archiveman.sh shifts and delegates to `cmd_restore`

#### Scenario: List subcommand

- **GIVEN** archiveman.sh is invoked as `archiveman.sh list`
- **WHEN** `$1` is `list`
- **THEN** archiveman.sh shifts and delegates to `cmd_list`

## Parity Tests: Exercise Default Path

### Requirement: Parity tests SHOULD exercise the default-to-archive path

The bash side of the archive parity test SHALL invoke `archiveman.sh` without the explicit `archive` subcommand for the archive operation, exercising the new default-to-archive fallback. The `list` test needs no change (already passes `list` directly).

#### Scenario: Parity test for archive

- **GIVEN** the archive parity test runs
- **WHEN** the bash side invokes `archiveman.sh <changeID> --description "test archive"` (without explicit `archive`)
- **THEN** the bash output matches the Go binary output (`fab-go archive <changeID> --description "test archive"`)

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Default to `cmd_archive` when $1 isn't a known subcommand | Confirmed from intake #1 — user chose Option 3 | S:95 R:90 A:95 D:95 |
| 2 | Certain | Keep explicit `archive` subcommand working | Confirmed from intake #2 — backwards compatibility | S:90 R:95 A:90 D:95 |
| 3 | Certain | Go backend needs no changes | Confirmed from intake #3 — Cobra structure already correct | S:95 R:95 A:95 D:95 |
| 4 | Confident | Empty args case stays an error | Confirmed from intake #4 — no reason to change | S:80 R:90 A:85 D:80 |
| 5 | Certain | Parity test bash side updated to exercise default path | Upgraded from intake #5 — test the new behavior, not just backwards compat | S:90 R:85 A:90 D:90 |
| 6 | Certain | Bats tests need no changes | Bats tests invoke archiveman.sh with explicit `archive` subcommand which still works | S:90 R:95 A:90 D:95 |

6 assumptions (5 certain, 1 confident, 0 tentative, 0 unresolved).
