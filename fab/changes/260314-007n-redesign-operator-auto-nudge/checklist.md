# Quality Checklist: Create fab-operator4 — Redesigned Auto-Nudge Operator

**Change**: 260314-007n-redesign-operator-auto-nudge
**Generated**: 2026-03-14
**Spec**: `spec.md`

## Functional Completeness
- [x] CHK-001 Routing discipline: Operator4 skill file contains a routing discipline section that prohibits direct execution of user instructions
- [x] CHK-002 Autopilot override: Operator4 uses `/fab-fff` (not `/fab-ff`) for autopilot pipeline invocations
- [x] CHK-003 Simplified answer model: Decision list items 1-6 are present with correct priority ordering, two-tier classification removed
- [x] CHK-004 Capture window: Terminal capture uses `-l 20` throughout the skill
- [x] CHK-005 Claude turn boundary guard: `>` cursor detection in last 2 lines skips question detection
- [x] CHK-006 Tightened `?` pattern: Matches last non-empty line only, <120 chars, skips `#`, `//`, `*`, `>`, and timestamp prefixes
- [x] CHK-007 New indicator patterns: `:` endings, enumerated options (`[1-9]\)`), `Press.*key`/`press.*enter`/`hit.*enter` patterns present
- [x] CHK-008 Blank capture guard: Blank/whitespace output skips detection
- [x] CHK-009 Re-capture before send: Terminal re-captured before sending auto-answer, abort on change
- [x] CHK-010 Per-answer logging: Inline reporting format specified for both auto-answers and escalations
- [x] CHK-011 Spec file: `docs/specs/skills/SPEC-fab-operator4.md` exists with all required sections per spec
- [x] CHK-012 Launcher script: `fab/.kit/scripts/fab-operator4.sh` exists with singleton tab pattern

## Behavioral Correctness
- [x] CHK-013 Inheritance chain: Skill file directs reader to read `_preamble.md`, then `fab-operator2.md`, then `fab-operator3.md` before operator4's content
- [x] CHK-014 No escalation tier: Two-tier auto-answer/escalate classification from operator3 is replaced by all-auto-answer model
- [x] CHK-015 Monitoring tick delta-only: Describes only changes from operator3's tick, not the full inherited step listing

## Scenario Coverage
- [x] CHK-016 Binary confirmation: Scenario for binary yes/no prompt auto-answer with `y`
- [x] CHK-017 Numbered menu: Scenario for multi-choice prompt auto-answer with `1`
- [x] CHK-018 Undeterminable question: Scenario for escalation when operator can't determine keystrokes
- [x] CHK-019 Claude boundary guard: Scenarios for both false positive prevention and genuine question passthrough
- [x] CHK-020 Race condition: Scenarios for re-capture abort (output changed) and proceed (output unchanged)
- [x] CHK-021 Singleton enforcement: Scenario for launcher script reusing existing tab

## Edge Cases & Error Handling
- [x] CHK-022 Question mark in comments/logs: `?` pattern correctly skips lines with `#`, `//`, `*`, `>`, or timestamp prefixes
- [x] CHK-023 Blank terminal: Blank capture guard prevents false detection
- [x] CHK-024 User instruction without target: Operator asks for disambiguation, doesn't execute directly

## Code Quality
- [x] CHK-025 Pattern consistency: Skill file follows operator3's structure and naming patterns (frontmatter, preamble directive, section organization)
- [x] CHK-026 No unnecessary duplication: Operator4 references inherited behavior via preamble directive rather than repeating it

## Documentation Accuracy
- [x] CHK-027 Spec file accuracy: SPEC-fab-operator4.md accurately reflects all behavior defined in the skill file
- [x] CHK-028 Primitives table: Spec primitives table uses `-l 20` flags (not `-l 10`)

## Cross References
- [x] CHK-029 Operator3 unchanged: No modifications to `fab/.kit/skills/fab-operator3.md`, `docs/specs/skills/SPEC-fab-operator3.md`, or `fab/.kit/scripts/fab-operator3.sh`
- [x] CHK-030 Inheritance correctness: Operator4 correctly extends operator3 (which extends operator2) — all inherited behavior preserved

## Notes

- Check items as you review: `- [x]`
- All items must pass before `/fab-continue` (hydrate)
- If an item is not applicable, mark checked and prefix with **N/A**: `- [x] CHK-008 **N/A**: {reason}`
