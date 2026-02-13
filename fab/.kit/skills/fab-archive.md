---
name: fab-archive
description: "Archive a completed change — move to archive folder, update index, mark backlog items, clear pointer."
---

# /fab-archive [<change-name>]

> Read and follow the instructions in `fab/.kit/skills/_context.md` before proceeding.

---

## Purpose

Archive a completed change after hydrate. Moves to archive, updates index, marks backlog items done, clears active pointer. Safe to re-run after interruption — detects already-moved folders and completes remaining steps.

---

## Arguments

- **`<change-name>`** *(optional)* — target a specific change. Resolution per `_context.md` (Change-name override).

---

## Pre-flight

1. Run `fab/.kit/scripts/fab-preflight.sh [change-name]` per `_context.md`
2. **Hydrate Guard**: If `progress.hydrate` is not `done`, STOP: `Hydrate has not completed. Run /fab-continue to hydrate memory first.`

---

## Context Loading

Minimal: `brief.md` (for backlog ID + keywords), `.status.yaml`, `fab/backlog.md` (if exists), `fab/current`.

---

## Behavior

### Resumability

If folder already at `fab/changes/archive/{name}/`, skip move and complete remaining steps.

### Step 1: Move Change Folder

`fab/changes/{name}/` → `fab/changes/archive/{name}/`. Create `archive/` if needed. Do NOT rename.

### Step 2: Update Archive Index

Maintain `fab/changes/archive/index.md`:
- Doesn't exist → create with header, backfill all existing archived folders
- Exists → prepend entry (most-recent-first)

Entry format: `- **{folder-name}** — {1-2 sentence description from brief Why}`

### Step 3: Mark Backlog Items Done

Skip silently if `fab/backlog.md` doesn't exist.

**3a — Exact-ID**: If brief has backlog ID, find and mark done (`- [ ]` → `- [x]`), move to Done section.

**3b — Keyword Scan**: Extract keywords from brief title/Why (filter stop words). Match against unchecked items — candidate when ≥2 significant keywords overlap (exclude 3a matches). No candidates → proceed silently.

**3c — Interactive Confirmation** (if 3b found candidates):

```
Backlog matches found:
  1. [ID] {description (~80 chars)}
  2. [ID] {description (~80 chars)}

Mark as done? (comma-separated numbers, or "none")
```

### Step 4: Clear Pointer (Conditional)

If `fab/current` contains the archived change → delete `fab/current`. Otherwise no-op.

Steps execute 1→4 for safety. If interrupted, re-run completes remaining.

---

## Output

```
Archive: {change name}

Moved:    ✓ fab/changes/archive/{name}/        (or: ✓ already in archive)
Index:    ✓ fab/changes/archive/index.md updated
Backlog:  ✓ [ID] marked done                   (or: — no backlog file)
Scan:     ✓ {N} candidates, {M} marked done    (or: ✓ no matches)
Pointer:  ✓ fab/current cleared                 (or: — skipped, not active)

Archive complete.

Next: /fab-new <description>
```
