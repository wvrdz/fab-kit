# Tasks: Operator5 — Use Case Registry, Branch Fallback, and Proactive Monitoring

**Change**: 260317-yrgo-operator5-branch-fallback
**Spec**: `spec.md`
**Intake**: `intake.md`

## Phase 1: Setup

- [x] T001 Copy `fab/.kit/skills/fab-operator4.md` to `fab/.kit/skills/fab-operator5.md` — update frontmatter (name: `fab-operator5`, description updated to mention use case registry, branch fallback, and proactive monitoring)
- [x] T002 [P] Copy `docs/specs/skills/SPEC-fab-operator4.md` to `docs/specs/skills/SPEC-fab-operator5.md` — update title and summary to reflect operator5

## Phase 2: Core Implementation

- [x] T003 In `fab/.kit/skills/fab-operator5.md` Section 3 (Safety Model): add "Branch Fallback Resolution" subsection after "Pre-Send Validation" and before "Bounded Retries & Escalation". Content: trigger (user-initiated only), resolution flow (`git for-each-ref` on `refs/heads/` + `refs/remotes/`), single/multi/no match handling, read-only vs action response, worktree creation via `wt create`
- [x] T004 In `fab/.kit/skills/fab-operator5.md`: replace Section 4 (Monitoring System) with "Use Case Registry" section. Content: `.fab-operator.yaml` config schema, conversational toggling, tick-start status roster format (🟢/⚪), loop lifecycle (runs while any use case enabled). Retain monitored set, enrollment/removal triggers, `/loop` lifecycle under the `monitor-changes` use case subsection
- [x] T005 In `fab/.kit/skills/fab-operator5.md`: add `monitor-changes` use case subsection within the Use Case Registry section. Content: operator4's existing monitoring tick (6 steps), auto-nudge (question detection + answer model), stuck detection — all preserved identically, just housed under the registry
- [x] T006 In `fab/.kit/skills/fab-operator5.md`: add `linear-inbox` use case subsection. Content: detection via MCP `list_issues`, deduplication against active/archived changes' `.status.yaml` `issues` arrays, action on new issue (report + spawn on confirm), config schema (`assignee`)
- [x] T007 In `fab/.kit/skills/fab-operator5.md`: add `pr-freshness` use case subsection. Content: detection via `gh pr list --author @me --state open --json number,headRefName,mergeStateStatus`, action matrix (idle agent → send rebase, busy → skip, no agent → offer spawn, DIRTY → flag conflicts), operator does NOT rebase directly
- [x] T008 In `fab/.kit/skills/fab-operator5.md`: add "Tab Preparation Procedure" section. Content: 5-step procedure (verify pane, check idle, check active change → fab-switch, check branch → git-branch, dispatch). Reference from playbooks and use cases that send commands
- [x] T009 In `fab/.kit/skills/fab-operator5.md` Section 6 (Modes of Operation): rename to "Playbooks". Update section title and introductory text. Add note that all playbooks use Tab Preparation Procedure before dispatching. Content of individual playbooks unchanged

## Phase 3: Integration & Edge Cases

- [x] T010 Update `docs/specs/skills/SPEC-fab-operator5.md` to reflect all operator5 additions: use case registry, branch fallback, 3 use cases, tab preparation, playbooks rename, resolved design decisions
- [x] T011 In `fab/.kit/skills/fab-operator5.md` Section 2 (Startup): update orientation to mention use case registry — on startup, read `.fab-operator.yaml` and display roster. Update "outside tmux" degradation note if needed
- [x] T012 In `fab/.kit/skills/fab-operator5.md` Section 8 (Configuration): add `.fab-operator.yaml` to the configuration table with default values for each use case
- [x] T013 In `fab/.kit/skills/fab-operator5.md` Section 9 (Key Properties): update properties table — add "Uses `.fab-operator.yaml`?" = Yes, update description to reflect use case registry

## Phase 4: Cleanup

- [x] T014 [P] Delete `fab/.kit/scripts/fab-operator1.sh`
- [x] T015 [P] Delete `fab/.kit/scripts/fab-operator2.sh`
- [x] T016 [P] Delete `fab/.kit/scripts/fab-operator3.sh`

---

## Execution Order

- T001 and T002 are independent (parallel)
- T003 through T009 depend on T001 (all modify operator5.md)
- T003-T009 are sequential (each modifies the same file, ordered by section number)
- T010 depends on T003-T009 (spec reflects final skill content)
- T011-T013 depend on T004 (reference use case registry structure)
- T014-T016 are independent of all other tasks (parallel, can run anytime)
