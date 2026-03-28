# Tasks: Fix wt open "Open Here" cd Mechanism

**Change**: 260328-3tds-fix-wt-open-here-cd
**Spec**: `spec.md`
**Intake**: `intake.md`

## Phase 1: Setup

- [x] T001 Create `src/go/wt/cmd/shell_setup.go` with `shellSetupCmd()` returning a `*cobra.Command` (Use: `shell-setup`, Short description, no flags)
- [x] T002 Register `shellSetupCmd()` in `src/go/wt/cmd/main.go` via `root.AddCommand(shellSetupCmd())`

## Phase 2: Core Implementation

- [x] T003 Implement shell wrapper output in `src/go/wt/cmd/shell_setup.go` — detect shell from `$SHELL` basename, output the wrapper function with `WT_WRAPPER=1` env var, `cd` last-line detection, and `eval`. Exit 1 with stderr error for unsupported shells (not bash/zsh)
- [x] T004 Add stderr hint to `src/go/wt/internal/worktree/apps.go` `OpenInApp()` `open_here` case — after printing `cd` to stdout, check `os.Getenv("WT_WRAPPER")`; if empty, print `hint: run "wt shell-setup" for setup instructions` to stderr
- [x] T005 Update `wt --help` long description in `src/go/wt/cmd/main.go` — replace raw wrapper function with `eval "$(wt shell-setup)"` setup instructions

## Phase 3: Integration & Edge Cases

- [x] T006 Add tests for `shell-setup` in `src/go/wt/cmd/shell_setup_test.go` — bash output, zsh output, unsupported shell error, wrapper contains `WT_WRAPPER=1`, wrapper contains `eval`
- [x] T007 Add test for stderr hint in `src/go/wt/cmd/open_test.go` or `src/go/wt/internal/worktree/apps_test.go` — verify hint printed when `WT_WRAPPER` unset, suppressed when set

---

## Execution Order

- T001 blocks T003
- T002 blocks T003 (registration needed before testing)
- T003 blocks T006
- T004 blocks T007
- T003 and T004 are independent, can run in parallel
- T005 is independent
