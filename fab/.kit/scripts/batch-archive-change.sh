#!/usr/bin/env bash
set -euo pipefail

# batch-archive-change.sh — Per completed change: open a tmux tab in its
# worktree and start a Claude Code session that runs /fab-archive <change>.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KIT_DIR="$(dirname "$SCRIPT_DIR")"
FAB_DIR="$(dirname "$KIT_DIR")"
CHANGES_DIR="${FAB_DIR}/changes"
CONFIG_FILE="${FAB_DIR}/config.yaml"

usage() {
  cat <<'EOF'
Usage: batch-archive-change <change> [<change>...]

Per completed change (hydrate:done): opens a new tmux tab in its worktree
and starts a Claude Code session that runs /fab-archive <change>.

Options:
  --list    Show archivable changes (hydrate:done)
  --all     Open tabs for all archivable changes

Examples:
  batch-archive-change v3rn
  batch-archive-change v3rn ab12
  batch-archive-change --all
EOF
}

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

is_hydrate_done() {
  local status_file="$1"
  [[ -f "$status_file" ]] || return 1
  grep -qE '^\s*hydrate:\s*done' "$status_file"
}

list_archivable() {
  echo "Archivable changes (hydrate:done):"
  echo ""
  local count=0
  for dir in "$CHANGES_DIR"/*/; do
    [[ -d "$dir" ]] || continue
    local name
    name=$(basename "$dir")
    [[ "$name" == "archive" ]] && continue
    if is_hydrate_done "${dir}.status.yaml"; then
      printf "  %s\n" "$name"
      ((count++))
    fi
  done
  if [[ $count -eq 0 ]]; then
    echo "  (none)"
  fi
}

all_archivable_names() {
  for dir in "$CHANGES_DIR"/*/; do
    [[ -d "$dir" ]] || continue
    local name
    name=$(basename "$dir")
    [[ "$name" == "archive" ]] && continue
    if is_hydrate_done "${dir}.status.yaml"; then
      printf '%s\n' "$name"
    fi
  done
}

get_branch_prefix() {
  if [[ -f "$CONFIG_FILE" ]]; then
    grep -E '^\s*branch_prefix:' "$CONFIG_FILE" | \
      sed -E 's/^\s*branch_prefix:\s*"?([^"]*)"?.*/\1/' || echo ""
  else
    echo ""
  fi
}

# ---------------------------------------------------------------------------
# Validation
# ---------------------------------------------------------------------------

if [[ ! -d "$CHANGES_DIR" ]]; then
  echo "Error: changes directory not found at $CHANGES_DIR" >&2
  exit 1
fi

if [[ -z "${TMUX:-}" ]]; then
  echo "Error: not inside a tmux session" >&2
  exit 1
fi

if [[ $# -eq 0 ]]; then
  usage
  exit 1
fi

# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------

changes=()
case "$1" in
  --list)  list_archivable; exit 0 ;;
  --all)
    mapfile -t changes < <(all_archivable_names)
    if [[ ${#changes[@]} -eq 0 ]]; then
      echo "No archivable changes found." >&2
      exit 1
    fi
    echo "Opening ${#changes[@]} tabs for all archivable changes..."
    ;;
  -h|--help) usage; exit 0 ;;
  *)         changes=("$@") ;;
esac

# ---------------------------------------------------------------------------
# Open tabs
# ---------------------------------------------------------------------------

for change in "${changes[@]}"; do
  # Resolve: exact match first, then substring match (e.g. "v3rn")
  match=""
  if [[ -d "${CHANGES_DIR}/${change}" ]]; then
    match="$change"
  else
    for dir in "$CHANGES_DIR"/*/; do
      [[ -d "$dir" ]] || continue
      name=$(basename "$dir")
      [[ "$name" == "archive" ]] && continue
      if [[ "$name" == *"${change}"* ]]; then
        if [[ -n "$match" ]]; then
          echo "Warning: '$change' matches multiple changes, skipping (be more specific)" >&2
          match=""
          break
        fi
        match="$name"
      fi
    done
  fi

  if [[ -z "$match" ]]; then
    echo "Warning: '$change' not found in changes, skipping" >&2
    continue
  fi

  # Verify hydrate:done
  local_status="${CHANGES_DIR}/${match}/.status.yaml"
  if ! is_hydrate_done "$local_status"; then
    echo "Warning: '$match' not ready for archive (hydrate not done), skipping" >&2
    continue
  fi

  printf "  %s\n" "$match"

  # Get branch prefix from config and construct the branch name
  branch_prefix=$(get_branch_prefix)
  branch_name="${branch_prefix}${match}"

  # Create worktree with the branch name
  wt_path=$(wt-create --non-interactive --worktree-name "$match" "$branch_name" | tail -1) || {
    echo "Error: failed to create worktree for '$match', skipping" >&2
    continue
  }

  # Escape single quotes for the nested shell: ' → '\''
  safe="${match//"'"/"'\''"}"

  tmux new-window -n "fab-${match}" -c "$wt_path" \
    "claude --dangerously-skip-permissions '/fab-archive ${safe}'"
done
