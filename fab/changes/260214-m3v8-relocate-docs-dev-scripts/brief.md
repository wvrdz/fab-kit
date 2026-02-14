# Brief: Reorganize docs, lib, and dev scripts

**Change**: 260214-m3v8-relocate-docs-dev-scripts
**Created**: 2026-02-14
**Status**: Draft

## Origin

> Move memory/ and specs/ from fab/ to docs/, and move fab-release.sh from fab/.kit/scripts/ to src/scripts/. See plan at .claude/plans/merry-roaming-anchor.md for full details.

## Why

`fab/` currently conflates three concerns: shipped kit (`.kit/`), workflow state (changes, config, constitution), and reference documentation (memory, specs). Moving memory and specs to `docs/` makes `fab/` focused on workflow machinery and puts documentation at a standard, discoverable location. Separately, dev-only scripts like `fab-release.sh` ship in the tarball even though end users never need them.

## What Changes

- Move `fab/memory/` to `docs/memory/`
- Move `fab/specs/` to `docs/specs/`
- Move `fab/.kit/scripts/fab-release.sh` to `src/scripts/fab-release.sh`
- Move `src/{calc-score,preflight,resolve-change,stageman}/` to `src/lib/{calc-score,preflight,resolve-change,stageman}/`
- Update `justfile` test glob from `src/*/test.sh` to `src/lib/*/test.sh`
- Update symlinks inside each `src/lib/*/` directory (they point to `../../fab/.kit/scripts/` — depth changes to `../../../fab/.kit/scripts/`)
- Update all path references across skills, templates, scaffold, scripts, config, constitution, README, and memory/specs files themselves
- Add `src/scripts` to the repo `.envrc` PATH (not the scaffold `.envrc` that ships)
- Add a migration entry for existing users
- Archived changes (68 folders) are frozen records and will NOT be updated

## Affected Memory

- `fab-workflow/kit-architecture`: (modify) directory structure changes — docs/ is a new top-level directory
- `fab-workflow/distribution`: (modify) tarball no longer includes fab-release.sh, docs live outside fab/
- `fab-workflow/init`: (modify) _init_scaffold.sh creates docs/ instead of fab/memory/ and fab/specs/
- `fab-workflow/context-loading`: (modify) context loading paths change from fab/memory/ to docs/memory/
- `fab-workflow/specs-index`: (modify) path references update
- `fab-workflow/hydrate`: (modify) hydration targets change from fab/memory/ to docs/memory/
- `fab-workflow/hydrate-specs`: (modify) path references update

## Impact

- **Skills**: ~14 skill files reference `fab/memory/` or `fab/specs/` paths — all need updating
- **Templates**: brief.md and spec.md templates reference `fab/memory/` paths
- **Scripts**: _init_scaffold.sh, fab-help.sh, _stageman.sh need path updates
- **Scaffold**: _init_scaffold.sh must create `docs/` structure instead of `fab/memory/` and `fab/specs/`
- **Cross-links preserved**: Both directories move together under `docs/`, so relative links between memory and specs (e.g., `../memory/index.md`) remain valid
- **Constitution**: Principle II and VI reference `fab/memory/` and `fab/specs/` explicitly
- **Justfile**: test glob changes from `src/*/test.sh` to `src/lib/*/test.sh`
- **Symlinks**: each `src/lib/*/` script symlink needs depth adjustment (one extra `../`)

## Open Questions

None — all decisions were resolved during planning discussion (docs/ over doc/, src/scripts/ for dev scripts, batch scripts stay in .kit/scripts/).
