# Spec: Replace Template-Driven Language Detection with Agent-Inferred Conventions

**Change**: 260306-143f-setup-language-inference
**Created**: 2026-03-06
**Affected memory**: `docs/memory/fab-workflow/setup.md`

## Non-Goals

- Changing any other `/fab-setup` steps (0, 1a, 1b, 1b2–1h, config, constitution, migrations)
- Modifying the Go binary or any other skills
- Adding new `fab/project/*` file types — only writing to the existing set

## Setup: Language Inference Flow (Step 1b-lang)

### Requirement: Replace Template Lookup with Agent Inference

Step 1b-lang in `fab/.kit/skills/fab-setup.md` SHALL replace the current template-lookup logic (detection table + `fab/.kit/templates/constitutions/` and `fab/.kit/templates/configs/` references) with a three-phase prompt-driven flow: **Detection → Inference → Write**.

The detection phase MUST read project marker files to identify language, framework, and tooling. The inference phase MUST use the agent's training knowledge combined with actual config values read from the project to derive conventions. The write phase MUST output inferred conventions to the appropriate `fab/project/*` files.

#### Scenario: Rust project detected
- **GIVEN** a project with `Cargo.toml` at the repo root
- **WHEN** step 1b-lang executes
- **THEN** the agent reads `Cargo.toml` for edition, dependencies, and features
- **AND** scans for `clippy.toml`, `rustfmt.toml`, `.cargo/config.toml`
- **AND** infers Rust conventions from the detected configuration
- **AND** writes enforcement rules to `fab/project/constitution.md`, stack info to `fab/project/context.md`, and coding standards to `fab/project/code-quality.md`

#### Scenario: TypeScript + React project detected
- **GIVEN** a project with `package.json` containing `"react"` in dependencies and `tsconfig.json` present
- **WHEN** step 1b-lang executes
- **THEN** the agent reads `package.json` for dependencies, scripts, and engine constraints
- **AND** reads `tsconfig.json` for compiler options
- **AND** scans for `biome.json`, `.eslintrc*`, `.prettierrc*`, `vitest.config.*`, `jest.config.*`
- **AND** infers both TypeScript and React conventions layered together
- **AND** writes to the appropriate `fab/project/*` files

#### Scenario: No language markers found
- **GIVEN** a project with no recognized marker files
- **WHEN** step 1b-lang executes
- **THEN** the step completes silently with no output and no modifications to `fab/project/*` files

### Requirement: Detection Phase — Marker File Scanning

The detection phase SHALL check for the following marker files at the repo root (checked in order, multiple can match):

| Marker file(s) | Language/Framework |
|----------------|-------------------|
| `Cargo.toml` | Rust |
| `tsconfig.json` + `package.json` | TypeScript |
| `package.json` (no `tsconfig.json`) | Node.js/JavaScript |
| `go.mod` | Go |
| `pyproject.toml` or `setup.py` or `requirements.txt` | Python |

After language detection, the agent SHALL scan for framework signals within detected files:
- `"react"` or `"next"` in `package.json` dependencies → React/Next.js
- `"vue"` in `package.json` dependencies → Vue
- `"svelte"` in `package.json` dependencies → Svelte

The agent SHALL also scan for linter/formatter/tooling configs:
- `.eslintrc*`, `eslint.config.*`, `biome.json`, `.prettierrc*`
- `clippy.toml`, `rustfmt.toml`
- `ruff.toml`, `pyproject.toml` `[tool.ruff]`/`[tool.black]` sections
- `vitest.config.*`, `jest.config.*`, `pytest.ini`, `Cargo.toml` `[dev-dependencies]` for test crates

#### Scenario: Multiple languages in monorepo
- **GIVEN** a project with both `Cargo.toml` and `package.json` + `tsconfig.json`
- **WHEN** step 1b-lang executes
- **THEN** the agent detects both Rust and TypeScript
- **AND** infers conventions for each language
- **AND** writes combined conventions to `fab/project/*` files with clear per-language sections

### Requirement: Inference Phase — Convention Derivation

For each detected language/framework, the agent SHALL:

1. Read the detected marker files and config files to extract concrete values (e.g., compiler options, linter rules, dependency versions)
2. Use its training knowledge to derive standard conventions for the detected stack, grounded in the actual config values read
3. Distinguish between enforcement rules (MUST/SHOULD) and descriptive context

The inference phase MUST NOT hard-code conventions in the skill file. The skill text SHALL describe the *process* (what to read, what to infer, where to write) — the agent's training knowledge provides the *content*.

#### Scenario: Biome config with specific settings
- **GIVEN** a project with `biome.json` containing `"semicolons": "asNeeded"` and `"indentStyle": "space"`
- **WHEN** the agent infers conventions
- **THEN** the inferred conventions reflect the actual config values (e.g., "semicolons as needed per Biome config")
- **AND** the conventions do not contradict or override the existing linter settings

