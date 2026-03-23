## Operator

- [ ] [gogt] 2026-03-19: Write operator6 from hand. Remove references to byobu in it.
- [ ] [02eh] 2026-03-18: When operator5 works in branch mode for matching a change, it asks for confirmation. Don't do so, just proceed
- [ ] [rd3u] 2026-03-19: operator should mention on each tick the current timestamp
- [ ] [djkp] 2026-03-19: operator's instuction of rebase to latest is misleading. Should mention - rebase to latest origin/main
## Open

- [ ] [ngaw] 2026-02-23: Quality gate - how to decide which PR has had deep thought vs just surface level?
- [ ] [v34t] 2026-02-23: A timeline or user journey mermaid diagram showing which commands are typed in the main repo vs the worktree
- [ ] [lc3m] 2026-03-01: Switch fab/.kit/LICENSE from PolyForm Internal Use to MIT. Reasoning: fab-kit is a process (prompts, markdown, shell scripts) — not protectable IP in the copyright sense. The US Copyright Office ruled prompts convey unprotectable ideas. PolyForm Internal Use blocks community contributions and ecosystem growth while providing only illusory control. MIT gives attribution (copyright notice preserved), maximum adoption, zero friction, and lets the community keep the tool evolving — the real moat is speed and opinionation, not licensing restrictions.
- [ ] [q0lw] 2026-03-11: If fab binary or wt or idea binary not found, stop. Add to preamble?
- [ ] [heht] 2026-03-19: Ensure randomness in worktree names is actually getting followed - inrease the name universe even further
- [ ] [p4ki] 2026-03-19: Allow 'idea <text of idea>' format once again
- [ ] [1tch] 2026-03-20: Change fab-switch --blank to fab-switch --none or just fab-switch. --blank isn't easy to remember

## Done

- [x] [t13m] 2026-03-18: Configurable agent spawn command in config.yaml — centralize the agent binary/flags so all spawn scripts and operator read from one place. Default: `claude --dangerously-skip-permissions --effort max -n "$(basename "$(pwd)")"`. Currently hardcoded in 4+ locations.
