# Intake: Standalone Operator4 Rewrite

**Change**: 260315-a2b2-standalone-operator4-rewrite
**Created**: 2026-03-15
**Status**: Draft

## Origin

> Conversational — extended `/fab-discuss` session exploring operator4's inheritance chain (operator1→2→3→4), skill-creator best practices, and the `_` file ecosystem. User identified that reading operator4 requires mentally merging ~800 lines across 4 files, and that the `_` file ecosystem has gaps (no external tool reference, no naming conventions file). Key decisions were made collaboratively about structure, naming, and branch handling.

## Why

1. **Operator4 is unreadable as-is.** Understanding it requires reading 4 files in sequence (operator1→2→3→4), mentally applying overrides (operator4 replaces operator3's two-tier answer model entirely), and discarding dead content (operator3's answer confidence section is read only to be overridden). An agent loading operator4 wastes context window on ~800 lines when ~300 are unique.

2. **The `_` file ecosystem has gaps.** External tools (`wt`, `tmux`, `/loop`) and naming conventions (branch, worktree, change folder patterns) are either undocumented in shared files or embedded in individual skills (`git-branch`, `git-pr`). Operator4 needs these conventions but has no shared file to import — it either reinvents them or inherits them through the operator chain.

3. **`_scripts.md` naming is misleading.** It documents fab CLI commands, not scripts. Renaming to `_cli-fab.md` aligns with its actual content and pairs naturally with the new `_cli-external.md`.

## What Changes

### 1. Rewrite `fab/.kit/skills/fab-operator4.md` as standalone

Remove the inheritance directive ("read operator2, then operator3, then return here"). Inline all behavior into a single self-contained skill file organized as:

1. **Principles** — what the operator is (coordination + auto-nudge proxy), what it isn't (not a lifecycle enforcer, not an implementer), routing discipline (coordinate don't execute), context discipline (never load change artifacts), state re-derivation (why: panes die, stages advance — stale state = wrong actions)
2. **Startup** — context loading (always-load layer), orientation (pane map → ready signal), degraded mode outside tmux
3. **Safety model** — confirmation tiers (read-only / recoverable / destructive), pre-send validation (pane exists + agent idle, why: dead pane errors silently, busy agent corruption), bounded retries & escalation table (why: unbounded retries compound errors), context discipline
4. **Monitoring system** — monitored set (fields, enrollment triggers, removal triggers), `/loop` lifecycle (start/extend/stop, one-loop invariant), monitoring tick with all 6 steps:
   - Stage advance detection
   - Pipeline completion detection
   - Review failure detection
   - Pane death detection
   - Auto-nudge (→ dedicated section)
   - Stuck detection (excludes input-waiting agents, why: waiting ≠ stuck)
5. **Auto-nudge** — question detection (capture `-l 20`, guards: Claude turn boundary / blank capture / idle-only, pattern matching: `?` on last non-empty line <120 chars with exclusions, additional indicator patterns, bottom-most indicator rule), answer model (simplified decision list items 1-6, all auto-answer except undeterminable keystrokes), re-capture before send (why: race condition between detection and send), per-answer logging
6. **Modes of operation** — shared rhythm (interpret → refresh → validate → execute → report → enroll) plus compact table: broadcast, sequenced rebase, merge PRs, spawn agent, status dashboard, unstick agent, notification, autopilot (→ dedicated section)
7. **Autopilot** — queue ordering (user-provided / confidence-based / hybrid), per-change loop (spawn → gate → monitor → merge → rebase → cleanup → progress), failure matrix, interruptibility (stop/skip/pause/resume), resumability, pipeline uses `/fab-fff`
8. **Configuration** — monitoring interval (5m), stuck threshold (15m), autopilot tick (2m), session-scoped
9. **Key properties** — standard table

**Writing style** (per skill-creator): explain the "why" behind constraints instead of heavy-handed MUSTs. Reserve imperatives for true safety constraints only. Target ~300 lines for the main file — tools, naming, and CLI references are all in shared `_` files.

### 2. Add `fab/.kit/skills/_cli-external.md`

New shared `_` file documenting external (non-fab) CLI tools:

- **wt** (worktree manager): `wt list`, `wt list --path <name>`, `wt create --non-interactive [--worktree-name <n>] [branch] [--base <ref>] [--reuse]`, `wt delete <name>`
- **tmux**: `tmux capture-pane -t <pane> -p [-l N]`, `tmux send-keys -t <pane> "<text>" Enter`, `tmux new-window -n <name> -c <dir> "<cmd>"`
- **`/loop`**: `/loop <interval> "<prompt>"` — recurring check skill

### 3. Add `fab/.kit/skills/_naming.md`

New shared `_` file with naming conventions extracted from `git-branch.md`, `git-pr.md`, and `docs/specs/naming.md`:

- **Change folder**: `{YYMMDD}-{XXXX}-{slug}` — generated by `fab change new`
- **Git branch**: identical to change folder name — created by `/git-branch`
- **Worktree directory**: `{adjective}-{noun}` — auto-generated by `wt create`, overridable via `--worktree-name`
- **Operator spawning rules**:
  - Known change (already exists): use change folder name as branch arg to `wt create`
  - New change (from backlog): `wt create` auto-generates worktree name, agent runs `/fab-new`, operator sends `/git-branch` after intake stage detected — this aligns the branch name

### 4. Rename `fab/.kit/skills/_scripts.md` to `fab/.kit/skills/_cli-fab.md`

Content unchanged. Name aligns with actual content (fab CLI reference) and pairs with `_cli-external.md`.

### 5. Update `_preamble.md` to reference new file names

