# Tasks: Add resolve --pane, drop send-keys, evolve pane-map

**Change**: 260312-kvng-resolve-pane-evolve-panemap
**Spec**: `spec.md`
**Intake**: `intake.md`

## Phase 1: Setup — Move reusable functions

- [x] T001 Move `resolvePaneChange()` from `src/go/fab/cmd/fab/sendkeys.go` to `src/go/fab/cmd/fab/panemap.go` (before `resolvePane` function)
- [x] T002 Move `matchPanesByFolder()` from `src/go/fab/cmd/fab/sendkeys.go` to `src/go/fab/cmd/fab/panemap.go`
- [x] T003 Move `TestMatchPanesByFolder` and `TestResolvePaneChange` from `src/go/fab/cmd/fab/sendkeys_test.go` to `src/go/fab/cmd/fab/panemap_test.go`
- [x] T004 Verify build and tests pass after moves: `cd src/go/fab && go build ./cmd/fab/... && go test ./cmd/fab/...`

## Phase 2: Core Implementation

- [x] T005 Delete `src/go/fab/cmd/fab/sendkeys.go` entirely
- [x] T006 Delete `src/go/fab/cmd/fab/sendkeys_test.go` entirely (remaining send-keys-only tests: `TestBuildSendKeysArgs`, `TestBuildSendKeysArgsWithSpaces`, `TestValidateSendKeysInputs`)
- [x] T007 Remove `sendKeysCmd()` from `root.AddCommand()` list in `src/go/fab/cmd/fab/main.go`
- [x] T008 Add `--pane` flag to `resolveCmd()` in `src/go/fab/cmd/fab/resolve.go`: register `Bool` flag, add `--pane` case in `PreRunE` priority chain (after `--status`, before default `--id`), implement pane resolution in `RunE` switch using `discoverPanes()`, `matchPanesByFolder()`, and `resolvePaneChange()`
- [x] T009 Update `resolvePane()` in `src/go/fab/cmd/fab/panemap.go` to always return `true`: for non-git panes show `filepath.Base(p.cwd) + "/"` as worktree and em dashes for change/stage/agent; for git-without-fab panes show worktree path and em dashes
- [x] T010 Update empty-rows message in `runPaneMap()` from `"No fab worktrees found in tmux panes."` to `"No tmux panes found."`
- [x] T011 Verify build and tests pass after all Go changes: `cd src/go/fab && go build ./cmd/fab/... && go test ./cmd/fab/...`

## Phase 3: Documentation Updates

- [x] T012 [P] Update `fab/.kit/skills/_scripts.md`: remove `fab send-keys` section and command table entry, add `--pane` to `fab resolve` flag table, update `fab pane-map` section to note all-panes behavior and new empty message
- [x] T013 [P] Update `fab/.kit/skills/fab-operator1.md`: replace all `fab send-keys` invocations with `fab resolve --pane` + raw `tmux send-keys` pattern, update Available Tools table, use case examples, pre-send validation
- [x] T014 [P] Update `fab/.kit/skills/fab-operator2.md`: same replacements as T013
- [x] T015 [P] Update `docs/specs/skills/SPEC-fab-operator1.md`: replace send-keys primitives with resolve --pane
- [x] T016 [P] Update `docs/specs/skills/SPEC-fab-operator2.md`: replace send-keys primitives with resolve --pane

---

## Execution Order

- T001-T003 must complete before T004 (verify moves)
- T004 must pass before T005-T007 (delete send-keys files and registration)
- T005-T007 must complete before T008 (resolve --pane can reference moved functions)
- T008-T010 are independent of each other but all depend on T005-T007
- T011 verifies all Go changes
- T012-T016 are parallelizable and independent of each other, but depend on T011 passing
