# Quality Checklist: fab-proceed Orchestrator

**Change**: 260325-kxw7-fab-proceed-orchestrator
**Generated**: 2026-03-26
**Spec**: `spec.md`

## Functional Completeness

- [x] CHK-001 State Detection Pipeline: Skill detects active change, branch, unactivated intake, and conversation context in the specified order
- [x] CHK-002 Dispatch Table: Each detected state maps to the correct set of prefix steps per the spec table
- [x] CHK-003 Conversation Context Synthesis: When no intake exists, skill synthesizes description from conversation for `/fab-new`
- [x] CHK-004 fab-fff Terminal Delegation: `/fab-fff` is invoked via Skill tool (not subagent) as the final step
- [x] CHK-005 No Arguments: Skill header is `# /fab-proceed` with no argument placeholders
- [x] CHK-006 Subagent Dispatch: Prefix steps (fab-new, fab-switch, git-branch) dispatched as subagents with standard context files
- [x] CHK-007 Progress Reporting: Output includes detecting state header, per-step reports, and handoff line

## Behavioral Correctness

- [x] CHK-008 Error on empty context: Skill outputs error message and stops when no conversation context and no intake
- [x] CHK-009 Idempotent re-run: Re-invoking detects completed steps and skips them
- [x] CHK-010 No force passthrough: Skill does not pass `--force` or any flags to `/fab-fff`
- [x] CHK-011 Multiple unactivated intakes: Most recent by date prefix is selected

## Scenario Coverage

- [x] CHK-012 Active change + matching branch: Only `/fab-fff` dispatched
- [x] CHK-013 Active change, no branch: `/git-branch` → `/fab-fff`
- [x] CHK-014 Unactivated intake: `/fab-switch` → `/git-branch` → `/fab-fff`
- [x] CHK-015 Conversation context, no intake: `/fab-new` → `/fab-switch` → `/git-branch` → `/fab-fff`
- [x] CHK-016 No context, no intake: Error message, no sub-skills invoked

## Edge Cases & Error Handling

- [x] CHK-017 Sub-skill failure propagation: Errors from fab-new, fab-switch, git-branch surface to user and stop execution
- [x] CHK-018 fab-fff gate failure: Gate failure from fab-fff stops normally, skill does not retry
- [x] CHK-019 Spurious arguments: Extra arguments are silently ignored

## Code Quality

- [x] CHK-020 Pattern consistency: Skill file follows existing skill conventions (frontmatter, preamble reference, section structure)
- [x] CHK-021 No unnecessary duplication: Delegates to existing skills rather than reimplementing their logic

## Documentation Accuracy

- [x] CHK-022 State table updated: `_preamble.md` includes `/fab-proceed` in initialized and intake states
- [x] CHK-023 Skills spec updated: `docs/specs/skills.md` has `/fab-proceed` entry
- [x] CHK-024 Per-skill flow diagram: `docs/specs/skills/SPEC-fab-proceed.md` created


## Cross References

- [x] CHK-025 Skill registration: Skill file exists at `fab/.kit/skills/fab-proceed.md` with correct frontmatter
- [x] CHK-026 Deployed copy: `fab-sync.sh` produces `.claude/skills/fab-proceed/` directory

## Notes

- Check items as you review: `- [x]`
- All items must pass before `/fab-continue` (hydrate)
- If an item is not applicable, mark checked and prefix with **N/A**: `- [x] CHK-008 **N/A**: {reason}`
