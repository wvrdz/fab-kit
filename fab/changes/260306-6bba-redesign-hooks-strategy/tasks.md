# Tasks: Redesign Hooks Strategy

**Change**: 260306-6bba-redesign-hooks-strategy
**Spec**: `spec.md`
**Intake**: `intake.md`

## Phase 1: Setup

- [x] T001 Create Go `runtime` command scaffold at `src/fab-go/cmd/fab/runtime.go` — add `runtimeCmd()` with `set-idle` and `clear-idle` subcommands, register in `src/fab-go/cmd/fab/main.go`

## Phase 2: Core Implementation

- [x] T002 Implement `fab runtime set-idle` in `src/fab-go/internal/runtime/runtime.go` — resolve change via `resolve` package, read/create `.fab-runtime.yaml`, write `{folder}.agent.idle_since` timestamp using `gopkg.in/yaml.v3`
- [x] T003 Implement `fab runtime clear-idle` in `src/fab-go/internal/runtime/runtime.go` — resolve change, read `.fab-runtime.yaml`, delete `{folder}.agent` block, handle missing file as no-op
- [x] T004 Add tests for runtime package at `src/fab-go/internal/runtime/runtime_test.go` — cover set-idle (create file, update existing), clear-idle (existing entry, no entry, missing file), invalid change reference
- [x] T005 Create `fab/.kit/hooks/on-artifact-write.sh` — read stdin JSON, extract `file_path`, pattern-match against `fab/changes/*/intake.md|spec.md|tasks.md|checklist.md`, derive change name, dispatch per-artifact bookkeeping, return `additionalContext` JSON, exit 0 always
- [x] T006 [P] Update `fab/.kit/hooks/on-stop.sh` — replace yq-based idle timestamp write with `"$fab_cmd" runtime set-idle "$change_folder" 2>/dev/null || true`, remove `command -v yq` guard and yq invocations
- [x] T007 [P] Update `fab/.kit/hooks/on-session-start.sh` — replace yq-based agent block deletion with `"$fab_cmd" runtime clear-idle "$change_folder" 2>/dev/null || true`, remove `command -v yq` guard and yq invocations
- [x] T008 Update `fab/.kit/sync/5-sync-hooks.sh` — extend `map_event()` to return event+matcher pairs, add mapping for `on-artifact-write.sh` → `PostToolUse Write` and `PostToolUse Edit`, update JSON construction to set `"matcher"` field per mapping

## Phase 3: Skill File Updates

- [x] T009 [P] Remove bookkeeping from `fab/.kit/skills/fab-new.md` — delete Step 6 (infer change type) and Step 7 (indicative confidence), renumber subsequent steps
- [x] T010 [P] Remove bookkeeping from `fab/.kit/skills/fab-continue.md` — delete the `fab score <change>` instruction after spec generation
- [x] T011 [P] Remove bookkeeping from `fab/.kit/skills/fab-ff.md` — delete the three `set-checklist` calls from Step 4, remove score-after-spec instruction (keep gate checks)
- [x] T012 [P] Remove bookkeeping from `fab/.kit/skills/fab-fff.md` — delete the three `set-checklist` calls from Step 4
- [x] T013 [P] Remove bookkeeping from `fab/.kit/skills/fab-clarify.md` — delete Step 7 (recompute confidence via `fab score` in suggest mode)
- [x] T014 [P] Remove bookkeeping from `fab/.kit/skills/_generation.md` — delete Checklist Generation Procedure step 6 (three `set-checklist` CLI commands)

## Phase 4: Documentation

- [x] T015 Update `fab/.kit/skills/_scripts.md` — add `fab runtime` command reference with `set-idle` and `clear-idle` subcommands, usage, and purpose (required by constitution: "Changes to the `fab` CLI MUST update `_scripts.md`")

---

## Execution Order

- T001 blocks T002, T003
- T002 and T003 block T004
- T002 and T003 block T006 and T007 (hooks need the commands to exist)
- T005 is independent of T002-T004 (hook script calls fab CLI which already exists in PATH)
- T008 depends on T005 (sync script registers the new hook file)
- T009-T014 are independent of each other (all [P]) and independent of T001-T008
- T015 depends on T001 (documents the new commands)
