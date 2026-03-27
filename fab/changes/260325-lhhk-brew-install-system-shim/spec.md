# Spec: Brew Install System Shim

**Change**: 260325-lhhk-brew-install-system-shim
**Created**: 2026-03-27
**Affected memory**: `docs/memory/fab-workflow/distribution.md` (new section), `docs/memory/fab-workflow/kit-architecture.md` (modify)

## Non-Goals

- Removing `fab/.kit/bin/fab` (the per-repo shell dispatcher) — it stays as the versioned runtime
- Automatic cache eviction — versions accumulate until manually cleaned
- `fab self-update` — rely on `brew upgrade fab-kit`
- Core Homebrew submission — using a tap (`wvrdz/homebrew-tap`)
- Modifying the existing release archive format — the shim downloads existing archives as-is

## Shim: Version-Aware Dispatcher

### Requirement: Config Discovery

The `fab` shim SHALL walk up from the current working directory to find `fab/project/config.yaml`. The search starts at CWD and checks each parent directory up to the filesystem root.

#### Scenario: Config found in current repo
- **GIVEN** CWD is `/home/user/myproject/src/`
- **WHEN** `fab/project/config.yaml` exists at `/home/user/myproject/fab/project/config.yaml`
- **THEN** the shim reads `fab_version` from that file

#### Scenario: Config not found (not in a fab-managed repo)
- **GIVEN** CWD is `/home/user/scratch/` with no `fab/project/config.yaml` in any ancestor
- **WHEN** the user runs `fab init`
- **THEN** the shim serves the non-repo command directly (no version resolution needed)

#### Scenario: Config not found and non-init command
- **GIVEN** CWD is not inside a fab-managed repo
- **WHEN** the user runs `fab status`
- **THEN** the shim SHALL exit with error: `"Not in a fab-managed repo. Run 'fab init' to get started."`
- **AND** commands `--version` and `--help` SHALL still work without a repo

### Requirement: Version Resolution

The shim SHALL read the `fab_version` field from `config.yaml` and resolve it to a cached kit version.

#### Scenario: fab_version present and cached
- **GIVEN** `config.yaml` contains `fab_version: "0.39.0"`
- **AND** `~/.fab-kit/versions/0.39.0/` exists with a valid kit
- **WHEN** the user runs `fab status`
- **THEN** the shim execs `~/.fab-kit/versions/0.39.0/fab/.kit/bin/fab status`

#### Scenario: fab_version present but not cached
- **GIVEN** `config.yaml` contains `fab_version: "0.40.0"`
- **AND** `~/.fab-kit/versions/0.40.0/` does not exist
- **WHEN** the user runs `fab status`
- **THEN** the shim downloads the `v0.40.0` release from GitHub, caches it at `~/.fab-kit/versions/0.40.0/`, then execs the cached runtime

#### Scenario: fab_version absent from config.yaml
- **GIVEN** `config.yaml` exists but has no `fab_version` field
- **WHEN** the user runs any fab command
- **THEN** the shim SHALL exit with error: `"No fab_version in config.yaml. Run 'fab init' to set one."`

### Requirement: Argument Passthrough

The shim SHALL pass all arguments verbatim to the per-repo runtime via `exec`. No argument parsing, transformation, or interception beyond `--version`, `--help`, and `init`.

#### Scenario: Full argument passthrough
- **GIVEN** a resolved version at `~/.fab-kit/versions/0.39.0/`
- **WHEN** the user runs `fab status finish mychange intake`
- **THEN** the shim execs `~/.fab-kit/versions/0.39.0/fab/.kit/bin/fab status finish mychange intake` with identical arguments

### Requirement: Version Download

The shim SHALL download kit releases from `github.com/wvrdz/fab-kit` using platform detection.

#### Scenario: Platform-specific download
- **GIVEN** the shim detects `darwin/arm64`
- **WHEN** version `0.40.0` is not cached
- **THEN** the shim downloads `https://github.com/wvrdz/fab-kit/releases/download/v0.40.0/kit-darwin-arm64.tar.gz`
- **AND** extracts it to `~/.fab-kit/versions/0.40.0/`
- **AND** the cached version contains `fab/.kit/bin/fab` (the per-repo runtime)

#### Scenario: Download failure
- **GIVEN** the network is unavailable or the release tag does not exist
- **WHEN** the shim attempts to download
- **THEN** the shim SHALL exit with a clear error message including the URL attempted and a suggestion to check network connectivity or verify the version exists

