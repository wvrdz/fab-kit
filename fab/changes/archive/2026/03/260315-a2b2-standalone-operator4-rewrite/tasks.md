# Tasks: Standalone Operator4 Rewrite

**Change**: 260315-a2b2-standalone-operator4-rewrite
**Spec**: `spec.md`
**Intake**: `intake.md`

## Phase 1: Setup

- [x] T001 [P] Create `fab/.kit/skills/_cli-external.md` — document wt commands (`list`, `list --path`, `create` with all flags, `delete`), tmux commands (`capture-pane`, `send-keys`, `new-window`), and `/loop` usage. Internal `_` file frontmatter (not user-invocable).
- [x] T002 [P] Create `fab/.kit/skills/_naming.md` — document change folder pattern (`{YYMMDD}-{XXXX}-{slug}`), git branch convention (= change folder name), worktree directory pattern (`{adjective}-{noun}`), and operator spawning rules (known change: use folder name as branch; new change: wt auto-generates, operator sends `/git-branch` after `/fab-new`). Internal `_` file frontmatter.
- [x] T003 [P] Rename `fab/.kit/skills/_scripts.md` to `fab/.kit/skills/_cli-fab.md` — `git mv`, content unchanged.

## Phase 2: Core Implementation

- [x] T004 Rewrite `fab/.kit/skills/fab-operator4.md` as standalone — replace the entire file with the 9-section structure (Principles, Startup, Safety Model, Monitoring System, Auto-Nudge, Modes of Operation, Autopilot, Configuration, Key Properties). Inline all behavior from operator1/2/3. Startup section includes `_cli-external.md` load instruction. Autopilot section uses `/fab-fff` (not `/fab-ff`). Monitoring tick includes sending `/git-branch` after detecting new change creation (per spec). Writing style: explain the why, not heavy-handed MUSTs. Target ~300 lines. <!-- clarified: made /fab-fff and /git-branch requirements explicit in task description since they are new spec requirements not inherited from operator1/2/3 -->
- [x] T005 Update `fab/.kit/skills/_preamble.md` — change `_scripts.md` reference to `_cli-fab.md`, add `_naming.md` to the always-load list. Do NOT add `_cli-external.md` to always-load.

## Phase 3: Integration & Cross-References

- [x] T006 [P] Add cross-reference to `fab/.kit/skills/git-branch.md` — add note near the top: `> Branch naming conventions are defined in _naming.md.` No procedural changes.
- [x] T007 [P] Add cross-reference to `fab/.kit/skills/git-pr.md` — add note near the top: `> Branch naming conventions are defined in _naming.md.` No procedural changes.
- [x] T008 [P] Delete `fab/.kit/skills/fab-operator1.md`, `fab/.kit/skills/fab-operator2.md`, `fab/.kit/skills/fab-operator3.md` — `git rm` all three files.
- [x] T009 [P] Delete `docs/specs/skills/SPEC-fab-operator1.md`, `docs/specs/skills/SPEC-fab-operator2.md`, `docs/specs/skills/SPEC-fab-operator3.md` — `git rm` all three files.
- [x] T010 Rewrite `docs/specs/skills/SPEC-fab-operator4.md` — update to reflect standalone structure (remove inheritance references, document 9-section layout).

## Phase 4: Verification

- [x] T011 Run `fab/.kit/scripts/fab-sync.sh` — verify stale files cleaned (`_scripts.md`, operator1/2/3 deployed copies removed), new files deployed (`_cli-external.md`, `_naming.md`, `_cli-fab.md`), operator4 updated.

---

## Execution Order

- T001, T002, T003 are independent (Phase 1 parallel)
- T004 depends on T001, T002, T003 (needs `_cli-external.md` and `_naming.md` to exist, needs `_cli-fab.md` renamed)
- T005 depends on T003 (references the renamed file)
- T006, T007, T008, T009 are independent of each other (Phase 3 parallel)
- T010 depends on T004 (spec reflects the rewritten skill)
- T011 depends on all previous tasks