### Requirement: Adaptive Questions

The agent MAY ask up to 3 free-form questions about aspects that genuinely cannot be inferred from the project files. Questions SHALL be adaptive — if a signal is already present in the project files, the agent MUST NOT ask about it.

SRAD does not apply to this question flow. The questions are free-form and contextual, not scored.

Examples of question-worthy gaps (when not inferable):
- Error handling philosophy (when no error handling patterns are visible)
- Architecture pattern (when directory structure is ambiguous)
- Testing expectations (when no test runner or test files are present)

#### Scenario: Test runner already in package.json
- **GIVEN** a project with `vitest` in `package.json` devDependencies
- **WHEN** the agent considers adaptive questions
- **THEN** the agent does NOT ask about test runner preference
- **AND** infers Vitest as the test runner convention

#### Scenario: No test infrastructure visible
- **GIVEN** a project with no test runner in dependencies and no test config files
- **WHEN** the agent considers adaptive questions
- **THEN** the agent MAY ask about testing expectations as one of its up-to-3 questions

### Requirement: Write Phase — Target Files

The write phase SHALL route inferred conventions to the appropriate `fab/project/*` files:

| Content type | Target file | Example |
|-------------|------------|---------|
| MUST/SHOULD enforcement rules | `fab/project/constitution.md` | "All public APIs MUST have doc comments" |
| Descriptive stack info | `fab/project/context.md` | "Tech stack: Rust 2021 edition with Tokio async runtime" |
| Coding standards, anti-patterns | `fab/project/code-quality.md` | "Prefer `thiserror` for error types" |
| Review policy additions | `fab/project/code-review.md` | "Unsafe blocks require justification comment" |
| Source paths, checklist categories | `fab/project/config.yaml` | `source_paths: [src/]` |

The agent SHALL insert language conventions into `constitution.md` as a new section before the `## Governance` heading, using the heading format `## {Language} Conventions` (e.g., `## Rust Conventions`).

For `context.md`, `code-quality.md`, and `code-review.md`, the agent SHALL append to existing content under appropriate headings, or create new headings if none match.

For `config.yaml`, the agent SHALL merge values using targeted string replacement (consistent with existing config update approach): union `source_paths`, union `checklist.extra_categories`, merge `stage_directives`.

#### Scenario: Constitution already has language section
- **GIVEN** `fab/project/constitution.md` already contains `## Rust Conventions`
- **WHEN** step 1b-lang detects Rust and infers conventions
- **THEN** the agent reads the existing section, identifies what's already covered
- **AND** appends only new conventions not already present
- **AND** does not duplicate or overwrite existing content

#### Scenario: First-run on fresh project
- **GIVEN** `fab/project/constitution.md` was just created by step 1b (no language sections)
- **WHEN** step 1b-lang detects TypeScript
- **THEN** the agent inserts `## TypeScript Conventions` before `## Governance`
- **AND** populates `context.md`, `code-quality.md` with inferred conventions
- **AND** updates `config.yaml` source paths if detectable

### Requirement: Idempotency on Re-run

Step 1b-lang MUST be idempotent. On re-run, the agent SHALL:

1. Read existing `fab/project/*` content before writing
2. Detect what conventions are already present
3. Merge/update without duplicating or overwriting user edits
4. Skip sections that are already complete

#### Scenario: Re-run after user edited constitution
- **GIVEN** a user manually edited `## Rust Conventions` in `constitution.md` to add a custom rule
- **WHEN** step 1b-lang re-runs
- **THEN** the agent preserves the user's custom rule
- **AND** does not revert or duplicate any content

## Setup: Template Deletion

### Requirement: Remove Constitution and Config Templates

The directories `fab/.kit/templates/constitutions/` and `fab/.kit/templates/configs/` SHALL be deleted entirely. This removes:

- `fab/.kit/templates/constitutions/rust.md`
- `fab/.kit/templates/constitutions/node-typescript.md`
- `fab/.kit/templates/constitutions/react.md`
- `fab/.kit/templates/configs/rust.yaml`
- `fab/.kit/templates/configs/node-typescript.yaml`
- `fab/.kit/templates/configs/react.yaml`

The `fab/.kit/templates/` directory itself SHALL remain — it contains artifact templates (`intake.md`, `spec.md`, `tasks.md`, `checklist.md`, `status.yaml`).

#### Scenario: Templates directory after deletion
- **GIVEN** the template deletion is applied
- **WHEN** listing `fab/.kit/templates/`
- **THEN** the `constitutions/` and `configs/` subdirectories do not exist
- **AND** artifact templates (`intake.md`, `spec.md`, etc.) remain untouched

## Sync Script: Remove Template Advisory

