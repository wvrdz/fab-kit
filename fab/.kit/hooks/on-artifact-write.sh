#!/usr/bin/env bash
# fab/.kit/hooks/on-artifact-write.sh — Claude Code PostToolUse hook (Write + Edit)
#
# Automatically runs bookkeeping when fab artifacts are written or edited:
#   intake.md  → infer change type + compute indicative confidence
#   spec.md    → compute confidence score
#   tasks.md   → count task items, update checklist total
#   checklist.md → set generated + count total/completed
#
# Registered for PostToolUse Write and PostToolUse Edit matchers.
# MUST exit 0 always — bookkeeping failures must never interrupt the agent.

# ── Read stdin JSON ──────────────────────────────────────────────────
input="$(cat)"

# Use jq if available (robust), fall back to grep/sed (best effort)
if command -v jq >/dev/null 2>&1; then
  file_path="$(echo "$input" | jq -r '.tool_input.file_path // empty' 2>/dev/null)"
else
  file_path="$(echo "$input" | grep -o '"file_path"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*: *"//;s/"$//')"
fi

[ -n "$file_path" ] || exit 0

# ── Pattern match: is this a fab artifact? ───────────────────────────
# Match fab/changes/{name}/artifact.md
case "$file_path" in
  */fab/changes/*/intake.md)   artifact="intake" ;;
  */fab/changes/*/spec.md)     artifact="spec" ;;
  */fab/changes/*/tasks.md)    artifact="tasks" ;;
  */fab/changes/*/checklist.md) artifact="checklist" ;;
  fab/changes/*/intake.md)     artifact="intake" ;;
  fab/changes/*/spec.md)       artifact="spec" ;;
  fab/changes/*/tasks.md)      artifact="tasks" ;;
  fab/changes/*/checklist.md)  artifact="checklist" ;;
  *) exit 0 ;;  # fast path — not an artifact
esac

# ── Derive change name from path ────────────────────────────────────
# Extract folder between fab/changes/ and /artifact.md
change_name="$(echo "$file_path" | sed -n 's|.*fab/changes/\([^/]*\)/.*|\1|p')"
[ -n "$change_name" ] || exit 0

# ── Locate fab CLI ──────────────────────────────────────────────────
repo_root="$(git rev-parse --show-toplevel 2>/dev/null)" || exit 0
fab_cmd="$repo_root/fab/.kit/bin/fab"
[ -x "$fab_cmd" ] || exit 0

# ── Get file content ────────────────────────────────────────────────
# Try stdin JSON content first, fall back to reading from disk
if command -v jq >/dev/null 2>&1; then
  content="$(echo "$input" | jq -r '.tool_input.content // empty' 2>/dev/null)"
else
  content=""
fi
if [ -z "$content" ]; then
  abs_path="$file_path"
  # If relative, make absolute from repo root
  case "$file_path" in
    /*) ;;
    *) abs_path="$repo_root/$file_path" ;;
  esac
  [ -f "$abs_path" ] && content="$(cat "$abs_path" 2>/dev/null)"
fi

# ── Per-artifact bookkeeping ────────────────────────────────────────
context_msg=""

case "$artifact" in
  intake)
    # Infer change type from content using keyword matching
    type="feat"
    content_lower="$(echo "$content" | tr '[:upper:]' '[:lower:]')"
    if echo "$content_lower" | grep -qE '\b(fix|bug|broken|regression)\b'; then
      type="fix"
    elif echo "$content_lower" | grep -qE '\b(refactor|restructure|consolidate|split|rename)\b'; then
      type="refactor"
    elif echo "$content_lower" | grep -qE '\b(docs|document|readme|guide)\b'; then
      type="docs"
    elif echo "$content_lower" | grep -qE '\b(test|spec|coverage)\b'; then
      type="test"
    elif echo "$content_lower" | grep -qE '\b(ci|pipeline|deploy|build)\b'; then
      type="ci"
    elif echo "$content_lower" | grep -qE '\b(chore|cleanup|maintenance|housekeeping)\b'; then
      type="chore"
    fi
    "$fab_cmd" status set-change-type "$change_name" "$type" 2>/dev/null || true
    # Compute indicative confidence
    score_out="$("$fab_cmd" score --stage intake "$change_name" 2>/dev/null)" || true
    context_msg="Bookkeeping: type=$type, indicative score computed"
    ;;

  spec)
    # Compute confidence score
    score_out="$("$fab_cmd" score "$change_name" 2>/dev/null)" || true
    context_msg="Bookkeeping: confidence score computed"
    ;;

  tasks)
    # Count task items (both checked and unchecked)
    total="$(echo "$content" | grep -cE '^\s*- \[([ x])\] ' 2>/dev/null || echo 0)"
    "$fab_cmd" status set-checklist "$change_name" total "$total" 2>/dev/null || true
    context_msg="Bookkeeping: tasks total=$total"
    ;;

  checklist)
    # Set generated flag
    "$fab_cmd" status set-checklist "$change_name" generated true 2>/dev/null || true
    # Count total items and completed items
    total="$(echo "$content" | grep -cE '^\s*- \[([ x])\] ' 2>/dev/null || echo 0)"
    completed="$(echo "$content" | grep -cE '^\s*- \[x\] ' 2>/dev/null || echo 0)"
    "$fab_cmd" status set-checklist "$change_name" total "$total" 2>/dev/null || true
    "$fab_cmd" status set-checklist "$change_name" completed "$completed" 2>/dev/null || true
    context_msg="Bookkeeping: checklist generated=true, total=$total, completed=$completed"
    ;;
esac

# ── Return additionalContext to agent ───────────────────────────────
if [ -n "$context_msg" ]; then
  printf '{"additionalContext": "%s"}\n' "$context_msg"
fi

exit 0
