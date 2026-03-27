# Intake: Brew Install System Shim

**Change**: 260325-lhhk-brew-install-system-shim
**Created**: 2026-03-25
**Status**: Draft

## Origin

> System-level installation for fab-kit via Homebrew. `brew install fab-kit` installs three binaries on PATH: `fab` (version-aware shim that reads fab_version from repo config.yaml, finds/fetches the matching fab/.kit/ release, and dispatches to the per-repo fab/.kit/bin/fab runtime), `wt` (worktree management), and `idea` (backlog management). The per-repo `fab/.kit/bin/fab` stays as the versioned runtime — never on PATH, invoked by the shim via absolute path. Each repo pins its own fab-kit version in config.yaml. The shim handles version resolution, download/caching of releases, and dispatch. wt and idea move out of fab/.kit/bin/ to system-only install.

This change was preceded by an exploratory `/fab-discuss` session covering distribution strategy, naming, and version management tradeoffs. Key decisions from that discussion are captured in the Assumptions table below.

## Why

Today, fab-kit is distributed by copying `fab/.kit/` into each repo. This works but creates friction:

1. **First-run friction** — new repos require manual `cp -r` or running an upgrade script to bootstrap `fab/.kit/`. There's no standard system-level command to get started.
2. **No version management** — repos vendor their own copy of `.kit/`, but there's no mechanism for a repo to declare which version it needs and have the tooling automatically resolve it. Upgrades are manual per-repo.
3. **Standalone tools trapped in repos** — `wt` (worktree management) and `idea` (backlog management) are general-purpose utilities useful outside fab-managed repos, but currently live inside `fab/.kit/bin/` and are only accessible within repos that have fab installed.

Without this change, onboarding remains manual, version drift across repos is invisible, and standalone utilities are unnecessarily scoped to fab-managed repos.

## What Changes

### System-level Homebrew formula (`fab-kit`)

A Homebrew formula named `fab-kit` that installs three binaries to the system PATH:

- **`fab`** — a version-aware shim/dispatcher (see below)
- **`wt`** — the worktree management binary (currently at `fab/.kit/bin/wt`)
- **`idea`** — the backlog management binary (currently at `fab/.kit/bin/idea`)

```
brew install fab-kit
# Installs: /usr/local/bin/fab, /usr/local/bin/wt, /usr/local/bin/idea
```

### The `fab` shim (version-aware dispatcher)

The system-installed `fab` binary acts as a thin shim. When invoked:

1. Walk up from CWD to find `fab/project/config.yaml`
2. Read `fab_version` from `config.yaml` (e.g., `fab_version: "0.39.0"`)
   - If `config.yaml` found but `fab_version` absent: error with actionable message (e.g., `"No fab_version in config.yaml. Run 'fab init' to set one."`)
3. Check the local cache for the matching version (`~/.fab-kit/versions/0.39.0/`)
4. If not cached, download the release from GitHub (`wvrdz/fab-kit` releases) and cache it
5. Exec `~/.fab-kit/versions/0.39.0/fab/.kit/bin/fab <original args>` — full passthrough of all arguments

If no `config.yaml` is found (not in a fab-managed repo), the shim serves non-repo commands: `fab init` (primary use case — scaffolds a new project), `fab --version`, etc.

```yaml
# fab/project/config.yaml — new field
fab_version: "0.39.0"
```

### Cache layout

```
~/.fab-kit/
  versions/
    0.39.0/
      fab/.kit/bin/fab      # the versioned runtime
      fab/.kit/bin/fab-go   # the Go backend
      fab/.kit/skills/      # skill files for this version
      fab/.kit/templates/   # templates for this version
    0.38.3/
      ...
```

### `wt` and `idea` become system-only binaries

These binaries move out of `fab/.kit/bin/` and are distributed exclusively through the Homebrew formula. They are not version-coupled to the per-repo fab-kit version — they're standalone utilities.

### Per-repo runtime stays at `fab/.kit/bin/fab`

The per-repo binary at `fab/.kit/bin/fab` remains unchanged. It is the versioned runtime that the shim dispatches to. It is never on PATH — the shim invokes it by absolute path. The binary name stays `fab` (not renamed).

