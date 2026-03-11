# Tasks: Operator Autopilot UC8

**Change**: 260310-1ttn-operator-autopilot-uc8
**Spec**: `spec.md`
**Intake**: `intake.md`

## Phase 1: Structural Updates

- [x] T001 Rename `## Seven Use Cases` heading to `## Use Cases` in `fab/.kit/skills/fab-operator1.md`
- [x] T002 Add UC8 stub after UC7 (Notification surface) in `fab/.kit/skills/fab-operator1.md` — accepts list of changes, references ordering strategies, notes destructive confirmation, delegates to Autopilot Behavior section
- [x] T003 Add autopilot entry to the Confirmation Model table in `fab/.kit/skills/fab-operator1.md` — Destructive risk, "Autopilot (merge after each success)", confirm full queue at start

## Phase 2: Core Implementation

- [x] T004 Add new `## Autopilot Behavior` top-level section in `fab/.kit/skills/fab-operator1.md` after the existing `## Terminal Output Inspection` section — include ordering strategies subsection with all three strategies (user-provided, confidence-based, hybrid)
- [x] T005 Add the per-change autopilot loop to the Autopilot Behavior section — 8-step sequence (spawn, open tab, gate check, monitor, merge, rebase next, cleanup, progress)
- [x] T006 Add failure matrix to the Autopilot Behavior section — table with all 6 failure types and their actions/resume behavior
- [x] T007 Add interruptibility subsection — 4 interrupt commands (stop after current, skip, pause, resume) with immediate acknowledgment requirement
- [x] T008 Add resumability subsection — state reconstruction from `fab pane-map`, resume from first non-completed change
- [x] T009 Add progress reporting subsection — per-change one-line status and final summary

## Phase 3: Spec Alignment

- [x] T010 Renumber UC7 → UC8 and UC8 → UC7 in `docs/specs/skills/SPEC-fab-operator1.md` to match skill numbering (UC7 = Notification surface, UC8 = Autopilot)
- [x] T011 Update any internal cross-references within `docs/specs/skills/SPEC-fab-operator1.md` that reference the old UC numbers

## Phase 4: Sync

- [x] T012 Run `fab/.kit/bin/fab hook sync` to regenerate deployed copies in `.claude/skills/`

---

## Execution Order

- T001 before T002 (heading rename before adding UC8 under new heading)
- T002 before T004 (UC8 stub references Autopilot Behavior section, which must exist)
- T004 before T005, T006, T007, T008, T009 (section header before subsections)
- T005 through T009 are sequential within the section (logical ordering)
- T010, T011 are independent of T001-T009 (different file)
- T012 depends on all prior tasks
