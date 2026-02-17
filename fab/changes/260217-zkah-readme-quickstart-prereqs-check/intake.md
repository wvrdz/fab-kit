# Intake: README Quick Start Restructure + fab-sync Prerequisites Check

**Change**: 260217-zkah-readme-quickstart-prereqs-check
**Created**: 2026-02-17
**Status**: Draft

## Origin

> Restructure README Quick Start and add prerequisites check to fab-sync.
>
> Discussion concluded with three agreed changes: (1) fold Initialize and Updating under Install as sub-sections, (2) update TOC to match, (3) add prerequisites checking to fab-sync.sh.

## Why

The README Quick Start currently presents Install and Initialize as separate top-level steps, but they're part of the same "getting fab into your project" flow. Separating them creates a false mental model — users think initialization is a distinct phase rather than part of setup. Similarly, the standalone Updating section is disconnected from the Install section it logically belongs with.

Additionally, `fab-sync.sh` has no prerequisites validation. If `yq`, `jq`, or other required tools are missing, the script either silently fails or produces confusing errors downstream. Users should get a clear, actionable error message up front.

## What Changes

### 1. Restructure Quick Start section in README.md

Collapse the current 4-step Quick Start (Install → Initialize → Your first change → Going parallel) into a 3-step flow:

```
### 1. Install
#### New project          (existing gh release download)
#### From a local clone   (existing cp -r)
#### Initialize           (existing fab-sync + direnv + /fab-setup — moved from step 2)
#### Updating from a previous version  (existing fab-upgrade.sh — moved from standalone section)

### 2. Your first change  (renumbered from 3)
### 3. Going parallel     (renumbered from 4)
```

Remove the standalone `## Updating` section entirely — its content moves under Install.

### 2. Update Contents TOC

Update the TOC line at the top to reflect the new structure. The `#updating` anchor should point to the sub-section within Install (GitHub auto-generates anchors from headings, so `#updating-from-a-previous-version` will work).

### 3. Add prerequisites check to fab-sync.sh

Add a check block early in fab-sync.sh (after the existing VERSION pre-flight, before directory creation) that validates required tools are available:

- **Fatal** (exit 1): `yq`, `jq` — these are used directly by fab-sync or downstream scripts
- **Warning** (non-fatal): `gh`, `direnv`, `bats` — useful but not required for sync itself

Output should point users to the Prerequisites section in the README.

## Affected Memory

- `fab-workflow/distribution`: (modify) Update to reflect prerequisites check in fab-sync and README restructuring

## Impact

- **README.md** — structural reorganization of Quick Start, TOC update, Updating section removal
- **fab/.kit/scripts/fab-sync.sh** — new prerequisites check block near top of script

## Open Questions

None — the approach was discussed and agreed in conversation.

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | yq and jq are fatal prerequisites | Both are used by fab-sync itself (jq for settings merge) and downstream scripts (yq for status files) | S:90 R:85 A:95 D:95 |
| 2 | Certain | gh, direnv, bats are non-fatal warnings | Not needed for sync to succeed — gh is for install, direnv for PATH, bats for testing | S:90 R:90 A:90 D:90 |
| 3 | Certain | GitHub auto-generates heading anchors | Standard GitHub markdown behavior — `#### Updating from a previous version` becomes `#updating-from-a-previous-version` | S:95 R:95 A:95 D:95 |
| 4 | Confident | Place prereqs check after VERSION check, before directory creation | Logical ordering: verify kit integrity first, then verify environment, then proceed with sync | S:80 R:90 A:85 D:80 |

4 assumptions (3 certain, 1 confident, 0 tentative, 0 unresolved).