#### Scenario: Concurrent downloads
- **GIVEN** two worktrees trigger a download of the same version simultaneously
- **WHEN** both attempt to write to `~/.fab-kit/versions/0.40.0/`
- **THEN** the shim SHALL use atomic extraction (download to temp dir, then rename) to prevent corruption

## Shim: Cache Management

### Requirement: Cache Layout

The cache SHALL live at `~/.fab-kit/versions/` with one subdirectory per version.

#### Scenario: Cache directory structure
- **GIVEN** versions 0.39.0 and 0.40.0 have been used
- **WHEN** the user inspects `~/.fab-kit/`
- **THEN** the layout is:
  ```
  ~/.fab-kit/
    versions/
      0.39.0/
        fab/.kit/bin/fab
        fab/.kit/bin/fab-go
        fab/.kit/skills/
        fab/.kit/templates/
        fab/.kit/VERSION
      0.40.0/
        ...
  ```

### Requirement: No Automatic Eviction

The shim SHALL NOT automatically evict cached versions. Users manage cache size manually (e.g., `rm -rf ~/.fab-kit/versions/0.38.0/`).

#### Scenario: Old versions persist
- **GIVEN** versions 0.38.0, 0.39.0, and 0.40.0 are cached
- **WHEN** the user only uses 0.40.0
- **THEN** all three versions remain in the cache indefinitely

## Shim: `fab init`

### Requirement: Project Scaffolding

`fab init` SHALL scaffold the `fab/project/` structure in the current directory and set `fab_version` to the latest release.

#### Scenario: Init in a fresh project
- **GIVEN** CWD is `/home/user/newproject/` with no `fab/` directory
- **WHEN** the user runs `fab init`
- **THEN** the shim fetches the latest release version from GitHub
- **AND** downloads and caches that version (if not already cached)
- **AND** runs the cached version's `fab-sync.sh` to bootstrap the project
- **AND** creates `fab/project/config.yaml` with `fab_version: "{latest}"`

#### Scenario: Init in a repo with existing fab/ but no fab_version
- **GIVEN** CWD has `fab/project/config.yaml` without `fab_version`
- **WHEN** the user runs `fab init`
- **THEN** the shim adds `fab_version: "{latest}"` to the existing `config.yaml`
- **AND** does NOT overwrite other config fields

#### Scenario: Init in an already-initialized repo
- **GIVEN** CWD has `fab/project/config.yaml` with `fab_version: "0.39.0"`
- **WHEN** the user runs `fab init`
- **THEN** the shim reports `"Already initialized (fab_version: 0.39.0). Edit fab_version in config.yaml to change versions."`
<!-- clarified: Replaced 'fab upgrade' reference with manual config edit — fab upgrade is not defined in this spec and Non-Goals explicitly exclude self-update machinery -->

### Requirement: Latest Version Detection

`fab init` SHALL determine the latest release by querying GitHub releases.

#### Scenario: Detect latest version
- **GIVEN** `wvrdz/fab-kit` has releases v0.39.0, v0.40.0, v0.41.0
- **WHEN** `fab init` queries GitHub
- **THEN** it resolves to `0.41.0` (the latest non-prerelease)

#### Scenario: Init with network failure
- **GIVEN** the network is unavailable or GitHub is unreachable
- **WHEN** the user runs `fab init`
- **THEN** the shim SHALL exit with a clear error message including a suggestion to check network connectivity
- **AND** no files are created or modified
<!-- clarified: Added missing edge case — fab init depends on GitHub for latest version detection and download, so network failure must be handled explicitly (mirrors the existing download failure scenario pattern) -->

## Homebrew Formula

### Requirement: Formula Name and Tap

The formula SHALL be named `fab-kit` and hosted in the `wvrdz/homebrew-tap` repository.

#### Scenario: Installation
- **GIVEN** the tap is added or auto-tapped
- **WHEN** the user runs `brew install wvrdz/tap/fab-kit`
- **THEN** three binaries are installed to the Homebrew prefix: `fab`, `wt`, `idea`

### Requirement: Three Binaries on PATH

The formula SHALL install `fab` (shim), `wt`, and `idea` to the Homebrew bin directory.

#### Scenario: All binaries available
- **GIVEN** `brew install fab-kit` completed successfully
- **WHEN** the user checks their PATH
- **THEN** `which fab` → Homebrew prefix, `which wt` → Homebrew prefix, `which idea` → Homebrew prefix

