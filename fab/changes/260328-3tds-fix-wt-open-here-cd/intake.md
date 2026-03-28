# Intake: Fix wt open "Open Here" cd Mechanism

**Change**: 260328-3tds-fix-wt-open-here-cd
**Created**: 2026-03-28
**Status**: Draft

## Origin

> Bug fix: wt open Open Here option prints cd command as text instead of changing directory. When selecting "Open here" from the editor menu, it outputs `cd -- /path/to/worktree` as a string to stdout instead of actually changing the shell working directory. The subprocess cannot change the parent shell directory, so "Open here" needs a different mechanism (e.g. printing a command for the shell to eval via a shell function wrapper, or using CDPATH).

## Why

When a user selects "Open here" from the `wt open` app menu, the Go binary prints `cd -- "/path/to/worktree"` to stdout (`apps.go:170`). This is by design — the binary is a child process and cannot change the parent shell's working directory.

The intended fix is already documented: a shell wrapper function in `main.go:21-31` that captures stdout, detects `cd` output, and `eval`s it in the parent shell. However, the wrapper must be manually added to the user's shell profile (`~/.bashrc` or `~/.zshrc`), and without it the `cd` command is just printed as text — confusing for users who expect "Open here" to actually change their directory.

The consequence: users who haven't installed the shell wrapper (which is most users, since it requires reading `wt --help` and manually copying the function) see a raw `cd` command printed and nothing happens. This makes "Open here" appear broken.

## What Changes

### Shell Wrapper Auto-Setup

Add a `wt init-shell` (or similar) subcommand that outputs the shell wrapper function to stdout, designed to be sourced by the user's shell profile:

```bash
# User adds to ~/.bashrc or ~/.zshrc:
eval "$(wt init-shell)"
```

This follows the established pattern used by tools like `direnv` (`eval "$(direnv hook bash)"`), `rbenv` (`eval "$(rbenv init -)"`), and `mise` (`eval "$(mise activate bash)"`).

The `init-shell` subcommand would:
1. Detect the current shell (bash/zsh via `$SHELL` or parent process)
2. Output the appropriate shell wrapper function
3. Include the `cd` detection logic already documented in `main.go`

### Improved User Guidance

When "Open here" is selected and no shell wrapper is detected (i.e., the binary is invoked directly rather than through the wrapper function):

1. Still print the `cd` command (for wrapper compatibility)
2. Optionally print a hint to stderr: `hint: Add 'eval "$(wt init-shell)"' to your shell profile to enable "Open here"`

### Detection of Wrapper

The wrapper can signal its presence by setting an environment variable (e.g., `WT_WRAPPER=1`) before calling `command wt "$@"`. When the binary sees `WT_WRAPPER=1`, it knows it's running through the wrapper and can skip the hint.

## Affected Memory

- `fab-workflow/distribution`: (modify) Document the shell wrapper setup as part of wt package distribution

## Impact

- `src/go/wt/cmd/main.go` — new `init-shell` subcommand registration
- `src/go/wt/cmd/init_shell.go` (new) — subcommand implementation
- `src/go/wt/internal/worktree/apps.go` — possible hint output on stderr when no wrapper detected
- `src/go/wt/cmd/open.go` — no structural changes, cd protocol unchanged
- `src/go/wt/cmd/create.go` — same cd protocol, benefits from same wrapper
- Shell wrapper function — enhanced version with `WT_WRAPPER=1` signaling

## Open Questions

- Should `wt init-shell` also handle fish shell, or just bash/zsh?
- Should the hint be printed every time, or only once (with a "don't show again" mechanism)?

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | cd protocol (print to stdout, eval in wrapper) is the correct mechanism | This is the standard pattern used by direnv, rbenv, mise, nvm — a child process cannot change parent shell CWD | S:90 R:95 A:95 D:95 |
| 2 | Certain | Shell wrapper must be user-installed (eval in profile) | No way to auto-inject into a running shell; user must opt-in by sourcing | S:85 R:90 A:95 D:95 |
| 3 | Confident | `wt init-shell` subcommand is the right UX for setup | Follows direnv/rbenv/mise convention; single eval line in profile | S:70 R:85 A:80 D:70 |
| 4 | Confident | Support bash and zsh only initially | These are the two shells used by the overwhelming majority of developers; fish can be added later | S:60 R:90 A:75 D:70 |
| 5 | Confident | Use WT_WRAPPER=1 env var to detect wrapper presence | Simple, zero-cost detection; allows binary to emit hint when invoked directly | S:65 R:90 A:80 D:75 |
| 6 | Tentative | Print hint to stderr when wrapper not detected | Helps users discover the setup, but could be noisy for scripted usage | S:50 R:85 A:60 D:50 |
<!-- assumed: stderr hint — common UX pattern for CLI tools, but may annoy users in scripts; easily removed later -->

6 assumptions (2 certain, 3 confident, 1 tentative, 0 unresolved).
