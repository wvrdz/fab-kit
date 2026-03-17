# Intake: Operator4 Branch Fallback

**Change**: 260317-yrgo-operator4-branch-fallback
**Created**: 2026-03-17
**Status**: Draft

## Origin

> Operator4 branch fallback: when the operator can't find a change via fab resolve, it should automatically search git branch names as a fallback. This is a skill-level behavior change to operator4 only — no CLI changes.

Conversational `/fab-discuss` session. The user initially considered adding branch-aware resolution to `fab resolve` (CLI level), but after discussion concluded that:

1. `fab` is orthogonal to git — mixing branch resolution into the CLI blurs boundaries
2. Branch name searching is trivially achievable by the agent (`git branch --list "*query*"`)
3. The fix belongs in operator4's behavior (skill level), not the CLI

Key rejected alternatives:
- **`fab resolve --search-branches`** — rejected because fab operates on change folders (filesystem/YAML), not git branches. Adding branch awareness to the CLI would complicate every caller expecting a folder path.
- **`fab resolve --branch` output mode** — rejected for the same boundary reason. Also would create awkward edge cases (e.g., `--dir` when a change only exists as a branch).
- **Automatic fallback in `fab resolve`** — rejected because it could surprise callers expecting only local results, and adds performance cost to every failed resolve.

## Why

1. **Problem**: When a change folder exists only in a git branch (not checked out in any worktree), the operator can't find it via `fab resolve`. The user must manually tell the operator which branch contains the change. This breaks the operator's "coordinate, don't execute" model — users shouldn't need to be the operator's search layer.

2. **Consequence**: The operator fails silently or reports "cannot resolve change" when the change clearly exists — just in a branch. The user has to context-switch from giving high-level instructions to debugging resolution failures and providing branch names.

3. **Approach**: The operator already re-derives state before every action (Section 4 principle: "state re-derivation"). Adding a branch name scan as a fallback when local resolution fails is a natural extension of this principle. Branch names follow the same `YYMMDD-{id}-{slug}` naming convention as change folders (`_naming.md`), so the same substring/ID matching works.

## What Changes

**Important**: Do NOT modify `fab/.kit/skills/fab-operator4.md`. This change creates a new **`fab-operator5`** skill that is operator4's full content plus the branch fallback behavior described below. Operator4 remains untouched as the previous version.

### New skill: `fab-operator5` (operator4 + branch fallback)

Create `fab/.kit/skills/fab-operator5.md` containing all of operator4's current content, plus a new subsection in Section 3 (Safety Model) for branch fallback resolution. The new behavior activates when `fab resolve` fails to find a change locally.

**Trigger**: User-initiated resolution only. Does not apply during monitoring ticks (pane death already handled by monitored set removal).
<!-- clarified: trigger scope — user-initiated only, not monitoring ticks -->

**Resolution flow** (after `fab resolve` returns non-zero):

1. Scan local and remote branch names:
   ```bash
   git for-each-ref --format='%(refname:short)' refs/heads/ refs/remotes/ | grep -i "<query>"
   ```
   <!-- clarified: searches both local and remote branches -->
2. If exactly one match: report the finding and choose response based on user intent:
   - **Read-only query** (status check, "what stage is X at?"): read `.status.yaml` directly from the branch via `git show <branch>:fab/changes/<folder>/.status.yaml` — no worktree needed
   - **Action query** (send command, resume work): offer to create a worktree
   ```
   Can't find {query} in any worktree. Found branch `{branch}`. Create a worktree for it?
   ```
   <!-- clarified: read-only queries use git show, worktree creation only for action queries -->
3. If multiple matches: present disambiguation list
4. If no match: report "not found locally or in any branch" (current behavior, just more explicit)

**After user confirms worktree creation**, the operator follows the existing "Known change" spawning rule from `_naming.md`:
```
wt create --non-interactive --worktree-name <name> <branch-name>
```

### Placement in the skill file

