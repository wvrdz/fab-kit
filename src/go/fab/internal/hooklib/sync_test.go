package hooklib

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupHooksDir(t *testing.T, scripts ...string) string {
	t.Helper()
	dir := t.TempDir()
	hooksDir := filepath.Join(dir, "hooks")
	os.MkdirAll(hooksDir, 0o755)
	for _, s := range scripts {
		os.WriteFile(filepath.Join(hooksDir, s), []byte("#!/bin/bash\nexit 0\n"), 0o755)
	}
	return hooksDir
}

func TestSync_FreshSettings(t *testing.T) {
	hooksDir := setupHooksDir(t,
		"on-session-start.sh",
		"on-stop.sh",
		"on-user-prompt.sh",
		"on-artifact-write.sh",
	)

	settingsPath := filepath.Join(t.TempDir(), ".claude", "settings.local.json")

	result, err := Sync(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "created" {
		t.Errorf("Status = %q, want %q", result.Status, "created")
	}

	// Verify the file was created with hook entries
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings: %v", err)
	}

	var settings map[string]json.RawMessage
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to parse settings: %v", err)
	}

	var hooks map[string][]hookEntry
	if err := json.Unmarshal(settings["hooks"], &hooks); err != nil {
		t.Fatalf("failed to parse hooks: %v", err)
	}

	// Should have SessionStart, Stop, UserPromptSubmit, PostToolUse
	if len(hooks["SessionStart"]) != 1 {
		t.Errorf("SessionStart entries = %d, want 1", len(hooks["SessionStart"]))
	}
	if len(hooks["Stop"]) != 1 {
		t.Errorf("Stop entries = %d, want 1", len(hooks["Stop"]))
	}
	if len(hooks["UserPromptSubmit"]) != 1 {
		t.Errorf("UserPromptSubmit entries = %d, want 1", len(hooks["UserPromptSubmit"]))
	}
	if len(hooks["PostToolUse"]) != 2 {
		t.Errorf("PostToolUse entries = %d, want 2 (Write + Edit)", len(hooks["PostToolUse"]))
	}
}

func TestSync_Deduplication(t *testing.T) {
	hooksDir := setupHooksDir(t,
		"on-session-start.sh",
		"on-stop.sh",
		"on-user-prompt.sh",
		"on-artifact-write.sh",
	)

	settingsPath := filepath.Join(t.TempDir(), ".claude", "settings.local.json")

	// First sync
	_, err := Sync(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("first sync failed: %v", err)
	}

	// Second sync should be OK (no changes)
	result, err := Sync(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("second sync failed: %v", err)
	}

	if result.Status != "ok" {
		t.Errorf("Status = %q, want %q", result.Status, "ok")
	}
}

func TestSync_MergeNewEntries(t *testing.T) {
	hooksDir := setupHooksDir(t,
		"on-session-start.sh",
		"on-stop.sh",
	)

	settingsPath := filepath.Join(t.TempDir(), ".claude", "settings.local.json")

	// First sync with only two scripts
	_, err := Sync(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("first sync failed: %v", err)
	}

	// Add more scripts
	os.WriteFile(filepath.Join(hooksDir, "on-user-prompt.sh"), []byte("#!/bin/bash\nexit 0\n"), 0o755)
	os.WriteFile(filepath.Join(hooksDir, "on-artifact-write.sh"), []byte("#!/bin/bash\nexit 0\n"), 0o755)

	// Second sync should add new entries
	result, err := Sync(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("second sync failed: %v", err)
	}

	if result.Status != "updated" {
		t.Errorf("Status = %q, want %q", result.Status, "updated")
	}
}

