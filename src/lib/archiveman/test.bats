#!/usr/bin/env bats

# Test suite for archiveman.sh
# Covers: archive, restore, list subcommands

REPO_ROOT="$(cd "$(dirname "$BATS_TEST_FILENAME")/../../.." && pwd)"
ARCHIVEMAN="$REPO_ROOT/fab/.kit/scripts/lib/archiveman.sh"

setup() {
  TEST_DIR="$(mktemp -d)"
  FAB_ROOT="$TEST_DIR/fab"
  mkdir -p "$FAB_ROOT/changes/archive" "$FAB_ROOT/.kit/scripts/lib"

  # Copy the actual archiveman.sh into the kit location
  cp "$ARCHIVEMAN" "$FAB_ROOT/.kit/scripts/lib/archiveman.sh"
  chmod +x "$FAB_ROOT/.kit/scripts/lib/archiveman.sh"

  # Copy resolve.sh (archiveman delegates archive resolution to it)
  cp "$REPO_ROOT/fab/.kit/scripts/lib/resolve.sh" "$FAB_ROOT/.kit/scripts/lib/resolve.sh"
  chmod +x "$FAB_ROOT/.kit/scripts/lib/resolve.sh"

  # Create a stub changeman.sh that handles resolve, switch, and switch --blank
  CHANGEMAN_LOG="$TEST_DIR/changeman-calls.log"
  cat > "$FAB_ROOT/.kit/scripts/lib/changeman.sh" <<STUB
#!/usr/bin/env bash
echo "\$@" >> "$CHANGEMAN_LOG"
if [ "\$1" = "resolve" ]; then
  # Read fab/current line 2 for active change
  if [ -f "$FAB_ROOT/current" ]; then
    sed -n '2p' "$FAB_ROOT/current" | tr -d '[:space:]'
  else
    exit 1
  fi
elif [ "\$1" = "switch" ] && [ "\${2:-}" = "--blank" ]; then
  rm -f "$FAB_ROOT/current"
  echo "No active change."
elif [ "\$1" = "switch" ]; then
  local_name="\$2"
  local_id=\$(echo "\$local_name" | cut -d'-' -f2)
  printf '%s\n%s' "\$local_id" "\$local_name" > "$FAB_ROOT/current"
  echo "fab/current → \$local_name"
fi
STUB
  chmod +x "$FAB_ROOT/.kit/scripts/lib/changeman.sh"

  SCRIPT="$FAB_ROOT/.kit/scripts/lib/archiveman.sh"
}

teardown() {
  rm -rf "$TEST_DIR"
}

# Helper: create a change folder in fab/changes/
create_change() {
  local name="$1"
  mkdir -p "$FAB_ROOT/changes/$name"
  echo "name: $name" > "$FAB_ROOT/changes/$name/.status.yaml"
}

# Helper: set active change in fab/current
set_active() {
  local name="$1"
  local id
  id=$(echo "$name" | cut -d'-' -f2)
  printf '%s\n%s' "$id" "$name" > "$FAB_ROOT/current"
}

# Helper: create an archived change
create_archived() {
  local name="$1"
  mkdir -p "$FAB_ROOT/changes/archive/$name"
  echo "name: $name" > "$FAB_ROOT/changes/archive/$name/.status.yaml"
}

# ── archive: happy path ──────────────────────────────────────────

@test "archive moves folder to archive/ and outputs YAML" {
  create_change "260303-hcq9-test-change"

  run bash "$SCRIPT" archive "260303-hcq9-test-change" --description "Test archive"
  [ "$status" -eq 0 ]
  [ -d "$FAB_ROOT/changes/archive/260303-hcq9-test-change" ]
  [ ! -d "$FAB_ROOT/changes/260303-hcq9-test-change" ]
  [[ "$output" == *"action: archive"* ]]
  [[ "$output" == *"name: 260303-hcq9-test-change"* ]]
  [[ "$output" == *"move: moved"* ]]
}