### New `config.yaml` field: `fab_version`

A new optional field in `fab/project/config.yaml` that declares the fab-kit version the repo expects. When present, the system shim uses it for version resolution. When absent, the shim errors with an actionable message directing the user to run `fab init`.

### `fab init` (primary use case)

The shim's main onboarding command. When run outside a fab-managed repo (or in a repo without `fab_version`), `fab init` scaffolds the `fab/project/` structure and sets `fab_version` to the latest release. This completes the "zero to working" story: `brew install fab-kit` → `fab init` → repo is fab-managed.

## Affected Memory

- `fab-workflow/distribution.md`: (new) Document the Homebrew distribution model, shim architecture, cache layout, and version resolution
- `fab-workflow/architecture.md`: (modify) Update to reflect wt/idea moving to system-only, shim dispatcher model

## Impact

- **Homebrew formula**: New formula in a tap or core homebrew (likely a tap: `wvrdz/tap`)
- **`fab/.kit/bin/`**: `wt` and `idea` removed from this directory in future releases
- **`config.yaml` schema**: New optional `fab_version` field
- **Go codebase**: New shim binary (small — walks up for config, resolves version, execs)
- **GitHub releases**: Release artifacts must be structured for the shim to download and cache
- **Existing repos**: No breaking change — repos without the shim continue to work via direct `fab/.kit/bin/fab` invocation

## Clarifications

### Session 2026-03-27

| # | Action | Detail |
|---|--------|--------|
| 7 | Confirmed | `wvrdz/homebrew-tap` (org-level tap) |
| 8 | Changed | Error with actionable message when `fab_version` absent |
| 9 | Resolved | `fab init` in scope — primary use case |
| 10 | Resolved | No automatic cache eviction |
| 11 | Resolved | No `fab self-update` — rely on `brew upgrade` |

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Homebrew formula name is `fab-kit`, binary name is `fab` | Discussed — user explicitly chose this split to avoid Python Fabric collision on formula name while keeping `fab` as the user-facing command | S:95 R:90 A:95 D:95 |
| 2 | Certain | `wt` and `idea` are system-only, not per-repo | Discussed — user confirmed these are standalone utilities not version-coupled to the repo | S:95 R:80 A:90 D:95 |
| 3 | Certain | Per-repo binary stays named `fab` at `fab/.kit/bin/fab` | Discussed — user decided against renaming the inner binary since the shim dispatches by absolute path | S:90 R:85 A:90 D:90 |
| 4 | Certain | Version pinned per-repo via `config.yaml` field | Discussed — user chose version-manager pattern (option 1) over wrapper pattern (option 2) | S:90 R:70 A:85 D:85 |
| 5 | Confident | Cache lives at `~/.fab-kit/versions/` | Standard XDG-style user cache location; could also be `~/.cache/fab-kit/` | S:60 R:90 A:70 D:60 |
| 6 | Confident | Shim downloads from GitHub releases | Natural fit given existing `wvrdz/fab-kit` repo; alternative would be a separate artifact store | S:65 R:85 A:75 D:70 |
| 7 | Certain | Homebrew tap at `wvrdz/homebrew-tap` | Clarified — user confirmed org-level tap | S:95 R:90 A:50 D:50 |
| 8 | Certain | Error when `fab_version` absent from `config.yaml` | Clarified — user chose strict mode; shim errors with actionable message directing user to set `fab_version` | S:95 R:70 A:50 D:45 |
| 9 | Certain | `fab init` is in scope as a primary use case | Clarified — user confirmed this is the main use case for the shim | S:95 R:70 A:80 D:90 |
| 10 | Certain | No automatic cache eviction — manual cleanup only | Clarified — user confirmed; versions are small, a cleanup command can be added later | S:95 R:90 A:80 D:90 |
| 11 | Certain | No `fab self-update` — rely on `brew upgrade fab-kit` | Clarified — user confirmed; don't reinvent the package manager | S:95 R:90 A:85 D:90 |

11 assumptions (9 certain, 2 confident, 0 tentative, 0 unresolved).
