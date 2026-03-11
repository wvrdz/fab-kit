# Spec: Add Operator2 Monitoring Skill

**Change**: 260311-5c11-add-operator2-monitoring-skill
**Created**: 2026-03-11
**Affected memory**: `docs/memory/fab-workflow/execution-skills.md`

## Non-Goals

- Modifying `fab-operator1.md` behavior — operator1 remains unchanged as the reactive model
- Adding new CLI subcommands — operator2 uses existing `fab pane-map`, `fab send-keys`, and `fab status` primitives
- Background daemon behavior — operator2 is still a Claude session that uses `/loop` for periodic input, not a polling service

## Operator2 Skill: Core Monitoring

### Requirement: Monitor-After-Action

When operator2 sends a command to an agent via `fab send-keys`, it SHALL enroll the target change in a monitoring set and start or extend a `/loop` at configurable interval (default 5m) to periodically check progress.

#### Scenario: First send enrolls monitoring

- **GIVEN** operator2 has no active monitoring loop
- **WHEN** the operator sends a command to agent r3m7 via `fab send-keys`
- **THEN** r3m7 is added to the monitored set with its current stage and agent state as baseline
- **AND** a `/loop 5m` is started to check progress periodically

#### Scenario: Subsequent send extends monitoring

- **GIVEN** a monitoring loop is already running for change r3m7
- **WHEN** the operator sends a command to agent k8ds
- **THEN** k8ds is added to the monitored set
- **AND** the existing loop continues (no second loop started)

#### Scenario: Send without monitoring (read-only actions)

- **GIVEN** the operator performs a read-only action (status check, pane map refresh)
- **WHEN** the action completes
- **THEN** no change is enrolled in monitoring (monitoring applies only to commands sent via `fab send-keys`)

### Requirement: Monitoring Tick Behavior

Each monitoring tick SHALL re-query `fab pane-map`, compare current state against the last known state for all monitored changes, and report transitions.

#### Scenario: Stage advance detected

- **GIVEN** change r3m7 was last seen at stage `spec`
- **WHEN** the monitoring tick runs and `fab pane-map` shows r3m7 at stage `tasks`
- **THEN** the operator reports: "r3m7: spec -> tasks"
- **AND** updates the baseline to `tasks`

#### Scenario: Pipeline completion detected

- **GIVEN** change r3m7 was last seen at stage `review`
- **WHEN** the monitoring tick runs and r3m7 is at `hydrate` or later
- **THEN** the operator reports: "r3m7: reached hydrate -- pipeline complete"
- **AND** removes r3m7 from the monitored set

#### Scenario: Review failure detected

- **GIVEN** change r3m7 was last seen at stage `review` with agent active
- **WHEN** the monitoring tick shows r3m7 at stage `apply` (reset after review failure)
- **THEN** the operator reports: "r3m7: review failed, reworking (back at apply)"

#### Scenario: Pane death detected

- **GIVEN** change r3m7 is in the monitored set mapped to pane %3
- **WHEN** the monitoring tick runs and r3m7 no longer appears in `fab pane-map`
- **THEN** the operator reports: "r3m7: pane %3 is gone -- agent exited"
- **AND** removes r3m7 from the monitored set

#### Scenario: Stuck detection

- **GIVEN** change r3m7 has been at stage `apply` with agent state `idle` for more than the stuck threshold (default 15m)
- **WHEN** the monitoring tick runs
- **THEN** the operator reports: "r3m7: idle at apply for 15m -- may be stuck"
- **AND** does NOT auto-nudge (nudging requires explicit UC6 invocation or user instruction)

### Requirement: Terminal State Removal

A change SHALL be removed from monitoring when it reaches a terminal state: `hydrate`, `ship`, or `review-pr` stage; or when the user explicitly says to stop monitoring it; or when its pane dies.

#### Scenario: User stops monitoring

- **GIVEN** change r3m7 is in the monitored set
- **WHEN** the user says "stop monitoring r3m7"
- **THEN** r3m7 is removed from the monitored set
- **AND** if no other changes are monitored, the loop is stopped

#### Scenario: All changes reach terminal state

- **GIVEN** r3m7 and k8ds are monitored
- **WHEN** both reach `hydrate` or later
- **THEN** both are removed from the monitored set
- **AND** the loop is stopped
- **AND** the operator reports: "All monitored changes complete."

### Requirement: Loop Lifecycle

The `/loop` SHALL be started when the first change is enrolled in monitoring and stopped when the monitored set becomes empty. Only one loop SHALL be active at a time.

#### Scenario: Loop stops on empty set

- **GIVEN** the monitoring loop is running with one change
- **WHEN** that change reaches hydrate
- **THEN** the change is removed from the monitored set
- **AND** the loop is stopped

## Operator2 Skill: Inherited Use Cases

### Requirement: Full Capability Parity with Operator1

Operator2 SHALL support all operator1 use cases (UC1-UC8) with identical behavior except where monitoring integration modifies them.

#### Scenario: UC1 broadcast with auto-enroll

