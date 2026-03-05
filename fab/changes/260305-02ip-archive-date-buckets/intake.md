# Intake: Archive Date Buckets

**Change**: 260305-02ip-archive-date-buckets
**Created**: 2026-03-05
**Status**: Draft

## Origin

> Discussion session exploring how to organize the growing archive folder. User noted the archive keeps getting bigger over time and needs `yyyy/mm/<change>` organization. Considered naming: "partition", "lifecycle" (AWS S3 metaphor), ultimately decided to build it directly into `fab-archive` rather than a separate command, with a one-time migration for existing flat archives.

Interaction mode: conversational (via `/fab-discuss`). Key decisions reached before `/fab-new`:
- Date-bucketed path structure: `archive/yyyy/mm/{name}`
- Date source: parse `YYMMDD` prefix from folder name (no git archaeology)
- Built into `fab-archive` directly, not a separate command
- One-time migration for existing flat archive entries

## Why

The `fab/changes/archive/` folder grows monotonically — every completed change lands there and never leaves. Over time this becomes a flat list of dozens or hundreds of folders, making it hard to browse and slowing filesystem operations. Date-bucketing by `yyyy/mm/` provides natural chronological organization that scales indefinitely and mirrors the folder naming convention already in use (`YYMMDD-XXXX-slug`).

If we don't fix it: the archive remains a flat directory that gets progressively harder to navigate. Not catastrophic, but a quality-of-life drag that compounds.

## What Changes

### 1. Archive destination path

`archiveman.sh archive` currently moves changes to:
```
fab/changes/archive/{name}
```

Change to:
```
fab/changes/archive/yyyy/mm/{name}
```

Where `yyyy` and `mm` are derived from the `YYMMDD` prefix in the folder name. For example:
- `260305-02ip-archive-date-buckets` → `archive/2026/03/260305-02ip-archive-date-buckets`
- `260207-09sj-autonomy-framework` → `archive/2026/02/260207-09sj-autonomy-framework`

Parsing logic: extract first 6 characters of folder name, split as `YY` (chars 1-2), `MM` (chars 3-4), `DD` (chars 5-6). Prefix `20` to `YY` for the 4-digit year.

### 2. Archive restore path

`archiveman.sh restore` must resolve changes from the new nested structure. The `resolve_archive` function currently scans `archive/*/` — it needs to scan `archive/*/*/*/` (or recursively find directories with `.status.yaml`).

### 3. Archive list

`archiveman.sh list` currently lists `archive/*/`. Update to walk the nested structure and output folder names (without the `yyyy/mm/` prefix path, just the change folder name — consistent with current output contract).

### 4. Archive index

`archive/index.md` remains flat (one entry per change, most recent first). No structural change needed — the index is a logical listing, not a filesystem mirror. The `backfill_index` function needs to scan the nested structure.

### 5. One-time migration

A migration script (or function within `archiveman.sh`) that:
1. Scans `fab/changes/archive/` for flat entries (directories directly under `archive/` that aren't `yyyy/` directories)
2. Parses the `YYMMDD` prefix from each folder name
3. `mkdir -p archive/yyyy/mm/` and `mv` the folder into the bucket
4. Leaves `archive/index.md` untouched (it doesn't encode paths)

This should be idempotent — running it twice is safe (already-bucketed entries won't be in the flat scan).

### 6. Collision detection

The existing collision check (`if [ -e "$archive_dir/$folder" ]`) needs to use the new bucketed path.

## Affected Memory

- `fab-workflow/change-lifecycle`: (modify) Update archive path convention from flat to date-bucketed

## Impact

- **`fab/.kit/scripts/lib/archiveman.sh`** — primary change target (archive, restore, list, resolve_archive, backfill_index)
- **`fab/.kit/skills/fab-archive.md`** — may need doc updates if it references archive paths
- **Go migration (gm08)** — the Go port should implement the new bucketed structure from the start
- **`fab/changes/archive/`** — existing entries restructured by migration

## Open Questions

- None — design was resolved during discussion.

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Date source is folder name YYMMDD prefix | Discussed — user confirmed, no git archaeology needed | S:95 R:90 A:95 D:95 |
| 2 | Certain | Archive path is `archive/yyyy/mm/{name}` | Discussed — user explicitly chose this structure | S:95 R:85 A:90 D:95 |
| 3 | Certain | Built into fab-archive, not a separate command | Discussed — user decided against separate lifecycle command | S:95 R:90 A:90 D:90 |
| 4 | Certain | Migration restructures existing flat entries | Discussed — user requested initial migration on version bump | S:90 R:85 A:85 D:90 |
| 5 | Confident | Year prefix is `20` + 2-digit year | Folder names use `YYMMDD` format; `20` prefix valid through 2099 | S:70 R:95 A:90 D:85 |
| 6 | Confident | `archive/index.md` stays flat (no path changes) | Index is logical, not a filesystem mirror; simplest approach | S:60 R:90 A:85 D:80 |
| 7 | Confident | Migration is idempotent and safe to re-run | Constitution requires idempotent operations (Principle III) | S:70 R:85 A:90 D:85 |

7 assumptions (4 certain, 3 confident, 0 tentative, 0 unresolved).
