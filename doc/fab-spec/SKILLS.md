# Fab Skills Reference

> Detailed behavior for each `/fab:*` skill. For a quick overview, see the [Quick Reference](README.md#quick-reference).

---

## `/fab:init`

**Purpose**: Bootstrap `fab/` in an existing project.

**Creates**:
- `fab/.kit/` — engine directory (templates, skills, scripts)
- `fab/config.yaml` — project configuration (prompts for name, tech stack, conventions)
- `fab/memory/constitution.md` — project principles and constraints (generated from conversation or existing docs)
- `fab/specs/` — empty, ready for centralized specs
- `fab/changes/` — empty, ready for change folders
- `.claude/skills/` — symlinks pointing into `fab/.kit/skills/`

**Example**:
```
/fab:init
→ "What's the project name?"
→ "Describe the tech stack and conventions..."
→ "fab/ initialized with config, templates, and empty specs."
```

**Behavior**:
1. Check if `fab/` already exists (abort if so, suggest manual edits)
2. Prompt for project name, description, tech stack
3. Create `fab/.kit/` with default templates, skills, and scripts
4. Generate `fab/config.yaml` from responses
5. Generate `fab/memory/constitution.md` from project context (README, existing docs, conversation)
6. Create symlinks in `.claude/skills/` pointing to `fab/.kit/skills/`
7. Optionally scaffold initial specs from existing code or documentation

---

## `/fab:new <description>`

**Purpose**: Start a new change from a natural language description.

**Creates**:
- Change folder named `{YYMMDD}-{XXXX}-{slug}`
- `.status.yaml` manifest
- `proposal.md` from template (with clarifying questions if ambiguous)

**Example**:
```
/fab:new Add OAuth2 support for Google and GitHub sign-in
→ Created fab/changes/260115-a7k2-add-oauth/
```

**Behavior**:
1. Generate folder name: today's date (`YYMMDD`) + 4 random alphanumeric chars + 2-4 word slug from description
2. Create `fab/changes/{name}/`
3. Write change name to `fab/current` (sets this as the active change)
4. Initialize `.status.yaml` with stage: proposal
5. Generate `proposal.md` using template (loading `fab/memory/constitution.md` and `fab/config.yaml` as context)
6. Ask clarifying questions if intent is ambiguous
7. Mark proposal complete when satisfied

---

## `/fab:continue`

**Purpose**: Create the next artifact in sequence.

**Example**:
```
/fab:continue
→ "Stage: proposal (complete). Next: Create delta specs."
```

**Behavior**:
1. Read `.status.yaml` to determine current stage
2. Identify next artifact to create
3. Load relevant template + context (including `fab/memory/constitution.md` for project principles)
4. Generate artifact (with clarification/research as needed)
5. **Plan decision** (when transitioning from specs to plan): evaluate whether a plan is warranted. If the change is small and the approach is obvious, propose skipping to the user: *"This change is straightforward — skip plan and go directly to tasks?"* If the user agrees, record `plan: skipped` in `.status.yaml` and proceed to tasks. If the user wants a plan, generate it normally.
6. Auto-generate checklist when creating tasks
7. Update `.status.yaml`

---

## `/fab:ff` (Fast Forward)

**Purpose**: Fast-forward through remaining planning stages in one pass. Requires an active change with a completed proposal (run `/fab:new` first).

**Flow**: specs → plan (if warranted) → tasks (+ checklist)

**When to use**:
- Small, well-understood changes
- Clear requirements upfront
- Want to reach implementation quickly

**Example**:
```
/fab:new Add a logout button to the navbar that clears session
/fab:ff
```

**Behavior**:
1. Read `fab/current` to resolve the active change; verify proposal is complete
2. Generate delta specs (resolve ambiguities inline, no blocking questions)
3. Evaluate whether a plan is warranted (see "Plan decision" below). If yes, draft plan with inline research. If no, skip directly to tasks
4. Produce task breakdown (referencing plan if it exists, otherwise referencing specs and proposal directly)
5. Auto-generate quality checklist
6. Update status to `tasks: complete`

**Plan decision**: The agent skips `plan.md` when the change is small and the implementation approach is obvious — e.g., single-file changes, straightforward CRUD, or well-known patterns. When skipped, `.status.yaml` records `plan: skipped`. Unlike `/fab:continue`, `/fab:ff` does **not** confirm with the user before skipping — it decides autonomously to maintain the fast-forward flow.

---

## `/fab:apply`

**Purpose**: Execute tasks from `tasks.md`.

**Example**:
```
/fab:apply
→ "Starting implementation. 12 tasks remaining."
```

**Behavior**:
1. Parse `tasks.md` for unchecked items `- [ ]`
2. Execute tasks in dependency order
3. Respect parallel markers `[P]`
4. After completing each task, run relevant tests (e.g., the test file for the module just modified). Fix failures before moving on.
5. Mark each task `[x]` immediately upon completion (not batched at the end)
6. Update `.status.yaml` progress after each task

**Resumability**: `/fab:apply` is inherently resumable. If the agent is interrupted mid-run, re-invoking `/fab:apply` picks up from the first unchecked item. The markdown checklist *is* the progress state — no separate tracking needed.

---

## `/fab:verify`

**Purpose**: Validate implementation against specs and checklists.

**Example**:
```
/fab:verify
→ "✓ 12/12 tasks complete"
→ "✓ 10/12 checklist items passed"
→ "✗ 2 items need attention: [CHK-007, CHK-011]"
```

**Checks** (the agent performs all of these):
1. All tasks in `tasks.md` marked `[x]`
2. All checklist items in `checklists/quality.md` verified and checked off — the agent re-reads each `CHK-*` item, inspects the relevant code/tests, and marks `[x]` or reports failure
3. Run tests affected by the change (scoped to modules touched, not the full suite)
4. Features match delta spec requirements (spot-check key scenarios from `specs/`)
5. No spec drift detected (implementation doesn't contradict centralized specs)

**On failure**, the agent presents the options and the user chooses where to loop back:

- **Fix code** → `/fab:apply`
  Implementation bug. The agent identifies which tasks need rework, unchecks them in `tasks.md` (marks `- [ ]` again with a `<!-- rework: reason -->` comment), and re-runs `/fab:apply` which picks up the unchecked items.

- **Revise tasks** → edit `tasks.md`, then `/fab:apply`
  Missing or wrong tasks. The agent adds/modifies tasks in `tasks.md` (new tasks get the next sequential ID). Completed tasks that are unaffected stay `[x]`. Only new or revised tasks are executed.

- **Revise plan** → `/fab:continue` from plan stage
  Architecture was wrong. The agent updates `plan.md` in place (not recreated from scratch). After the plan is revised, `tasks.md` is regenerated — all tasks are reset to `- [ ]` since the implementation approach changed. The checklist is also regenerated.

- **Revise specs** → `/fab:continue` from specs stage
  Requirements were wrong or incomplete. Delta specs are updated in place. Plan (if it exists) and tasks are subsequently regenerated. All downstream artifacts are reset.

The `.status.yaml` stage is reset to the chosen re-entry point. The general rule: **artifacts at and after the re-entry point are regenerated or updated; artifacts before it are preserved.**

---

## `/fab:archive`

**Purpose**: Complete the change and hydrate into centralized specs.

**Example**:
```
/fab:archive
→ "Archived to fab/changes/archive/260115-a7k2-add-oauth/"
→ "Hydrated specs: fab/specs/auth/authentication.md"
```

**Behavior**:
1. **Final validation** — verify must pass (all tasks `[x]`, all checklist items `[x]` or `[N/A]`)
2. **Concurrent change check** — scan `fab/changes/` for other active changes whose delta specs reference the same centralized spec files. If found, warn the user: *"Change {name} also modifies {spec}. After this archive, that change's delta was written against a now-stale base. Re-verify with `/fab:verify` after switching to it."*
3. **Hydrate delta specs** into `fab/specs/`:
   The agent reads the delta specs and the current centralized spec, then rewrites the centralized spec to incorporate the changes. The ADDED/MODIFIED/REMOVED markers are **semantic hints to the agent about intent**, not instructions for a text processor:
   - **ADDED** → agent integrates new requirements into the appropriate section
   - **MODIFIED** → agent updates the existing requirement in context, preserving surrounding content
   - **REMOVED** → agent removes the requirement, adjusting related content for coherence
   The agent should minimize edits to unchanged sections to prevent drift over successive archives.
4. **Update status** to `archived` in `.status.yaml`
5. **Move change folder** to `archive/` (no rename — date is already in the folder name)
6. **Clear pointer** — delete `fab/current` (no active change)

**Order of operations**: Steps 3–6 are ordered to fail safely. Status is updated *before* the folder move, so if the move is interrupted, the change is marked archived but still in `changes/` — the agent can detect and complete the move on next invocation. The pointer is cleared last so that mid-archive, `/fab:status` still reports the active change rather than "no active change" with a half-hydrated spec.

---

## `/fab:switch <change-name>`

**Purpose**: Switch the active change when multiple changes exist.

**Example**:
```
/fab:switch fix-checkout
→ "fab/current now points to 260202-m3x1-fix-checkout-bug"
```

**Behavior**:
1. Match `change-name` against `fab/changes/` (supports partial/slug match)
2. Write the full change name to `fab/current`
3. Display the switched change's status summary

---

## `/fab:status`

**Purpose**: Show current change state at a glance.

**Example output**:
```
Change: 260115-a7k2-add-oauth
Stage:  plan (3/7)

Progress:
  ✓ proposal    complete
  ✓ specs       complete
  ◉ plan        in_progress
  ○ tasks       pending
  ○ apply       pending
  ○ verify      pending
  ○ archive     pending

Checklist: not yet generated (created at tasks stage)

Next: Complete plan.md, then /fab:continue
```
