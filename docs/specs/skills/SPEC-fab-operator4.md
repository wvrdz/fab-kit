# fab-operator4

## Summary

Multi-agent coordination layer with redesigned auto-nudge capability. Extends `/fab-operator3` (which extends `/fab-operator2`) with a simplified answer model, improved question detection, a re-capture guard, routing discipline, and per-answer logging.

**Key difference from operator3**: Where operator3 uses a two-tier confidence model (auto-answer vs escalate), operator4 auto-answers all detected questions — the only escalation case is when the operator cannot determine what keystrokes to send. Detection is also tightened with a wider capture window, additional guards, and new indicator patterns.

Not a lifecycle enforcer — the operator answers operational questions as a proxy for the user, not advancing stages or making pipeline decisions.

---

## Primitives

Inherits all operator2 and operator3 primitives. Modified:

| Primitive | Source | Purpose |
|-----------|--------|---------|
| `tmux capture-pane -t <pane> -p -l 20` | tmux | Capture last ~20 lines of an agent's terminal (increased from operator3's `-l 10`) |
| `tmux send-keys -t <pane> "<answer>" Enter` | tmux | Send auto-answer to an agent's terminal (inherited, unchanged) |

All operator2 primitives remain: `fab pane-map`, `fab resolve --pane`, `wt list` + `fab change list`, `/loop`.

---

## Routing Discipline

The operator MUST NOT execute user instructions directly. When the user gives an instruction (e.g., "fix the tests", "add error handling"), the operator:

1. Determines which running tmux session/agent the request corresponds to
2. Routes the instruction to that agent via `tmux send-keys`
3. Enrolls the agent in monitoring

If the target agent is ambiguous, the operator asks the user which agent to route to. The operator's role is coordination, not implementation.

---

## Question Detection

### Terminal Heuristic

For each monitored agent that is **idle** (per pane-map Agent column), the monitoring tick captures the last ~20 lines of the agent's terminal via `tmux capture-pane -t <pane> -p -l 20` and scans for question indicators.

### Claude Turn Boundary Guard

If a Claude Code `>` prompt cursor (`^\s*>\s*$`) appears in the last 2 lines of captured output, skip question detection for that agent. The agent is at a normal human-turn boundary, not a blocking prompt. This prevents false positives from Claude's own conversational output.

### Blank Capture Guard

If captured output is entirely blank or whitespace, skip question detection for that tick. Treat as "cannot determine", not "no question". Stuck detection proceeds normally.

### Idle-Only Guard

Question detection runs ONLY for agents that are idle in the pane map. Active agents are not scanned. (Inherited from operator3.)

### Question Indicator Patterns

Inherited from operator3:
- Lines ending with `?`
- `[Y/n]`, `[y/N]`, `(y/n)`, `(yes/no)` patterns
- `Allow?`, `Approve?`, `Confirm?`, `Proceed?` keywords
- Claude Code permission prompts (tool approval patterns)
- `Do you want to...`, `Should I...`, `Would you like...` phrasing

New in operator4:
- Lines ending with `:` or `:\s*$` (CLI input prompts)
- Enumerated options (`[1-9]\)` patterns)
- `Press.*key`, `press.*enter`, `hit.*enter` (case-insensitive)

### Tightened `?` Pattern

The `?` pattern matches only on the **last non-empty line** of captured output (not any of the 20 lines). The matching line MUST be <120 characters. Lines starting with `#`, `//`, `*`, `>`, or timestamp patterns (e.g., `[2026-03-14`, `2026-03-14T`) are skipped as comments, log output, or search results.

### Bottom-Most Indicator Rule

When multiple question indicators appear in the captured output, evaluate the most recent (bottom-most) indicator. Consistent with standard terminal interaction patterns. (Inherited from operator3.)

---

## Answer Model

### Simplified Decision List

Replaces operator3's two-tier classification. All detected questions are auto-answered. Evaluate in order:

1. Binary yes/no or confirmation prompt -> `y`
2. `[Y/n]` or `[y/N]` prompt -> `y`
3. Claude Code permission/approval prompt -> `y`
4. Numbered menu or multi-choice -> `1` (first/default option)
5. Open-ended question where a concrete answer is determinable from visible terminal context -> send that answer
6. Question where the operator cannot determine what keystrokes to send -> escalate

### No Nudge Budget

No cooldown or retry limit. Each question is evaluated independently. The PR review step serves as the safety net.

---

## Re-Capture Before Send

Before sending an auto-answer via `tmux send-keys`, the operator re-captures the terminal (`tmux capture-pane -t <pane> -p -l 20`). If the output changed since the initial capture, the send is aborted — the agent is no longer waiting. This eliminates the race condition between detection and send.

