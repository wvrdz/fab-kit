# Intake: Expand Worktree Name Universe

**Change**: 260326-kpbc-expand-worktree-name-universe
**Created**: 2026-03-26
**Status**: Draft

## Origin

> Ensure randomness in worktree names is actually getting followed - increase the name universe even further

One-shot request to expand the word lists used for random worktree name generation (`adjective-noun` pattern), increasing the combinatorial space beyond the current 14,400 combinations.

## Why

1. **Collision risk grows with usage**: As more worktrees are created over time (especially in multi-agent operator workflows that spawn many concurrent worktrees), the 14,400-combination namespace increases the chance of retries in `GenerateUniqueName`. More combinations mean fewer retries and faster worktree creation.
2. **If left as-is**: The current 120x120 space works for moderate usage but becomes a limiting factor for power users running many parallel agents. The birthday paradox means collisions start appearing sooner than intuition suggests — with ~120 active worktrees, there's already a ~50% chance of a collision on any new name draw.
3. **Approach**: Expand both adjectives and nouns lists to ~200+ entries each, giving 40,000+ combinations. This is a pure additive change — no structural modifications needed.

## What Changes

### Word List Expansion in `src/go/wt/internal/worktree/names.go`

**Adjectives** (currently 120 → target ~200):
- Add new categories or expand existing ones with more entries
- Maintain the categorical grouping comment style for readability
- All words must be: short (1-3 syllables preferred), positive/neutral connotation, commonly understood English, no hyphens or special characters

**Nouns** (currently 120 → target ~200):
- Add more animal categories or expand existing ones
- Could also introduce non-animal nature nouns (e.g., river, summit, canyon, reef) to diversify
- Same constraints: short, recognizable, no special characters

### Test Update in `src/go/wt/internal/worktree/names_test.go`

- Update `TestWordListsNonEmpty` minimum thresholds from 120 to the new minimum (e.g., 200)
- Update `TestGenerateRandomName_Variety` comment to reflect new combo count
- The exhaustive collision test (`TestGenerateUniqueName_RetryExhaustion`) creates all combinations — verify it still runs in reasonable time with the larger space (200x200 = 40,000 dirs in temp). If too slow, consider capping or restructuring that test.

## Affected Memory

- `fab-workflow/distribution`: (modify) Update if worktree naming conventions are documented there

## Impact

- **`src/go/wt/internal/worktree/names.go`** — primary change: expand word lists
- **`src/go/wt/internal/worktree/names_test.go`** — update thresholds and validate perf
- No API changes, no config changes, no migration needed
- Backward compatible — existing worktree names are unaffected

## Open Questions

- None — the scope is clear and self-contained.

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Keep the adjective-noun format | Format is documented in specs and naming conventions; no reason to change it | S:90 R:95 A:95 D:95 |
| 2 | Certain | Expand both lists rather than just one | Combinatorial growth is multiplicative; expanding both lists maximizes the space increase | S:85 R:90 A:90 D:90 |
| 3 | Confident | Target ~200 entries per list (40,000+ combos) | Description says "even further" — doubling feels like a reasonable expansion. Could be 150 or 250 but 200 is a natural milestone | S:60 R:90 A:70 D:65 |
| 4 | Confident | Include only animals and nature nouns (no abstract nouns) | Existing list is all animals; nature nouns (river, canyon) fit the aesthetic. Abstract or tech nouns would break the style | S:55 R:85 A:80 D:70 |
| 5 | Certain | Exhaustive collision test may need restructuring | 200x200=40,000 temp dirs is slow; the test likely needs adjustment | S:80 R:90 A:85 D:85 |

5 assumptions (3 certain, 2 confident, 0 tentative, 0 unresolved).
