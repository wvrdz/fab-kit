# Quality Checklist: Add Operator2 Monitoring Skill

**Change**: 260311-5c11-add-operator2-monitoring-skill
**Generated**: 2026-03-11
**Spec**: `spec.md`

## Functional Completeness

- [x] CHK-001 Monitor-After-Action: Operator2 skill enrolls targets in monitoring after every `fab send-keys` invocation
- [x] CHK-002 Monitoring Tick Behavior: Each tick re-queries `fab pane-map` and reports stage advances, completions, failures, pane deaths, and stuck agents
- [x] CHK-003 Terminal State Removal: Monitored changes are removed on hydrate/ship/review-pr, user stop, or pane death
- [x] CHK-004 Loop Lifecycle: Loop starts on first enrollment, stops when monitored set empties, only one loop at a time
- [x] CHK-005 Full Capability Parity: All UC1-UC8 from operator1 are present in operator2
- [x] CHK-006 UC1 Broadcast Auto-Enroll: Broadcasting auto-enrolls all recipients in monitoring
- [x] CHK-007 UC2 Sequenced Rebase via Monitoring: Monitoring tick triggers rebase when trigger change reaches target stage
- [x] CHK-008 UC6 Unstick with Monitoring: Nudged agents are enrolled in monitoring for recovery tracking
- [x] CHK-009 UC7 Notification via Monitoring: "Tell me when X finishes" enrolls in monitoring loop automatically
- [x] CHK-010 Configurable Monitoring Interval: Default 5m, user-overridable via natural language
- [x] CHK-011 Configurable Stuck Threshold: Default 15m, user-overridable
- [x] CHK-012 New Launcher fab-operator2.sh: Singleton tab `operator`, invokes `/fab-operator2`
- [x] CHK-013 Rename Launcher: `fab-operator.sh` renamed to `fab-operator1.sh` with tab name `operator`
- [x] CHK-014 Operator1 Launcher Reference: `fab-operator1.md` references `fab-operator1.sh`
- [x] CHK-015 Spec File: `docs/specs/skills/SPEC-fab-operator2.md` exists with required sections

## Behavioral Correctness

- [x] CHK-016 Confirmation Model: Three-tier model (read-only/recoverable/destructive) matches operator1
- [x] CHK-017 Pre-Send Validation: Pane existence and idle state checked before every send
- [x] CHK-018 State Re-derivation: `fab pane-map` re-queried before every action, no stale state
- [x] CHK-019 Context Discipline: Operator2 never loads change artifacts (intakes, specs, tasks)
- [x] CHK-020 Bounded Retries: Same retry limits and escalation as operator1

## Scenario Coverage

- [x] CHK-021 First send starts monitoring loop
- [x] CHK-022 Subsequent sends extend monitored set without creating new loop
- [x] CHK-023 Stage advance reported correctly on tick
- [x] CHK-024 Pane death detected and change removed from monitoring
- [x] CHK-025 Stuck agent flagged after threshold (advisory, no auto-nudge)
- [x] CHK-026 All changes complete stops loop with summary message
- [x] CHK-027 Tab already exists — launcher switches instead of creating

## Edge Cases & Error Handling

- [x] CHK-028 Send to busy agent warns and requires confirmation
- [x] CHK-029 Pane disappeared between actions — reports gracefully
- [x] CHK-030 Outside tmux — status-only mode with warning

## Code Quality

- [x] CHK-031 Pattern consistency: Operator2 skill structure matches operator1 conventions (sections, formatting, key properties table)
- [x] CHK-032 No unnecessary duplication: Shared behavior described by reference to operator1 patterns, not copy-pasted prose

## Documentation Accuracy

- [x] CHK-033 Spec accurately describes all monitoring behavior and inherited UCs
- [x] CHK-034 Launcher scripts have correct metadata comments

## Cross References

- [x] CHK-035 Operator1 skill references updated launcher name
- [x] CHK-036 Operator2 spec references correct primitives (`fab pane-map`, `fab send-keys`, `/loop`)

## Notes

- Check items as you review: `- [x]`
- All items must pass before `/fab-continue` (hydrate)
- If an item is not applicable, mark checked and prefix with **N/A**: `- [x] CHK-008 **N/A**: {reason}`