@test "archive creates index entry with description" {
  create_change "260303-hcq9-test-change"

  run bash "$SCRIPT" archive "260303-hcq9-test-change" --description "My description here"
  [ "$status" -eq 0 ]
  grep -qF '**260303-hcq9-test-change** — My description here' "$FAB_ROOT/changes/archive/index.md"
}

@test "archive removes .pr-done when present" {
  create_change "260303-hcq9-test-change"
  touch "$FAB_ROOT/changes/260303-hcq9-test-change/.pr-done"

  run bash "$SCRIPT" archive "260303-hcq9-test-change" --description "Test"
  [ "$status" -eq 0 ]
  [[ "$output" == *"clean: removed"* ]]
  [ ! -f "$FAB_ROOT/changes/archive/260303-hcq9-test-change/.pr-done" ]
}

@test "archive reports clean: not_present when no .pr-done" {
  create_change "260303-hcq9-test-change"

  run bash "$SCRIPT" archive "260303-hcq9-test-change" --description "Test"
  [ "$status" -eq 0 ]
  [[ "$output" == *"clean: not_present"* ]]
}

@test "archive creates archive/ directory if missing" {
  rm -rf "$FAB_ROOT/changes/archive"
  create_change "260303-hcq9-test-change"

  run bash "$SCRIPT" archive "260303-hcq9-test-change" --description "Test"
  [ "$status" -eq 0 ]
  [ -d "$FAB_ROOT/changes/archive" ]
  [ -d "$FAB_ROOT/changes/archive/260303-hcq9-test-change" ]
}

@test "archive clears pointer when change is active" {
  create_change "260303-hcq9-test-change"
  set_active "260303-hcq9-test-change"

  run bash "$SCRIPT" archive "260303-hcq9-test-change" --description "Test"
  [ "$status" -eq 0 ]
  [[ "$output" == *"pointer: cleared"* ]]
  [ ! -f "$FAB_ROOT/current" ]
}

@test "archive skips pointer when change is not active" {
  create_change "260303-hcq9-test-change"
  # Set a different change as active
  set_active "260303-xxxx-other-change"

  run bash "$SCRIPT" archive "260303-hcq9-test-change" --description "Test"
  [ "$status" -eq 0 ]
  [[ "$output" == *"pointer: skipped"* ]]
  # fab/current should still exist
  [ -f "$FAB_ROOT/current" ]
}

@test "archive skips pointer when no active change" {
  create_change "260303-hcq9-test-change"
  rm -f "$FAB_ROOT/current"

  run bash "$SCRIPT" archive "260303-hcq9-test-change" --description "Test"
  [ "$status" -eq 0 ]
  [[ "$output" == *"pointer: skipped"* ]]
}

@test "archive YAML output has all required fields" {
  create_change "260303-hcq9-test-change"

  run bash "$SCRIPT" archive "260303-hcq9-test-change" --description "Test"
  [ "$status" -eq 0 ]
  [[ "$output" == *"action: archive"* ]]
  [[ "$output" == *"name:"* ]]
  [[ "$output" == *"clean:"* ]]
  [[ "$output" == *"move:"* ]]
  [[ "$output" == *"index:"* ]]
  [[ "$output" == *"pointer:"* ]]
}

# ── archive: index management ────────────────────────────────────

@test "archive creates index.md with header when missing" {
  rm -f "$FAB_ROOT/changes/archive/index.md"
  create_change "260303-hcq9-test-change"

  run bash "$SCRIPT" archive "260303-hcq9-test-change" --description "First entry"
  [ "$status" -eq 0 ]
  [[ "$output" == *"index: created"* ]]
  head -1 "$FAB_ROOT/changes/archive/index.md" | grep -q "# Archive Index"
}

