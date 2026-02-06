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
3. Initialize `.status.yaml` with stage: proposal
4. Generate `proposal.md` using template
5. Ask clarifying questions if intent is ambiguous
6. Mark proposal complete when satisfied

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
3. Load relevant template + context
4. Generate artifact (with clarification/research as needed)
5. Auto-generate checklist when creating tasks
6. Update `.status.yaml`

---

## `/fab:ff` (Fast Forward)

**Purpose**: Generate all planning artifacts in one pass.

**Flow**: proposal → specs → plan → tasks (+ checklist)

**When to use**:
- Small, well-understood changes
- Clear requirements upfront
- Want to reach implementation quickly

**Example**:
```
/fab:ff Add a logout button to the navbar that clears session
```

**Behavior**:
1. Create proposal from description
2. Generate delta specs (ask clarifying questions inline)
3. Draft plan (do research inline)
4. Produce task breakdown
5. Auto-generate quality checklist
6. Update status to `tasks: complete`

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

**On failure**, the user chooses where to loop back:
- **Fix code** → `/fab:apply` (implementation bug — re-run uncompleted/fixed tasks)
- **Revise tasks** → edit `tasks.md`, then `/fab:apply` (missing or wrong tasks)
- **Revise plan** → `/fab:continue` from plan stage (architecture was wrong)
- **Revise specs** → `/fab:continue` from specs stage (requirements were wrong or incomplete)

The `.status.yaml` stage is reset to the chosen re-entry point. Existing artifacts at that stage are updated in place, not recreated from scratch.

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
1. Final validation (verify must pass)
2. Hydrate delta specs into `fab/specs/`:
   The agent reads the delta specs and the current centralized spec, then rewrites the centralized spec to incorporate the changes. The ADDED/MODIFIED/REMOVED markers are **semantic hints to the agent about intent**, not instructions for a text processor:
   - **ADDED** → agent integrates new requirements into the appropriate section
   - **MODIFIED** → agent updates the existing requirement in context, preserving surrounding content
   - **REMOVED** → agent removes the requirement, adjusting related content for coherence
   The agent should minimize edits to unchanged sections to prevent drift over successive archives.
3. Move change folder to `archive/` (no rename — date is already in the folder name)
4. Update status to `archived`

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
