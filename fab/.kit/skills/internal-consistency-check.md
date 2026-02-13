# Internal Consistency Check

Scan for inconsistencies between the three sources of truth in this project:

- **Implementation** — the project's source code (paths declared in `source_paths` in `fab/config.yaml`)
- **Docs** (`fab/memory/`) — centralized documentation (generated/hydrated)
- **Design** (`fab/specs/`) — human-curated design specs and architecture

These layers can drift apart over time — stale references, renamed concepts, missing coverage, contradicted behavior. Use a team of agents to audit all three in parallel.

---

## Pre-flight

1. Read `fab/config.yaml`
2. Extract the `source_paths` list — these are the implementation directories to scan
3. If `source_paths` is missing or empty, STOP with: `No source_paths defined in fab/config.yaml. Add source_paths to tell this command where your implementation lives.`

Let `{IMPL_PATHS}` refer to the resolved `source_paths` entries for the rest of this command.

---

## Execution

Spawn **three parallel agents** using the Task tool (subagent_type: `Explore`, thoroughness: `very thorough`). Each agent audits one dimension of consistency. Run all three concurrently.

Include the resolved `{IMPL_PATHS}` in each agent prompt so they know exactly which directories to scan.

### Agent 1: Design ↔ Implementation Drift

Prompt:

> Audit consistency between design specs (`fab/specs/`) and the implementation (directories: `{IMPL_PATHS}`).
>
> 1. Read `fab/specs/index.md` to understand the intended architecture and all design docs
> 2. Read every file in `fab/specs/` to catalog the specified skills, stages, workflow steps, naming conventions, and templates
> 3. Read every file in the implementation directories: `{IMPL_PATHS}`
> 4. Report inconsistencies in these categories:
>    - **Missing implementations**: things described in design that don't exist in the implementation
>    - **Undocumented implementations**: things in the implementation not covered by any design doc
>    - **Naming mismatches**: different names for the same concept between design and implementation
>    - **Behavioral contradictions**: where implementation behavior contradicts design spec
>    - **Stale references**: design docs referencing files, paths, or concepts that no longer exist
>
> For each finding, cite the specific files and lines involved.

### Agent 2: Docs ↔ Implementation Drift

Prompt:

> Audit consistency between centralized docs (`fab/memory/`) and the implementation (directories: `{IMPL_PATHS}`).
>
> 1. Read `fab/memory/index.md` to understand the documentation landscape
> 2. Read every doc in `fab/memory/` recursively
> 3. Read every file in the implementation directories: `{IMPL_PATHS}`
> 4. Report inconsistencies in these categories:
>    - **Stale docs**: docs describing behavior/features that no longer exist or work differently
>    - **Missing docs**: implemented features not covered by any doc
>    - **Wrong paths/names**: docs referencing files, commands, or concepts by outdated names
>    - **Contradicted behavior**: docs claiming behavior X when implementation does Y
>    - **Orphaned docs**: doc files not referenced from any index or other doc
>
> For each finding, cite the specific files and lines involved.

### Agent 3: Design ↔ Docs Drift

Prompt:

> Audit consistency between design specs (`fab/specs/`) and centralized docs (`fab/memory/`).
>
> 1. Read `fab/specs/index.md` and `fab/memory/index.md`
> 2. Read every file in both `fab/specs/` and `fab/memory/` recursively
> 3. Report inconsistencies in these categories:
>    - **Terminology drift**: same concept described with different names across design and docs
>    - **Coverage gaps**: design concepts not reflected in docs, or doc topics not grounded in design
>    - **Contradictions**: where docs and design disagree on workflow, stages, behavior, or structure
>    - **Stale cross-references**: either layer referencing the other with outdated paths, names, or structure
>    - **Glossary drift**: terms defined in `fab/specs/glossary.md` that are used inconsistently in docs
>
> For each finding, cite the specific files and lines involved.

---

## Synthesis

After all three agents return, synthesize a unified report:

### 1. Summary Table

| Dimension | Findings | Critical | Minor |
|-----------|----------|----------|-------|
| Design ↔ Implementation | {count} | {count} | {count} |
| Docs ↔ Implementation | {count} | {count} | {count} |
| Design ↔ Docs | {count} | {count} | {count} |

### 2. Critical Findings

List all findings where behavior is contradicted or key concepts are missing/broken. These should be fixed first.

### 3. Minor Findings

List naming mismatches, missing coverage, stale references, and orphaned content. These are cleanup opportunities.

### 4. Suggested Actions

Concrete next steps, grouped by type:

- **Fix**: things that are actively wrong and should be corrected
- **Add**: missing documentation or design coverage
- **Remove**: stale/orphaned content that should be deleted
- **Rename**: terminology alignment opportunities

---

## Classification

A finding is **Critical** if:
- Implementation contradicts design intent (not just naming — actual behavior)
- Docs instruct users to do something that will fail or produce wrong results
- A referenced file, command, or path does not exist

A finding is **Minor** if:
- Naming mismatch without behavioral impact
- Coverage gap (missing docs, not wrong docs)
- Stale reference that doesn't cause user-facing confusion
- Orphaned content