New subsection in **Section 3: Safety Model**, after "Pre-Send Validation" and before "Bounded Retries & Escalation". The branch fallback is structurally similar to pre-send validation — a precondition check before the operator can act on a change.
<!-- clarified: placement — Section 3 Safety Model, after Pre-Send Validation -->

The resolution fallback applies to user-initiated operator actions that need to resolve a change — status queries, command routing, spawn. It does not apply during monitoring ticks.

### Cleanup: delete legacy operator scripts

Delete the following obsolete launcher scripts:

- `fab/.kit/scripts/fab-operator1.sh`
- `fab/.kit/scripts/fab-operator2.sh`
- `fab/.kit/scripts/fab-operator3.sh`

These are leftovers from the operator inheritance chain (operator1→2→3→4) that was replaced by the standalone operator4 rewrite. `fab/.kit/scripts/fab-operator4.sh` remains as the active launcher.

## Affected Memory

- `fab-workflow/execution-skills`: (modify) Add branch fallback behavior to operator4 documentation

## Impact

- **New skill file**: `fab/.kit/skills/fab-operator5.md` — operator4's full content + branch fallback subsection
- **New spec file**: `docs/specs/skills/SPEC-fab-operator5.md` — corresponding spec
- **Operator4 untouched**: `fab/.kit/skills/fab-operator4.md` and `docs/specs/skills/SPEC-fab-operator4.md` remain as-is
- **No CLI changes**: `fab resolve`, `fab change`, and all other `fab` subcommands remain unchanged
- **No template changes**: No new templates or status fields
- **Deleted scripts**: `fab/.kit/scripts/fab-operator{1,2,3}.sh` — legacy launchers from the inheritance chain
- **No migration needed**: Pure skill addition + cleanup

## Open Questions

- Where exactly in the skill file should this behavior live — as part of an existing section or a new section?

## Clarifications

### Session 2026-03-17

| # | Action | Detail |
|---|--------|--------|
| 7 | Clarified | Trigger scope: user-initiated resolution only, not monitoring ticks |
| 8 | Clarified | Search both local and remote branches |
| 9 | Clarified | Read-only queries use `git show`; worktree creation only for action queries |
| 6 | Clarified | Placement: Section 3 (Safety Model), after Pre-Send Validation |

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | No CLI changes — branch fallback is skill-level only. Create new operator5, do not modify operator4 | Discussed — user explicitly rejected CLI-level approaches; later clarified to create operator5 rather than edit operator4 | S:95 R:90 A:95 D:95 |
| 2 | Certain | Use `git for-each-ref` + grep for branch name scanning | Discussed — user confirmed branch name matching is the right approach, not branch content inspection | S:90 R:95 A:90 D:90 |
| 3 | Certain | Same substring/ID matching as existing resolution | Branch names follow `YYMMDD-{id}-{slug}` convention per `_naming.md`, same as folder names | S:85 R:90 A:95 D:95 |
| 4 | Confident | Offer worktree creation when a branch match is found | Natural next step per existing operator spawning rules in `_naming.md` — but user didn't explicitly confirm this UX | S:75 R:80 A:85 D:80 |
| 5 | Confident | Single-match auto-report, multi-match disambiguation | Standard resolution pattern consistent with `fab resolve` behavior | S:70 R:85 A:85 D:80 |
| 6 | Certain | Place as new subsection in Section 3 (Safety Model), after Pre-Send Validation | Clarified — user chose option 1 | S:95 R:90 A:60 D:50 |
| 7 | Certain | Branch fallback triggers on user-initiated resolution only, not monitoring ticks | Clarified — user agreed. Monitoring already handles pane death via set removal | S:95 R:90 A:90 D:95 |
| 8 | Certain | Search both local (`refs/heads/`) and remote (`refs/remotes/`) branches | Clarified — user requested remote branch search | S:95 R:85 A:90 D:95 |
| 9 | Certain | Read-only queries use `git show` for status; worktree creation only for action queries | Clarified — user agreed. Response proportional to the ask | S:95 R:90 A:90 D:90 |
9 assumptions (7 certain, 2 confident, 0 tentative).
