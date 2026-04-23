# Intake: Merge Tasks and Checklist into a Single Plan Artifact

**Change**: 260423-qszh-merge-tasks-checklist
**Created**: 2026-04-23
**Status**: Draft

## Origin

> Discussion question (paraphrased): "tasks.md and checklist.md are both derived from spec — why two files?"

Raised during a `/fab-discuss` workflow retrospective on 2026-04-23, after ~2 months of production use. Analysis of `src/kit/skills/_generation.md` showed that Tasks Generation Procedure and Checklist Generation Procedure both take `spec.md` as input, both produce line-item lists, and differ primarily by consumer: apply reads tasks, review reads checklist. The two files are generated sequentially by the same orchestrator call (`/fab-continue` at spec-ready, `/fab-ff`/`/fab-fff` Steps 2+3) and have no independent lifecycle — you never touch one without the other.

**Related draft**: `260423-xvaz-skip-tasks-simple-types` — skips tasks generation for simple change types. Orthogonal but complementary. If this change ships first, that change rephrases as "skip `plan.md` generation for simple types" (checklist portion always required, tasks portion optional per policy). If that ships first, this change operates on whatever types still generate the tasks artifact. Agents picking this up should read that intake before starting.

## Why

**Problem**:
1. **Drift risk**: A spec requirement might be tracked as a task but missed from the checklist (or vice versa). Today, both generations are independent LLM passes over the same spec — the failure mode is silent gap in review coverage.
2. **Generation cost**: Two separate artifact-generation rounds for the same source material. Measured on recent changes, checklist generation is ~30–40% of the planning-stage wall time — duplicating work.
3. **Cognitive load**: Two files with overlapping but non-identical item IDs (T001 vs CHK-001), two templates, two parsers (apply vs review), two `.status.yaml` surfaces (`tasks` stage + `checklist.total/completed`).
4. **Review-rework friction**: When review fails, rework may touch tasks AND/OR checklist; agents sometimes update one and forget the other, causing a second review failure.

**Consequence if unfixed**: Continued silent drift + doubled generation cost on every pipeline run. As the project grows, the ratio of "ceremony-tax tokens" to "real work tokens" degrades further.

**Why this approach over alternatives**:
- **Alternative A — keep two files, add cross-check step**: adds more ceremony, not less. Agents already struggle to keep them in sync; a cross-check pass is another place to fail.
- **Alternative B — drop checklist, review reads tasks directly**: loses review's acceptance-criteria framing. Tasks are imperative ("do X"); checklist items are declarative ("X is done and correct"). Review wants the declarative framing.
- **Alternative C — merge into one file with two sections** (chosen): preserves the consumer distinction (apply reads `## Tasks`, review reads `## Acceptance`) while guaranteeing both sections are generated from the same spec in the same pass, with the same model, in the same context window. Drift becomes mechanically impossible at generation time.

## What Changes

### 1. New artifact: `plan.md`

Replaces `tasks.md` + `checklist.md`. Location: `fab/changes/{name}/plan.md`.

**Structure**:

```markdown
# Plan: {CHANGE_NAME}

**Change**: {YYMMDD-XXXX-slug}
**Status**: In Progress
**Intake**: [intake.md](./intake.md)
**Spec**: [spec.md](./spec.md)

## Tasks

<!-- Sequential work items for the apply stage. Checked off [x] as completed. -->

### Phase 1: Setup
- [ ] T001 [P] {description with file path}

### Phase 2: Core Implementation
- [ ] T002 {description with file path}

### Phase 3: Integration & Edge Cases
- [ ] T003 {description with file path}

## Execution Order

<!-- Non-obvious dependencies between tasks. Omit if order is self-evident. -->

## Acceptance

<!-- Declarative acceptance criteria used by the review stage. -->

### Functional Completeness
- [ ] CHK-001 {acceptance item derived from a spec requirement}

### Behavioral Correctness
- [ ] CHK-002 {item}

### Code Quality
- [ ] CHK-003 {item from code-quality.md Principles}

### Edge Cases & Error Handling
- [ ] CHK-004 {item}
```

Section headings (`## Tasks`, `## Acceptance`) are the parser contract — stable identifiers. Phase/category subheadings under each are presentational and may vary per change.

### 2. Unified generation procedure

`src/kit/skills/_generation.md` — replace **Tasks Generation Procedure** + **Checklist Generation Procedure** with a single **Plan Generation Procedure** that produces both sections in one pass over the spec. The procedure enumerates requirements once, then for each requirement emits:
- A Task entry (what to implement, in which file)
- An Acceptance entry (what must be true for review to pass)

Cross-linking via shared IDs is optional; the co-generation invariant is what guarantees alignment.

### 3. Apply behavior

`src/kit/skills/fab-continue.md` § Apply Behavior, Task Execution step 1:

