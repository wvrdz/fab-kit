# Quality Checklist: Operator Autopilot UC8

**Change**: 260310-1ttn-operator-autopilot-uc8
**Generated**: 2026-03-11
**Spec**: `spec.md`

## Functional Completeness

- [ ] CHK-001 Heading rename: `## Seven Use Cases` replaced with `## Use Cases` in skill file
- [ ] CHK-002 UC8 stub: UC8 entry exists after UC7 in skill file, references Autopilot Behavior section
- [ ] CHK-003 Confirmation Model: Autopilot row added to Confirmation Model table as Destructive
- [ ] CHK-004 Ordering strategies: All three strategies (user-provided, confidence-based, hybrid) documented in Autopilot Behavior section
- [ ] CHK-005 Autopilot loop: 8-step per-change sequence (spawn → progress) fully described
- [ ] CHK-006 Failure matrix: All 6 failure types with actions and resume behavior documented
- [ ] CHK-007 Interruptibility: All 4 interrupt commands documented with immediate acknowledgment requirement
- [ ] CHK-008 Resumability: State reconstruction from `fab pane-map` documented
- [ ] CHK-009 Progress reporting: Per-change status and final summary documented
- [ ] CHK-010 Spec renumbering: UC7 ↔ UC8 swapped in `docs/specs/skills/SPEC-fab-operator1.md`

## Behavioral Correctness

- [ ] CHK-011 UC8 stub delegates to Autopilot Behavior section (not inlined)
- [ ] CHK-012 Confirmation model autopilot entry specifies full queue confirmation at start
- [ ] CHK-013 Spawn pattern matches validated pattern: `wt create` + `tmux new-window` + `claude --dangerously-skip-permissions`
- [ ] CHK-014 Gate check uses `fab status show` for confidence, not `fab score`
- [ ] CHK-015 Merge runs from operator shell (not via send-keys to agent)
- [ ] CHK-016 Rebase conflict handling: flag to user, never auto-resolve

## Scenario Coverage

- [ ] CHK-017 Successful change through autopilot scenario: spawn → ff → merge → rebase next
- [ ] CHK-018 Below-gate scenario: flag with score and threshold, wait for user input
- [ ] CHK-019 Review failure scenario: flag and skip after rework budget exhaustion
- [ ] CHK-020 Rebase conflict scenario: flag, never auto-resolve, skip to next
- [ ] CHK-021 Pane death scenario: 1 respawn attempt, then flag and skip
- [ ] CHK-022 Stop/skip/pause/resume interrupt scenarios covered

## Edge Cases & Error Handling

- [ ] CHK-023 Stage timeout (>30 min): flags regardless of agent active/idle state
- [ ] CHK-024 Total timeout (>2 hr): flags for user review
- [ ] CHK-025 Operator session restart: reconstructs from pane-map, resumes correctly
- [ ] CHK-026 "all idle" as change list input: resolved correctly

## Code Quality

- [ ] CHK-027 Pattern consistency: New sections follow existing skill file structure (heading levels, formatting, table style)
- [ ] CHK-028 No unnecessary duplication: Reuses existing primitives (pane-map, send-keys, runtime) without re-documenting their behavior
- [ ] CHK-029 Readability: Autopilot Behavior section is scannable with clear subsections

## Documentation Accuracy

- [ ] CHK-030 Skill file references correct CLI commands with correct argument syntax
- [ ] CHK-031 Spec cross-references updated after renumbering (no stale UC7/UC8 references)

## Cross References

- [ ] CHK-032 Spec file (`SPEC-fab-operator1.md`) aligns with skill file numbering
- [ ] CHK-033 Deployed copy (`.claude/skills/fab-operator1.md`) matches source after sync
