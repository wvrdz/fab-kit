# Tasks: Add Operator2 Monitoring Skill

**Change**: 260311-5c11-add-operator2-monitoring-skill
**Spec**: `spec.md`
**Intake**: `intake.md`

## Phase 1: Setup

- [x] T001 Rename `fab/.kit/scripts/fab-operator.sh` to `fab/.kit/scripts/fab-operator1.sh` — update `TAB_NAME` from `operator1` to `operator`, keep skill invocation as `/fab-operator1`
- [x] T002 Update launcher reference in `fab/.kit/skills/fab-operator1.md` — change `fab-operator.sh` to `fab-operator1.sh` in the Orientation section

## Phase 2: Core Implementation

- [x] T003 Create `fab/.kit/skills/fab-operator2.md` — full operator skill with all sections: Purpose, Arguments, Context Loading, Command Logging, Orientation, State Re-derivation, Monitoring State, Monitoring Tick Behavior, Use Cases (UC1-UC8 with monitoring integration), Confirmation Model, Pre-Send Validation, Bounded Retries and Escalation, Context Discipline, Not a Lifecycle Enforcer, Terminal Output Inspection, Autopilot Behavior, Key Properties
- [x] T004 Create `fab/.kit/scripts/fab-operator2.sh` — singleton launcher with tab name `operator`, invokes `/fab-operator2`

## Phase 3: Integration & Edge Cases

- [x] T005 Create `docs/specs/skills/SPEC-fab-operator2.md` — spec covering summary, primitives, discovery, use cases, monitoring behavior, interaction model, and guardrails

## Execution Order

- T001 and T002 are independent setup tasks
- T003 is the core skill — no dependency on T001/T002 but logically follows
- T004 depends on T003 (launcher references the skill name)
- T005 depends on T003 (spec describes the skill)
