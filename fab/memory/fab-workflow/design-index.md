# Design Index

**Domain**: fab-workflow

## Overview

`fab/specs/index.md` is the centralized index for pre-implementation design artifacts. It complements `fab/memory/index.md` (post-implementation truth) by providing a persistent home for conceptual design intent — what was planned, the "why" behind features.

## Requirements

### Design vs Docs Distinction

Design files are pre-implementation artifacts — what you planned. They capture conceptual design intent, high-level decisions, and the "why" behind features. Docs are post-implementation artifacts — what actually happened, the authoritative source of truth for system behavior.

- `fab/specs/index.md` boilerplate clearly states design files are pre-implementation / planning artifacts
- `fab/memory/index.md` boilerplate clearly states docs are post-implementation / authoritative truth
- Both index files cross-reference each other with relative links

### Flat Structure

The design index does not prescribe a domain-based directory hierarchy. Design files may be organized by the human in any structure they choose. The index simply lists what exists.

### Human-Curated Ownership

Design files are written and maintained by humans. No automated tooling creates or enforces structure in `fab/specs/`. `/fab-hydrate-design` provides assisted reverse-hydration — it identifies structural gaps between docs and design and proposes concise additions, but every insertion requires explicit user confirmation. Design files remain human-curated.

### Bootstrap Integration

`/fab-init` creates `fab/specs/index.md` during structural bootstrap (after docs/index.md). The creation is idempotent — if the file already exists, init skips it with a status message.

### Context Loading Integration

`fab/specs/index.md` is included in the "Always Load" context layer in `_context.md`, alongside `config.yaml`, `constitution.md`, and `fab/memory/index.md`. This gives every skill baseline awareness of the design landscape.

## Changelog

| Change | Date | Summary |
|--------|------|---------|
| 260211-r3k8-simplify-planning-stages | 2026-02-11 | Consistent design terminology, updated from specs-index references |
| 260209-h3v7-fab-backfill | 2026-02-09 | Updated Human-Curated Ownership section to reference `/fab-hydrate-design` as assisted reverse-hydration |
| 260207-bb1q-add-specs-index | 2026-02-07 | Initial creation — added `fab/specs/` directory, design index with boilerplate, bootstrap and context loading integration |