@test "archive prepends new entry after header in existing index" {
  create_change "260303-hcq9-new-change"
  printf '# Archive Index\n\n- **260301-xxxx-old-change** — Old entry\n' > "$FAB_ROOT/changes/archive/index.md"

  run bash "$SCRIPT" archive "260303-hcq9-new-change" --description "New entry"
  [ "$status" -eq 0 ]
  [[ "$output" == *"index: updated"* ]]
  # New entry should be on line 3 (after header + blank line)
  line3=$(sed -n '3p' "$FAB_ROOT/changes/archive/index.md")
  [[ "$line3" == *"260303-hcq9-new-change"* ]]
  # Old entry should follow
  line4=$(sed -n '4p' "$FAB_ROOT/changes/archive/index.md")
  [[ "$line4" == *"260301-xxxx-old-change"* ]]
}

@test "archive entry format is correct" {
  create_change "260303-hcq9-test-change"

  run bash "$SCRIPT" archive "260303-hcq9-test-change" --description "My description"
  [ "$status" -eq 0 ]
  grep -q '^- \*\*260303-hcq9-test-change\*\* — My description$' "$FAB_ROOT/changes/archive/index.md"
}

# ── archive: backfill ────────────────────────────────────────────

@test "archive backfills unindexed archived folders" {
  # Pre-existing archived folder with no index entry
  create_archived "260201-aaaa-pre-existing"
  create_change "260303-hcq9-new-change"

  run bash "$SCRIPT" archive "260303-hcq9-new-change" --description "New"
  [ "$status" -eq 0 ]
  # Both should be in the index
  grep -qF "**260303-hcq9-new-change**" "$FAB_ROOT/changes/archive/index.md"
  grep -qF "**260201-aaaa-pre-existing**" "$FAB_ROOT/changes/archive/index.md"
  grep -qF "(no description — pre-index archive)" "$FAB_ROOT/changes/archive/index.md"
}

@test "archive backfill is no-op when all folders indexed" {
  create_archived "260201-aaaa-already-indexed"
  printf '# Archive Index\n\n- **260201-aaaa-already-indexed** — Already here\n' > "$FAB_ROOT/changes/archive/index.md"
  create_change "260303-hcq9-new-change"

  run bash "$SCRIPT" archive "260303-hcq9-new-change" --description "New"
  [ "$status" -eq 0 ]
  # Count entries for the pre-existing change — should be exactly 1
  local count
  count=$(grep -cF "**260201-aaaa-already-indexed**" "$FAB_ROOT/changes/archive/index.md")
  [ "$count" -eq 1 ]
}

# ── archive: error cases ─────────────────────────────────────────

@test "archive errors when --description missing" {
  create_change "260303-hcq9-test-change"

  run bash "$SCRIPT" archive "260303-hcq9-test-change"
  [ "$status" -ne 0 ]
  [[ "$output" == *"--description is required"* ]]
}

@test "archive errors when no <change> argument" {
  run bash "$SCRIPT" archive --description "Test"
  [ "$status" -ne 0 ]
  [[ "$output" == *"<change> argument is required"* ]]
}

@test "archive errors when change not found" {
  run bash "$SCRIPT" archive "nonexistent" --description "Test"
  [ "$status" -ne 0 ]
  [[ "$output" == *"No change matches"* ]] || [[ "$output" == *"No active changes"* ]]
}

# ── restore: happy path ──────────────────────────────────────────

@test "restore moves folder from archive/ to changes/" {
  create_archived "260303-hcq9-test-change"

  run bash "$SCRIPT" restore "260303-hcq9-test-change"
  [ "$status" -eq 0 ]
  [ -d "$FAB_ROOT/changes/260303-hcq9-test-change" ]
  [ ! -d "$FAB_ROOT/changes/archive/260303-hcq9-test-change" ]
  [[ "$output" == *"action: restore"* ]]
  [[ "$output" == *"move: restored"* ]]
}

