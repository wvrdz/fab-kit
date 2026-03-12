# Spec: Add resolve --pane, drop send-keys, evolve pane-map

**Change**: 260312-kvng-resolve-pane-evolve-panemap
**Created**: 2026-03-12
**Affected memory**: `docs/memory/fab-workflow/kit-architecture.md`, `docs/memory/fab-workflow/execution-skills.md`

## Non-Goals

- Rust binary changes — the Rust binary is not maintained
- Changes to `dispatch.sh` or `run.sh` — they use raw `tmux send-keys` with pane IDs, not `fab send-keys`

## CLI: resolve --pane

### Requirement: Pane output mode

`fab resolve <change> --pane` SHALL output the tmux pane ID (e.g., `%5`) for the pane running the resolved change. The `--pane` flag SHALL be registered as a `Bool` flag on the resolve command and integrated into the existing `PreRunE` priority chain after `--status` and before the default `--id`.

#### Scenario: Successful pane resolution
- **GIVEN** a tmux session with pane `%3` whose CWD is inside a worktree with active change `260306-r3m7-add-retry-logic`
- **WHEN** `fab resolve r3m7 --pane` is executed
- **THEN** stdout prints `%3` followed by a newline
- **AND** exit code is 0

#### Scenario: No tmux session
- **GIVEN** `$TMUX` environment variable is unset
- **WHEN** `fab resolve <change> --pane` is executed
- **THEN** stderr prints `Error: not inside a tmux session`
- **AND** exit code is 1

#### Scenario: No matching pane
- **GIVEN** a tmux session with no pane matching change `r3m7`
- **WHEN** `fab resolve r3m7 --pane` is executed
- **THEN** stderr prints `no tmux pane found for change "<folder>"`
- **AND** exit code is 1

#### Scenario: Multiple panes for same change
- **GIVEN** panes `%3` and `%7` both have active change `260306-r3m7-add-retry-logic`
- **WHEN** `fab resolve r3m7 --pane` is executed
- **THEN** stdout prints the first matching pane ID (`%3`)
- **AND** stderr prints `Warning: multiple panes found for r3m7, using %3`

### Requirement: Pane resolution implementation

The `--pane` mode SHALL reuse `discoverPanes()` and `matchPanesByFolder()` from `panemap.go` (moved there from `sendkeys.go` as part of this change). The `resolvePaneChange()` function SHALL also be reused from `panemap.go`. No new tmux discovery logic SHALL be introduced.

#### Scenario: Reuse of shared functions
- **GIVEN** the resolve command's `--pane` handler
- **WHEN** it resolves a change to a pane
- **THEN** it calls `discoverPanes()` for tmux pane discovery
- **AND** it calls `matchPanesByFolder()` with `resolvePaneChange` as the resolver function

## CLI: send-keys removal

### Requirement: Delete send-keys subcommand

The `fab send-keys` subcommand SHALL be removed from the Go binary entirely. `sendKeysCmd()` SHALL be removed from `main.go`'s `AddCommand` list. The files `sendkeys.go` and `sendkeys_test.go` SHALL be deleted.

#### Scenario: Unknown command after removal
- **GIVEN** the compiled `fab` binary without the send-keys subcommand
- **WHEN** a user runs `fab send-keys r3m7 "/fab-continue"`
- **THEN** Cobra returns an unknown command error
- **AND** exit code is non-zero

### Requirement: Preserve reusable functions

Before deleting `sendkeys.go`, the following functions SHALL be moved to `panemap.go`:
- `resolvePaneChange(p paneEntry) string`
- `matchPanesByFolder(panes []paneEntry, folder string, resolveFunc func(paneEntry) string) ([]string, string)`

The corresponding tests SHALL be moved from `sendkeys_test.go` to `panemap_test.go`:
- `TestMatchPanesByFolder`
- `TestResolvePaneChange`