- Change `_scripts.md` references to `_cli-fab.md`
- Add `_naming.md` to the always-load list (alongside `_cli-fab.md`)
- `_cli-external.md` is NOT added to always-load — it is loaded only by operator4's own startup section (wt/tmux/loop are operator-specific tools, not needed by pipeline skills)

### 6. Update `git-branch.md` and `git-pr.md`

Add a cross-reference to `_naming.md` at the top of each skill: `> Branch naming conventions are defined in _naming.md.` Keep all procedural logic intact — the procedures *apply* the naming conventions but are not duplication of them.

### 7. Delete `fab-operator1.md`, `fab-operator2.md`, `fab-operator3.md`

Delete outright — git history preserves them. Dead files in the skills directory risk agents loading them (especially operator1/2 which share the "operator" keyword). Sync script must also remove deployed copies from `.claude/skills/`.

### 8. Update `fab-sync.sh`

- Handle rename: deploy `_cli-fab.md`, clean up stale `_scripts.md` from `.claude/skills/`
- Deploy new files: `_cli-external.md`, `_naming.md`
- Remove deleted files: `fab-operator1.md`, `fab-operator2.md`, `fab-operator3.md` from `.claude/skills/`
- Invariant: deployed skill set in `.claude/skills/` must exactly match source set in `fab/.kit/skills/` — no more, no less

## Affected Memory

- `fab-workflow/execution-skills`: (modify) Update operator skill documentation to reflect standalone operator4
- `fab-workflow/kit-architecture`: (modify) Document `_` file ecosystem additions (`_cli-external.md`, `_naming.md`, `_cli-fab.md` rename)

## Impact

- **Skills**: `fab-operator4.md` (rewrite), `_cli-external.md` (new), `_naming.md` (new), `_scripts.md` → `_cli-fab.md` (rename), `_preamble.md` (update references), `git-branch.md` (cross-ref), `git-pr.md` (cross-ref), `fab-operator1.md` (delete), `fab-operator2.md` (delete), `fab-operator3.md` (delete)
- **Scripts**: `fab-sync.sh` (handle renames, new files, deletions — enforce exact-match invariant)
- **Specs**: `docs/specs/skills.md` should be updated to reflect the standalone operator4 and retired operator1-3
- **No behavioral changes**: The operator4 rewrite is a documentation/structure change — all runtime behavior (auto-nudge, monitoring, autopilot) is preserved exactly as-is

## Open Questions

None — all resolved via clarification.

## Clarifications

### Session 2026-03-15

| # | Action | Detail |
|---|--------|--------|
| 11 | Changed | Split: `_cli-fab.md` + `_naming.md` always-load; `_cli-external.md` operator-only load |
| 10 | Changed | Delete outright (git history preserves them; dead files risk ghost triggers) |
| Q3 | Resolved | Sync script updates included in scope — exact-match invariant for deployed skills |
| Q4 | Resolved | git-branch/git-pr keep procedures, add cross-reference to `_naming.md` only |
| Q5 | Changed | `_cli-external.md` NOT always-load — loaded only by operator4's startup section |

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Operator4 becomes fully standalone — no inheritance from operator1/2/3 | Discussed — user explicitly requested this as the primary goal | S:95 R:80 A:90 D:95 |
| 2 | Certain | `_preamble.md` stays intact (not split into framework + context) | Discussed — splitting saves ~60 lines but adds import overhead to every pipeline skill; not worth it | S:90 R:85 A:85 D:90 |
| 3 | Certain | `_scripts.md` renamed to `_cli-fab.md` | Discussed — user confirmed the `_cli-fab` / `_cli-external` split | S:95 R:90 A:90 D:95 |
| 4 | Certain | New `_cli-external.md` for wt, tmux, /loop — loaded only by operator4 | Discussed — user confirmed as its own file; clarified as operator-only, not always-load | S:95 R:85 A:90 D:95 |
| 5 | Certain | New `_naming.md` with conventions extracted from git-branch/git-pr — always-load | Discussed — user proposed this to share conventions with operator4 | S:95 R:85 A:90 D:95 |
| 6 | Certain | `_generation.md` unchanged | Discussed — not relevant to operator4, no changes needed | S:90 R:95 A:90 D:95 |
| 7 | Certain | Operator sends `/git-branch` after detecting new change creation | Discussed — user confirmed operator (not the agent) is responsible for this step | S:90 R:75 A:85 D:85 |
| 8 | Certain | Skill-creator writing style: explain why, not heavy-handed MUSTs | Discussed — incorporated from skill-creator analysis | S:85 R:85 A:80 D:85 |
| 9 | Confident | Operator4 target ~300 lines main file | Discussed — tools and naming offloaded to `_` files shrinks from ~350 to ~280-300 | S:75 R:90 A:80 D:80 |
| 10 | Certain | Delete operator1/2/3 outright | Clarified — user confirmed deletion; git history preserves; dead files risk ghost triggers via sync | S:95 R:70 A:90 D:95 |
| 11 | Certain | `_cli-fab.md` + `_naming.md` always-load; `_cli-external.md` operator-only | Clarified — user specified wt/tmux/loop are operator-specific, shouldn't load for pipeline skills | S:95 R:85 A:90 D:95 |
| 12 | Certain | Sync script updated — exact-match invariant (deployed = source, no more, no less) | Clarified — user confirmed sync updates in scope with this principle | S:95 R:80 A:90 D:95 |
| 13 | Certain | git-branch/git-pr keep procedures, add cross-reference to `_naming.md` | Clarified — procedures apply conventions, not duplication; only add a reference line | S:90 R:90 A:85 D:90 |

13 assumptions (12 certain, 1 confident, 0 tentative, 0 unresolved).
