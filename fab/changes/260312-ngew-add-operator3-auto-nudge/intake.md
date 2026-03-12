# Intake: Add Operator3 — Auto-Nudge for Blocked Agents

**Change**: 260312-ngew-add-operator3-auto-nudge
**Created**: 2026-03-12
**Status**: Draft

## Origin

> Add operator3 skill — extends operator2 with auto-nudge capability for agents waiting on user input. When monitoring detects an agent is idle AND terminal content shows a question/prompt (via tmux capture-pane heuristic), operator3 assesses question complexity: auto-answers routine operational questions and escalates judgment calls to the user.

Preceded by a `/fab-discuss` conversation that explored the design space. Key decisions from that discussion are captured in the Assumptions table below.

## Why

1. **Problem**: When running multiple agents in parallel (via operator1/2 + autopilot), agents frequently block on user input — permission prompts ("Should I commit these files before rebasing?"), confirmation requests ("Proceed with the changes?"), and simple yes/no questions. Each blocked agent stalls its entire pipeline until a human notices and responds.

2. **Consequence**: Without operator3, the user must constantly monitor all agent panes for input prompts. In autopilot mode with 3-5 agents, this defeats the purpose of automation — the user becomes the bottleneck. Operator2's monitoring detects stuck agents but only reports them; it doesn't resolve the block.

3. **Approach**: Extend operator2's monitoring infrastructure to detect "waiting for user input" (distinct from "slow but working") and auto-answer routine questions. The operator reads the agent's terminal via `tmux capture-pane`, assesses question complexity, and either answers directly or escalates to the user. This is NOT lifecycle enforcement — the operator isn't advancing stages or making pipeline decisions. It's acting as a proxy for the user on routine operational questions.

## What Changes

### New Skill: `fab/.kit/skills/fab-operator3.md`

A new operator skill that inherits all of operator2's behavior (monitoring, enrollment, `/loop` ticks, all use cases) and adds:

#### Question Detection (Terminal Heuristic)

During each monitoring tick, for each monitored agent that is **idle** (per pane-map Agent column):

1. Capture the last ~10 lines of the agent's terminal via `tmux capture-pane -t <pane> -p -l 10`
2. Scan for question indicators:
   - Lines ending with `?`
   - `[Y/n]`, `[y/N]`, `(y/n)`, `(yes/no)` patterns
   - `Allow?`, `Approve?`, `Confirm?`, `Proceed?` keywords
   - Claude Code permission prompts (tool approval patterns)
   - `Do you want to...`, `Should I...`, `Would you like...` phrasing
3. If no question indicators found → treat as normal idle (operator2 behavior: stuck detection after threshold)
4. If question indicators found → proceed to Answer Confidence assessment

The heuristic is intentionally broad — false positives are handled by the confidence model (escalate when unsure), not by tightening detection.

#### Answer Confidence Model

When a question is detected, operator3 classifies it into one of two tiers:

**Auto-answer** — one obvious right answer given the pipeline context:
- Permission/approval prompts → approve (the agent is running in `--dangerously-skip-permissions` anyway in autopilot; in non-autopilot, the user dispatched the work)
- "Should I commit before rebasing?" → yes
- "Proceed with the changes?" → yes
- "Continue?" → yes
- Simple operational confirmations where saying "no" would just stall the pipeline

**Escalate** — multiple valid paths, judgment required:
- "Tests are failing, should I fix them or skip?" → escalate
- "Which approach should I take: A or B?" → escalate
- "Should I delete these files?" → escalate (destructive)
- Any question involving tradeoffs, architecture decisions, or irreversible actions

The classification heuristic: if the question is a binary yes/no where "yes" continues the work and "no" stalls it, and the question doesn't involve destructive or branching choices → auto-answer. Everything else → escalate.

#### Updated Monitoring Tick Behavior

The monitoring tick gains a new step between existing detection steps:

1. *(existing)* Stage advance detection
2. *(existing)* Pipeline completion detection
3. *(existing)* Review failure detection
4. *(existing)* Pane death detection
5. **NEW: Input-waiting detection** — for idle agents, run the terminal heuristic. If question detected:
   - Auto-answer tier: send the answer via `tmux send-keys`, report "{change}: auto-answered '{summary of question}' → {answer}"
   - Escalate tier: report "{change}: waiting for input — '{summary of question}'. Please respond."
