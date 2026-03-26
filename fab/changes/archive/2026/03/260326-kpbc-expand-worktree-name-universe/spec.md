# Spec: Expand Worktree Name Universe

**Change**: 260326-kpbc-expand-worktree-name-universe
**Created**: 2026-03-26
**Affected memory**: `docs/memory/fab-workflow/distribution.md`

## Non-Goals

- Changing the `adjective-noun` naming format — the pattern is documented and stable
- Adding city names as nouns — cultural bias, length, and aesthetic mismatch
- Adding statistical randomness tests — Go's `math/rand` provides uniform distribution; testing it would test the stdlib, not our code
- Restructuring `GenerateRandomName()` or `GenerateUniqueName()` functions — only the word lists change

## Worktree Naming: Word List Expansion

### Requirement: Adjective List Size

The `adjectives` slice in `src/go/wt/internal/worktree/names.go` SHALL contain at least 200 entries.

#### Scenario: Adjective count meets minimum
- **GIVEN** the `names.go` file is compiled
- **WHEN** `len(adjectives)` is evaluated
- **THEN** the result SHALL be >= 200

### Requirement: Noun List Size

The `nouns` slice in `src/go/wt/internal/worktree/names.go` SHALL contain at least 200 entries.

#### Scenario: Noun count meets minimum
- **GIVEN** the `names.go` file is compiled
- **WHEN** `len(nouns)` is evaluated
- **THEN** the result SHALL be >= 200

### Requirement: Combinatorial Space

The total name space (`len(adjectives) * len(nouns)`) SHALL be at least 40,000.

#### Scenario: Name space exceeds target
- **GIVEN** adjectives has >= 200 entries and nouns has >= 200 entries
- **WHEN** the product is computed
- **THEN** the result SHALL be >= 40,000

### Requirement: Adjective Expansion Strategy

New adjectives SHALL primarily expand existing categories (Quality & character, Nature & space, Temperament & energy, Sensory & aesthetic, Motion & state, Abstract positive) with ~8-10 additional entries per category. New categories (e.g., Time & weather, Texture & material) MAY be added to reach the target.

#### Scenario: Existing categories grow
- **GIVEN** the current 6 adjective categories each have ~16-24 entries
- **WHEN** new adjectives are added
- **THEN** each existing category SHOULD have at least 5 additional entries
- **AND** the categorical grouping comment style SHALL be maintained

### Requirement: Noun Expansion Strategy

New nouns SHALL primarily expand existing animal categories with ~4-5 additional entries per category. One new nature/geography category SHALL be added with ~15-20 entries (e.g., `river`, `summit`, `canyon`, `reef`, `grove`, `meadow`, `delta`, `ridge`).

#### Scenario: Existing animal categories grow
- **GIVEN** the current 15 noun categories each have ~8 entries
- **WHEN** new nouns are added
- **THEN** each existing animal category SHOULD have at least 3 additional entries
- **AND** the categorical grouping comment style SHALL be maintained

#### Scenario: Nature/geography category added
- **GIVEN** the current nouns are all animals
- **WHEN** the noun list is expanded
- **THEN** a new "Nature & geography" category SHALL be added
- **AND** it SHALL contain at least 15 entries
- **AND** entries SHALL be short (1-2 syllables preferred), universally recognizable nature/geography terms

### Requirement: Word Quality Constraints

All new words (adjectives and nouns) MUST meet these criteria:
- Short: 1-3 syllables preferred, no words longer than 4 syllables
- Recognizable: commonly understood English words
- Positive/neutral connotation: no negative, violent, or offensive words
- No hyphens, spaces, or special characters
- No duplicates within either list

#### Scenario: Word validation
- **GIVEN** a candidate word for either list
- **WHEN** it is evaluated for inclusion
- **THEN** it MUST be a single lowercase ASCII word with no special characters
- **AND** it MUST NOT already exist in the same list

### Requirement: Comment Style Preservation

The categorical grouping comments (e.g., `// Canids & small mammals`) SHALL be preserved and extended for new entries. New categories SHALL follow the same comment pattern.

#### Scenario: New category comment format
- **GIVEN** a new noun category "Nature & geography" is added
- **WHEN** the category is written in `names.go`
- **THEN** it SHALL have a comment in the format `// Nature & geography`
- **AND** entries SHALL follow on subsequent lines in the same 8-per-line comma-separated format

