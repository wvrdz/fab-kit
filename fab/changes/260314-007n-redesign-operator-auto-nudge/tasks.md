# Tasks: Create fab-operator4 — Redesigned Auto-Nudge Operator

**Change**: 260314-007n-redesign-operator-auto-nudge
**Spec**: `spec.md`
**Intake**: `intake.md`

## Phase 1: Setup

- [x] T001 Create launcher script `fab/.kit/scripts/fab-operator4.sh` following the singleton tab pattern from `fab/.kit/scripts/fab-operator3.sh` — same structure, invoking `/fab-operator4` instead of `/fab-operator3`

## Phase 2: Core Implementation

- [x] T002 Create the operator4 skill file `fab/.kit/skills/fab-operator4.md` with frontmatter, preamble directive (read `_preamble.md` then `fab-operator2.md` then `fab-operator3.md`), and condensed Purpose section (drop "Key difference" paragraph, inline launcher info)
- [x] T003 Add Routing Discipline section to `fab/.kit/skills/fab-operator4.md` — operator MUST NOT execute user instructions directly, route via `tmux send-keys`, enroll in monitoring
- [x] T004 Add Autopilot Pipeline Override section to `fab/.kit/skills/fab-operator4.md` — use `/fab-fff` instead of `/fab-ff` for autopilot gate checks
- [x] T005 Add Simplified Answer Model section to `fab/.kit/skills/fab-operator4.md` — decision list (items 1-6) replacing operator3's two-tier classification, no cooldown or retry limit
- [x] T006 Add Question Detection Improvements section to `fab/.kit/skills/fab-operator4.md` — capture window `-l 20`, Claude turn boundary guard, tightened `?` pattern (last non-empty line, <120 chars, skip prefixes), new indicator patterns (`:` endings, enumerated options, `Press.*key`), blank capture guard. Integrate idle-only guard and bottom-most indicator rule inline
- [x] T007 Add Re-Capture Before Send section to `fab/.kit/skills/fab-operator4.md` — re-capture terminal before sending auto-answer, abort if output changed
- [x] T008 Add Per-Answer Logging section to `fab/.kit/skills/fab-operator4.md` — inline reporting format for auto-answers and escalations
- [x] T009 Add Updated Monitoring Tick section to `fab/.kit/skills/fab-operator4.md` — delta-only description of changes from operator3's tick (simplified answer model, detection improvements)
- [x] T010 Add Key Properties table to `fab/.kit/skills/fab-operator4.md` — only ~4 rows covering novel/modified properties

## Phase 3: Integration & Edge Cases

- [x] T011 [P] Create spec file `docs/specs/skills/SPEC-fab-operator4.md` following `SPEC-fab-operator3.md` structure — summary, primitives table (with `-l 20`), routing discipline, question detection (guards as named subsections), answer model, re-capture, logging, monitoring tick changes, relationship to operator3 table, launcher, one-operator-at-a-time, key properties, resolved design decisions
- [x] T012 [P] Sync deployed skill copy by running `fab/.kit/scripts/fab-sync.sh` to update `.claude/skills/fab-operator4.md` from the source at `fab/.kit/skills/fab-operator4.md`

---

## Execution Order

- T001 is independent (setup)
- T002 blocks T003–T010 (skill file must exist before adding sections)
- T003–T010 are sequential within the file (each builds on previous content)
- T011 and T012 are parallel and independent of each other, but both depend on T010 (skill file complete)