---

## Per-Answer Logging

Every auto-answer is reported inline: `"{change}: auto-answered '{summary}' -> {answer}"`. For escalated questions (decision list item 6): `"{change}: can't determine answer for '{summary}'. Please respond."`.

No cooldown or retry limit — each question is evaluated independently.

---

## Monitoring Tick Changes

The monitoring tick inherits all 6 steps from operator2/operator3. Modifications to step 5 (input-waiting detection):

- **Answer model**: Uses the simplified decision list (items 1-6) instead of operator3's two-tier classification
- **Detection**: Uses improved question detection — `-l 20` capture, Claude turn boundary guard, blank capture guard, tightened `?` pattern, additional indicator patterns
- **Pre-send**: Adds the re-capture guard before every `tmux send-keys`

All other tick steps (stage advance, pipeline completion, review failure, pane death, stuck detection) are inherited unchanged.

---

## Relationship to Operator3

| Aspect | Operator3 | Operator4 |
|--------|-----------|-----------|
| Inherits from | `/fab-operator2` | `/fab-operator3` (and transitively `/fab-operator2`) |
| Answer model | Two-tier (auto-answer vs escalate) | Simplified — all auto-answer, escalate only when undeterminable |
| Capture window | `-l 10` | `-l 20` |
| `?` pattern | Any line in capture | Last non-empty line only, <120 chars, skip prefixes |
| Additional indicators | None | `:` endings, enumerated options, `Press.*key` |
| Claude turn boundary guard | Not present | New — `>` cursor in last 2 lines skips detection |
| Blank capture guard | Not present | New — skip detection on blank output |
| Re-capture before send | Not present | New — abort if output changed since detection |
| Per-answer logging | Inline report only | Same format, all questions logged (no escalation tier omission) |
| Routing discipline | Not present | New — operator routes, never executes directly |
| Autopilot pipeline | `/fab-ff` | `/fab-fff` |
| Monitoring | Proactive via `/loop` | Same — inherited |
| Use cases (UC1-UC8) | All 8 | Same — inherited |
| Confirmation model | 3-tier | Same — inherited |

### Launcher

Start via `fab/.kit/scripts/fab-operator4.sh` — creates a singleton tmux tab named `operator` and invokes `/fab-operator4` in a new Claude session.

### One Operator at a Time

Like operator1, operator2, and operator3, operator4 uses the shared singleton tmux tab name `operator`. Only one operator (1, 2, 3, or 4) runs at a time.

---

## Key Properties

| Property | Value |
|----------|-------|
| Inherits from | `/fab-operator3` (which inherits `/fab-operator2`) |
| Answer model | Simplified — all auto-answer, escalate only when keystrokes undeterminable |
| Capture window | `-l 20` (operator3 uses `-l 10`) |
| Autopilot pipeline | `/fab-fff` (operator3 uses `/fab-ff`) |

All other properties (requires active change, runs preflight, read-only, idempotent, advances stage, outputs `Next:` line, loads change artifacts, requires tmux, uses `/loop`) are inherited from operator3/operator2.

---

## Resolved Design Decisions

1. **All-auto-answer over two-tier classification.** Worktree isolation and human PR merge provide the safety gate. The operator should not be a bottleneck — all questions are auto-answered unless the operator literally cannot determine what to send. The two-tier model added pipeline latency without meaningful safety improvement given the PR safety net.

2. **Decision list over prose heuristic.** A numbered decision list (items 1-6) provides deterministic priority ordering for answer selection. Prose heuristics are ambiguous; a priority-ordered list ensures consistent, deterministic processing of answer patterns.

3. **Re-capture before send over single-tick grace period.** The operator re-captures terminal output immediately before sending, eliminating the race condition between idle check and send. Single-tick grace period was rejected — it adds latency without fully solving the race condition.

4. **Independent operator4 over modifying operator3.** Operator4 is a new skill file, not a modification to operator3. Preserves operator3 as-is and avoids regression risk. No shared base extraction — each operator remains independent.

5. **Claude turn boundary guard.** Checks last 2 lines for Claude's `>` prompt cursor to prevent false positives from Claude's conversational output. Rejecting the alternative of excluding all question-mark lines from Claude output — too broad, would miss genuine blocking prompts.

6. **`/fab-fff` for autopilot.** `/fab-fff` is the more autonomous pipeline variant, fitting for operator-driven autopilot where human interaction is minimized. `/fab-ff`'s interactive fallback on review failure conflicts with the operator's autonomous mode.