### Requirement: wt and idea Are Standalone

`wt` and `idea` installed via Homebrew SHALL be standalone binaries, not version-coupled to any per-repo fab-kit version. They are the same binaries currently at `fab/.kit/bin/wt` and `fab/.kit/bin/idea`.

#### Scenario: wt works outside any fab repo
- **GIVEN** `fab-kit` is installed via Homebrew
- **WHEN** the user runs `wt create myproject` outside any fab-managed repo
- **THEN** the worktree management command works (it only needs git, not fab)

## Config Schema: `fab_version`

### Requirement: New Optional Field

`fab/project/config.yaml` SHALL support an optional `fab_version` field declaring the pinned fab-kit version.

#### Scenario: Field present
- **GIVEN** `config.yaml` contains `fab_version: "0.39.0"`
- **WHEN** the system shim reads the config
- **THEN** it resolves to version 0.39.0

#### Scenario: Field absent (shim installed)
- **GIVEN** `config.yaml` exists without `fab_version`
- **WHEN** the system shim reads the config
- **THEN** the shim errors with guidance to run `fab init`

#### Scenario: Field absent (direct invocation — no shim)
- **GIVEN** a user invokes `fab/.kit/bin/fab` directly (no shim)
- **WHEN** `config.yaml` has no `fab_version`
- **THEN** behavior is unchanged — the per-repo runtime does not read this field

## Go Binary: Shim Implementation

### Requirement: Shim as a Go Binary

The `fab` shim SHALL be implemented as a Go binary in a new module at `src/go/shim/`.

#### Scenario: Build the shim
- **GIVEN** the Go toolchain is available
- **WHEN** `cd src/go/shim && go build -o fab ./cmd`
- **THEN** a `fab` binary is produced that implements the shim behavior

### Requirement: Minimal Dependencies

The shim SHALL have minimal dependencies — standard library plus `gopkg.in/yaml.v3` for config parsing. No Cobra (the shim does not need a full command framework).

#### Scenario: No Cobra dependency
- **GIVEN** `src/go/shim/go.mod`
- **WHEN** inspecting the dependency list
- **THEN** `github.com/spf13/cobra` is NOT present

## Release and Distribution

### Requirement: Shim in Release Archives

The shim binary SHALL NOT be included in the existing per-repo kit archives (`kit-{os}-{arch}.tar.gz`). It is distributed exclusively through Homebrew.

#### Scenario: Kit archives unchanged
- **GIVEN** a release is cut
- **WHEN** `just package-kit` runs
- **THEN** the kit archives contain `fab-go`, `idea`, and `wt` at `.kit/bin/` — no shim binary

### Requirement: Homebrew Formula Build

The Homebrew formula SHALL build the shim, `wt`, and `idea` from source using Go.

#### Scenario: Formula build from source
- **GIVEN** the formula references a tagged release of `wvrdz/fab-kit`
- **WHEN** `brew install fab-kit` runs
- **THEN** it compiles `src/go/shim/cmd` → `fab`, `src/go/wt/cmd` → `wt`, `src/go/idea/cmd` → `idea`

### Requirement: Justfile Recipes for Shim

New justfile recipes SHALL be added for building the shim locally.

#### Scenario: Local shim build
- **GIVEN** the developer runs `just build-shim`
- **WHEN** the build completes
- **THEN** a `fab` shim binary is produced (separate from the per-repo `fab-go`)

## Backward Compatibility

### Requirement: No Breaking Changes

Existing repos that invoke `fab/.kit/bin/fab` directly SHALL continue to work without modification. The shim is additive — it does not replace or modify the per-repo runtime.

#### Scenario: Direct invocation still works
- **GIVEN** a repo with `fab/.kit/bin/fab` and `fab/.kit/bin/fab-go`
- **WHEN** the user runs `fab/.kit/bin/fab status`
- **THEN** it works exactly as before, regardless of whether the shim is installed

### Requirement: wt and idea Remain in Kit Archives (Transition)

During the transition period, `wt` and `idea` SHALL continue to be included in per-repo kit archives. Removal is deferred to a future change once Homebrew adoption is sufficient.

#### Scenario: Kit archives during transition
- **GIVEN** this change is shipped
- **WHEN** a repo upgrades to the new kit version
- **THEN** `fab/.kit/bin/wt` and `fab/.kit/bin/idea` are still present in the archive