Functions specific to send-keys (`buildSendKeysArgs`, `validateSendKeysInputs`, `runSendKeys`, `resolveChangePane`) and their tests (`TestBuildSendKeysArgs`, `TestBuildSendKeysArgsWithSpaces`, `TestValidateSendKeysInputs`) SHALL be deleted.

#### Scenario: Moved functions still pass tests
- **GIVEN** `matchPanesByFolder` and `resolvePaneChange` have been moved to `panemap.go`
- **AND** their tests have been moved to `panemap_test.go`
- **WHEN** `go test ./src/go/fab/cmd/fab/...` is executed
- **THEN** all moved tests pass

## CLI: pane-map shows all panes

### Requirement: Include non-fab panes

`resolvePane()` in `panemap.go` SHALL always return `true` (include the pane in output) instead of returning `false` for non-git or non-fab panes. The function SHALL populate row fields with fallback values for panes that lack git or fab context.

#### Scenario: Non-git directory pane
- **GIVEN** a tmux pane `%5` with CWD `/tmp/scratch` (not a git repo)
- **WHEN** `fab pane-map` is executed
- **THEN** the output includes pane `%5`
- **AND** the Worktree column shows `scratch/` (basename of CWD + `/`)
- **AND** the Change column shows `—` (em dash)
- **AND** the Stage column shows `—` (em dash)
- **AND** the Agent column shows `—` (em dash)

#### Scenario: Git repo without fab directory
- **GIVEN** a tmux pane `%8` with CWD inside a git repo at `/home/user/other-project` that has no `fab/` directory
- **WHEN** `fab pane-map` is executed
- **THEN** the output includes pane `%8`
- **AND** the Worktree column shows the computed worktree path (using `worktreeDisplayPath`)
- **AND** the Change column shows `—` (em dash)
- **AND** the Stage column shows `—` (em dash)
- **AND** the Agent column shows `—` (em dash)

#### Scenario: Fab-aware pane (unchanged)
- **GIVEN** a tmux pane with CWD inside a git repo with `fab/` directory and an active change
- **WHEN** `fab pane-map` is executed
- **THEN** the pane shows full change info (worktree path, change name, stage, agent state)

### Requirement: Update empty message

The empty-rows message SHALL change from `"No fab worktrees found in tmux panes."` to `"No tmux panes found."`.

#### Scenario: No panes discovered
- **GIVEN** `discoverPanes()` returns an empty list
- **WHEN** `fab pane-map` is executed
- **THEN** output is `No tmux panes found.`

### Requirement: Non-git Worktree column display

For panes not inside a git repo, the Worktree column SHALL show `filepath.Base(p.cwd) + "/"`. For panes inside a git repo (with or without `fab/`), the existing `worktreeDisplayPath` logic SHALL be used.

#### Scenario: Non-git CWD display
- **GIVEN** a pane with CWD `/home/user/downloads`
- **WHEN** the pane is resolved for display
- **THEN** Worktree shows `downloads/`

#### Scenario: Git repo CWD display (no fab)
- **GIVEN** a pane with CWD inside a git worktree at `/home/user/myrepo.worktrees/alpha` and main root `/home/user/myrepo`
- **WHEN** the pane is resolved for display
- **THEN** Worktree shows `myrepo.worktrees/alpha/` (same as existing logic)

## Documentation: _scripts.md

### Requirement: Remove send-keys, add --pane, update pane-map

`fab/.kit/skills/_scripts.md` SHALL be updated:
- Remove the `fab send-keys` section entirely (current lines ~306-340)
- Remove `fab send-keys` from the command reference table
- Add `--pane` flag to the `fab resolve` flag table with description `Output tmux pane ID`
- Update the `fab pane-map` section to note it shows all tmux panes (not just fab worktrees)
- Update the empty-panes message documentation

#### Scenario: _scripts.md reflects new state
- **GIVEN** the updated `_scripts.md`
- **WHEN** an agent reads the command reference
- **THEN** `fab send-keys` is not mentioned
- **AND** `fab resolve --pane` is documented in the flag table
- **AND** `fab pane-map` documentation states it shows all panes

