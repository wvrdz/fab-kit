# Shared Context Preamble

> This file defines shared conventions for all Fab skills. Each skill file should begin with:
> `Read and follow the instructions in fab/.kit/skills/_context.md before proceeding.`

---

## Context Loading

Before generating or validating any artifact, load the relevant context layers below. This ensures output is grounded in the actual project state, not assumptions.

### 1. Always Load (every skill except `/fab:init`, `/fab:switch`, `/fab:status`)

Read these files first — they define the project's identity and constraints:

- **`fab/config.yaml`** — project configuration, tech stack, naming conventions, stage configuration
- **`fab/constitution.md`** — project principles and constraints (MUST/SHOULD/MUST NOT rules)

### 2. Change Context (when operating on an active change)

Resolve the active change and load its state:

1. Read `fab/current` to get the active change name
2. If `fab/current` does not exist, there is no active change — inform the user and suggest `/fab:new`
3. Load `fab/changes/{name}/.status.yaml` — current stage, progress, branch info
4. Load all completed artifacts in the change folder (e.g., `proposal.md`, `spec.md`, `plan.md`, `tasks.md`) — read each file that exists so you have full context of what has been decided so far

### 3. Centralized Doc Lookup (when writing or validating specs)

Load the documentation landscape to ensure specs reflect the actual current state:

1. Read `fab/docs/index.md` to understand which domains and docs exist
2. Read the specific centralized doc(s) referenced by the proposal's **Affected Docs** section (the New, Modified, and Removed entries)
3. Use this context to write specs against the real current state, not assumptions

### 4. Source Code Loading (during implementation and review)

Load only the source files relevant to the current work:

1. Read the relevant source files referenced in the plan's **File Changes** section (New, Modified, Deleted) or in the task descriptions
2. Scope to files actually touched by the change — do not load the entire codebase
3. This applies primarily to `/fab:apply` and `/fab:review`

---

## Next Steps Convention

Every skill MUST end its output with a `Next:` line suggesting the available follow-up commands. This keeps the user oriented in the workflow without needing to memorize the stage graph.

**Format**: `Next: /fab:command` or `Next: /fab:commandA or /fab:commandB (description)`

### Lookup Table

| After skill | Stage reached | Next line |
|-------------|---------------|-----------|
| `/fab:init` | initialized | `Next: /fab:new <description>` |
| `/fab:new` | proposal done | `Next: /fab:continue or /fab:ff (fast-forward all planning)` |
| `/fab:continue` → specs | specs done | `Next: /fab:continue (plan) or /fab:ff (fast-forward) or /fab:clarify (refine spec)` |
| `/fab:continue` → plan | plan done | `Next: /fab:continue (tasks) or /fab:clarify (refine plan)` |
| `/fab:continue` → tasks | tasks done | `Next: /fab:apply` |
| `/fab:ff` | tasks done | `Next: /fab:apply` |
| `/fab:clarify` | same stage | `Next: /fab:clarify (refine further) or /fab:continue or /fab:ff` |
| `/fab:apply` | apply done | `Next: /fab:review` |
| `/fab:review` (pass) | review done | `Next: /fab:archive` |
| `/fab:review` (fail) | review failed | *(contextual — see /fab:review for fix options)* |
| `/fab:archive` | archived | `Next: /fab:new <description> (start next change)` |
