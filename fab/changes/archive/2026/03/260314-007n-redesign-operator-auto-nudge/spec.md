# Spec: Create fab-operator4 — Redesigned Auto-Nudge Operator

**Change**: 260314-007n-redesign-operator-auto-nudge
**Created**: 2026-03-14
**Affected memory**: `docs/memory/fab-workflow/execution-skills.md`

## Non-Goals

- Modify operator1, operator2, or operator3 — operator4 is purely additive
- Extract shared base or merge operators — each operator remains independent
- Add persistent audit trail — inline logging is sufficient for v1

## Skill: Routing Discipline

### Requirement: Operator SHALL NOT Execute User Instructions Directly

The operator4 skill MUST NOT execute user instructions directly. When the user gives an instruction (e.g., "fix the tests", "add error handling"), the operator SHALL:

1. Determine which running tmux session/agent the request corresponds to
2. Route the instruction to that agent via `tmux send-keys`
3. Enroll the agent in monitoring

The operator's role is coordination, not implementation.

#### Scenario: User Gives Implementation Instruction

- **GIVEN** a monitored agent running in pane `%3` for change `r3m7`
- **WHEN** the user says "fix the failing tests in r3m7"
- **THEN** the operator routes the instruction to pane `%3` via `tmux send-keys`
- **AND** enrolls the agent in monitoring
- **AND** does NOT attempt to fix the tests itself

#### Scenario: User Instruction Without Clear Target

- **GIVEN** multiple monitored agents running
- **WHEN** the user gives an instruction without specifying a change (e.g., "add error handling")
- **THEN** the operator asks the user which agent to route to
- **AND** does NOT execute the instruction itself

## Skill: Autopilot Pipeline Override

### Requirement: Autopilot SHALL Use /fab-fff

Operator4 SHALL use `/fab-fff` instead of `/fab-ff` for autopilot gate checks and pipeline invocations. All other autopilot behavior (inherited from operator2) is unchanged.

#### Scenario: Autopilot Spawns Agent

- **GIVEN** the operator is running autopilot with a queued change
- **WHEN** the change passes the confidence gate
- **THEN** the operator sends `/fab-fff` (not `/fab-ff`) to the agent

## Skill: Simplified Answer Model

### Requirement: All Questions SHALL Be Auto-Answered

Operator4 SHALL remove the two-tier auto-answer/escalate classification from operator3. All detected questions SHALL be auto-answered. The only escalation case is when the operator literally cannot determine what keystrokes to send.

