# Quality Checklist: Standalone Operator4 Rewrite

**Change**: 260315-a2b2-standalone-operator4-rewrite
**Generated**: 2026-03-15
**Spec**: `spec.md`

## Functional Completeness
- [x] CHK-001 Self-contained skill: operator4 has no inheritance directive referencing operator1/2/3
- [x] CHK-002 Section structure: operator4 has all 9 sections in order (Principles, Startup, Safety, Monitoring, Auto-Nudge, Modes, Autopilot, Config, Key Properties)
- [x] CHK-003 Monitoring tick: all 6 steps fully specified inline (stage advance, pipeline completion, review failure, pane death, auto-nudge, stuck detection)
- [x] CHK-004 Auto-nudge: decision list (items 1-6), all guards, capture window, pattern matching, re-capture guard, logging format all present
- [x] CHK-005 Operator-only loading: operator4 startup loads `_cli-external.md`; `_preamble.md` does NOT list `_cli-external.md`
- [x] CHK-006 Autopilot pipeline: operator4 references `/fab-fff` (not `/fab-ff`)
- [x] CHK-007 `/git-branch` after new change: monitoring tick includes sending `/git-branch` after detecting new change from backlog
- [x] CHK-008 `_cli-external.md` exists with wt, tmux, and `/loop` documentation
- [x] CHK-009 `_naming.md` exists with change folder, branch, worktree, and operator spawning conventions
- [x] CHK-010 `_cli-fab.md` exists (renamed from `_scripts.md`) with unchanged content
- [x] CHK-011 `_preamble.md` references `_cli-fab.md` (not `_scripts.md`) and includes `_naming.md` in always-load
- [x] CHK-012 `git-branch.md` and `git-pr.md` have `_naming.md` cross-reference
- [x] CHK-013 Operator1/2/3 source files deleted from `fab/.kit/skills/`
- [x] CHK-014 Operator1/2/3 spec files deleted from `docs/specs/skills/`
- [x] CHK-015 `SPEC-fab-operator4.md` updated to reflect standalone structure

## Behavioral Correctness
- [x] CHK-016 Operator4 behavior preserved: auto-nudge, monitoring, autopilot logic unchanged from flattened operator1→2→3→4 chain
- [x] CHK-017 Writing style: constraints explain "why" rather than using heavy-handed MUSTs (except safety constraints)

## Scenario Coverage
- [x] CHK-018 Agent loads operator4: complete behavior from single file + `_` files
- [x] CHK-019 Pipeline skill does not load `_cli-external.md`: only operator4 loads it
- [x] CHK-020 Sync handles rename: `_scripts.md` stale copy removed, `_cli-fab.md` deployed
- [x] CHK-021 Sync handles deletions: operator1/2/3 stale copies removed from all agent directories

## Code Quality
- [x] CHK-022 Pattern consistency: new `_` files follow existing `_` file conventions (frontmatter, structure)
- [x] CHK-023 No unnecessary duplication: operator4 does not duplicate tool tables from `_cli-fab.md` or `_cli-external.md`

## Documentation Accuracy
- [x] CHK-024 `_naming.md` conventions match `docs/specs/naming.md` (authoritative spec)
- [x] CHK-025 `_cli-external.md` wt flags match actual `wt` binary behavior

## Cross References
- [x] CHK-026 Spec files index (`docs/specs/skills/`) consistent with actual files present

## Notes

- Check items as you review: `- [x]`
- All items must pass before `/fab-continue` (hydrate)
- If an item is not applicable, mark checked and prefix with **N/A**: `- [x] CHK-008 **N/A**: {reason}`