- **GIVEN** operator2 broadcasts `/fab-continue` to 3 idle agents
- **WHEN** the broadcast completes
- **THEN** all 3 agents are enrolled in monitoring
- **AND** the operator reports the sends and starts/extends the monitoring loop

#### Scenario: UC2 sequenced rebase via monitoring

- **GIVEN** the user says "when r3m7 finishes, rebase k8ds on main"
- **WHEN** the monitoring tick detects r3m7 at `hydrate` or later
- **THEN** the operator automatically sends the rebase command to k8ds via `fab send-keys`
- **AND** reports: "r3m7 finished. Sending rebase to k8ds."

#### Scenario: UC3 merge completed PRs (unchanged)

- **GIVEN** changes at `ship` or `review-pr` stage exist
- **WHEN** the user asks to merge PRs
- **THEN** the operator confirms before executing (destructive), same as operator1

#### Scenario: UC4 spawn worktree (unchanged)

- **GIVEN** the user asks to start a new idea
- **WHEN** the operator spawns a worktree and opens a new tab
- **THEN** behavior is identical to operator1

#### Scenario: UC5 status dashboard (unchanged)

- **GIVEN** the user asks for status
- **WHEN** the operator refreshes the pane map
- **THEN** it presents the same human-readable summary as operator1

#### Scenario: UC6 unstick with monitoring

- **GIVEN** the user says "nudge r3m7"
- **WHEN** the operator sends `/fab-continue` to r3m7
- **THEN** r3m7 is enrolled in monitoring to check recovery
- **AND** standard bounded retry limits apply (1 nudge max)

#### Scenario: UC7 notification via monitoring

- **GIVEN** the user says "tell me when r3m7 finishes"
- **WHEN** operator2 receives this instruction
- **THEN** r3m7 is enrolled in monitoring (if not already)
- **AND** the monitoring loop handles notification automatically (no "check on next user interaction" caveat)

#### Scenario: UC8 autopilot (unchanged)

- **GIVEN** the user requests autopilot with a list of changes
- **WHEN** autopilot runs
- **THEN** behavior is identical to operator1's autopilot (already uses `/loop`)

### Requirement: Confirmation Model

Operator2 SHALL use the same three-tier confirmation model as operator1: read-only (no confirmation), recoverable (announce before sending), destructive (confirm before executing).

#### Scenario: Destructive action requires confirmation

- **GIVEN** the user asks to merge a PR
- **WHEN** the operator prepares to execute
- **THEN** it confirms: "Will merge PR for {change}. Proceed?"

### Requirement: Pre-Send Validation

Operator2 SHALL verify pane existence and agent idle state before every `fab send-keys` invocation, identical to operator1.

#### Scenario: Send to busy agent

- **GIVEN** agent r3m7 is active (not idle)
- **WHEN** the operator attempts to send a command
- **THEN** it warns: "{change} is currently active. Sending may corrupt its work. Send anyway?"
- **AND** only sends on user confirmation

### Requirement: State Re-derivation

Operator2 SHALL re-query `fab pane-map` before every action. It SHALL NOT rely on stale values from conversation memory.

#### Scenario: Pane died between actions

- **GIVEN** the operator previously saw r3m7 at pane %3
- **WHEN** the user asks to send a command to r3m7 and the pane map no longer shows it
- **THEN** the operator reports: "Pane for r3m7 is gone (agent exited or tab closed)."

### Requirement: Context Discipline

Operator2 SHALL NOT load change artifacts (intakes, specs, tasks). Its context window is reserved for coordination state. Terminal output inspection uses `tmux capture-pane` for diagnosis.

#### Scenario: User asks what went wrong

- **GIVEN** agent r3m7 appears stuck
- **WHEN** the user asks for diagnostics
- **THEN** the operator uses `tmux capture-pane -t <pane> -p` and reports the output
- **AND** does NOT read r3m7's spec or tasks files

## Operator2 Skill: Configuration

### Requirement: Configurable Monitoring Interval

The monitoring interval SHALL default to 5 minutes and MAY be overridden by the user via natural language (e.g., "check every 2 minutes").

#### Scenario: Custom interval

- **GIVEN** the user says "monitor every 2m"
- **WHEN** the operator starts the loop
- **THEN** it uses `/loop 2m` instead of the default `/loop 5m`

### Requirement: Configurable Stuck Threshold

The stuck detection threshold SHALL default to 15 minutes. The operator MAY adjust it per user instruction.

#### Scenario: Custom stuck threshold

- **GIVEN** the user says "flag agents stuck for more than 10 minutes"
- **WHEN** the monitoring tick runs
- **THEN** agents idle at a non-terminal stage for >10 minutes are flagged

## Launcher Scripts

### Requirement: New Launcher `fab-operator2.sh`

A new launcher script `fab/.kit/scripts/fab-operator2.sh` SHALL create a singleton tmux tab named `operator` and invoke `/fab-operator2` in a new Claude session. The singleton pattern SHALL match `fab-operator.sh`.

#### Scenario: Launch operator2

- **GIVEN** no `operator` tab exists in the tmux session
- **WHEN** the user runs `fab-operator2.sh`
- **THEN** a new tmux tab named `operator` is created
- **AND** it runs `claude --dangerously-skip-permissions '/fab-operator2'`

