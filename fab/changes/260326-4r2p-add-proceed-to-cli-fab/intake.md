# Intake: Add fab-proceed to _cli-fab.md

**Change**: 260326-4r2p-add-proceed-to-cli-fab
**Created**: 2026-03-26
**Status**: Draft

## Origin

> Add fab-proceed to _cli-fab.md — the new context-aware orchestrator skill was implemented in PR 278 but its entry was missed from the kit script invocation guide. Document it alongside the other skill references.

One-shot request. The user noticed the omission while reviewing the PR 278 merge (`1c956ac feat: fab-proceed orchestrator (#278)`).

## Why

`_cli-fab.md` is the authoritative reference loaded by every skill via `_preamble.md` to understand how to invoke CLI commands. When a new skill that invokes CLI commands ships without being documented here, the reference becomes incomplete. `/fab-proceed` calls `fab resolve --folder`, `fab change switch`, and relies on other skills that in turn call `fab status`, `fab score`, etc. Its calling patterns should be documented so that future skill authors and the `_cli-fab.md` reference itself remain accurate.

The constitution (§ Additional Constraints) requires: "Changes to the `fab` CLI (Go binary) MUST include corresponding test updates and MUST update `fab/.kit/skills/_cli-fab.md` with any new or changed command signatures." While fab-proceed didn't change CLI signatures, it introduced a new skill that consumes CLI commands — the spirit of this rule is that `_cli-fab.md` stays current as the definitive guide.

## What Changes

### Add fab-proceed to the Callers table under `fab log`

The `fab log` section has a Callers table (lines 210–218 of `_cli-fab.md`) listing which skills invoke `fab log command`. `/fab-proceed` itself is exempt from preflight (it delegates to `/fab-fff`), so it doesn't directly call `fab log command` via the preamble §2 path. However, if `/fab-proceed` invokes `/fab-new` as a subagent, that subagent will call `fab log command` via `fab change new --log-args`. This is already covered by the existing `fab change new` auto-log entry — no new Callers row needed for `fab log` specifically.

### Add fab-proceed to the `fab resolve` section

`/fab-proceed` directly invokes `fab resolve --folder` in its Step 1 (Active Change Check) to detect whether an active change exists. This is a notable consumer that should be documented.

### Add fab-proceed to the `fab change switch` reference

`/fab-proceed` dispatches `fab change switch` via a subagent when an unactivated intake is detected (dispatch table row 3). This usage should be noted.

### Cross-reference in Command Reference table

The Command Reference table at the top of `_cli-fab.md` lists commands and their purposes. No new commands were added by PR 278 — this table needs no changes. The documentation additions are to existing command sections, adding `/fab-proceed` as a caller/consumer.

## Affected Memory

- `fab-workflow/execution-skills`: (modify) Add fab-proceed's CLI invocation patterns to the execution skills memory

## Impact

- `fab/.kit/skills/_cli-fab.md` — primary file being updated
- No CLI commands changed — purely documentation additions
- No other skills affected — this is additive cross-referencing

## Open Questions

None — the scope is clear from the PR 278 diff and the existing `_cli-fab.md` structure.

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | No new CLI commands to document | PR 278 added a skill file only; no Go binary changes | S:95 R:95 A:95 D:95 |
| 2 | Certain | Document fab-proceed as a consumer of existing commands | _cli-fab.md tracks callers of CLI commands; fab-proceed calls fab resolve and dispatches fab change switch | S:90 R:90 A:90 D:90 |
| 3 | Confident | No new Callers row needed under fab log | fab-proceed doesn't call fab log command directly — its subagents do, but those are already covered by existing entries | S:80 R:85 A:80 D:75 |
| 4 | Confident | Affected Memory includes execution-skills | execution-skills.md was already updated in PR 278 but may need the CLI invocation detail added | S:75 R:85 A:70 D:80 |
| 5 | Certain | This is a docs change type | Only markdown documentation files are modified | S:95 R:95 A:95 D:95 |

5 assumptions (3 certain, 2 confident, 0 tentative, 0 unresolved).