@test "restore removes entry from index.md" {
  create_archived "260303-hcq9-test-change"
  printf '# Archive Index\n\n- **260303-hcq9-test-change** — Some description\n- **260201-aaaa-other** — Other\n' > "$FAB_ROOT/changes/archive/index.md"

  run bash "$SCRIPT" restore "260303-hcq9-test-change"
  [ "$status" -eq 0 ]
  [[ "$output" == *"index: removed"* ]]
  ! grep -qF "**260303-hcq9-test-change**" "$FAB_ROOT/changes/archive/index.md"
  # Other entry preserved
  grep -qF "**260201-aaaa-other**" "$FAB_ROOT/changes/archive/index.md"
}

@test "restore with --switch activates the change" {
  create_archived "260303-hcq9-test-change"

  run bash "$SCRIPT" restore "260303-hcq9-test-change" --switch
  [ "$status" -eq 0 ]
  [[ "$output" == *"pointer: switched"* ]]
  grep -q "changeman-calls" "$TEST_DIR/changeman-calls.log" 2>/dev/null || true
  grep -qF "switch 260303-hcq9-test-change" "$CHANGEMAN_LOG"
}

@test "restore without --switch skips pointer" {
  create_archived "260303-hcq9-test-change"

  run bash "$SCRIPT" restore "260303-hcq9-test-change"
  [ "$status" -eq 0 ]
  [[ "$output" == *"pointer: skipped"* ]]
}

@test "restore YAML output has all required fields" {
  create_archived "260303-hcq9-test-change"

  run bash "$SCRIPT" restore "260303-hcq9-test-change"
  [ "$status" -eq 0 ]
  [[ "$output" == *"action: restore"* ]]
  [[ "$output" == *"name:"* ]]
  [[ "$output" == *"move:"* ]]
  [[ "$output" == *"index:"* ]]
  [[ "$output" == *"pointer:"* ]]
}

# ── restore: resolution ──────────────────────────────────────────

@test "restore resolves exact match" {
  create_archived "260303-hcq9-test-change"

  run bash "$SCRIPT" restore "260303-hcq9-test-change"
  [ "$status" -eq 0 ]
  [[ "$output" == *"name: 260303-hcq9-test-change"* ]]
}

@test "restore resolves substring match (case-insensitive)" {
  create_archived "260303-hcq9-test-change"

  run bash "$SCRIPT" restore "HCQ9"
  [ "$status" -eq 0 ]
  [[ "$output" == *"name: 260303-hcq9-test-change"* ]]
}

@test "restore resolves 4-char ID match" {
  create_archived "260303-hcq9-test-change"

  run bash "$SCRIPT" restore "hcq9"
  [ "$status" -eq 0 ]
  [[ "$output" == *"name: 260303-hcq9-test-change"* ]]
}

@test "restore errors on multiple matches" {
  create_archived "260303-hcq9-fix-one"
  create_archived "260303-ab12-fix-two"

  run bash "$SCRIPT" restore "fix"
  [ "$status" -ne 0 ]
  [[ "$output" == *"Multiple archives match"* ]]
}

@test "restore errors when no match" {
  create_archived "260303-hcq9-test-change"

  run bash "$SCRIPT" restore "nonexistent"
  [ "$status" -ne 0 ]
  [[ "$output" == *"No archive matches"* ]]
}

@test "restore errors when archive is empty" {
  # archive/ exists but has no folders
  run bash "$SCRIPT" restore "something"
  [ "$status" -ne 0 ]
  [[ "$output" == *"No archived changes"* ]]
}

@test "restore errors when archive dir missing" {
  rm -rf "$FAB_ROOT/changes/archive"

  run bash "$SCRIPT" restore "something"
  [ "$status" -ne 0 ]
  [[ "$output" == *"No archive folder"* ]]
}

# ── restore: resumability ────────────────────────────────────────

