# Tasks: Expand Worktree Name Universe

**Change**: 260326-kpbc-expand-worktree-name-universe
**Spec**: `spec.md`
**Intake**: `intake.md`

## Phase 1: Core Implementation

- [x] T001 [P] Expand adjectives list in `src/go/wt/internal/worktree/names.go` — add ~80 new entries across existing categories (~8-10 per group) and 2 new categories (Time & weather, Texture & material) to reach >= 200 total
- [x] T002 [P] Expand nouns list in `src/go/wt/internal/worktree/names.go` — add ~80 new entries: ~3-5 per existing animal category plus a new "Nature & geography" category with ~15-20 entries, reaching >= 200 total

## Phase 2: Test Updates

- [x] T003 Update `TestWordListsNonEmpty` in `src/go/wt/internal/worktree/names_test.go` — change minimum thresholds from 120 to 200 for both adjectives and nouns
- [x] T004 Update `TestGenerateRandomName_Variety` comment in `src/go/wt/internal/worktree/names_test.go` — change "120*120=14400" to reflect new combo count
- [x] T005 Evaluate and restructure `TestGenerateUniqueName_RetryExhaustion` in `src/go/wt/internal/worktree/names_test.go` — if creating 40K+ temp dirs is too slow (> 30s), restructure to use a smaller subset or mock approach while still testing the retry exhaustion error path
- [x] T006 Update comment at top of adjectives/nouns slices in `src/go/wt/internal/worktree/names.go` — change "~120" to reflect actual new count

## Phase 3: Verification

- [x] T007 Run `go test ./src/go/wt/internal/worktree/...` and verify all tests pass, including the restructured collision test

---

## Execution Order

- T001 and T002 are independent (different slices), can run in parallel
- T003-T006 depend on T001+T002 being complete
- T007 depends on all prior tasks
