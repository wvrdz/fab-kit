# _cli-rk

## Summary

Run-kit (rk) capability reference — iframe windows, proxy URL pattern, visual display recipe. Always loaded via `_preamble.md`; fails silently if rk is not installed. Centralized recipe enables any skill to show HTML content to the user via rk iframe windows.

## Flow

```
Skill loads always-load layer (via _preamble.md)
│
├─ Read: .claude/skills/_cli-rk/SKILL.md
│        (skip gracefully if missing)
│
└─ At use-time (when skill wants to display HTML):
   ├─ Bash: command -v rk (detection — skip all if absent)
   ├─ Bash: rk context (server URL discovery)
   ├─ Bash: python3 -m http.server <port> (serve HTML)
   └─ Bash: tmux new-window + set-option @rk_type/@rk_url (iframe)
```

### Tools used

| Tool | Purpose |
|------|---------|
| Read | Skill file (always-load layer) |
| Bash | `command -v rk`, `rk context`, HTTP server, tmux iframe commands |

### Sub-agents

None.