## Worktree Naming: Test Updates

### Requirement: Updated Minimum Thresholds

`TestWordListsNonEmpty` in `src/go/wt/internal/worktree/names_test.go` SHALL assert `len(adjectives) >= 200` and `len(nouns) >= 200` (up from 120).

#### Scenario: Test threshold enforces new minimum
- **GIVEN** the updated test file
- **WHEN** `TestWordListsNonEmpty` runs
- **THEN** it SHALL fail if either list has fewer than 200 entries

### Requirement: Updated Variety Comment

The comment in `TestGenerateRandomName_Variety` SHALL be updated to reflect the new combo count (200*200=40000+).

#### Scenario: Comment accuracy
- **GIVEN** the updated test file
- **WHEN** `TestGenerateRandomName_Variety` is read
- **THEN** the comment SHALL reference the updated combinatorial space (e.g., "200*200=40000")

### Requirement: Exhaustive Collision Test Performance

The `TestGenerateUniqueName_RetryExhaustion` test creates all possible name combinations as directories. With 40,000+ combinations, this test SHOULD be restructured to avoid creating all combinations if it becomes too slow (> 30 seconds).

#### Scenario: Collision test with large namespace
- **GIVEN** 200+ adjectives and 200+ nouns (40,000+ combinations)
- **WHEN** `TestGenerateUniqueName_RetryExhaustion` runs
- **THEN** it SHALL still verify that exhaustion is detected
- **AND** it SHOULD complete in under 30 seconds

#### Scenario: Test restructuring approach
- **GIVEN** creating 40,000+ temp directories is too slow
- **WHEN** the test is restructured
- **THEN** it MAY use a smaller subset or mock the filesystem check
- **AND** it MUST still verify the retry exhaustion error path

## Design Decisions

1. **Expand existing categories over adding new category types**: Keeps the established aesthetic consistent. Each existing group has natural room to grow (many recognizable animals/adjectives remain). User confirmed this approach.
   - *Why*: Maintains cohesive naming style without forced categories
   - *Rejected*: Adding many new category types (e.g., mythological creatures, foods) — dilutes the aesthetic

2. **One new nature/geography noun category**: Animals alone can reach 200, but nature terms (river, canyon, reef) add semantic diversity while fitting the established vibe. User agreed.
   - *Why*: Short, universal, aesthetically compatible with animal names
   - *Rejected*: City names — cultural bias, multi-syllable, aesthetic clash

3. **No randomness testing needed**: Go's `math/rand.Intn` provides uniform distribution over slice indices. The existing variety test is sufficient as a sanity check.
   - *Why*: Would be testing Go's stdlib, not our code
   - *Rejected*: Chi-squared or distribution tests — unnecessary complexity

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Keep adjective-noun format | Confirmed from intake #1 — documented convention, no reason to change | S:95 R:95 A:95 D:95 |
| 2 | Certain | Expand both lists symmetrically | Confirmed from intake #2 — multiplicative combinatorial growth | S:90 R:90 A:90 D:90 |
| 3 | Certain | Target ~200 entries per list | Discussed — user confirmed ~200 × ~200 = ~40,000 as the target | S:90 R:90 A:85 D:90 |
| 4 | Certain | Animals + nature/geography nouns only | Discussed — user agreed on animals + one nature/geo category, rejected cities | S:90 R:85 A:90 D:90 |
| 5 | Certain | Primarily expand existing categories | Discussed — user chose Option B (expand existing) over Option A (new category types) | S:90 R:90 A:85 D:90 |
| 6 | Certain | No statistical randomness tests | Discussed — user agreed Go's math/rand is sufficient | S:90 R:95 A:90 D:95 |
| 7 | Confident | Collision test may need restructuring for 40K dirs | Spec-level analysis — 40K temp dirs may be slow; approach TBD at implementation | S:70 R:85 A:75 D:70 |
| 8 | Certain | New adjective categories: time/weather, texture/material | Discussed — user agreed to expand A symmetrically with B | S:85 R:90 A:85 D:85 |

8 assumptions (7 certain, 1 confident, 0 tentative, 0 unresolved).
