# Intake: Window prefix primitives and done-marker on removal

**Change**: 260423-rxu3-window-prefix-primitives
**Created**: 2026-04-23
**Status**: Draft

## Origin

Conversational follow-up to `260422-jyyg-operator-prefix-enrolled-windows` (merged PR #345, 2026-04-22). After that change shipped, the user raised a design question during the closing discussion:

> Should the chevron be removed once the tracking is over?

The working conclusion reached in conversation:

- Simply **removing** `»` on removal reintroduces the signal-staleness concern (the tab bar lies about what's currently tracked), but the original "no restore" rule protected against a real edge case (user renaming the window mid-monitoring and our restore clobbering their intent).
- The user proposed **replacing** `»` with a different single-width BMP character on removal — e.g., `›` (U+203A, single guillemet, nice parity with `»`) or `✓` (U+2713, explicit "done"). Keeps the tab bar honest AND retains the "operator-touched" trail.
- The user then asked whether this should be **mechanized** into `fab pane` subcommands rather than inline tmux shell in the skill, noting the operator's §4 algorithm is already copy-paste-prone prose.

I argued for a **minimum set of two verbs** rather than four:

- `ensure-prefix <pane> <char>` — idempotent add. If name starts with `<char>`, no-op; else prepend.
- `replace-prefix <pane> <from> <to>` — atomic swap guarded on current name starting with `<from>`. Empty `<to>` gives removal.

Add/remove as separate verbs are redundant: `add-prefix` without idempotency is almost never what you want (double-prefixing), and `remove-prefix` collapses into `replace-prefix X ""`. Two verbs cover every use case discussed (enrollment, done-marker swap, potential future pause-marker).

The user responded: "Ok, create a draft using /fab-draft here" — hence this intake as a draft (not activated).

## Why

**Problem 1 — Signal staleness.** `260422-jyyg` extended the `»` prefix to all enrolled windows but kept it forever. After a monitored change reaches a terminal stage (hydrate / ship / review-pr) and leaves the monitored set, the window still shows `»`. A user glancing at the tab bar cannot distinguish "actively tracked" from "operator-touched-once-upon-a-time." For a convention whose entire purpose is *at-a-glance coordination*, this is a self-inflicted UX degradation.

**Problem 2 — Prose-encoded primitive.** The enrollment rename is described as four lines of shell in `src/kit/skills/fab-operator.md` §4:

```sh
name=$(tmux display-message -p -t <pane> '#W')
case "$name" in »*) ;; *) tmux rename-window -t <pane> "»${name}" ;; esac
```

If any second skill (or a third, or the operator itself for the done-marker transform) needs the same guarded-rename semantics, that shell will get copied and drift. The existing `fab pane` command group exists *specifically* so skills call tested primitives instead of inline shell — cf. `fab pane-map`, `fab pane capture`, `fab pane send`, `fab pane process`, `fab operator tick-start`, `fab operator time`. The rename algorithm belongs in that group, not in prose.

**Why not fix just one.** The fixes are coupled:

- Replace-on-removal needs a guarded atomic swap that is awkward to express inline (read current name → prefix-match → rewrite preserving suffix → single tmux rename). Inline shell doubles the drift surface from today's one-site algorithm.
- Adding `fab pane window-name` verbs without the replace-on-removal behavior ships infrastructure with only one user (the existing enrollment site), which rightly fails the "build the minimum" test. Bundling gives the primitive two use sites on day one, validating the abstraction.

**Rejected alternatives** (from the conversation):

- *Do nothing.* Leaves the signal-staleness complaint unaddressed. Not ruled out on correctness grounds, but the user explicitly opened the reconsideration.
- *Remove `»` on removal (naive restore).* Reintroduces the clobber risk for user-renamed windows. Ruled out in prior spec (see `260422-jyyg/spec.md` Design Decision 4).
- *Store `original_name` on the monitored entry and restore.* Adds schema state and an ambiguity ("user renamed at t=now+5m — is the original still the intent?"). Avoided by the current-prefix-check approach: transform only if the current name still starts with `»`, which implicitly detects user drift without storing anything.
- *Four-verb set (add / remove / replace / ensure).* Redundancy noted above. Two verbs suffice.
- *Keep the algorithm inline.* Prose-encoded primitives copy-paste drift; the operator file will accumulate two variants of the guarded-rename if we do nothing, and more if other skills want the same.

## What Changes

### 1. New `fab pane window-name` subcommand group

Two verbs, both tmux-dependent (fail cleanly if `$TMUX` unset or the pane ID does not resolve):

#### `fab pane window-name ensure-prefix <pane> <char>`

Idempotent prepend.

- **Behavior**: Read the current window name via `tmux display-message -p -t <pane> '#W'`. If it starts with `<char>` (literal string prefix, no regex): no-op, exit 0 with no stdout. Else: run `tmux rename-window -t <pane> "<char><current-name>"`, exit 0 with stdout `renamed: <old> -> <new>`.
- **Errors**: `$TMUX` unset → exit 1 with `tmux not running`. Pane doesn't exist → exit 2 with tmux's stderr. Rename fails for any other reason → exit 3 with tmux's stderr.
- **Arguments**: `<pane>` is a tmux pane ID (e.g., `%3`, `%foo`), same format accepted by the existing `fab pane send` / `fab pane capture` commands. `<char>` is any non-empty string; the command does not restrict to single characters.

#### `fab pane window-name replace-prefix <pane> <from> <to>`

Atomic guarded swap.

- **Behavior**: Read the current window name. If it starts with `<from>` (literal string prefix): run `tmux rename-window -t <pane> "<to><name-without-from-prefix>"`, exit 0 with stdout `renamed: <old> -> <new>`. If it does not start with `<from>`: no-op, exit 0 with no stdout (the guard protects user-renamed windows without any additional state).
- **Empty `<to>` (= removal)**: Supported. Strips the `<from>` prefix.
- **Errors**: Same triad as `ensure-prefix`.
- **Arguments**: `<pane>` as above. `<from>` must be non-empty. `<to>` may be empty.

Source location (following existing patterns): `src/go/fab/cmd/fab/pane_window_name.go` with a subroutine in `internal/pane/` if the tmux read-and-rename logic is worth factoring. Wire into the existing `fab pane` cobra group defined in `src/go/fab/cmd/fab/pane.go`.

#### Parent group help

Extend `fab pane`'s help output to list `window-name` alongside the existing four subcommands (`map`, `capture`, `send`, `process`). Keep alphabetical order if the existing list uses it; otherwise match the current ordering convention.

### 2. Operator uses the primitives at two sites

Update `src/kit/skills/fab-operator.md`:

**Site A — §4 Enrollment bullet**. Replace the existing inline `tmux display-message` + `case` + `tmux rename-window` block with a single call:

```sh
fab pane window-name ensure-prefix <pane> »
```

…inheriting the idempotent semantics from the subcommand. The existing skipped-rename log line format and the "enrollment is already durable" property are preserved — the operator treats a non-zero exit from `ensure-prefix` the same way it currently treats a non-zero exit from the raw `tmux rename-window`: log `"{change}: window rename skipped ({error})."` and continue.

**Site B — §4 Removal bullet (new behavior)**. Remove the current "window name is not restored" rule. In its place: on every removal path (terminal stage, stop_stage, pane death, explicit stop) the operator calls

```sh
fab pane window-name replace-prefix <pane> » <done-char>
```

…where `<done-char>` is the chosen done-marker. The guard inside `replace-prefix` handles the pane-death case (the command will exit non-zero; operator logs and continues) AND the user-rename-mid-monitoring case (if the user changed the name so it no longer starts with `»`, the swap is a no-op — nothing gets clobbered).

Update §6 step 4's parenthetical to reference the `ensure-prefix` primitive instead of the inline shell, and add a brief mention of the removal swap to §4's Removal bullet.

### 3. Done-marker character selection

Two candidates surfaced in conversation, both single-width BMP (per the 260328/260416 decisions — no SMP emoji, no double-width glyphs):

- `›` (U+203A, SINGLE RIGHT-POINTING ANGLE QUOTATION MARK) — direct parity with `»` (U+00BB, DOUBLE), visually lighter, reads as "was-chevron, now reduced." Keeps the guillemet family.
- `✓` (U+2713, CHECK MARK) — strong, unambiguous "done" semantics. Trade-off: breaks the guillemet family; could clash with any other user convention that already uses `✓`.

**Tentative lean**: `›` — the semantic is "monitoring complete, trail preserved," not "task done." The completion signal for a change is already `✓` in the operator status frame (`● apply → review ✓`), so reusing `✓` on window names creates a light visual collision.

Spec stage will lock this in. Open question #1 below.

### 4. Memory updates (hydrate stage)

- `docs/memory/fab-workflow/execution-skills.md` — rewrite §4 Enrollment/Removal paragraphs shipped in 260422-jyyg; revise the "`»` Prefix Extends to Enrolled Windows" Design Decision (or add a successor entry) to capture the replace-on-removal and primitive-extraction decisions; add changelog entry.
- `docs/memory/fab-workflow/pane-commands.md` — add a `### Subcommand: fab pane window-name` section mirroring the existing `map` / `capture` / `send` / `process` sections (behavior, source path, errors). Update the parent-group summary at the top to list five subcommands instead of four.
- `docs/memory/fab-workflow/index.md` — date bump for both affected files.

### 5. Spec file sync

Per constitution "Additional Constraints", `docs/specs/skills/SPEC-fab-operator.md` gets its one-bullet Monitoring System line updated to reflect the replace-on-removal behavior (replacing the current "Removal does not restore the original name"). No broader rewrite.

## Affected Memory

- `fab-workflow/execution-skills`: (modify) — rewrite §4 Enrollment/Removal paragraphs; revise/extend the `»` prefix Design Decision; add changelog entry.
- `fab-workflow/pane-commands`: (modify) — add `fab pane window-name` subcommand section (first new entry since the 260419 rewrite); update parent-group subcommand count.

## Impact

- **Go binary**: new subcommand group, likely ~80–120 lines in `src/go/fab/cmd/fab/pane_window_name.go` plus a test file. Possibly small shared helper in `internal/pane/` if the tmux read-and-rename sequence is worth factoring. `fab_version` bump (minor: adds a new subcommand group without breaking existing surface).
- **Skill file**: `src/kit/skills/fab-operator.md` — §4 Enrollment paragraph replaces 4 lines of shell with 1 line of `fab pane`; §4 Removal bullet gets new rule; §6 step 4 parenthetical updated.
- **Spec file**: `docs/specs/skills/SPEC-fab-operator.md` — one-bullet Monitoring System line updated.
- **Memory files**: `execution-skills.md` + `pane-commands.md` + `index.md`.
- **Schema**: none. The `replace-prefix` guard on `<from>` means no `original_name` field is needed on monitored entries — state stays where it is.
- **Migration**: none at the user-visible level — the existing `»` prefix on live monitored windows remains valid and the new enrollment step is idempotent over it. The first run of the new operator on a windowscape that already has `»`-prefixed names does nothing unexpected.
- **Backwards compatibility**: `fab pane` gains a new subcommand; existing subcommands untouched. Old skill versions that still inline the tmux shell continue to work.

## Open Questions

1. **Done-marker character.** `›` (U+203A, guillemet parity) vs `✓` (U+2713, explicit done) vs something else (`·`, `¦`, `‥`)? See §3 above for tentative lean toward `›`.
2. **Should the primitives validate `<char>` / `<from>` / `<to>` for width?** The 260328/260416 decisions require single-width BMP characters for tmux tab rendering. Enforcing that in the Go subcommand (reject multi-codepoint or double-width input) is one option; leaving it to the caller is another. Lean toward *no validation* in the primitive — the verb is a mechanical shell-replacement tool, not a style enforcer — and keeping the single-width requirement as operator-skill guidance.
3. **Should `ensure-prefix` and `replace-prefix` emit structured (JSON) output?** Current `fab pane` subcommands mix plain text and `--json`. Lean toward following the pattern: plain `renamed: <old> -> <new>` by default, `--json` flag for `{"pane": …, "old": …, "new": …, "action": renamed|noop}`.
4. **Does pane death need a distinct exit code?** Currently suggested as generic exit 2 ("pane doesn't exist"). The operator's removal-path caller wants to treat "pane gone" as successful removal (the window is gone anyway) and "pane alive but rename failed" as a warning. Separate exit codes (2 for "no such pane" vs 3 for "tmux other error") allow this discrimination.
5. **Stage fence.** Pure infra + behavior in one PR is nice for validating the abstraction, but doubles the rework risk if review catches an issue in either half. Acceptable because: (a) both halves land in the same `.md` hydrate file, (b) the skill change is textually small (~6 lines net).

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Scope covers three deliverables: `fab pane window-name` subcommands, operator skill consumption at two sites, replace-on-removal behavior | User: "one PR that (a) adds the subcommands, (b) updates /fab-operator to call them, (c) introduces the replace-on-removal rule." Confirmed via "Ok, create a draft." | S:95 R:85 A:90 D:90 |
| 2 | Certain | Minimum verb set is two: `ensure-prefix` and `replace-prefix`; add-prefix and remove-prefix are subsumed (add → ensure, remove → replace with empty `<to>`) | Discussed explicitly. User endorsed by moving to draft | S:90 R:85 A:90 D:90 |
| 3 | Certain | Subcommands live under `fab pane` (not `fab operator`, not top-level) | Existing `fab pane {map, capture, send, process}` group is the tmux-pane-operation home; consistent with 260419 rewrite | S:95 R:90 A:95 D:95 |
| 4 | Certain | Both verbs use a literal string-prefix check, not regex | Consistent with `260422-jyyg` Design Decision 2 (literal `»` guard). Regex absorbs legacy markers silently | S:95 R:85 A:90 D:95 |
| 5 | Certain | `<pane>` argument format matches existing `fab pane send` / `capture` (tmux pane ID like `%3`) | Consistency with the existing subcommand group | S:95 R:95 A:100 D:95 |
| 6 | Certain | Replace-on-removal replaces the current "no rename on removal" rule wholesale | User directly revisited the prior decision. The new rule is strictly more informative (honest tab-bar signal) | S:90 R:80 A:90 D:90 |
| 7 | Certain | Guard-on-current-prefix (not `original_name` field) is how user-rename-mid-monitoring is protected | Discussed and endorsed. Avoids schema churn and avoids the "which original is authoritative?" ambiguity | S:95 R:80 A:85 D:90 |
| 8 | Confident | Done-marker candidates are single-width BMP; `›` (U+203A) is the tentative lead over `✓` (U+2713) | `✓` already appears in operator status frame as stage-done signal → visual collision. `›` has parity with `»`. Both BMP. Spec-stage decision | S:70 R:85 A:85 D:75 |
| 9 | Confident | Changes land in a single PR, not split infra-then-behavior | User: "one PR that …". Splitting loses validation value (primitive shipped with only one user) | S:80 R:80 A:85 D:85 |
| 10 | Confident | No `original_name` or equivalent state added to `.fab-operator.yaml` | Implied by #7. Spec-stage to confirm no edge case forces schema addition | S:75 R:80 A:85 D:85 |
| 11 | Confident | Enrollment site call shape is `fab pane window-name ensure-prefix <pane> »` (literal `»` arg, not quoted shell) | One-liner matches existing `fab pane send` / `capture` ergonomics; pane ID first, payload after | S:75 R:90 A:85 D:80 |
| 12 | Confident | Removal site applies to every removal path (terminal stage, stop_stage, pane death, explicit stop) uniformly | `260422-jyyg` spec established a single Removal-triggers set; this change doesn't split that | S:80 R:85 A:85 D:85 |
| 13 | Tentative | Stored `<done-char>` is a skill constant, not a config option | A config surface (e.g., `config.yaml: operator.done_marker`) is plausible but speculative. Spec stage to decide | S:55 R:70 A:70 D:60 |
| 14 | Tentative | Pane-death error path uses a distinct exit code (2 = no such pane, 3 = other tmux error) | Open Question #4. Operator benefits from the discrimination but the value is small and can be deferred | S:50 R:75 A:75 D:65 |
| 15 | Tentative | Primitives do not validate width/BMP/codepoint of `<char>`, `<from>`, `<to>` | Open Question #2. Leaning toward primitive-stays-mechanical, but spec stage to confirm | S:55 R:80 A:70 D:60 |
| 16 | Tentative | Primitives provide `--json` alongside plain output, matching existing subcommands | Open Question #3. Most existing `fab pane` subcommands offer `--json`; plausible default but not explicitly required | S:60 R:85 A:75 D:70 |
| 17 | Tentative | Done-marker character (`›` vs `✓` vs other) | Open Question #1. Spec stage to lock in after considering the operator status frame's existing `✓` usage | S:55 R:75 A:70 D:50 |

17 assumptions (7 certain, 5 confident, 5 tentative, 0 unresolved).