## Documentation: Operator skills

### Requirement: Update operator1 and operator2 skills

`fab/.kit/skills/fab-operator1.md` and `fab/.kit/skills/fab-operator2.md` SHALL replace all `fab send-keys <change> "<text>"` invocations with the `fab resolve <change> --pane` + raw `tmux send-keys` pattern:

```bash
tmux send-keys -t "$(fab/.kit/bin/fab resolve <change> --pane)" "<text>" Enter
```

Available Tools tables, use case examples, and pre-send validation sections SHALL be updated accordingly. `send-keys` tool entries SHALL be replaced with `resolve --pane` entries.

#### Scenario: Operator skill uses resolve --pane
- **GIVEN** the updated `fab-operator1.md`
- **WHEN** it describes sending keys to a change's pane
- **THEN** the pattern is `tmux send-keys -t "$(fab/.kit/bin/fab resolve <change> --pane)" "<text>" Enter`
- **AND** `fab send-keys` is not referenced

## Documentation: Operator spec files

### Requirement: Update operator spec files

`docs/specs/skills/SPEC-fab-operator1.md` and `docs/specs/skills/SPEC-fab-operator2.md` SHALL update their primitives sections to replace `send-keys` references with `resolve --pane`.

#### Scenario: Spec files reflect new primitives
- **GIVEN** the updated operator spec files
- **WHEN** an agent or human reads the specs
- **THEN** `send-keys` is not listed as a primitive
- **AND** `resolve --pane` is documented as the change→pane lookup primitive

## Deprecated Requirements

### send-keys subcommand

**Reason**: Redundant — `resolve --pane` + `tmux send-keys` achieves the same result while being composable. The user explicitly chose removal over coexistence.
**Migration**: Replace `fab send-keys <change> "<text>"` with `tmux send-keys -t "$(fab resolve <change> --pane)" "<text>" Enter`

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Extend `fab resolve` with `--pane` flag rather than new subcommand | Confirmed from intake #1 — resolve already handles all change→handle lookups | S:95 R:90 A:95 D:95 |
| 2 | Certain | Delete `fab send-keys` entirely | Confirmed from intake #2 — user explicitly chose removal, no duplicate mechanisms | S:95 R:85 A:90 D:95 |
| 3 | Certain | `pane-map` shows all tmux panes, not just fab ones | Confirmed from intake #3 — operator needs visibility into all tabs | S:90 R:90 A:90 D:90 |
| 4 | Certain | Go binary only, ignore Rust binary | Confirmed from intake #4 — Rust binary not maintained | S:95 R:95 A:95 D:95 |
| 5 | Certain | Non-fab panes show em dashes for Change/Stage/Agent columns | Confirmed from intake #5 — consistent with existing pane-map convention | S:85 R:95 A:90 D:90 |
| 6 | Certain | Move `matchPanesByFolder` and `resolvePaneChange` to panemap.go | Confirmed from intake #6 — both depend on panemap.go types | S:85 R:90 A:90 D:85 |
| 7 | Confident | Non-git panes show `filepath.Base(cwd) + "/"` for Worktree | Confirmed from intake #7 — consistent with `worktreeDisplayPath` fallback | S:75 R:90 A:80 D:80 |
| 8 | Confident | Shell scripts (dispatch.sh, run.sh) unaffected | Confirmed from intake #8 — they use raw tmux, not fab send-keys | S:85 R:90 A:85 D:90 |
| 9 | Certain | `--pane` integrates into PreRunE priority chain after `--status` | Codebase shows clear priority pattern in resolve.go PreRunE | S:90 R:95 A:95 D:90 |
| 10 | Certain | Error message format for --pane matches existing resolve patterns | Codebase convention — resolve errors use lowercase with folder in quotes | S:85 R:95 A:90 D:95 |

10 assumptions (8 certain, 2 confident, 0 tentative, 0 unresolved).