6. *(existing)* Stuck detection (only for agents NOT detected as input-waiting)

Note: input-waiting detection runs **before** stuck detection. An agent waiting for input is not "stuck" — it's blocked. Stuck detection should only flag agents that are idle without a visible question.

#### Nudge Budget

Auto-answers have no retry limit per se — each question is a distinct event. However, if operator3 auto-answers a question and the **same agent asks another question within 2 minutes**, operator3 escalates instead of auto-answering. This prevents runaway auto-answer loops where the agent keeps asking follow-up questions that compound.

### New Launcher Script: `fab/.kit/scripts/fab-operator3.sh`

Mirror of `fab-operator2.sh`, launching `/fab-operator3` in the singleton `operator` tmux tab.

### Updated Memory: `docs/memory/fab-workflow/execution-skills.md`

Add `/fab-operator3` section documenting the new behavior, how it extends operator2, and the question detection + confidence model.

### Updated Spec: `docs/specs/skills/SPEC-fab-operator3.md`

New per-skill spec for operator3.

## Affected Memory

- `fab-workflow/execution-skills`: (modify) Add `/fab-operator3` section covering question detection, answer confidence model, and monitoring tick changes

## Impact

- **Skill files**: New `fab/.kit/skills/fab-operator3.md`
- **Scripts**: New `fab/.kit/scripts/fab-operator3.sh`
- **Specs**: New `docs/specs/skills/SPEC-fab-operator3.md`
- **No CLI changes**: Uses existing `tmux capture-pane`, `fab resolve --pane`, `tmux send-keys`, `fab pane-map`
- **No schema changes**: No new fields in `.status.yaml` or `.fab-runtime.yaml`
- **Backward compatible**: Operator1 and operator2 remain unchanged; users choose which operator to run

## Open Questions

- Should the terminal heuristic patterns be configurable (e.g., in `config.yaml`), or is the hardcoded set sufficient for v1?

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Operator3 extends operator2 (not replaces) | Discussed — user confirmed incremental evolution model matching operator1→2 pattern | S:95 R:90 A:95 D:95 |
| 2 | Certain | "Stuck" means waiting for user input, not slow | Discussed — user explicitly defined: "the definition of nudge is to give inputs to agents that are waiting for user input - not agents that are slow" | S:95 R:95 A:95 D:95 |
| 3 | Certain | Detection via terminal heuristic (tmux capture-pane) | Discussed — user chose option 1 (terminal heuristic) over time-gated heuristic or autopilot-only | S:90 R:80 A:85 D:90 |
| 4 | Certain | Two-tier confidence model: auto-answer vs escalate | Discussed — user agreed confidence model is the right approach for question complexity | S:90 R:85 A:85 D:90 |
| 5 | Certain | No audit trail for now | Discussed — user agreed audit trail is overkill for v1 | S:95 R:95 A:90 D:95 |
| 6 | Certain | "Not a Lifecycle Enforcer" still holds | Discussed — user confirmed "not enforcing, just nudging recovery" | S:90 R:90 A:90 D:90 |
| 7 | Confident | Nudge is not /fab-continue but answering the actual question | Discussed — user clarified with the "commit before rebasing" example; operator3 reads and responds to the specific question | S:85 R:80 A:80 D:75 |
| 8 | Confident | No skip-on-failure behavior | Discussed — user said "No, don't skip" when asked about nudge-then-skip | S:85 R:85 A:80 D:80 |
| 9 | Tentative | Rapid follow-up questions (same agent within 2m) trigger escalation | Reasonable guard against auto-answer loops, not explicitly discussed | S:50 R:70 A:70 D:60 |
<!-- assumed: 2-minute cooldown for follow-up questions — prevents runaway auto-answer chains without explicit user discussion -->
| 10 | Tentative | Terminal heuristic patterns are hardcoded in the skill file (not configurable) | Simpler for v1; config.yaml extension can come later if needed | S:55 R:85 A:75 D:65 |
<!-- assumed: hardcoded patterns — no user discussion on configurability, but high reversibility -->

10 assumptions (6 certain, 2 confident, 2 tentative, 0 unresolved).