Current:
> Parse tasks: `- [ ]` = remaining, `- [x]` = skip

New:
> Parse `plan.md` `## Tasks` section (everything between `## Tasks` and `## Acceptance` / `## Execution Order` / end-of-file). Apply ignores the `## Acceptance` section entirely. Checkboxes in `## Tasks` are marked `[x]` as tasks complete (behavior unchanged).

### 4. Review behavior

`src/kit/skills/_review.md` Preconditions:

Current:
> `tasks.md` and `checklist.md` MUST exist

New:
> `plan.md` MUST exist with both `## Tasks` and `## Acceptance` sections populated.

Inward sub-agent currently does:
> 2. Quality checklist: Inspect code/tests per CHK item. Mark `[x]` if met...

New:
> 2. Quality checklist: Inspect code/tests per CHK item in `plan.md` `## Acceptance` section. Mark `[x]` in place.

### 5. `.status.yaml` schema

Current:

```yaml
progress:
  tasks: done
checklist:
  generated: true
  total: 12
  completed: 12
```

Proposed:

```yaml
progress:
  plan: done       # renamed from `tasks`
plan:
  generated: true
  task_count: 8
  acceptance_count: 12
  acceptance_completed: 12
```

The `tasks` stage in progress is renamed to `plan` — small consumer-churn cost, but aligns artifact name and stage name. Rename is performed via a migration.

`fab status set-checklist` CLI → renamed/aliased to `fab status set-plan`. Existing command kept as deprecated alias for one release cycle.

### 6. Template

- Add `src/kit/templates/plan.md` (new, combined structure above).
- Deprecate `src/kit/templates/tasks.md` and `src/kit/templates/checklist.md`. Keep the files for one release cycle so deployed kits don't break mid-migration; remove in the next minor version.

### 7. Migration (required per Constitution § Additional Constraints)

New migration file at `src/kit/migrations/{NNN}-merge-tasks-checklist.md`:

Purpose: For every in-flight change (`fab/changes/*/` excluding `archive/`) that has `tasks.md` and/or `checklist.md` but no `plan.md`:
1. Read both files.
2. Produce `plan.md` by concatenating: tasks content under `## Tasks`, checklist content under `## Acceptance`.
3. Rewrite `.status.yaml`: rename `progress.tasks` → `progress.plan`, transform `checklist.{total,completed}` → `plan.{acceptance_count, acceptance_completed}`.
4. Leave `tasks.md` and `checklist.md` in place (do not delete) so users can verify the merge, with a one-line note appended: `<!-- Migrated to plan.md on {DATE} — safe to delete. -->`
5. Archived changes (`fab/changes/archive/**`) left untouched.

Idempotent: re-running the migration is a no-op when `plan.md` already exists.

### 8. Skill updates

- `src/kit/skills/fab-continue.md` — dispatch table row for `spec ready` collapses two actions (`generate tasks.md + checklist`) into one (`generate plan.md`). Apply preconditions and review preconditions updated as above.
- `src/kit/skills/fab-ff.md` / `fab-fff.md` — Steps 2 and 3 merge into a single **Step 2: Generate `plan.md`**.
- `src/kit/skills/fab-clarify.md` — target disambiguation: accept both `plan` and legacy `tasks` (alias). Scan `## Tasks` for task-level clarifications, `## Acceptance` for acceptance-level.
- `src/kit/skills/_generation.md` — merged procedure replaces the two existing ones.
- `src/kit/skills/_review.md` — precondition + parsing updates.
- `src/kit/skills/_preamble.md` — no direct change (State Table stages unchanged if we keep `plan` as the stage name).

### 9. Specs

- `docs/specs/templates.md` — replace tasks.md + checklist.md entries with plan.md entry.
- `docs/specs/skills.md` — per-skill flow updates.
- `docs/specs/overview.md` — stage list / artifacts per stage.

## Affected Memory

- `fab-workflow/templates`: (modify) — plan.md replaces tasks.md + checklist.md in template list
- `fab-workflow/planning-skills`: (modify) — unified plan generation
- `fab-workflow/execution-skills`: (modify) — apply reads plan.md `## Tasks`; review reads plan.md `## Acceptance`
- `fab-workflow/change-lifecycle`: (modify) — artifact file list
- `fab-workflow/schemas`: (modify) — `.status.yaml` `progress.plan` replaces `progress.tasks`
- `fab-workflow/migrations`: (modify) — new migration file registered
- `fab-workflow/hydrate`: (modify, if applicable) — hydrate may reference plan.md

## Impact

