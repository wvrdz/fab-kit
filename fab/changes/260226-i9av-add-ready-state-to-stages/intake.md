# Intake: Add Ready State to Stage Lifecycle

**Change**: 260226-i9av-add-ready-state-to-stages
**Created**: 2026-02-26
**Status**: Draft

## Origin

> Add a `ready` state to the stage lifecycle: `pending → active → ready → done`. Currently `active` is overloaded — it means both "artifact is being generated" and "artifact exists and awaits advancement or clarification." This makes it impossible for agents to distinguish between resuming generation and advancing to the next stage.

Conversational — identified during a `/fab-continue` session where the agent confused the preflight `stage` field (forward-looking: "what to produce next") with the dispatch table's stage column (current-stage: "what stage is active"). The user proposed a three-state model (`pending → active → iterating → done`); discussion refined it to use `ready` instead of `iterating` for clearer semantics.

Key decisions from conversation:
- `ready` chosen over `iterating` — the semantic is "artifact produced, awaiting decision" not "ongoing work"
- The user wants to think more on edge cases within this change — the intake captures the agreed direction, spec stage will refine

## Why

1. **`active` is overloaded**: Currently, `spec: active` could mean either "spec.md is being generated right now" or "spec.md exists and the user can run `/fab-clarify` or `/fab-continue`." An agent resuming a session can't tell which. This caused a concrete bug: the `/fab-continue` dispatch table says `spec` → generate tasks.md (assuming spec.md exists), but preflight returns `stage: spec` when spec.md doesn't exist yet. The mismatch led to a failed `stageman.sh transition` call.

2. **Preflight/dispatch semantic mismatch**: Preflight's `stage` field is forward-looking ("what needs work"), while the dispatch table treats it as current-stage ("what's active"). Adding `ready` eliminates the ambiguity: `active` always means "work in progress", `ready` always means "work product exists, eligible for advancement."

3. **Resume clarity**: When an agent picks up a session, `active` = resume generation (maybe mid-stream), `ready` = artifact exists, ask user what to do next. No guessing.

## What Changes

### 1. New `ready` State in workflow.yaml

Add `ready` to the `states` list in `fab/.kit/schemas/workflow.yaml`. Update each stage's `allowed_states` to include `ready`:

```yaml
states:
  - id: pending
  - id: active
  - id: ready
  - id: done
  - id: failed
```

### 2. Stage Lifecycle: pending → active → ready → done

The transition flow becomes:

- `pending` → stage hasn't started
- `active` → work in progress (generating artifact, executing tasks)
- `ready` → stage's work product exists, eligible for:
  - `/fab-continue` → advance to next stage (`ready → done`, next stage `pending → active`)
  - `/fab-clarify` → deepen current artifact (stays `ready`)
- `done` → advanced past, locked

### 3. Transition Logic in stageman.sh

- `set-state` accepts `ready` as a valid state
- `transition` expects `from_stage` to be `ready` (not `active`) when advancing — this is the key behavioral change
- New convenience: `stageman.sh complete <file> <stage>` sets a stage from `active` to `ready` without advancing (for use after artifact generation)

### 4. Preflight Output Disambiguation

With the `ready` state, preflight's `stage` field becomes unambiguous:
- `stage: spec` + `spec: active` → agent should generate spec.md
- `stage: spec` + `spec: ready` → spec.md exists, `/fab-continue` advances to tasks

### 5. Skill Updates

All skills that transition stages need updating:
- `/fab-continue`: After generating an artifact, set to `ready` (not directly `done`). Separate "generate" dispatch from "advance" dispatch.
- `/fab-ff` and `/fab-fff`: Same pattern — generate → `ready` → advance in sequence
- `/fab-clarify`: Verify stage is `ready` (artifact exists) before scanning

### 6. Execution Stages

For apply/review/hydrate:
- `apply: active` → tasks being executed
- `apply: ready` → all tasks checked, eligible for review advancement
- `review: active` → sub-agent review in progress
- `review: ready` → review passed, eligible for hydrate advancement

## Affected Memory

- `fab-workflow/schemas`: (modify) document new `ready` state in workflow.yaml
- `fab-workflow/change-lifecycle`: (modify) update state vocabulary and transition rules
- `fab-workflow/planning-skills`: (modify) update stage transition descriptions
- `fab-workflow/execution-skills`: (modify) update apply/review/hydrate transition descriptions

## Impact

- **`fab/.kit/schemas/workflow.yaml`**: New state, updated allowed_states per stage
- **`fab/.kit/scripts/lib/stageman.sh`**: Transition logic, new `complete` subcommand
- **`fab/.kit/scripts/lib/preflight.sh`**: Disambiguation logic using ready state
- **`fab/.kit/skills/fab-continue.md`**: Split generate/advance dispatch
- **`fab/.kit/skills/fab-ff.md`**: Same transition pattern updates
- **`fab/.kit/skills/fab-fff.md`**: Same transition pattern updates
- **`fab/.kit/skills/fab-clarify.md`**: Stage guard uses `ready` instead of `active`
- **`fab/.kit/skills/_preamble.md`**: State table, context loading references

## Open Questions

- Should `/fab-continue` on a `ready` stage auto-advance without user confirmation, or should it prompt? (Current behavior with `active` auto-advances — keeping same UX seems right, but worth confirming.)
- Does `failed` need a parallel: `review: failed` means "go back to apply" — does it stay as-is or interact with `ready`?
- For `/fab-ff` and `/fab-fff`, should the pipeline pause at each `ready` state or auto-advance through? (Auto-advance seems right for pipeline commands — `ready` is primarily useful for `/fab-continue` single-step flow.)

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Four states: `pending → active → ready → done` | Discussed — user proposed three-state extension, naming refined from `iterating` to `ready` | S:95 R:70 A:85 D:90 |
| 2 | Certain | `ready` means "artifact exists, eligible for advancement or clarification" | Discussed — user specified the semantic: can fab-continue (advance) or fab-clarify (deepen) | S:90 R:70 A:85 D:90 |
| 3 | Certain | `active` means "work in progress, generation/execution ongoing" | Discussed — user specified: if active, we know artifact is still to be created | S:90 R:70 A:85 D:90 |
| 4 | Confident | `transition` requires `ready` (not `active`) for advancement | Logical consequence — prevents advancing before artifact exists. Major behavioral change but eliminates the root cause of the dispatch confusion | S:80 R:50 A:80 D:75 |
| 5 | Tentative | New `stageman.sh complete` subcommand for active→ready | Reasonable convenience to avoid two separate set-state calls, but may be unnecessary if skills just call `set-state <stage> ready` | S:60 R:85 A:70 D:55 |
| 6 | Tentative | Pipeline commands (`/fab-ff`, `/fab-fff`) auto-advance through `ready` | Pipeline semantics suggest no pause at `ready` — the pause is for single-step `/fab-continue`. But user hasn't confirmed | S:55 R:60 A:70 D:50 |
| 7 | Tentative | `failed` state unchanged — remains as-is alongside `ready` | No discussion yet — `failed` is only used by review, orthogonal to ready. But interaction deserves explicit confirmation | S:50 R:55 A:65 D:55 |

7 assumptions (3 certain, 1 confident, 3 tentative, 0 unresolved). Run /fab-clarify to review.
