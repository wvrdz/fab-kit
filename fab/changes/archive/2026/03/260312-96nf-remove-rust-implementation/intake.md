# Intake: Remove Rust Implementation

**Change**: 260312-96nf-remove-rust-implementation
**Created**: 2026-03-12
**Status**: Draft

## Origin

> remove the rust implementation (src/rust) - its confusing the agent on any new work. The rust version isn't being maintained only the go version is. Also remove any related commands in justfile, in build workflows.

One-shot request. The user wants to eliminate the unmaintained Rust backend entirely — source code, build commands, and any references that cause agents to consider it during new work.

## Why

The Rust backend (`src/rust/fab/`) was built as an alternative to the Go binary but is no longer maintained. Only Go is actively developed and shipped in releases. The presence of the Rust source and build targets causes confusion:

1. **Agent confusion**: When agents explore the codebase for new work, they encounter `src/rust/` and the dispatcher's `fab-rust` priority logic, leading them to consider Rust compatibility unnecessarily.
2. **Dead code**: The Rust implementation is stale — it doesn't implement newer commands (e.g., `hook`, `pane-map`, `send-keys`, `idea`) and will only drift further.
3. **Build complexity**: The justfile carries 5 Rust-specific recipes (`test-rust`, `build-rust`, `_rust-target`, `build-rust-target`, `build-rust-all`) and a helper script (`src/scripts/just/rust-target.sh`) that serve no purpose.

If not addressed, agents will continue wasting context and making incorrect assumptions about multi-backend support.

## What Changes

### 1. Delete `src/rust/` directory

Remove the entire `src/rust/` directory tree. This includes:
- `src/rust/fab/Cargo.toml`, `Cargo.lock`
- All Rust source code and tests

### 2. Remove Rust recipes from `justfile`

Delete the following recipes and the `rust_src` variable:
- Line 2: `rust_src := "src/rust/fab"`
- Lines 48: comment referencing Rust in `build-all`
- Lines 54-55: `test-rust` recipe
- Lines 58-60: `build-rust` recipe
- Lines 63-64: `_rust-target` recipe
- Lines 67-70: `build-rust-target` recipe
- Lines 73-77: `build-rust-all` recipe

### 3. Remove `src/scripts/just/rust-target.sh`

Delete the helper script used by the `_rust-target` justfile recipe.

### 4. Simplify the dispatcher (`fab/.kit/bin/fab`)

Remove `fab-rust` references from the dispatcher script:
- Remove the `rust` backend detection in `--version` handling (lines 14-15)
- Remove the `rust` case in the backend override block (lines 29-30)
- Remove the `fab-rust` fallback in the default priority block (lines 38-39)
- Update the error message to reference only `fab-go` (line 41)

The dispatcher should only look for `fab-go`. The backend override mechanism (`FAB_BACKEND`, `.fab-backend`) can stay for Go-only use if desired, or be simplified to just check for `fab-go` directly.

### 5. Update `fab/.kit/skills/_scripts.md`

- Remove "Backend Priority" section mentioning `fab-rust` (line 28)
- Update the dispatcher description to reference only `fab-go`

### 6. Update documentation references

- `docs/memory/fab-workflow/kit-architecture.md` — remove/update the "Rust Binary (`fab-rust`)" section, benchmark data referencing Rust, and dispatcher description mentioning Rust
- `docs/memory/fab-workflow/distribution.md` — remove any `fab-rust` references
- `docs/specs/packages.md` — remove the `fab-rust` line from the directory tree

### 7. Clean up `package-kit.sh`

The `--exclude='.kit/bin/fab-rust'` in the generic archive tar command (line 25) can be removed since the file won't exist.

## Affected Memory

- `fab-workflow/kit-architecture`: (modify) Remove Rust Binary section, update dispatcher description, update benchmark section
- `fab-workflow/distribution`: (modify) Remove fab-rust exclusion/reference if present

## Impact

- **Source code**: `src/rust/` directory deleted
- **Build system**: justfile simplified, `rust-target.sh` script removed
- **Dispatcher**: `fab/.kit/bin/fab` simplified to Go-only
- **Skills**: `_scripts.md` updated to remove Rust backend mention
- **Documentation**: Memory and spec files updated to reflect Go-only architecture
- **CI**: Release workflow (`release.yml`) already Go-only — no changes needed
- **No user impact**: The Rust binary was never shipped in release archives; users only receive Go binaries

## Open Questions

None — the scope is clear and all artifacts to modify are identified.

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Delete entire `src/rust/` directory | User explicitly requested removal of the Rust implementation | S:95 R:60 A:95 D:95 |
| 2 | Certain | Remove all Rust recipes from justfile | User explicitly requested removal of related justfile commands | S:90 R:85 A:90 D:95 |
| 3 | Certain | Simplify dispatcher to Go-only | Direct consequence of removing Rust — dispatcher must not reference a backend that doesn't exist | S:85 R:80 A:90 D:90 |
| 4 | Confident | Keep backend override mechanism but Go-only | The `FAB_BACKEND` / `.fab-backend` mechanism still works for Go; removing it entirely is out of scope | S:60 R:85 A:70 D:70 |
| 5 | Confident | Update `_scripts.md` and memory docs | Constitution requires `_scripts.md` updates for CLI changes; memory should reflect current state | S:70 R:90 A:85 D:85 |
| 6 | Certain | No CI workflow changes needed | `release.yml` already builds Go-only; no Rust steps present | S:95 R:95 A:95 D:95 |
| 7 | Confident | Remove `package-kit.sh` fab-rust exclusion | The exclusion is harmless but dead — cleaning it keeps the script accurate | S:65 R:95 A:80 D:85 |

7 assumptions (3 certain, 4 confident, 0 tentative, 0 unresolved).