### Code
- `src/kit/skills/_generation.md` — merge two procedures into one
- `src/kit/skills/_review.md` — preconditions + parser
- `src/kit/skills/fab-continue.md` — dispatch + apply + review
- `src/kit/skills/fab-ff.md` — Step 2+3 merge
- `src/kit/skills/fab-fff.md` — Step 2+3 merge
- `src/kit/skills/fab-clarify.md` — target disambiguation + scan
- `src/kit/skills/_preamble.md` — State Table sanity check (stage named `tasks` → `plan` if renamed)

### Binary (`fab` Go CLI)
- `fab status set-plan` (new) — alias for set-checklist
- `fab status set-checklist` (deprecate, keep as alias) — emit deprecation notice
- `fab status finish/advance/reset/start` — accept `plan` stage name (alongside legacy `tasks` for backwards-compat during migration window)
- `fab change list` / `fab-status` display — reference `plan` stage

### Templates
- `src/kit/templates/plan.md` — new
- `src/kit/templates/tasks.md` — deprecate
- `src/kit/templates/checklist.md` — deprecate

### Migrations
- `src/kit/migrations/{NNN}-merge-tasks-checklist.md` — new

### Specs
- `docs/specs/templates.md`, `docs/specs/skills.md`, `docs/specs/overview.md`, `docs/specs/user-flow.md` (if it diagrams per-file artifacts)

## Open Questions

- Rename `progress.tasks` → `progress.plan` or keep the legacy name? Rename is clearer but has more consumer churn. Alternative: keep `progress.tasks` as the stage key, just change what artifact gets generated. Leaning: rename, per the "artifact name and stage name should match" design principle.
- Single file `plan.md` vs two files with a manifest linking them? Single file is simpler and matches the co-generation invariant. Two files re-introduces drift risk. Leaning: single file.
- Do we keep `tasks.md` + `checklist.md` templates indefinitely as "legacy format" or remove them after one release? Removal is cleaner; keeping forever encourages opt-out of the new flow. Leaning: remove after one release cycle with deprecation notice.
- Acceptance section IDs: keep `CHK-001` or renumber to `A001`? CHK is recognizable from 2 months of user habit. Leaning: keep CHK-.
- Does `fab score` need any update? (Scoring is spec-level, so probably not — but worth verifying the CLI doesn't read any tasks/checklist file directly.)

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Two files → one file eliminates drift risk at generation time | Tautologically true: a single-pass generation over the same context cannot produce divergent sub-artifacts | S:95 R:90 A:95 D:95 |
| 2 | Certain | Constitution § Additional Constraints mandates migration for user-data schema change | Direct constitution quote | S:95 R:95 A:95 D:95 |
| 3 | Certain | Apply and review need different sub-sections of the artifact | Apply = imperative tasks, review = declarative acceptance. Design principle preserved. | S:90 R:85 A:90 D:90 |
| 4 | Confident | `plan.md` is the right name | Matches SpecKit/industry naming; "plan" encompasses both tasks and acceptance cleanly | S:80 R:85 A:75 D:75 |
| 5 | Confident | `## Tasks` and `## Acceptance` are the stable parser contract | Heading-based parsing is what existing skills already use implicitly | S:80 R:80 A:85 D:75 |
| 6 | Confident | Rename `progress.tasks` → `progress.plan` is worth the consumer churn | Stage and artifact names should match; migration handles existing data | S:70 R:60 A:75 D:65 |
| 7 | Confident | `fab status set-checklist` becomes deprecated alias for `set-plan` | Standard deprecation pattern; one release cycle | S:75 R:75 A:80 D:70 |
| 8 | Confident | Migration leaves legacy `tasks.md` / `checklist.md` on disk with a note, doesn't delete | Safer for users to verify manually; cleanup is cheap later | S:80 R:80 A:80 D:75 |
| 9 | Tentative | Keep `CHK-` prefix for acceptance items (not renumber to `A-`) | Preserves 2 months of user visual memory; minor aesthetic cost | S:60 R:80 A:65 D:55 |
| 10 | Tentative | Remove legacy templates after one release cycle | Encourages migration; risk of breaking third-party tooling that reads them | S:55 R:60 A:65 D:50 |
| 11 | Tentative | Migration is idempotent no-op when `plan.md` exists | Standard migration pattern but must be explicitly verified in the migration file | S:65 R:75 A:70 D:60 |
| 12 | Unresolved | Should `fab-continue tasks` remain a valid reset target (alias for `plan`) or error? | Alias is friendlier; erroring forces users to learn new vocab. Depends on deprecation aggressiveness. | S:40 R:55 A:55 D:40 |
| 13 | Unresolved | Interaction with intake #1 (`260423-xvaz-skip-tasks-simple-types`): if that ships first, does this change rewrite the skip-policy key from `generate_tasks` → `generate_plan`? | Depends on ship order; agent picking this up must coordinate with other draft | S:40 R:50 A:55 D:40 |

13 assumptions (3 certain, 5 confident, 3 tentative, 2 unresolved).
