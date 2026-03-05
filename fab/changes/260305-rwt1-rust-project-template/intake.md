# Intake: Rust Project Template for Fab Kit

**Change:** 260305-rwt1-rust-project-template
**Type:** feat
**Issue:** fa-wh2
**Date:** 2026-03-05

---

## Problem Statement

When users run `/fab-setup` on a new project, Fab Kit scaffolds generic `fab/project/constitution.md` and `fab/project/config.yaml` files. These contain no language-specific conventions, forcing teams to manually add their own coding standards after setup. For Rust projects in particular, there is a well-established set of idioms and toolchain expectations that should be part of the scaffolded defaults from day one.

## Proposed Solution

Add a Rust-specific template layer to Fab Kit so that `/fab-setup` auto-detects Rust projects and merges Rust conventions into the constitution and config. This is the first language-specific preset, establishing the pattern for future language templates (Node.js, Go, etc.).

---

## What Changes

### A. New File: `fab/.kit/templates/constitutions/rust.md`

A Rust-specific constitution fragment containing the following principles:

- **No `unsafe` blocks** unless justified with an inline `// SAFETY:` comment explaining the invariant being upheld.
- **Error handling strategy:** Use `thiserror` for library crate errors (structured, typed) and `anyhow` for application-level errors (context-rich, ergonomic).
- **No `.unwrap()` in library code.** `.unwrap()` is permitted only in tests and `main.rs` / binary entry points. Library code must propagate errors with `?` or return `Result`.
- **Pure core pattern:** Domain logic crates must have zero I/O dependencies. All file system, network, and database access lives in adapter crates that depend on the core, never the reverse.
- **State machine coverage:** Every state machine transition must have a dedicated unit test asserting the before-state, trigger, and after-state.
- **Prefer std over external crates** when the standard library provides a reasonable implementation. External crates are justified when they offer meaningful ergonomic or correctness advantages (e.g., `regex`, `serde`).
- **Shell-outs via `std::process::Command`** with proper error handling: check exit status, capture stderr, and return structured errors on failure.
- **No hardcoded durations or thresholds.** All timeouts, retry counts, intervals, and numeric thresholds must come from configuration (config file, environment variable, or builder pattern default).

### B. New File: `fab/.kit/templates/configs/rust.yaml`

A Rust-specific config overlay containing:

```yaml
language: rust
source_paths:
  - "src/"
  - "crates/"

checklist:
  extra_categories:
    - compilation_clean
    - cargo_test_pass
    - clippy_clean

stage_directives:
  spec:
    - "Reference Cargo.toml for dependency constraints"
  review:
    - "Verify cargo check passes with no warnings"
    - "Verify cargo test passes"
    - "Run cargo clippy -- -D warnings"
```

### C. Modified: `fab/.kit/skills/fab-setup.md`

After scaffolding the default constitution and config, add a language detection step:

1. Check for marker files at the repository root:
   - `Cargo.toml` exists → **Rust project**
   - `package.json` exists → Node.js project (future template, no-op for now)
   - `go.mod` exists → Go project (future template, no-op for now)
2. For Rust detection:
   - Merge `templates/constitutions/rust.md` content into the scaffolded `fab/project/constitution.md` under a `## Rust Conventions` section.
   - Merge `templates/configs/rust.yaml` values into `fab/project/config.yaml`, adding to `source_paths`, `checklist.extra_categories`, and `stage_directives` without overwriting user-provided values.
3. Print to the user: `"Detected Rust project. Applied Rust conventions to constitution and config."`

### D. Modified: `fab/.kit/sync/2-sync-workspace.sh`

During workspace sync, add a detection check:

1. If `fab/.kit/templates/constitutions/` directory exists, AND
2. The project root contains `Cargo.toml`, AND
3. The existing `fab/project/constitution.md` does not contain the marker `## Rust Conventions`
4. Then print a suggestion: `"Rust project detected but constitution lacks Rust conventions. Run /fab-setup --refresh to apply."`

This is advisory only — sync does not modify the constitution automatically.

---

## Impact Assessment

| Aspect | Detail |
|---|---|
| New files | 2 template files (~50-80 lines each) in `fab/.kit/templates/` |
| Modified files | `fab/.kit/skills/fab-setup.md`, `fab/.kit/sync/2-sync-workspace.sh` |
| Breaking changes | None. Existing projects retain their current constitution and config unchanged. |
| Risk | Low. Template merging is additive. Detection is file-existence based (no parsing). |
| Rollback | Remove template files and revert skill/sync modifications. |

---

## Assumptions

| # | Assumption | SRAD | Rationale |
|---|---|---|---|
| A1 | `Cargo.toml` at repo root is a reliable signal for a Rust project | S | Standard Rust convention. Cargo is the universal build tool. Projects without `Cargo.toml` at root are non-standard and should not be auto-detected. |
| A2 | Constitution merging is append-only (add a new section, never rewrite existing content) | R | Users may have already customized their constitution. Overwriting would destroy their work. Appending a clearly marked section is safe and reversible. |
| A3 | Config merging uses additive semantics (add to lists, don't replace) | R | Same rationale as A2. `source_paths` and `stage_directives` should union with existing values, not replace them. |
| A4 | Only one language template applies per project (no polyglot detection) | A | Polyglot projects (e.g., Rust backend + Node.js frontend) are a real scenario, but handling them adds complexity. Single-language detection is sufficient for the first iteration. Future changes can layer multiple templates. |
| A5 | The `## Rust Conventions` heading serves as an idempotency marker | S | Simple string presence check prevents duplicate application. If the marker exists, the template has already been applied. |
| A6 | Future language templates (Node.js, Go) will follow the same pattern: a constitution fragment and a config overlay | A | The directory structure (`templates/constitutions/`, `templates/configs/`) assumes this pattern generalizes. If a future language needs a fundamentally different approach, the structure may need revision. |
| A7 | `fab-setup --refresh` flag exists or will be implemented as part of this change | D | The sync script suggests running `--refresh`, so the flag must work. If `--refresh` is not yet implemented, it needs to be added in fab-setup.md as part of this change. |
| A8 | Template files are static (no variable interpolation or templating engine needed) | S | The Rust conventions are universal enough that no project-specific variable substitution is required. The constitution fragment and config overlay are verbatim content. |

**SRAD Legend:**
- **S (Safe):** High confidence, well-grounded in existing conventions or constraints.
- **R (Reasonable):** Logical inference from context, low risk of being wrong.
- **A (Assumed):** Plausible but could go either way. Needs validation if scope grows.
- **D (Dangerous):** Could be wrong and would require rework. Flag for early clarification.

---

## Open Questions

1. Should `--refresh` re-apply the full template or only fill in missing sections? (Relates to A7)
2. If a user removes the `## Rust Conventions` section intentionally, sync will keep suggesting refresh. Should there be an opt-out marker (e.g., `<!-- fab:no-rust-conventions -->`)?
3. Should the Rust config overlay specify a minimum Rust edition (e.g., `edition: "2021"`) or leave that entirely to `Cargo.toml`?