@test "restore reports already_in_changes when folder exists in changes/" {
  create_archived "260303-hcq9-test-change"
  # Also create the folder in changes/ (simulates interrupted restore)
  mkdir -p "$FAB_ROOT/changes/260303-hcq9-test-change"

  run bash "$SCRIPT" restore "260303-hcq9-test-change"
  [ "$status" -eq 0 ]
  [[ "$output" == *"move: already_in_changes"* ]]
}

@test "restore still runs index/pointer even when move is skipped" {
  create_archived "260303-hcq9-test-change"
  mkdir -p "$FAB_ROOT/changes/260303-hcq9-test-change"
  printf '# Archive Index\n\n- **260303-hcq9-test-change** — Some desc\n' > "$FAB_ROOT/changes/archive/index.md"

  run bash "$SCRIPT" restore "260303-hcq9-test-change" --switch
  [ "$status" -eq 0 ]
  [[ "$output" == *"move: already_in_changes"* ]]
  [[ "$output" == *"index: removed"* ]]
  [[ "$output" == *"pointer: switched"* ]]
}

# ── restore: index cleanup ───────────────────────────────────────

@test "restore index: not_found when entry missing from index" {
  create_archived "260303-hcq9-test-change"
  printf '# Archive Index\n\n- **260201-aaaa-other** — Other\n' > "$FAB_ROOT/changes/archive/index.md"

  run bash "$SCRIPT" restore "260303-hcq9-test-change"
  [ "$status" -eq 0 ]
  [[ "$output" == *"index: not_found"* ]]
}

@test "restore preserves index file even when last entry removed" {
  create_archived "260303-hcq9-test-change"
  printf '# Archive Index\n\n- **260303-hcq9-test-change** — Last entry\n' > "$FAB_ROOT/changes/archive/index.md"

  run bash "$SCRIPT" restore "260303-hcq9-test-change"
  [ "$status" -eq 0 ]
  [ -f "$FAB_ROOT/changes/archive/index.md" ]
  # Header should remain
  head -1 "$FAB_ROOT/changes/archive/index.md" | grep -q "# Archive Index"
}

# ── list ─────────────────────────────────────────────────────────

@test "list outputs one folder name per line" {
  create_archived "260301-aaaa-change-one"
  create_archived "260302-bbbb-change-two"
  create_archived "260303-cccc-change-three"

  run bash "$SCRIPT" list
  [ "$status" -eq 0 ]
  local line_count
  line_count=$(echo "$output" | wc -l)
  [ "$line_count" -eq 3 ]
}

@test "list excludes index.md" {
  create_archived "260301-aaaa-change-one"
  printf '# Archive Index\n' > "$FAB_ROOT/changes/archive/index.md"

  run bash "$SCRIPT" list
  [ "$status" -eq 0 ]
  [[ "$output" != *"index.md"* ]]
  [[ "$output" == *"260301-aaaa-change-one"* ]]
}

@test "list returns empty output for empty archive" {
  # archive/ exists but no folders
  run bash "$SCRIPT" list
  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "list returns empty output when archive dir missing" {
  rm -rf "$FAB_ROOT/changes/archive"

  run bash "$SCRIPT" list
  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

# ── CLI edge cases ───────────────────────────────────────────────

@test "--help prints usage information" {
  run bash "$SCRIPT" --help
  [ "$status" -eq 0 ]
  [[ "$output" == *"USAGE"* ]]
  [[ "$output" == *"archive"* ]]
  [[ "$output" == *"restore"* ]]
  [[ "$output" == *"list"* ]]
}

@test "no subcommand produces error" {
  run bash "$SCRIPT"
  [ "$status" -ne 0 ]
  [[ "$output" == *"No subcommand"* ]]
}

@test "unknown subcommand produces error" {
  run bash "$SCRIPT" delete
  [ "$status" -ne 0 ]
  [[ "$output" == *"Unknown subcommand 'delete'"* ]]
}

@test "restore errors when no <change> argument" {
  run bash "$SCRIPT" restore
  [ "$status" -ne 0 ]
  [[ "$output" == *"<change> argument is required"* ]]
}