#### Scenario: Tab already exists

- **GIVEN** an `operator` tab already exists
- **WHEN** the user runs `fab-operator2.sh`
- **THEN** the script switches to the existing tab
- **AND** does NOT create a new one

### Requirement: Rename Existing Launcher

`fab/.kit/scripts/fab-operator.sh` SHALL be renamed to `fab/.kit/scripts/fab-operator1.sh`. The tab name SHALL be changed from `operator1` to `operator`. The skill invocation (`/fab-operator1`) SHALL remain unchanged.

#### Scenario: Renamed launcher works

- **GIVEN** `fab-operator1.sh` exists (renamed from `fab-operator.sh`)
- **WHEN** the user runs it
- **THEN** a singleton tmux tab named `operator` is created running `/fab-operator1`

### Requirement: Operator1 Skill Launcher Reference Update

The launcher reference in `fab/.kit/skills/fab-operator1.md` SHALL be updated from `fab-operator.sh` to `fab-operator1.sh`.

#### Scenario: Launcher reference is correct

- **GIVEN** `fab-operator1.md` contains a launcher reference
- **WHEN** a user or agent reads the skill
- **THEN** the reference points to `fab-operator1.sh` (not the old `fab-operator.sh`)

## Spec File

### Requirement: New Spec for Operator2

A new spec file `docs/specs/skills/SPEC-fab-operator2.md` SHALL be created per the constitution requirement that skill changes include corresponding spec updates.

#### Scenario: Spec exists alongside skill

- **GIVEN** `fab/.kit/skills/fab-operator2.md` is created
- **WHEN** the change is complete
- **THEN** `docs/specs/skills/SPEC-fab-operator2.md` exists with a summary, primitives, discovery, use cases, monitoring behavior, interaction model, and guardrails sections

## Design Decisions

1. **Monitoring via conversation-held state, not persistent file**: The monitored set (change ID, last-known stage, last-known agent state, enrolled-at timestamp) is held in conversation context, not in a YAML file. This matches operator1's pattern for standing orders (UC7) and avoids introducing a new persistent state file.
   - *Why*: Simplicity — the monitored set is session-scoped. If the operator session ends, the state is lost, but `fab pane-map` allows full reconstruction on restart.
   - *Rejected*: Persisting to `.fab-runtime.yaml` — adds complexity, and the state is ephemeral by nature.

2. **Advisory stuck detection, no auto-nudge**: Stuck detection reports to the user but does not automatically send `/fab-continue`. Auto-nudging could corrupt an agent that's blocked on a legitimate long operation.
   - *Why*: Safety — sending commands to an agent that might be mid-operation violates the pre-send validation contract.
   - *Rejected*: Auto-nudge after idle threshold — too risky without understanding why the agent is idle.

3. **Shared `operator` tab name**: Both launchers use tab name `operator` (not `operator1`/`operator2`). Since only one runs at a time, this prevents stale tabs.
   - *Why*: The intake specified one operator at a time in a shared tab.
   - *Rejected*: Separate tab names — would allow accidental concurrent operators.

## Assumptions

| # | Grade | Decision | Rationale | Scores |
|---|-------|----------|-----------|--------|
| 1 | Certain | Separate skill file, not a modification to operator1 | Confirmed from intake #1 — user explicitly chose this | S:95 R:90 A:95 D:95 |
| 2 | Certain | Full capability parity with operator1 plus monitoring | Confirmed from intake #2 — workflow iteration, not stripped-down | S:95 R:85 A:95 D:95 |
| 3 | Certain | Use `/loop` for periodic monitoring | Confirmed from intake #3 — proven pattern from autopilot | S:90 R:90 A:90 D:95 |
| 4 | Certain | Pane-map only, no `fab runtime is-idle` | Confirmed from intake #4 — cross-worktree issue | S:95 R:90 A:95 D:95 |
| 5 | Certain | Default 5m monitoring interval | Confirmed from intake #5 — user proposed | S:90 R:95 A:85 D:90 |
| 6 | Certain | New launcher `fab-operator2.sh` | Confirmed from intake #6 | S:95 R:90 A:90 D:95 |
| 7 | Certain | Shared tab name `operator` | Confirmed from intake #7 — one operator at a time | S:90 R:85 A:90 D:90 |
| 8 | Certain | One operator at a time | Confirmed from intake #8 | S:90 R:85 A:90 D:90 |
| 9 | Certain | Create spec per constitution | Confirmed from intake #9 — constitutional requirement | S:95 R:80 A:95 D:95 |
| 10 | Confident | Stuck threshold 15m, advisory only | Confirmed from intake #10 — carried from operator1 autopilot guardrails | S:75 R:90 A:80 D:85 |
| 11 | Certain | Monitored set held in conversation context, not persistent file | Matches operator1's pattern for standing orders; session-scoped state | S:85 R:95 A:90 D:90 |
| 12 | Certain | Rename fab-operator.sh to fab-operator1.sh | Explicitly specified in intake — symmetric naming | S:95 R:90 A:95 D:95 |

12 assumptions (11 certain, 1 confident, 0 tentative, 0 unresolved).
