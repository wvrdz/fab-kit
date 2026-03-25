# Tasks: fab-proceed Orchestrator

**Change**: 260325-kxw7-fab-proceed-orchestrator
**Spec**: `spec.md`
**Intake**: `intake.md`

## Phase 1: Setup

- [x] T001 Create skill source file `fab/.kit/skills/fab-proceed.md` with frontmatter (name, description) and the `# /fab-proceed` header. Include the preamble reference line.

## Phase 2: Core Implementation

- [x] T002 Write the State Detection section in `fab/.kit/skills/fab-proceed.md` — document the 4-step detection pipeline (active change via `fab resolve`, branch check via `git branch --show-current`, unactivated intake scan of `fab/changes/`, conversation context evaluation). Include the dispatch table mapping detected state → steps to run.
- [x] T003 Write the Dispatch Behavior section — document subagent dispatch for prefix steps (fab-new, fab-switch, git-branch) per `_preamble.md` § Subagent Dispatch. Document `/fab-fff` terminal delegation via Skill tool (not subagent). Include standard subagent context requirements.
- [x] T004 Write the Error Handling section — document the "nothing to proceed with" error for empty context, and error propagation from sub-skill failures (fab-new, fab-switch, git-branch, fab-fff).
- [x] T005 Write the Output section — document the progress reporting format: detecting state header, per-step reports (created intake, activated, branch), handoff line, then fab-fff takes over.
- [x] T006 Write the Key Properties table and skill metadata — arguments (none), flags (none), requires active change (no), runs preflight (no), read-only (no), idempotent (yes), advances stage (no directly), outputs Next line (inherits from fab-fff).

## Phase 3: Integration

- [x] T007 Update `fab/.kit/skills/_preamble.md` state table — add `/fab-proceed` as available command in `initialized` state (after `/fab-new`) and `intake` state (after `/fab-fff`).
- [x] T008 [P] Run `fab/.kit/scripts/fab-sync.sh` to deploy the new skill to `.claude/skills/fab-proceed/`.
- [x] T009 [P] Update `docs/specs/skills.md` — add `/fab-proceed` entry to the skills spec table with description.

## Phase 4: Polish

- [x] T010 Create per-skill flow diagram at `docs/specs/skills/SPEC-fab-proceed.md` following existing patterns in `docs/specs/skills/`.

---

## Execution Order

- T001 blocks T002-T006 (need the file to exist before writing sections)
- T002-T006 are sequential (sections build on each other within the skill file)
- T007 is independent of T002-T006 (edits a different file)
- T008 requires T001-T006 complete (need final skill file for sync)
- T009, T010 are independent documentation tasks
