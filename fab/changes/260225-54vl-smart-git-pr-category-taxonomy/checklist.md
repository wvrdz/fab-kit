# Quality Checklist: Smart git-pr Category Taxonomy

**Change**: 260225-54vl-smart-git-pr-category-taxonomy
**Generated**: 2026-02-25
**Spec**: `spec.md`

## Functional Completeness
- [ ] CHK-001 PR Category Taxonomy: All 7 types (feat, fix, refactor, docs, test, ci, chore) are defined and referenced in the skill
- [ ] CHK-002 Type Resolution Chain: Three-step chain (explicit → intake → diff) is implemented in Step 0
- [ ] CHK-003 Two-Tier PR Template: Tier 1 (fab-linked) and Tier 2 (lightweight) templates are both defined in Step 3c
- [ ] CHK-004 PR Title Prefix: All PR titles use `{type}: {title}` format
- [ ] CHK-005 Fix Broken Artifact Links: Blob URLs use `https://github.com/{owner}/{repo}/blob/{branch}/...` format

## Behavioral Correctness
- [ ] CHK-006 Explicit argument override: `/git-pr chore` with a fab change folder still uses type `chore` (not feat/fix/refactor)
- [ ] CHK-007 Intake inference: Pattern matching is case-insensitive and uses the correct keyword lists (fix/bug/broken/regression → fix; refactor/restructure/consolidate/split/rename → refactor; else → feat)
- [ ] CHK-008 Diff inference: File path analysis correctly maps `.github/` → ci, `docs/` → docs, test dirs → test, else → chore
- [ ] CHK-009 Lightweight skip: Steps 4-4c gracefully skip when no fab change exists (no errors, no warnings)

## Scenario Coverage
- [ ] CHK-010 Scenario: User ships a feature — fab change with intake/spec → type inferred, Tier 1 template, blob URL links
- [ ] CHK-011 Scenario: User ships a chore without fab folder — no fab change → type `chore`, Tier 2 template, "no design artifacts" note
- [ ] CHK-012 Scenario: Explicit type override — `/git-pr docs` → type `docs`, Tier 2 template regardless of fab state
- [ ] CHK-013 Scenario: PR from feature branch — blob URLs resolve against feature branch, not main

## Edge Cases & Error Handling
- [ ] CHK-014 Invalid type argument: If user passes an unrecognized type (e.g., `/git-pr foo`), the skill should handle gracefully (either reject with valid types list or fall back to inference)
- [ ] CHK-015 Missing `gh` CLI: Blob URL construction depends on `gh repo view`; if `gh` is missing, the existing "gh CLI not found" guard catches this before URL construction
- [ ] CHK-016 Spec.md conditional: Tier 1 template only includes spec link if `spec.md` exists in the change folder

## Code Quality
- [ ] CHK-017 Pattern consistency: New skill sections follow existing formatting (### headings, numbered steps, code blocks for commands)
- [ ] CHK-018 No unnecessary duplication: Type resolution logic is defined once in Step 0, not duplicated in Step 3c

## Documentation Accuracy
- [ ] CHK-019 Type reference table: Inline reference lists all 7 types with descriptions and tier mapping
- [ ] CHK-020 Existing step references: Steps 1, 1b, 2, 4, 4b, 4c remain accurate after changes

## Cross References
- [ ] CHK-021 Autonomous contract: Skill remains "no prompts, no questions" — type resolution never asks the user
- [ ] CHK-022 Constitution compliance: No binary formats, no build steps — changes are pure markdown