The decision list (replacing operator3's prose heuristic) SHALL be:

1. Binary yes/no or confirmation prompt → `y`
2. `[Y/n]` or `[y/N]` prompt → `y`
3. Claude Code permission/approval prompt → `y`
4. Numbered menu or multi-choice → `1` (first/default option)
5. Open-ended question where a concrete answer is determinable from visible terminal context → send that answer
6. Question where the operator cannot determine what keystrokes to send → escalate: `"{change}: can't determine answer for '{summary}'. Please respond."`

#### Scenario: Binary Confirmation Prompt

- **GIVEN** a monitored idle agent with terminal showing "Proceed with the changes? [Y/n]"
- **WHEN** the monitoring tick detects this as a question
- **THEN** the operator sends `y` via `tmux send-keys`
- **AND** reports: `"{change}: auto-answered 'Proceed with the changes?' -> y"`

#### Scenario: Numbered Menu

- **GIVEN** a monitored idle agent with terminal showing a numbered list of options (e.g., "1) Option A\n2) Option B\nSelect:")
- **WHEN** the monitoring tick detects this as a question
- **THEN** the operator sends `1` via `tmux send-keys`
- **AND** reports: `"{change}: auto-answered 'Select option' -> 1"`

#### Scenario: Open-Ended Question with Determinable Answer

- **GIVEN** a monitored idle agent with terminal showing "Which branch should I rebase on?" and visible context indicating the main branch is `main`
- **WHEN** the monitoring tick detects this as a question
- **THEN** the operator sends `main` via `tmux send-keys`
- **AND** reports the auto-answer

#### Scenario: Undeterminable Question

- **GIVEN** a monitored idle agent with terminal showing a question where the operator cannot determine what keystrokes to send
- **WHEN** the monitoring tick processes the question
- **THEN** the operator does NOT send any answer
- **AND** reports: `"{change}: can't determine answer for '{summary}'. Please respond."`

#### Scenario: No Nudge Budget

- **GIVEN** the operator has auto-answered multiple questions for the same agent
- **WHEN** a new question is detected
- **THEN** the operator evaluates it independently on its own merits
- **AND** there is no cooldown or retry limit between auto-answers

## Skill: Question Detection Improvements

### Requirement: Capture Window SHALL Be 20 Lines

The terminal capture command SHALL use `-l 20` instead of operator3's `-l 10`. The command is: `tmux capture-pane -t <pane> -p -l 20`.

#### Scenario: Long Prompt Captured

- **GIVEN** a monitored idle agent with a question prompt preceded by verbose output that pushes the prompt past line 10
- **WHEN** the monitoring tick captures terminal output with `-l 20`
- **THEN** the question prompt is included in the captured output
- **AND** question detection succeeds

### Requirement: Claude Turn Boundary Guard SHALL Prevent False Positives

If a Claude Code `>` prompt cursor (`^\s*>\s*$`) appears in the last 2 lines of the captured output, question detection SHALL be skipped for that agent. The agent is at a normal human-turn boundary, not a blocking prompt.

#### Scenario: Claude Conversational Output with Question Mark

- **GIVEN** a monitored idle agent whose terminal shows Claude's conversational output ending with "Would you like me to run the tests?" followed by a `>` prompt cursor
- **WHEN** the monitoring tick captures terminal output
- **THEN** the Claude turn boundary guard detects the `>` cursor in the last 2 lines
- **AND** question detection is skipped

#### Scenario: Genuine Question Without Claude Prompt

- **GIVEN** a monitored idle agent whose terminal shows a genuine blocking prompt (e.g., "Allow this tool call? [Y/n]") without a `>` cursor in the last 2 lines
- **WHEN** the monitoring tick captures terminal output
- **THEN** the Claude turn boundary guard does NOT fire
- **AND** question detection proceeds normally

### Requirement: Question Mark Pattern SHALL Match Last Non-Empty Line Only

The `?` pattern SHALL match only on the **last non-empty line** of the captured output (not any of the 20 lines). The matching line MUST be <120 characters. Lines starting with `#`, `//`, `*`, `>`, or timestamp patterns (e.g., `[2026-03-14`, `2026-03-14T`) SHALL be skipped as comments, log output, or search results.

#### Scenario: Question Mark in Log Output

- **GIVEN** a monitored idle agent whose captured output contains "Warning: deprecated API?" on line 5 and the last non-empty line is "Processing complete."
- **WHEN** the monitoring tick scans for question indicators
- **THEN** the `?` pattern does NOT match (not the last non-empty line)

#### Scenario: Question Mark in Comment Line

- **GIVEN** a monitored idle agent whose last non-empty line is "# TODO: should we refactor this?"
- **WHEN** the monitoring tick scans for question indicators
- **THEN** the `?` pattern does NOT match (line starts with `#`)

#### Scenario: Genuine Question on Last Line

- **GIVEN** a monitored idle agent whose last non-empty line is "Which database should I use?" (< 120 chars, no skip prefix)
- **WHEN** the monitoring tick scans for question indicators
- **THEN** the `?` pattern matches

### Requirement: Additional Question Indicator Patterns

Operator4 SHALL add the following question indicator patterns to operator3's existing list:

- Lines ending with `:` or `:\s*$` (CLI input prompts)
- Enumerated options (`[1-9]\)` patterns)
- `Press.*key`, `press.*enter`, `hit.*enter` (case-insensitive)

#### Scenario: Colon-Ending Input Prompt

- **GIVEN** a monitored idle agent whose last non-empty line is "Enter your choice:"
- **WHEN** the monitoring tick scans for question indicators
- **THEN** the colon pattern matches

#### Scenario: Press Enter Prompt

- **GIVEN** a monitored idle agent whose captured output contains "Press Enter to continue"
- **WHEN** the monitoring tick scans for question indicators
- **THEN** the `Press.*enter` pattern matches (case-insensitive)

### Requirement: Blank Capture Guard SHALL Skip Detection

If the captured output from `tmux capture-pane` is entirely blank or whitespace, question detection SHALL be skipped for that tick. This is treated as "cannot determine", not "no question".

#### Scenario: Blank Terminal Output

- **GIVEN** a monitored idle agent whose captured output is entirely whitespace
- **WHEN** the monitoring tick processes the capture
- **THEN** question detection is skipped
- **AND** stuck detection proceeds normally (blank capture does NOT count as input-waiting)

## Skill: Re-Capture Before Send

### Requirement: Terminal SHALL Be Re-Captured Before Sending Auto-Answer

Before sending an auto-answer via `tmux send-keys`, the operator SHALL re-capture the terminal (`tmux capture-pane -t <pane> -p -l 20`). If the output changed since the initial capture, the send SHALL be aborted — the agent is no longer waiting.

#### Scenario: Agent Transitions Between Detection and Send

- **GIVEN** the monitoring tick detected a question in pane `%3`
- **WHEN** the operator re-captures terminal output before sending
- **AND** the output has changed since the initial capture
- **THEN** the auto-answer is aborted
- **AND** no keystrokes are sent to the pane

#### Scenario: Terminal Unchanged Between Detection and Send

- **GIVEN** the monitoring tick detected a question in pane `%3`
- **WHEN** the operator re-captures terminal output before sending
- **AND** the output is identical to the initial capture
- **THEN** the auto-answer proceeds normally via `tmux send-keys`

## Skill: Per-Answer Logging

### Requirement: Every Auto-Answer SHALL Be Reported Inline

Every auto-answer MUST be reported inline in the operator's terminal output using the format: `"{change}: auto-answered '{summary}' -> {answer}"`. For escalated questions (decision list item 6), the format is: `"{change}: can't determine answer for '{summary}'. Please respond."`.

There SHALL be no cooldown or retry limit — each question is evaluated independently.

#### Scenario: Auto-Answer Reported

- **GIVEN** the operator auto-answers a question for change `r3m7`
- **WHEN** the answer is sent
- **THEN** the operator outputs: `"r3m7: auto-answered 'Proceed?' -> y"`

## Skill: Structure Optimization

### Requirement: Monitoring Tick SHALL Use Delta-Only Description

The monitoring tick section SHALL describe only the two modifications to operator2's tick (simplified answer model replacing the two-tier classification, and the detection improvements), not the full 6-step listing inherited from operator2/operator3.

#### Scenario: Monitoring Tick Description

- **GIVEN** the operator4 skill file
- **WHEN** a reader examines the monitoring tick section
- **THEN** it describes only the changes from operator3's tick, not the inherited steps

### Requirement: Key Properties Table SHALL Include Only Novel Properties

The Key Properties table SHALL contain only 4 rows covering novel or modified properties, not the full 11-row table from operator3.

#### Scenario: Key Properties Table Size

- **GIVEN** the operator4 skill file
- **WHEN** a reader examines the Key Properties table
- **THEN** it contains approximately 4 rows

### Requirement: Purpose Section SHALL Be Condensed

The Purpose section SHALL drop the redundant "Key difference" paragraph and inline the launcher information. The overall section SHOULD be shorter than operator3's equivalent.

### Requirement: Guards SHALL Be Integrated Into Detection Steps

The idle-only guard and bottom-most indicator rule SHALL be integrated into the detection steps rather than appearing as separate paragraphs or subsections.

## Spec File

### Requirement: Spec File SHALL Follow Established Pattern

A spec file SHALL be created at `docs/specs/skills/SPEC-fab-operator4.md` following the structure of `SPEC-fab-operator3.md`. It SHALL include:

- Summary
- Primitives table (with `-l 20` flags reflecting the new capture window)
- Routing discipline section
- Question detection (with all guards as named subsections: Claude Turn Boundary Guard, Blank Capture Guard, Idle-Only Guard)
- Answer model (simplified, decision list)
- Re-capture before send
- Logging
- Monitoring tick changes
- Relationship to operator3 table
- Launcher details
- One-operator-at-a-time note
- Key properties
- Resolved design decisions

#### Scenario: Spec File Created

- **GIVEN** the apply stage executes
- **WHEN** the spec file task is completed
- **THEN** `docs/specs/skills/SPEC-fab-operator4.md` exists with all required sections

## Launcher Script

### Requirement: Launcher Script SHALL Follow operator3 Pattern

A launcher script SHALL be created at `fab/.kit/scripts/fab-operator4.sh` following the same singleton tab pattern as `fab-operator3.sh`:

- Requires tmux/byobu (`$TMUX` check)
- Singleton: switches to existing `operator` tab if it exists
- Creates new tmux window named `operator` running `claude --dangerously-skip-permissions '/fab-operator4'`

#### Scenario: First Launch

- **GIVEN** no `operator` tmux tab exists
- **WHEN** `fab-operator4.sh` is executed
- **THEN** a new tmux window named `operator` is created
- **AND** it runs `claude --dangerously-skip-permissions '/fab-operator4'`

#### Scenario: Singleton Enforcement

- **GIVEN** an `operator` tmux tab already exists
- **WHEN** `fab-operator4.sh` is executed
- **THEN** it switches to the existing `operator` tab
- **AND** does NOT create a new window

## Design Decisions

1. **All-auto-answer over two-tier classification**: Worktree isolation and human PR merge provide the safety gate. The operator should not be a bottleneck — all questions are auto-answered unless the operator literally cannot determine what to send.
   - *Why*: Operator3's escalation tier stalls the pipeline for questions that are safe to answer. The real safety gate is at the PR level, not the operator level.
   - *Rejected*: Multi-tier confidence model — adds complexity without meaningful safety improvement given the PR safety net.

2. **Decision list over prose heuristic for "what to send"**: A numbered decision list (items 1-6) provides deterministic priority ordering for answer selection.
   - *Why*: Prose heuristics are ambiguous; a priority-ordered list ensures the operator processes answer patterns in a consistent, deterministic order.
   - *Rejected*: Prose heuristic (operator3 style) — harder to reason about edge cases and priority conflicts.

3. **Re-capture before send over single-tick grace period**: The operator re-captures terminal output immediately before sending. This eliminates the race condition between idle check and send.
   - *Why*: Single-tick grace period was considered but rejected — it adds latency without fully solving the race condition.
   - *Rejected*: Single-tick grace period — delays answers by one full monitoring cycle and doesn't guarantee safety.

4. **Independent operator4 over modifying operator3**: Operator4 is a new skill file, not a modification to operator3.
   - *Why*: User explicitly decided to keep all operators independent. This preserves operator3 as-is and avoids regression risk.
   - *Rejected*: Modifying operator3 in place, extracting shared base between operators.

5. **Claude turn boundary guard (`>` cursor detection)**: Checks last 2 lines for Claude's prompt cursor to prevent false positives from Claude's own conversational output.
   - *Why*: Claude's output often contains question-like phrasing ("Would you like me to...?") that triggers detection. The `>` cursor indicates the agent is at a normal human-turn boundary, not a blocking prompt.
   - *Rejected*: Excluding all question-mark lines from Claude — too broad, would miss genuine blocking prompts from Claude.

6. **`/fab-fff` for autopilot over `/fab-ff`**: An upcoming change will make `/fab-fff` confidence-gated, aligning it with the operator's autopilot needs.
   - *Why*: `/fab-fff` is the more autonomous pipeline variant, fitting for operator-driven autopilot where human interaction is minimized.
   - *Rejected*: Keeping `/fab-ff` — its interactive fallback on review failure conflicts with the operator's autonomous mode.

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Auto-answer all questions, no escalation tier | Confirmed from intake #1 — worktree isolation + human PR merge is the safety gate | S:95 R:85 A:90 D:95 |
| 2 | Certain | Add routing discipline section | Confirmed from intake #2 — operator must route, never execute directly | S:95 R:90 A:90 D:95 |
| 3 | Certain | Add per-answer inline logging | Confirmed from intake #3 — every auto-answer must be reported | S:90 R:90 A:85 D:95 |
| 4 | Certain | Add Claude turn boundary guard | Confirmed from intake #4 — prevents false positives from Claude output | S:90 R:90 A:85 D:90 |
| 5 | Certain | Add re-capture before send guard | Confirmed from intake #5 — eliminates race condition | S:90 R:85 A:85 D:90 |
| 6 | Certain | Increase capture window to `-l 20` | Confirmed from intake #6 — compensates for wrapping | S:85 R:90 A:85 D:90 |
| 7 | Certain | Tighten `?` to last non-empty line only, <120 chars, skip comment prefixes | Confirmed from intake #7 — reduces false positives | S:90 R:90 A:85 D:90 |
| 8 | Certain | Add `:` endings, enumerated options, `Press.*key` patterns | Confirmed from intake #8 — covers common prompt formats | S:85 R:90 A:80 D:85 |
| 9 | Certain | Apply all optimization suggestions (~26% token reduction) | Confirmed from intake #9 — delta-only tick, reduced key properties, condensed purpose, integrated guards | S:90 R:90 A:90 D:95 |
| 10 | Certain | Keep all operators independent | Confirmed from intake #10 — no merging or extracting shared base | S:95 R:85 A:85 D:95 |
| 11 | Certain | Create as operator4, not modify operator3 | Confirmed from intake #11 — purely additive | S:95 R:90 A:90 D:95 |
| 12 | Certain | Launcher script follows operator3.sh pattern | Clarified — codebase confirms identical singleton pattern in operator1/2/3 launchers <!-- clarified: launcher pattern verified from codebase precedent --> | S:95 R:90 A:85 D:90 |
| 13 | Certain | Numbered decision list for "what to send" (items 1-6) | Clarified — decision list items and priority ordering locked in spec <!-- clarified: decision list finalized in spec with explicit items 1-6 --> | S:95 R:85 A:80 D:80 |
| 14 | Certain | Autopilot uses `/fab-fff` instead of `/fab-ff` | Confirmed from intake #14 — fff will become confidence-gated | S:95 R:80 A:85 D:95 |
| 15 | Certain | Spec file follows SPEC-fab-operator3.md structure | Codebase pattern — all operator specs follow the same structure | S:90 R:90 A:90 D:95 |
| 16 | Certain | Bottom-most indicator rule integrated into detection steps | Spec-level decision — optimization from intake #9 | S:85 R:90 A:90 D:95 |
| 17 | Certain | Idle-only guard integrated into detection steps | Spec-level decision — optimization from intake #9 | S:85 R:90 A:90 D:95 |

17 assumptions (17 certain, 0 confident, 0 tentative, 0 unresolved).
