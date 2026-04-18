# git-pr-review

## Summary

Processes PR review comments from any reviewer (human or bot). Fully autonomous — detects reviews, requests an automated Copilot review and polls up to 10 minutes for it to appear when no existing reviews are found, triages comments with disposition intent (fix/defer/skip), applies fixes, commits, pushes, and posts reply comments confirming outcomes.

## Arguments

- **`--tool <name>`** *(optional)* — Forces a specific review tool. Valid values: `copilot` only.

## Configuration

The `review_tools` block in `fab/project/config.yaml` controls whether Copilot is attempted:

```yaml
review_tools:
  copilot: true    # try GitHub Copilot (remote) — default when key is absent
```

Setting `copilot` to `false` skips Phase 2 entirely. When the `review_tools` key is absent, Copilot defaults to enabled.

## Flow

```
/git-pr-review [--tool <name>] invoked (user or sub-agent)
│
├─ Step 0: Start Review-PR Stage
│  └─ Bash: fab status start <change> review-pr git-pr-review
│
├─ Step 1: Resolve PR
│  ├─ Bash: gh pr view --json number,url
│  └─ Bash: gh repo view --json nameWithOwner
│
├─ Step 1.5: Parse --tool Flag
│  └─ Validate tool name (copilot only) or STOP on invalid
│
├─ Step 2: Detect Reviews and Route
│  ├─ Phase 1: Check existing reviews
│  │  ├─ Bash: gh api .../pulls/{n}/reviews
│  │  └─ Bash: gh api .../pulls/{n}/comments
│  │     └─ [if comments exist] → Step 3
│  │
│  └─ Phase 2: Copilot Review Request (no reviews found)
│     ├─ Read config: review_tools.copilot from fab/project/config.yaml
│     ├─ [copilot: false] "No automated reviewer available" → STOP (clean finish)
│     ├─ Bash: gh pr edit {n} --add-reviewer copilot-pull-request-reviewer
│     │  ├─ [success] Print "Copilot review requested. Waiting up to 10 minutes..."
│     │  │  └─ Poll: gh pr view --json reviews every 30s, up to 20 attempts
│     │  │     ├─ [review appears] → Step 3
│     │  │     └─ [20 attempts, no review] "...not yet available. Re-run /git-pr-review..." → STOP (clean finish)
│     │  └─ [failure] "No automated reviewer available..." → STOP (clean finish)
│
├─ Step 3: Fetch Comments (with id, node_id)
│  └─ Bash: gh api .../pulls/{n}/comments
│
├─ Step 4: Triage Comments
│  ├─ Classify intent: fix, defer, skip, or informational
│  ├─ Read: source files at {path}
│  └─ Edit: source files (targeted fixes for "fixed" comments)
│
├─ Step 5: Commit and Push
│  ├─ Bash: git add {files}
│  ├─ Bash: git commit -m "fix: address review feedback"
│  ├─ Bash: git push
│  └─ [no modifications] → proceed to Step 5.5 (don't stop)
│
├─ Step 5.5: Post Replies
│  ├─ Deduplicate: skip comments with existing disposition replies
│  ├─ Bash: gh api .../pulls/{n}/comments -f body=... -F in_reply_to=...
│  └─ Best-effort: failed POSTs logged, not fatal
│
└─ Step 6: Update Review-PR Stage
   ├─ [pass] Bash: fab status finish <change> review-pr
   └─ [fail] Bash: fab status fail <change> review-pr

Phase tracking (via yq directly on .status.yaml):
  waiting → received → triaging → fixing → pushed → replying
```

### Copilot Review Request (Phase 2)

Phase 2 runs when Phase 1 finds no existing reviews with inline comments. It requests a Copilot review and polls for up to 10 minutes:

| Tool | Type | Detection | On Success | On Failure |
|------|------|-----------|------------|------------|
| Copilot | Remote | Attempt `gh pr edit --add-reviewer copilot-pull-request-reviewer` | Poll 30s/attempt up to 20× — proceed to Step 3 when review appears; clean finish on timeout | Clean finish: "No automated reviewer available..." |

The `--tool copilot` flag forces the Copilot path regardless of config — the config check is skipped entirely when this flag is present. Without the flag, if `review_tools.copilot: false`, Phase 2 exits cleanly without attempting the request.

### Disposition taxonomy

Triage assigns **intent** (action verb); replies confirm **outcome** (past-tense).

| Intent (triage) | Reply (outcome) |
|-----------------|-----------------|
| `fix` | `Fixed — {description}. ({sha})` |
| `defer` | `Deferred — {reason}.` |
| `skip` | `Skipped — {reason}.` |

Informational comments receive no reply.

### Tools used

| Tool | Purpose |
|------|---------|
| Read | Source files for applying fixes |
| Edit | Source files (targeted fixes from review comments) |
| Bash | gh API calls (REST only), git operations, fab status commands, yq phase tracking |

### Sub-agents

None.

### Direct .status.yaml writes (via yq, not fab CLI)

| Field | When |
|-------|------|
| `stage_metrics.review-pr.phase` | At each phase transition (including `replying`) |
| `stage_metrics.review-pr.reviewer` | When reviews detected |