## Design Decisions

1. **Shim as separate Go module (`src/go/shim/`)**: The shim is a distinct binary with different concerns (config discovery, version resolution, HTTP download) from the per-repo runtime (`fab-go`). Separate module prevents dependency bleed and keeps the shim minimal.
   - *Why*: The shim needs HTTP client and archive extraction but not Cobra or any fab internals.
   - *Rejected*: Adding shim as a subcommand of `fab-go` — this would couple the shim to the per-repo binary and defeat the purpose of version management.

2. **No Cobra for the shim**: The shim intercepts only `--version`, `--help`, and `init`. Everything else is passed through verbatim. A full command framework is unnecessary overhead.
   - *Why*: Minimizes binary size and startup time. The shim's job is to find the right version and exec into it.
   - *Rejected*: Using Cobra for consistency with `fab-go` — the shim has fundamentally different responsibilities.

3. **Transition period for wt/idea in kit archives**: Rather than immediately removing `wt` and `idea` from kit archives, keep them during a transition period. This avoids breaking repos that haven't adopted Homebrew.
   - *Why*: Gradual migration is safer. Repos that use `fab/.kit/bin/wt` directly continue working.
   - *Rejected*: Immediate removal — too disruptive, no way to know which repos have adopted Homebrew.

4. **Atomic version caching via temp dir + rename**: Prevents corruption when multiple worktrees trigger the same download concurrently. Standard pattern for filesystem atomicity.
   - *Why*: Multiple tmux panes / worktrees often run in parallel and could race on first use of a new version.
   - *Rejected*: File locking — more complex, platform-dependent, and rename is sufficient for this use case.

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Homebrew formula name is `fab-kit`, binary name is `fab` | Confirmed from intake #1 — user explicitly chose this split | S:95 R:90 A:95 D:95 |
| 2 | Certain | `wt` and `idea` are system-only, not per-repo | Confirmed from intake #2 — standalone utilities | S:95 R:80 A:90 D:95 |
| 3 | Certain | Per-repo binary stays named `fab` at `fab/.kit/bin/fab` | Confirmed from intake #3 — shim dispatches by absolute path | S:90 R:85 A:90 D:90 |
| 4 | Certain | Version pinned per-repo via `config.yaml` field | Confirmed from intake #4 — version-manager pattern | S:90 R:70 A:85 D:85 |
| 5 | Confident | Cache lives at `~/.fab-kit/versions/` | Carried from intake #5 — standard user cache location; could also be `~/.cache/fab-kit/` | S:60 R:90 A:70 D:60 |
| 6 | Confident | Shim downloads from GitHub releases | Carried from intake #6 — natural fit for existing release infrastructure | S:65 R:85 A:75 D:70 |
| 7 | Certain | Homebrew tap at `wvrdz/homebrew-tap` | Confirmed from intake #7 — user confirmed org-level tap | S:95 R:90 A:50 D:50 |
| 8 | Certain | Error when `fab_version` absent from `config.yaml` | Confirmed from intake #8 — strict mode, actionable error message | S:95 R:70 A:50 D:45 |
| 9 | Certain | `fab init` is in scope as a primary use case | Confirmed from intake #9 — main use case for the shim | S:95 R:70 A:80 D:90 |
| 10 | Certain | No automatic cache eviction | Confirmed from intake #10 — manual cleanup only | S:95 R:90 A:80 D:90 |
| 11 | Certain | No `fab self-update` — rely on `brew upgrade fab-kit` | Confirmed from intake #11 — don't reinvent the package manager | S:95 R:90 A:85 D:90 |
| 12 | Certain | Shim is a separate Go module at `src/go/shim/` | Per-repo runtime and shim have different dependencies and concerns; separation is clean | S:85 R:85 A:90 D:85 |
| 13 | Certain | Transition period for wt/idea in kit archives | Archives keep wt/idea during transition; removal deferred to future change | S:85 R:90 A:85 D:90 |
| 14 | Confident | Shim intercepts only `--version`, `--help`, and `init` | All other commands pass through to the resolved runtime; minimal surface area | S:70 R:90 A:80 D:70 |
| 15 | Confident | Platform detection via `runtime.GOOS` and `runtime.GOARCH` | Go provides these at compile time; no need for `uname` in a Go binary | S:80 R:95 A:90 D:85 |

15 assumptions (11 certain, 4 confident, 0 tentative, 0 unresolved).
