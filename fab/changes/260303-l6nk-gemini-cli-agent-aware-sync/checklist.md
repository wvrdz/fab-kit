# Quality Checklist: Gemini CLI Agent-Aware Sync

**Change**: 260303-l6nk-gemini-cli-agent-aware-sync
**Generated**: 2026-03-04
**Spec**: `spec.md`

## Functional Completeness

- [ ] CHK-001 Gemini CLI skill deployment: `.gemini/skills/<name>/SKILL.md` created for every skill when `gemini` available
- [ ] CHK-002 Agent detection: `command -v` used to check each agent's CLI command before syncing
- [ ] CHK-003 Skip message: `Skipping {agent}: {command} not found in PATH` printed for absent agents
- [ ] CHK-004 No-agent warning: warning printed when no agents detected, script continues (exit 0)
- [ ] CHK-005 Gitignore entry: `/.gemini` present in `fab/.kit/scaffold/fragment-.gitignore`

## Behavioral Correctness

- [ ] CHK-006 Existing agent sync unchanged: Claude Code, OpenCode, Codex sync behavior identical when their CLIs are present
- [ ] CHK-007 Existing dot folders preserved: absent agents do not cause deletion of existing dot folders
- [ ] CHK-008 Detection scope: scaffold tree-walk (section 2) remains unconditional

## Scenario Coverage

- [ ] CHK-009 Gemini skills synced on first run: verified via test
- [ ] CHK-010 Agent absent — sync skipped: verified via test
- [ ] CHK-011 No agents — warning printed: verified via test
- [ ] CHK-012 Partial availability — only detected agents synced: verified via test
- [ ] CHK-013 Stale Gemini skills cleaned: `clean_stale_skills` called for `.gemini/skills/`

## Edge Cases & Error Handling

- [ ] CHK-014 Script exits 0 even when no agents found
- [ ] CHK-015 Idempotency: running sync twice with same agents produces same result

## Code Quality

- [ ] CHK-016 Pattern consistency: agent block structure matches existing Claude/OpenCode/Codex pattern
- [ ] CHK-017 No unnecessary duplication: reuses `sync_agent_skills` and `clean_stale_skills` (no new copy functions)
- [ ] CHK-018 Readability: follows project principle of readability over cleverness
- [ ] CHK-019 No god functions: helper function is small and focused
- [ ] CHK-020 No magic strings: agent names and commands appear as clear literals

## Documentation Accuracy

- [ ] CHK-021 Gitignore scaffold includes `/.gemini` entry

## Cross References

- [ ] CHK-022 Spec matches implementation: all spec requirements have corresponding code

## Notes

- Check items as you review: `- [x]`
- All items must pass before `/fab-continue` (hydrate)
- If an item is not applicable, mark checked and prefix with **N/A**: `- [x] CHK-008 **N/A**: {reason}`
