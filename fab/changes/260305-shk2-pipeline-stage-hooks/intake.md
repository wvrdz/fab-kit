# Intake: Pipeline Stage Hooks

**Change:** `260305-shk2-pipeline-stage-hooks`
**Type:** feat
**Issue:** fa-d1r
**Date:** 2026-03-05

---

## Problem

Fab Kit's pipeline has 8 stages: intake, spec, tasks, apply, review, hydrate, ship, review-pr. Currently, stage transitions are purely agent-driven with no automated enforcement of quality gates. There is no mechanism to run validation commands (tests, linting, formatting checks) at stage boundaries. This means an agent can advance through stages without verifying that the codebase is in a healthy state.

## Solution

Add configurable pre/post hooks that run shell commands at stage boundaries. Hooks are defined per-stage in the project config and execute automatically during stage transitions.

---

## What Changes

### A. Config Schema Extension

File: `fab/project/config.yaml`

A new optional `stage_hooks` section is added to the project config schema:

```yaml
stage_hooks:
  apply:
    post:
      - command: "cargo test"
        on_failure: "block"    # block = fail the stage, warn = log and continue
      - command: "cargo clippy -- -D warnings"
        on_failure: "warn"
  review:
    pre:
      - command: "cargo test"
        on_failure: "block"
    post:
      - command: "cargo fmt --check"
        on_failure: "warn"
  hydrate:
    pre:
      - command: "cargo doc --no-deps"
        on_failure: "warn"
```

**Hook types:**

- `pre`: Runs BEFORE the stage transition (before the `start` event). If `on_failure: "block"`, the stage cannot start.
- `post`: Runs AFTER the stage completes (after the `finish` event). If `on_failure: "block"`, the stage transitions to `failed` instead of `done`.

**Hook fields:**

| Field        | Type   | Required | Default | Description                                      |
|-------------|--------|----------|---------|--------------------------------------------------|
| `command`    | string | yes      | —       | Shell command to execute                         |
| `on_failure` | string | no       | "block" | Behavior on failure: "block" or "warn"           |
| `timeout`    | string | no       | "5m"    | Max execution time before the process is killed  |

### B. Go Binary Changes

File: `src/fab-go/internal/hooks/hooks.go` (new)

- `RunHooks(stage string, phase string, config *Config, workdir string) (bool, error)`
  - `phase` is `"pre"` or `"post"`
  - Reads `stage_hooks` from config, finds matching stage+phase combination
  - For each hook command: executes via `os/exec.Command("sh", "-c", command)`, captures stdout/stderr
  - Returns `(allPassed bool, firstError error)`
  - Respects `on_failure`: `"block"` returns `false` on failure, `"warn"` logs but returns `true`
  - Timeout: optional `timeout` field per hook (default 5m), kills the process on exceed
  - Estimated size: ~150-200 lines

File: `src/fab-go/internal/status/status.go` (modified)

- `Start()` now calls `RunHooks(stage, "pre", ...)` before transitioning. If blocked, returns an error and the stage does not start.
- `Finish()` now calls `RunHooks(stage, "post", ...)` after transitioning. If blocked, the stage transitions to `failed` instead of `done`.
- Estimated changes: ~30 lines

### C. Skill Awareness

File: `fab/.kit/skills/_preamble.md` (modified)

- Add a note that stage hooks may run automatically at stage boundaries, and that the agent should expect hook output in the transition response.
- Estimated changes: ~10 lines

File: `fab/.kit/skills/fab-continue.md` (modified)

- Add guidance: if a post hook fails with `"block"`, the agent sees the failure output and should fix the issue before re-advancing the stage.
- Estimated changes: ~10 lines

### D. Scaffold Update

File: `fab/.kit/scaffold/fab/project/config.yaml` (modified)

- Add a commented-out `stage_hooks` example section so new projects see the option.

---

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| Hooks run in the change's worktree directory | The worktree root is the natural cwd for build/test commands |
| Hooks inherit the current shell environment | Avoids complexity of environment isolation; hooks need access to toolchains |
| Hook output is captured and displayed to the agent on failure | The agent needs to see what broke to fix it |
| Hooks are synchronous | The pipeline must block until validation completes; async would add complexity with no benefit |
| Multiple hooks per stage+phase run sequentially | Deterministic ordering; first "block" failure stops the chain to avoid cascading noise |
| `on_failure` defaults to "block" if not specified | Fail-safe: unintentional omission should not silently skip validation |
| Timeout defaults to 5 minutes | Prevents runaway processes; long enough for most test suites |
| Executed via `sh -c` | Allows shell features (pipes, redirects, env vars) in hook commands |

---

## Impact

| Area | Scope |
|------|-------|
| New Go code | ~150-200 lines (`hooks` package) |
| Modified Go code | ~30 lines (`status.go` integration points) |
| Modified skills | ~10 lines each (awareness notes) |
| Schema extension | `config.yaml` gets optional `stage_hooks` section |
| Backward compatibility | Fully compatible — no `stage_hooks` means no hooks run |

---

## Assumptions

| # | Assumption | S | R | A | D | Score |
|---|-----------|---|---|---|---|-------|
| 1 | All hook commands are available in the shell environment where `fab` runs (e.g., `cargo`, `npm`, language toolchains are on PATH) | 0.9 | 0.7 | 0.8 | 0.6 | 0.75 |
| 2 | 5 minutes is a sufficient default timeout for most hook commands (test suites, linters, formatters) | 0.8 | 0.6 | 0.7 | 0.7 | 0.70 |
| 3 | Sequential execution of hooks within a stage+phase is acceptable (no need for parallel execution) | 0.9 | 0.8 | 0.9 | 0.8 | 0.85 |
| 4 | Hooks only need to run shell commands (no need for Go plugin hooks or webhook callbacks) | 0.8 | 0.7 | 0.8 | 0.7 | 0.75 |
| 5 | The agent can parse and act on hook failure output without structured error formats (plain stdout/stderr is sufficient) | 0.7 | 0.6 | 0.7 | 0.6 | 0.65 |
| 6 | Hook configuration at the project level (not per-change) is the right granularity | 0.8 | 0.7 | 0.7 | 0.6 | 0.70 |
| 7 | `sh -c` is available and appropriate on all target platforms (Linux, macOS) | 0.95 | 0.9 | 0.9 | 0.9 | 0.91 |
| 8 | Inheriting the full shell environment for hooks does not pose a security concern in the expected usage context | 0.7 | 0.5 | 0.6 | 0.5 | 0.58 |
| 9 | The `status.go` integration points (`Start()` and `Finish()`) are the correct and only places where hooks need to be called | 0.8 | 0.7 | 0.8 | 0.7 | 0.75 |
| 10 | No existing stage transition callers bypass `Start()`/`Finish()` in a way that would skip hooks | 0.7 | 0.6 | 0.7 | 0.6 | 0.65 |

**SRAD Key:**
- **S** (Specificity): How precisely defined is the assumption?
- **R** (Robustness): How resilient is the system if the assumption is wrong?
- **A** (Agreement): How likely are stakeholders to agree?
- **D** (Durability): How long will this assumption remain valid?
- **Score**: Weighted average across dimensions

---

## Out of Scope

- Webhook/HTTP-based hooks (shell commands only for now)
- Per-change hook overrides (project-level config only)
- Parallel hook execution
- Hook result caching or memoization
- GUI/dashboard visibility of hook results (agent-only for now)