func TestSync_MissingScript(t *testing.T) {
	// Only create session-start and stop, NOT user-prompt or artifact-write
	hooksDir := setupHooksDir(t,
		"on-session-start.sh",
		"on-stop.sh",
	)

	settingsPath := filepath.Join(t.TempDir(), ".claude", "settings.local.json")

	result, err := Sync(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "created" {
		t.Errorf("Status = %q, want %q", result.Status, "created")
	}

	// Verify only 2 entries (no UserPromptSubmit, no PostToolUse)
	data, _ := os.ReadFile(settingsPath)
	var settings map[string]json.RawMessage
	json.Unmarshal(data, &settings)
	var hooks map[string][]hookEntry
	json.Unmarshal(settings["hooks"], &hooks)

	if _, ok := hooks["UserPromptSubmit"]; ok {
		t.Error("UserPromptSubmit should not be present (script missing)")
	}
	if _, ok := hooks["PostToolUse"]; ok {
		t.Error("PostToolUse should not be present (script missing)")
	}
}

func TestSync_PreserveNonHookSettings(t *testing.T) {
	settingsDir := t.TempDir()
	settingsPath := filepath.Join(settingsDir, "settings.local.json")

	// Write initial settings with non-hook data
	initial := `{"model":"claude-opus-4-6","permissions":{"allow":["Read"]}}`
	os.WriteFile(settingsPath, []byte(initial), 0o644)

	hooksDir := setupHooksDir(t, "on-session-start.sh")

	_, err := Sync(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify non-hook settings preserved
	data, _ := os.ReadFile(settingsPath)
	var settings map[string]json.RawMessage
	json.Unmarshal(data, &settings)

	var model string
	json.Unmarshal(settings["model"], &model)
	if model != "claude-opus-4-6" {
		t.Errorf("model = %q, want %q", model, "claude-opus-4-6")
	}

	if settings["permissions"] == nil {
		t.Error("permissions should be preserved")
	}
}

func TestSync_EmptySettings(t *testing.T) {
	settingsDir := t.TempDir()
	settingsPath := filepath.Join(settingsDir, "settings.local.json")
	os.WriteFile(settingsPath, []byte("{}"), 0o644)

	hooksDir := setupHooksDir(t, "on-session-start.sh")

	result, err := Sync(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "created" {
		t.Errorf("Status = %q, want %q", result.Status, "created")
	}
}

func TestSync_EmptyHooksDir(t *testing.T) {
	hooksDir := setupHooksDir(t) // no scripts
	settingsPath := filepath.Join(t.TempDir(), ".claude", "settings.local.json")

	result, err := Sync(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No entries added, but file should be written
	if result.Status != "ok" {
		t.Errorf("Status = %q, want %q", result.Status, "ok")
	}
}

func TestSync_UsesClaudeProjectDir(t *testing.T) {
	hooksDir := setupHooksDir(t, "on-stop.sh")
	settingsPath := filepath.Join(t.TempDir(), ".claude", "settings.local.json")

	_, err := Sync(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hooks := readHooks(t, settingsPath)
	cmd := hooks["Stop"][0].Hooks[0].Command
	want := `bash "$CLAUDE_PROJECT_DIR"/fab/.kit/hooks/on-stop.sh`
	if cmd != want {
		t.Errorf("command = %q, want %q", cmd, want)
	}
}

func TestSync_MigratesOldRelativePaths(t *testing.T) {
	settingsDir := t.TempDir()
	settingsPath := filepath.Join(settingsDir, "settings.local.json")

	// Write settings with old-format (relative path) hooks
	oldSettings := `{
  "hooks": {
    "Stop": [{"matcher": "", "hooks": [{"type": "command", "command": "bash fab/.kit/hooks/on-stop.sh"}]}],
    "SessionStart": [{"matcher": "", "hooks": [{"type": "command", "command": "bash fab/.kit/hooks/on-session-start.sh"}]}]
  }
}`
	os.WriteFile(settingsPath, []byte(oldSettings), 0o644)

	hooksDir := setupHooksDir(t, "on-stop.sh", "on-session-start.sh")

	result, err := Sync(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "updated" {
		t.Errorf("Status = %q, want %q", result.Status, "updated")
	}
	if !strings.Contains(result.Message, "migrated") {
		t.Errorf("Message should mention migration, got %q", result.Message)
	}

	// Verify commands were migrated
	hooks := readHooks(t, settingsPath)
	for _, event := range []string{"Stop", "SessionStart"} {
		cmd := hooks[event][0].Hooks[0].Command
		if !strings.Contains(cmd, "$CLAUDE_PROJECT_DIR") {
			t.Errorf("%s command not migrated: %q", event, cmd)
		}
		if strings.HasPrefix(cmd, "bash fab/") {
			t.Errorf("%s command still has old relative path: %q", event, cmd)
		}
	}

	// No duplicate entries — migration should update in place, not add new ones
	if len(hooks["Stop"]) != 1 {
		t.Errorf("Stop entries = %d, want 1 (no duplicates after migration)", len(hooks["Stop"]))
	}
}

func TestSync_MigrateAndDedup(t *testing.T) {
	settingsDir := t.TempDir()
	settingsPath := filepath.Join(settingsDir, "settings.local.json")

	// Old-format hook already present
	oldSettings := `{
  "hooks": {
    "Stop": [{"matcher": "", "hooks": [{"type": "command", "command": "bash fab/.kit/hooks/on-stop.sh"}]}]
  }
}`
	os.WriteFile(settingsPath, []byte(oldSettings), 0o644)

	hooksDir := setupHooksDir(t, "on-stop.sh")

	// Sync should migrate the old entry, then dedup (not add a second)
	_, err := Sync(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hooks := readHooks(t, settingsPath)
	if len(hooks["Stop"]) != 1 {
		t.Errorf("Stop entries = %d, want 1 (migrated entry should dedup with desired)", len(hooks["Stop"]))
	}

	// Second sync should be clean
	result, err := Sync(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "ok" {
		t.Errorf("Status = %q, want %q after migration+dedup", result.Status, "ok")
	}
}

func TestSync_MigratePreservesNonFabHooks(t *testing.T) {
	settingsDir := t.TempDir()
	settingsPath := filepath.Join(settingsDir, "settings.local.json")

	// Mix of fab hooks (old format) and non-fab hooks
	oldSettings := `{
  "hooks": {
    "Stop": [
      {"matcher": "", "hooks": [{"type": "command", "command": "bash fab/.kit/hooks/on-stop.sh"}]},
      {"matcher": "", "hooks": [{"type": "command", "command": "echo custom stop hook"}]}
    ]
  }
}`
	os.WriteFile(settingsPath, []byte(oldSettings), 0o644)

	hooksDir := setupHooksDir(t, "on-stop.sh")

	_, err := Sync(hooksDir, settingsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hooks := readHooks(t, settingsPath)
	if len(hooks["Stop"]) != 2 {
		t.Errorf("Stop entries = %d, want 2 (migrated fab + custom)", len(hooks["Stop"]))
	}

	// Find the custom hook — should be untouched
	found := false
	for _, entry := range hooks["Stop"] {
		for _, h := range entry.Hooks {
			if h.Command == "echo custom stop hook" {
				found = true
			}
		}
	}
	if !found {
		t.Error("custom non-fab hook was lost during migration")
	}
}

// readHooks is a test helper that reads and parses hooks from the settings file.
func readHooks(t *testing.T, settingsPath string) map[string][]hookEntry {
	t.Helper()
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings: %v", err)
	}
	var settings map[string]json.RawMessage
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to parse settings: %v", err)
	}
	var hooks map[string][]hookEntry
	if err := json.Unmarshal(settings["hooks"], &hooks); err != nil {
		t.Fatalf("failed to parse hooks: %v", err)
	}
	return hooks
}
