---
name: fab-operator4
description: "Multi-agent coordination with redesigned auto-nudge — extends operator3 with simplified answer model, improved question detection, re-capture guard, and per-answer logging."
---

# /fab-operator4

> Read `fab/.kit/skills/_preamble.md` first, then read `fab/.kit/skills/fab-operator2.md`, then read `fab/.kit/skills/fab-operator3.md`. Operator4 inherits ALL of operator2 and operator3's behavior — monitoring, enrollment, `/loop` ticks, all use cases (UC1-UC8), confirmation model, pre-send validation, context discipline, configuration, question detection, and answer confidence model. Then return here for the overrides and additions defined below.

---

## Purpose

Extends operator3 with a redesigned auto-nudge system: simplified answer model (all questions auto-answered unless keystrokes are undeterminable), improved question detection (wider capture window, tightened patterns, new guards), a re-capture guard eliminating the detection-to-send race condition, and per-answer inline logging. Start via `fab/.kit/scripts/fab-operator4.sh` (singleton tmux tab named `operator`).

---

## Routing Discipline

The operator MUST NOT execute user instructions directly. When the user gives an instruction (e.g., "fix the tests", "add error handling"), the operator SHALL:

1. Determine which running tmux session/agent the request corresponds to
2. Route the instruction to that agent via `tmux send-keys`
3. Enroll the agent in monitoring

If the target agent is ambiguous (multiple agents running, no change specified), ask the user which agent to route to. The operator's role is coordination, not implementation.

---

## Autopilot Pipeline Override

Operator4 uses `/fab-fff` instead of `/fab-ff` for autopilot gate checks and pipeline invocations. All other autopilot behavior (inherited from operator2) is unchanged.

---

## Simplified Answer Model

Replaces operator3's two-tier auto-answer/escalate classification. All detected questions are auto-answered. The only escalation case is when the operator cannot determine what keystrokes to send.

**Decision list** (evaluate in order):

1. Binary yes/no or confirmation prompt -> `y`
2. `[Y/n]` or `[y/N]` prompt -> `y`
3. Claude Code permission/approval prompt -> `y`
4. Numbered menu or multi-choice -> `1` (first/default option)
5. Open-ended question where a concrete answer is determinable from visible terminal context -> send that answer
6. Question where the operator cannot determine what keystrokes to send -> escalate

No cooldown or retry limit — each question is evaluated independently.

---

## Question Detection Improvements

Overrides operator3's question detection with a wider capture window, tightened patterns, and additional guards.

### Capture Window

`tmux capture-pane -t <pane> -p -l 20` — increased from operator3's `-l 10` to compensate for line wrapping and verbose preambles.

### Claude Turn Boundary Guard

If a Claude Code `>` prompt cursor (`^\s*>\s*$`) appears in the last 2 lines of captured output, skip question detection for that agent. The agent is at a normal human-turn boundary, not a blocking prompt.

### Blank Capture Guard

If captured output is entirely blank or whitespace, skip question detection for that tick. Treat as "cannot determine", not "no question". Stuck detection proceeds normally.

### Tightened `?` Pattern

The `?` pattern matches only on the **last non-empty line** of captured output (not any of the 20 lines). The matching line MUST be <120 characters. Skip lines starting with `#`, `//`, `*`, `>`, or timestamp patterns (e.g., `[2026-03-14`, `2026-03-14T`) — these are comments, log output, or search results.

### Additional Indicator Patterns

Added to operator3's existing list:

- Lines ending with `:` or `:\s*$` (CLI input prompts)
- Enumerated options (`[1-9]\)` patterns)
- `Press.*key`, `press.*enter`, `hit.*enter` (case-insensitive)

All indicator patterns apply to idle agents only and evaluate the bottom-most match when multiple indicators appear.

---

## Re-Capture Before Send

Before sending an auto-answer via `tmux send-keys`, the operator SHALL re-capture the terminal (`tmux capture-pane -t <pane> -p -l 20`). If the output changed since the initial capture, abort — the agent is no longer waiting. This eliminates the race condition between detection and send.

---

## Per-Answer Logging

Every auto-answer MUST be reported inline: `"{change}: auto-answered '{summary}' -> {answer}"`. For escalated questions (decision list item 6): `"{change}: can't determine answer for '{summary}'. Please respond."`.

---

## Updated Monitoring Tick

The monitoring tick inherits all 6 steps from operator2/operator3. The modifications to step 5 (input-waiting detection) are:

- **Answer model**: Uses the simplified decision list (items 1-6) instead of operator3's two-tier classification
- **Detection**: Uses the improved question detection — `-l 20` capture window, Claude turn boundary guard, blank capture guard, tightened `?` pattern, additional indicator patterns
- **Pre-send**: Adds the re-capture guard before every `tmux send-keys`

All other tick steps (stage advance, pipeline completion, review failure, pane death, stuck detection) are inherited unchanged.

---

## Key Properties

| Property | Value |
|----------|-------|
| Inherits from | `/fab-operator3` (which inherits `/fab-operator2`) |
| Answer model | Simplified — all auto-answer, escalate only when keystrokes undeterminable |
| Capture window | `-l 20` (operator3 uses `-l 10`) |
| Autopilot pipeline | `/fab-fff` (operator3 uses `/fab-ff`) |