### Requirement: Delete Template Advisory Block

The "Language template advisory" block in `fab/.kit/sync/2-sync-workspace.sh` (lines 219–247, section `2b`) SHALL be deleted. This block checks for `fab/.kit/templates/constitutions` existence and suggests running `/fab-setup` when conventions are missing. Since the templates are being deleted, this advisory becomes dead code.

#### Scenario: Sync script after advisory removal
- **GIVEN** the advisory block is removed from `2-sync-workspace.sh`
- **WHEN** `fab-sync.sh` runs
- **THEN** no "Note: {Language} project detected but constitution lacks {Language} conventions" messages are produced
- **AND** the rest of `2-sync-workspace.sh` functions identically (directories, scaffold, skills, cleanup, version stamp)

## Design Decisions

1. **Agent inference over templates**: The agent uses its training knowledge to derive conventions, grounded in actual project file contents. This avoids bundling opinionated language conventions in `fab/.kit/`, keeping the kit language-neutral per Constitution §V.
   - *Why*: Templates required maintenance for each language/framework and encoded opinions that may not match the project's actual setup.
   - *Rejected*: Keeping templates — violates portability, creates maintenance burden, makes judgment calls on behalf of users.

2. **Process description in skill, not content**: The skill file describes *what to detect*, *what to read*, and *where to write* — it does NOT contain hard-coded convention text. The agent provides the convention content from its training knowledge.
   - *Why*: Hard-coding conventions in the skill would recreate the template problem in a different location.
   - *Rejected*: Inline convention blocks in the skill — same maintenance burden as templates.

3. **Write to all `fab/project/*` files**: Conventions are routed to the appropriate file by content type, not all dumped into `constitution.md`.
   - *Why*: Each file has a specific purpose (enforcement vs. descriptive vs. quality standards). Mixing content types degrades the signal.
   - *Rejected*: Writing everything to `constitution.md` — conflates enforcement rules with descriptive context.

## Deprecated Requirements

### Template-Driven Language Detection (Step 1b-lang)
**Reason**: Replaced by agent-inferred conventions. Template files (`fab/.kit/templates/constitutions/`, `fab/.kit/templates/configs/`) are deleted.
**Migration**: Step 1b-lang now uses agent inference instead of template lookup. No user action required — the new flow produces equivalent output.

### Language Template Advisory (2-sync-workspace.sh Section 2b)
**Reason**: Referenced template files that no longer exist.
**Migration**: N/A — advisory was informational only.

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Kit stays language-neutral — no bundled convention templates | Confirmed from intake #1 — Constitution §V mandates portability | S:95 R:90 A:95 D:95 |
| 2 | Certain | Detection via marker files is kept (same approach, expanded) | Confirmed from intake #2 — detection is neutral, only content sourcing changes | S:90 R:95 A:90 D:95 |
| 3 | Certain | Convention content comes from agent inference + project file reading | Confirmed from intake #3 — user explicitly chose this over templates | S:95 R:85 A:90 D:90 |
| 4 | Certain | Write to all `fab/project/*` files as appropriate by content type | Confirmed from intake #4 — setup works on all project files | S:90 R:85 A:90 D:90 |
| 5 | Certain | Up to 3 adaptive free-form questions for unresolvable gaps | Confirmed from intake #5 — adaptive questioning, no hard-coded questions | S:90 R:90 A:85 D:90 |
| 6 | Certain | Re-runs detect existing content and merge/update idempotently | Confirmed from intake #6 — idempotent merge behavior | S:85 R:80 A:85 D:90 |
| 7 | Certain | SRAD does not apply to the setup question flow | Confirmed from intake #7 — user explicitly excluded SRAD | S:90 R:90 A:85 D:95 |
| 8 | Certain | Delete `fab/.kit/templates/constitutions/` and `fab/.kit/templates/configs/` | Confirmed from intake #8 — rejected template files | S:95 R:85 A:95 D:95 |
| 9 | Certain | Remove template advisory block from `2-sync-workspace.sh` | Upgraded from intake Confident #9 — block references deleted templates, must be removed. Verified lines 219-247 | S:90 R:90 A:95 D:90 |
| 10 | Certain | Detection table expands to include Python, Go, and more linter/framework configs | Upgraded from intake Confident #10 — natural extension aligned with intake's explicit file list | S:85 R:90 A:90 D:85 |
| 11 | Confident | Constitution conventions inserted before `## Governance` heading | Standard pattern for constitution sections — keeps governance at the end. Intake didn't specify insertion point | S:75 R:90 A:85 D:85 |
| 12 | Confident | Skill describes detection process, not hard-coded convention content | Logically follows from "agent inference" decision — skill is the process, agent is the content | S:80 R:85 A:85 D:90 |

12 assumptions (10 certain, 2 confident, 0 tentative, 0 unresolved).
