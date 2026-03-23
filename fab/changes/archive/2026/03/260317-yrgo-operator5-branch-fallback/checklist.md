# Quality Checklist: Operator5 — Use Case Registry, Branch Fallback, and Proactive Monitoring

**Change**: 260317-yrgo-operator5-branch-fallback
**Generated**: 2026-03-18
**Spec**: `spec.md`

## Functional Completeness
- [ ] CHK-001 Operator5 skill file: `fab/.kit/skills/fab-operator5.md` exists with all operator4 content plus additions
- [ ] CHK-002 Branch fallback: subsection in Section 3 with `git for-each-ref` on `refs/heads/` + `refs/remotes/`, single/multi/no match, read-only vs action
- [ ] CHK-003 Use case registry: `.fab-operator.yaml` schema documented, conversational toggling, tick-start roster with 🟢/⚪
- [ ] CHK-004 Loop lifecycle: runs while any use case enabled, stops when all disabled
- [ ] CHK-005 `monitor-changes` use case: operator4's monitoring system (6-step tick, auto-nudge, stuck detection) preserved
- [ ] CHK-006 `linear-inbox` use case: MCP detection, deduplication, spawn-on-confirm
- [ ] CHK-007 `pr-freshness` use case: `gh pr list` detection, action matrix (idle/busy/no-agent/dirty), operator routes only
- [ ] CHK-008 Tab preparation procedure: 5-step sequence (pane → idle → switch → branch → dispatch)
- [ ] CHK-009 Playbooks: Section 6 renamed from "Modes of Operation", all 9 patterns listed, references tab preparation
- [ ] CHK-010 Spec file: `docs/specs/skills/SPEC-fab-operator5.md` reflects all additions

## Behavioral Correctness
- [ ] CHK-011 Operator4 unchanged: `fab/.kit/skills/fab-operator4.md` and `docs/specs/skills/SPEC-fab-operator4.md` not modified
- [ ] CHK-012 Branch fallback trigger: user-initiated only, NOT during monitoring ticks
- [ ] CHK-013 Branch fallback scope: searches both local and remote branches
- [ ] CHK-014 PR rebase: operator sends instructions to agents, does NOT run `git rebase` directly

## Removal Verification
- [ ] CHK-015 `fab/.kit/scripts/fab-operator1.sh` deleted
- [ ] CHK-016 `fab/.kit/scripts/fab-operator2.sh` deleted
- [ ] CHK-017 `fab/.kit/scripts/fab-operator3.sh` deleted
- [ ] CHK-018 `fab/.kit/scripts/fab-operator4.sh` NOT deleted

## Scenario Coverage
- [ ] CHK-019 Branch fallback: single match read-only → `git show` for status
- [ ] CHK-020 Branch fallback: single match action → offer worktree creation
- [ ] CHK-021 Branch fallback: multiple matches → disambiguation list
- [ ] CHK-022 Branch fallback: no match → "not found locally or in any branch"
- [ ] CHK-023 PR freshness: stale + idle agent → send rebase
- [ ] CHK-024 PR freshness: stale + no agent → offer spawn
- [ ] CHK-025 Tab preparation: wrong change active → fab-switch first

## Edge Cases & Error Handling
- [ ] CHK-026 Config file missing → created with defaults
- [ ] CHK-027 All use cases disabled → loop stops

## Code Quality
- [ ] CHK-028 Pattern consistency: operator5 follows operator4's section numbering and structural patterns
- [ ] CHK-029 No unnecessary duplication: shared behavior (tab preparation) is referenced, not duplicated across playbooks

## Documentation Accuracy
- [ ] CHK-030 All spec requirements have corresponding sections in operator5 skill file
- [ ] CHK-031 SPEC-fab-operator5.md accurately summarizes operator5's structure and capabilities

## Cross References
- [ ] CHK-032 `_naming.md` conventions referenced for branch name matching and worktree creation
- [ ] CHK-033 `_cli-external.md` referenced for `wt`, `tmux`, `/loop` commands

## Notes

- Check items as you review: `- [x]`
- All items must pass before `/fab-continue` (hydrate)
- If an item is not applicable, mark checked and prefix with **N/A**: `- [x] CHK-008 **N/A**: {reason}`
