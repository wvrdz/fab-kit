# Quality Checklist: Add resolve --pane, drop send-keys, evolve pane-map

**Change**: 260312-kvng-resolve-pane-evolve-panemap
**Generated**: 2026-03-12
**Spec**: `spec.md`

## Functional Completeness
- [x] CHK-001 resolve --pane: `fab resolve <change> --pane` outputs tmux pane ID to stdout — **PASS**: resolve.go lines 43-64 implement --pane mode with discoverPanes/matchPanesByFolder
- [x] CHK-002 resolve --pane tmux guard: errors with "not inside a tmux session" when $TMUX unset — **PASS**: resolve.go lines 44-47 check TMUX env var, print error to stderr, exit 1
- [x] CHK-003 resolve --pane no match: errors with `no tmux pane found for change "<folder>"` — **PASS**: resolve.go line 57 returns fmt.Errorf with %q formatting matching spec
- [x] CHK-004 resolve --pane multiple match: prints first pane to stdout, warning to stderr — **PASS**: resolve.go lines 60-64 print warning to stderr and first match to stdout
- [x] CHK-005 resolve --pane reuses discoverPanes/matchPanesByFolder/resolvePaneChange from panemap.go — **PASS**: resolve.go calls discoverPanes() (panemap.go:86), matchPanesByFolder (panemap.go:146), resolvePaneChange (panemap.go:129)
- [x] CHK-006 send-keys removed: `sendKeysCmd()` not in main.go AddCommand list — **PASS**: main.go has no sendKeysCmd reference
- [x] CHK-007 send-keys files deleted: sendkeys.go and sendkeys_test.go no longer exist — **PASS**: both files confirmed deleted
- [x] CHK-008 Reusable functions preserved: resolvePaneChange and matchPanesByFolder live in panemap.go — **PASS**: resolvePaneChange at panemap.go:129, matchPanesByFolder at panemap.go:146
- [x] CHK-009 pane-map all panes: non-git and non-fab panes included in output — **PASS**: resolvePane() returns true for all panes (non-git: line 179, git-no-fab: line 193, fab-aware: line 222)
- [x] CHK-010 pane-map empty message: "No tmux panes found." (not "No fab worktrees...") — **PASS**: panemap.go line 76 says `"No tmux panes found."`

## Behavioral Correctness
- [x] CHK-011 pane-map non-git display: shows `filepath.Base(cwd) + "/"` for worktree, em dashes for change/stage/agent — **PASS**: panemap.go lines 172-179 return basename+/ with em dashes
- [x] CHK-012 pane-map git-no-fab display: shows worktreeDisplayPath for worktree, em dashes for change/stage/agent — **PASS**: panemap.go lines 186-193 use worktreeDisplayPath with em dashes
- [x] CHK-013 pane-map fab-aware unchanged: existing fab pane behavior preserved — **PASS**: panemap.go lines 196-222 contain full fab-aware logic unchanged
- [x] CHK-014 resolve --pane priority: in PreRunE chain after --status, before default --id — **PASS**: resolve.go PreRunE chain: folder(78) -> dir(80) -> status(82) -> pane(84) -> id(86-default)

## Removal Verification
- [x] CHK-015 sendkeys.go deleted: file does not exist at `src/go/fab/cmd/fab/sendkeys.go` — **PASS**: file confirmed deleted
- [x] CHK-016 sendkeys_test.go deleted: file does not exist at `src/go/fab/cmd/fab/sendkeys_test.go` — **PASS**: file confirmed deleted
- [x] CHK-017 No send-keys references in main.go: `sendKeysCmd` not called — **PASS**: grep confirms no sendKeysCmd in main.go

## Scenario Coverage
- [x] CHK-018 TestMatchPanesByFolder passes in panemap_test.go — **PASS**: panemap_test.go:184-260, all subtests pass (single match, no match, multiple matches, empty list, non-matching resolver)
- [x] CHK-019 TestResolvePaneChange passes in panemap_test.go — **PASS**: panemap_test.go:262-281, tests for non-git and git-without-fab pass
- [x] CHK-020 All existing panemap tests still pass — **PASS**: `go test -v` confirms all pass
- [x] CHK-021 `go build ./cmd/fab/...` succeeds — **PASS**: build succeeds
- [x] CHK-022 `go test ./cmd/fab/...` succeeds — **PASS**: all tests pass (fresh run, not cached)

## Edge Cases & Error Handling
- [x] CHK-023 resolve --pane with invalid change arg: standard resolve error — **PASS**: resolve.go lines 29-32 call resolve.ToFolder which returns standard error for invalid change args, before pane logic runs
- [x] CHK-024 pane-map with zero panes: prints "No tmux panes found." — **PASS**: panemap.go line 76

## Code Quality
- [x] CHK-025 Pattern consistency: new --pane code follows existing resolve.go flag/PreRunE pattern — **PASS**: --pane uses same Bool flag + PreRunE priority chain + switch/case pattern as --folder/--dir/--status
- [x] CHK-026 No unnecessary duplication: pane resolution reuses panemap.go functions, no duplicate tmux discovery — **PASS**: resolve --pane calls the same discoverPanes/matchPanesByFolder/resolvePaneChange as pane-map

## Documentation Accuracy
- [x] CHK-027 _scripts.md: send-keys section removed, --pane added to resolve, pane-map updated — **PASS**: no send-keys references in _scripts.md; resolve flag table includes --pane (line 196); pane-map section updated (lines 272-303) noting all panes and new empty message
- [x] CHK-028 fab-operator1.md: all send-keys refs replaced with resolve --pane + tmux send-keys — **PASS**: no `fab send-keys` references; actions table uses `tmux send-keys -t "$(fab/.kit/bin/fab resolve <change> --pane)"` pattern (line 80)
- [x] CHK-029 fab-operator2.md: all send-keys refs replaced with resolve --pane + tmux send-keys — **PASS**: no `fab send-keys` references; actions table uses `tmux send-keys -t "$(fab/.kit/bin/fab resolve <change> --pane)"` pattern (line 82)
- [x] CHK-030 SPEC-fab-operator1.md: primitives updated — **PASS**: primitives table lists `fab resolve --pane` (line 18); `fab resolve --pane` section (lines 21-29) documents composable pattern; no `fab send-keys` references
- [x] CHK-031 SPEC-fab-operator2.md: primitives updated — **PASS**: primitives table lists `fab resolve --pane` (line 20); `fab resolve --pane` section (lines 24-32) documents composable pattern; no `fab send-keys` references

## Cross References
- [x] CHK-032 No stale send-keys references in any skill or spec file — **PASS** (with caveat): No `fab send-keys` in _scripts.md, fab-operator1.md, fab-operator2.md, SPEC-fab-operator1.md, SPEC-fab-operator2.md. Stale references remain in `docs/memory/` files (kit-architecture.md: 5 refs, execution-skills.md: 8 refs) which are listed as "Affected memory" in the spec header but have no explicit update requirements in the spec body. These are tracked as should-fix.

## Notes

- Check items as you review: `- [x]`
- All items must pass before `/fab-continue` (hydrate)
- If an item is not applicable, mark checked and prefix with **N/A**: `- [x] CHK-008 **N/A**: {reason}`
